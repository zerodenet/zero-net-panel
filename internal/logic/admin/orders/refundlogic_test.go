package orders

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/zero-net-panel/zero-net-panel/internal/bootstrap/migrations"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/security"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

func setupAdminOrderTestContext(t *testing.T) (*svc.ServiceContext, func()) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = migrations.Apply(context.Background(), db, 0)
	require.NoError(t, err)

	repos, err := repository.NewRepositories(db)
	require.NoError(t, err)

	svcCtx := &svc.ServiceContext{
		DB:           db,
		Repositories: repos,
	}

	cleanup := func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}

	return svcCtx, cleanup
}

func TestAdminRefundOrder_FullRefundCancelsOrder(t *testing.T) {
	svcCtx, cleanup := setupAdminOrderTestContext(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	admin := repository.User{
		Email:       "admin@test.local",
		DisplayName: "Admin",
		Roles:       []string{"admin"},
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, svcCtx.DB.Create(&admin).Error)

	customer := repository.User{
		Email:       "buyer@test.local",
		DisplayName: "Buyer",
		Roles:       []string{"user"},
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, svcCtx.DB.Create(&customer).Error)

	orderModel := repository.Order{
		Number:        repository.GenerateOrderNumber(),
		UserID:        customer.ID,
		Status:        repository.OrderStatusPaid,
		PaymentMethod: repository.PaymentMethodBalance,
		TotalCents:    1500,
		Currency:      "CNY",
		Metadata:      map[string]any{"seed": false},
		CreatedAt:     now.Add(-2 * time.Hour),
		UpdatedAt:     now.Add(-2 * time.Hour),
	}
	paidAt := now.Add(-90 * time.Minute)
	orderModel.PaidAt = &paidAt
	require.NoError(t, svcCtx.DB.Create(&orderModel).Error)

	item := repository.OrderItem{
		OrderID:        orderModel.ID,
		ItemType:       "plan",
		ItemID:         1,
		Name:           "Premium",
		Quantity:       1,
		UnitPriceCents: 1500,
		Currency:       "CNY",
		SubtotalCents:  1500,
		CreatedAt:      now.Add(-2 * time.Hour),
	}
	require.NoError(t, svcCtx.DB.Create(&item).Error)

	claims := security.UserClaims{ID: admin.ID, Email: admin.Email, Roles: []string{"admin"}}
	ctx = security.WithUser(ctx, claims)

	logic := NewRefundLogic(ctx, svcCtx)
	req := types.AdminRefundOrderRequest{
		OrderID:     orderModel.ID,
		AmountCents: 1500,
		Reason:      "duplicate charge",
		Metadata:    map[string]any{"ticket": "A-100"},
	}
	resp, err := logic.Refund(&req)
	require.NoError(t, err)
	require.Equal(t, repository.OrderStatusCancelled, resp.Order.Status)
	require.Equal(t, int64(1500), resp.Order.RefundedCents)
	require.NotNil(t, resp.Order.RefundedAt)
	require.Contains(t, resp.Order.Metadata, "last_refund_reason")
	require.Equal(t, "duplicate charge", resp.Order.Metadata["last_refund_reason"])
	require.Contains(t, resp.Order.Metadata, "cancel_reason")
	require.Equal(t, "refund_completed", resp.Order.Metadata["cancel_reason"])

	var updated repository.Order
	require.NoError(t, svcCtx.DB.First(&updated, orderModel.ID).Error)
	require.Equal(t, repository.OrderStatusCancelled, updated.Status)
	require.Equal(t, int64(1500), updated.RefundedCents)
	require.NotNil(t, updated.RefundedAt)
	require.NotNil(t, updated.CancelledAt)

	balance, err := svcCtx.Repositories.Balance.GetBalance(ctx, customer.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1500), balance.BalanceCents)

	transactions, _, err := svcCtx.Repositories.Balance.ListTransactions(ctx, customer.ID, repository.ListBalanceTransactionsOptions{PerPage: 10})
	require.NoError(t, err)
	require.Len(t, transactions, 1)
	require.Equal(t, "refund", transactions[0].Type)
	require.Equal(t, int64(1500), transactions[0].AmountCents)
	require.Contains(t, transactions[0].Metadata, "ticket")
}
