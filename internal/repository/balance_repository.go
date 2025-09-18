package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserBalance captures wallet summary for用户余额。
type UserBalance struct {
	UserID            uint64  `gorm:"primaryKey"`
	BalanceCents      int64   `gorm:"column:balance_cents"`
	Currency          string  `gorm:"size:16"`
	LastTransactionID *uint64 `gorm:"column:last_transaction_id"`
	UpdatedAt         time.Time
	CreatedAt         time.Time
}

// TableName overrides default naming.
func (UserBalance) TableName() string { return "user_balances" }

// BalanceTransaction describes ledger records for充值/消费等。
type BalanceTransaction struct {
	ID                uint64         `gorm:"primaryKey"`
	UserID            uint64         `gorm:"index"`
	Type              string         `gorm:"size:32"`
	AmountCents       int64          `gorm:"column:amount_cents"`
	Currency          string         `gorm:"size:16"`
	BalanceAfterCents int64          `gorm:"column:balance_after_cents"`
	Reference         string         `gorm:"size:64"`
	Description       string         `gorm:"size:255"`
	Metadata          map[string]any `gorm:"serializer:json"`
	CreatedAt         time.Time
}

// TableName custom binding.
func (BalanceTransaction) TableName() string { return "balance_transactions" }

// ListBalanceTransactionsOptions controls pagination for ledger entries.
type ListBalanceTransactionsOptions struct {
	Page    int
	PerPage int
	Type    string
}

// BalanceRepository exposes wallet related operations.
type BalanceRepository interface {
	GetBalance(ctx context.Context, userID uint64) (UserBalance, error)
	ListTransactions(ctx context.Context, userID uint64, opts ListBalanceTransactionsOptions) ([]BalanceTransaction, int64, error)
	ApplyTransaction(ctx context.Context, userID uint64, tx BalanceTransaction) (BalanceTransaction, UserBalance, error)
}

type balanceRepository struct {
	db *gorm.DB
}

// NewBalanceRepository wires repository.
func NewBalanceRepository(db *gorm.DB) (BalanceRepository, error) {
	if db == nil {
		return nil, errors.New("repository: database connection is required")
	}
	return &balanceRepository{db: db}, nil
}

func (r *balanceRepository) GetBalance(ctx context.Context, userID uint64) (UserBalance, error) {
	if err := ctx.Err(); err != nil {
		return UserBalance{}, err
	}

	var balance UserBalance
	err := r.db.WithContext(ctx).First(&balance, "user_id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return UserBalance{UserID: userID, Currency: "CNY", BalanceCents: 0}, nil
	}
	if err != nil {
		return UserBalance{}, err
	}
	return balance, nil
}

func (r *balanceRepository) ListTransactions(ctx context.Context, userID uint64, opts ListBalanceTransactionsOptions) ([]BalanceTransaction, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	opts = normalizeTransactionOptions(opts)

	base := r.db.WithContext(ctx).Model(&BalanceTransaction{}).Where("user_id = ?", userID)
	if opts.Type != "" {
		base = base.Where("LOWER(type) = ?", opts.Type)
	}

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []BalanceTransaction{}, 0, nil
	}

	offset := (opts.Page - 1) * opts.PerPage
	var transactions []BalanceTransaction
	if err := base.Session(&gorm.Session{}).
		Order("created_at DESC, id DESC").
		Limit(opts.PerPage).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

func normalizeTransactionOptions(opts ListBalanceTransactionsOptions) ListBalanceTransactionsOptions {
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PerPage <= 0 || opts.PerPage > 100 {
		opts.PerPage = 20
	}
	opts.Type = strings.TrimSpace(strings.ToLower(opts.Type))
	return opts
}

// ApplyTransaction records a balance transaction and updates the aggregate balance atomically.
func (r *balanceRepository) ApplyTransaction(ctx context.Context, userID uint64, tx BalanceTransaction) (BalanceTransaction, UserBalance, error) {
	if err := ctx.Err(); err != nil {
		return BalanceTransaction{}, UserBalance{}, err
	}

	if tx.AmountCents == 0 {
		return BalanceTransaction{}, UserBalance{}, ErrInvalidArgument
	}

	var resultTx BalanceTransaction
	var resultBalance UserBalance

	err := r.db.WithContext(ctx).Transaction(func(gormTx *gorm.DB) error {
		var balance UserBalance
		lock := gormTx.Clauses(clause.Locking{Strength: "UPDATE"})
		err := lock.First(&balance, "user_id = ?", userID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			now := time.Now().UTC()
			balance = UserBalance{
				UserID:       userID,
				BalanceCents: 0,
				Currency:     tx.Currency,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if balance.Currency == "" {
				balance.Currency = "CNY"
			}
			if err := gormTx.Create(&balance).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		currency := balance.Currency
		if currency == "" {
			currency = tx.Currency
		}
		if currency == "" {
			currency = "CNY"
		}

		now := time.Now().UTC()
		newBalance := balance.BalanceCents + tx.AmountCents
		if newBalance < 0 {
			return ErrInsufficientBalance
		}

		txRecord := tx
		txRecord.UserID = userID
		txRecord.Currency = currency
		txRecord.BalanceAfterCents = newBalance
		txRecord.CreatedAt = now

		if err := gormTx.Create(&txRecord).Error; err != nil {
			return err
		}

		balance.BalanceCents = newBalance
		balance.Currency = currency
		balance.UpdatedAt = now
		balance.LastTransactionID = &txRecord.ID

		if err := gormTx.Model(&UserBalance{}).
			Where("user_id = ?", userID).
			Updates(map[string]any{
				"balance_cents":       newBalance,
				"currency":            currency,
				"last_transaction_id": txRecord.ID,
				"updated_at":          now,
			}).Error; err != nil {
			return err
		}

		resultTx = txRecord
		resultBalance = balance
		return nil
	})
	if err != nil {
		return BalanceTransaction{}, UserBalance{}, translateError(err)
	}

	return resultTx, resultBalance, nil
}
