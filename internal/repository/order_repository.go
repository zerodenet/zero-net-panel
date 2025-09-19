package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusCancelled = "cancelled"

	PaymentMethodBalance = "balance"
)

// Order represents a billing order.
type Order struct {
	ID            uint64         `gorm:"primaryKey"`
	Number        string         `gorm:"size:40;uniqueIndex"`
	UserID        uint64         `gorm:"index"`
	PlanID        *uint64        `gorm:"column:plan_id"`
	Status        string         `gorm:"size:32"`
	PaymentMethod string         `gorm:"size:32"`
	TotalCents    int64          `gorm:"column:total_cents"`
	Currency      string         `gorm:"size:16"`
	RefundedCents int64          `gorm:"column:refunded_cents"`
	RefundedAt    *time.Time     `gorm:"column:refunded_at"`
	PaidAt        *time.Time     `gorm:"column:paid_at"`
	CancelledAt   *time.Time     `gorm:"column:cancelled_at"`
	Metadata      map[string]any `gorm:"serializer:json"`
	PlanSnapshot  map[string]any `gorm:"serializer:json"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TableName declares table mapping.
func (Order) TableName() string { return "orders" }

// OrderItem represents individual items within an order.
type OrderItem struct {
	ID             uint64         `gorm:"primaryKey"`
	OrderID        uint64         `gorm:"index"`
	ItemType       string         `gorm:"size:32"`
	ItemID         uint64         `gorm:"index"`
	Name           string         `gorm:"size:255"`
	Quantity       int            `gorm:"column:quantity"`
	UnitPriceCents int64          `gorm:"column:unit_price_cents"`
	Currency       string         `gorm:"size:16"`
	SubtotalCents  int64          `gorm:"column:subtotal_cents"`
	Metadata       map[string]any `gorm:"serializer:json"`
	CreatedAt      time.Time
}

// TableName declares table mapping.
func (OrderItem) TableName() string { return "order_items" }

// ListOrdersOptions controls filtering.
type ListOrdersOptions struct {
	Page          int
	PerPage       int
	Status        string
	PaymentMethod string
	Number        string
	UserID        *uint64
	Sort          string
	Direction     string
}

// UpdateOrderStatusParams describes fields that can be changed when updating an order state.
type UpdateOrderStatusParams struct {
	Status        string
	PaymentMethod *string
	PaidAt        *time.Time
	CancelledAt   *time.Time
	MetadataPatch map[string]any
}

// AddRefundParams captures arguments when applying a refund to an order.
type AddRefundParams struct {
	AmountCents   int64
	RefundAt      time.Time
	MetadataPatch map[string]any
}

// OrderRepository exposes CRUD helpers.
type OrderRepository interface {
	Create(ctx context.Context, order Order, items []OrderItem) (Order, []OrderItem, error)
	Get(ctx context.Context, id uint64) (Order, []OrderItem, error)
	List(ctx context.Context, opts ListOrdersOptions) ([]Order, int64, error)
	ListItems(ctx context.Context, orderIDs []uint64) (map[uint64][]OrderItem, error)
	UpdateStatus(ctx context.Context, id uint64, params UpdateOrderStatusParams) (Order, error)
	AddRefund(ctx context.Context, id uint64, params AddRefundParams) (Order, error)
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository constructs repository.
func NewOrderRepository(db *gorm.DB) (OrderRepository, error) {
	if db == nil {
		return nil, errors.New("repository: database connection is required")
	}
	return &orderRepository{db: db}, nil
}

// GenerateOrderNumber builds a monotonic order number using timestamp.
func GenerateOrderNumber() string {
	return fmt.Sprintf("ORD-%d", time.Now().UTC().UnixNano())
}

func (r *orderRepository) Create(ctx context.Context, order Order, items []OrderItem) (Order, []OrderItem, error) {
	if err := ctx.Err(); err != nil {
		return Order{}, nil, err
	}

	now := time.Now().UTC()
	if order.CreatedAt.IsZero() {
		order.CreatedAt = now
	}
	if order.UpdatedAt.IsZero() {
		order.UpdatedAt = now
	}
	if order.Number == "" {
		order.Number = GenerateOrderNumber()
	}
	if order.Status == "" {
		order.Status = OrderStatusPending
	}
	if order.PaymentMethod == "" {
		order.PaymentMethod = PaymentMethodBalance
	}

	if err := r.db.WithContext(ctx).Create(&order).Error; err != nil {
		return Order{}, nil, translateError(err)
	}

	if len(items) > 0 {
		for i := range items {
			items[i].OrderID = order.ID
			if items[i].CreatedAt.IsZero() {
				items[i].CreatedAt = now
			}
			if items[i].Currency == "" {
				items[i].Currency = order.Currency
			}
			if items[i].SubtotalCents == 0 {
				items[i].SubtotalCents = int64(items[i].Quantity) * items[i].UnitPriceCents
			}
		}
		if err := r.db.WithContext(ctx).Create(&items).Error; err != nil {
			return Order{}, nil, translateError(err)
		}
	}

	return order, items, nil
}

func (r *orderRepository) Get(ctx context.Context, id uint64) (Order, []OrderItem, error) {
	if err := ctx.Err(); err != nil {
		return Order{}, nil, err
	}

	var order Order
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		return Order{}, nil, translateError(err)
	}

	var items []OrderItem
	if err := r.db.WithContext(ctx).Where("order_id = ?", id).Order("id ASC").Find(&items).Error; err != nil {
		return Order{}, nil, err
	}

	return order, items, nil
}

func (r *orderRepository) List(ctx context.Context, opts ListOrdersOptions) ([]Order, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	opts = normalizeListOrdersOptions(opts)

	base := r.db.WithContext(ctx).Model(&Order{})
	if opts.UserID != nil {
		base = base.Where("user_id = ?", *opts.UserID)
	}
	if opts.Status != "" {
		base = base.Where("LOWER(status) = ?", strings.ToLower(opts.Status))
	}
	if opts.PaymentMethod != "" {
		base = base.Where("LOWER(payment_method) = ?", strings.ToLower(opts.PaymentMethod))
	}
	if opts.Number != "" {
		like := fmt.Sprintf("%%%s%%", strings.TrimSpace(opts.Number))
		base = base.Where("number LIKE ?", like)
	}

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []Order{}, 0, nil
	}

	offset := (opts.Page - 1) * opts.PerPage
	orderClause := buildOrderSortClause(opts.Sort, opts.Direction)

	var orders []Order
	if err := base.Session(&gorm.Session{}).
		Order(orderClause).
		Limit(opts.PerPage).
		Offset(offset).
		Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *orderRepository) ListItems(ctx context.Context, orderIDs []uint64) (map[uint64][]OrderItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(orderIDs) == 0 {
		return map[uint64][]OrderItem{}, nil
	}

	var items []OrderItem
	if err := r.db.WithContext(ctx).
		Where("order_id IN ?", orderIDs).
		Order("id ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	grouped := make(map[uint64][]OrderItem, len(orderIDs))
	for _, item := range items {
		grouped[item.OrderID] = append(grouped[item.OrderID], item)
	}

	return grouped, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uint64, params UpdateOrderStatusParams) (Order, error) {
	if err := ctx.Err(); err != nil {
		return Order{}, err
	}

	status := strings.TrimSpace(strings.ToLower(params.Status))
	if status == "" {
		return Order{}, ErrInvalidArgument
	}

	var result Order
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, id).Error; err != nil {
			return translateError(err)
		}

		now := time.Now().UTC()
		order.Status = status
		order.UpdatedAt = now

		fields := []string{"Status", "UpdatedAt"}

		if params.PaymentMethod != nil {
			method := strings.TrimSpace(*params.PaymentMethod)
			if method != "" {
				order.PaymentMethod = method
				fields = append(fields, "PaymentMethod")
			}
		}

		if params.PaidAt != nil {
			paidAt := params.PaidAt.UTC()
			order.PaidAt = &paidAt
			fields = append(fields, "PaidAt")
		}

		if params.CancelledAt != nil {
			cancelledAt := params.CancelledAt.UTC()
			order.CancelledAt = &cancelledAt
			fields = append(fields, "CancelledAt")
		}

		metadataChanged := false
		if len(params.MetadataPatch) > 0 {
			order.Metadata = mergeMetadata(order.Metadata, params.MetadataPatch)
			metadataChanged = true
			fields = append(fields, "Metadata")
		}

		if err := tx.Model(&order).Select(fields).Updates(order).Error; err != nil {
			return translateError(err)
		}

		if !metadataChanged && order.Metadata == nil {
			order.Metadata = map[string]any{}
		}
		result = order
		return nil
	})
	if err != nil {
		return Order{}, err
	}

	return result, nil
}

func (r *orderRepository) AddRefund(ctx context.Context, id uint64, params AddRefundParams) (Order, error) {
	if err := ctx.Err(); err != nil {
		return Order{}, err
	}
	if params.AmountCents <= 0 {
		return Order{}, ErrInvalidArgument
	}

	refundAt := params.RefundAt
	if refundAt.IsZero() {
		refundAt = time.Now().UTC()
	} else {
		refundAt = refundAt.UTC()
	}

	var result Order
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, id).Error; err != nil {
			return translateError(err)
		}

		remaining := order.TotalCents - order.RefundedCents
		if remaining <= 0 || params.AmountCents > remaining {
			return ErrInvalidArgument
		}

		updateTime := time.Now().UTC()
		order.RefundedCents += params.AmountCents
		order.RefundedAt = &refundAt
		order.UpdatedAt = updateTime

		fields := []string{"RefundedCents", "RefundedAt", "UpdatedAt"}
		if len(params.MetadataPatch) > 0 {
			order.Metadata = mergeMetadata(order.Metadata, params.MetadataPatch)
			fields = append(fields, "Metadata")
		}

		if err := tx.Model(&order).Select(fields).Updates(order).Error; err != nil {
			return translateError(err)
		}

		result = order
		return nil
	})
	if err != nil {
		return Order{}, err
	}

	return result, nil
}

func normalizeListOrdersOptions(opts ListOrdersOptions) ListOrdersOptions {
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PerPage <= 0 || opts.PerPage > 100 {
		opts.PerPage = 20
	}
	opts.Status = strings.TrimSpace(strings.ToLower(opts.Status))
	opts.PaymentMethod = strings.TrimSpace(strings.ToLower(opts.PaymentMethod))
	opts.Sort = strings.TrimSpace(strings.ToLower(opts.Sort))
	opts.Direction = strings.TrimSpace(strings.ToLower(opts.Direction))
	return opts
}

func buildOrderSortClause(sort, direction string) string {
	column := "created_at"
	switch sort {
	case "updated":
		column = "updated_at"
	case "total":
		column = "total_cents"
	}

	dir := "DESC"
	if strings.EqualFold(direction, "asc") {
		dir = "ASC"
	}

	return fmt.Sprintf("%s %s, id DESC", column, dir)
}

func mergeMetadata(base map[string]any, patch map[string]any) map[string]any {
	if len(patch) == 0 {
		return base
	}

	merged := make(map[string]any, len(base)+len(patch))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range patch {
		merged[k] = v
	}
	return merged
}
