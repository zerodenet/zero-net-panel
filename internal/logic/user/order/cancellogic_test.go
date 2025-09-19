package order

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

func setupUserOrderTestContext(t *testing.T) (*svc.ServiceContext, func()) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = migrations.Apply(context.Background(), db, 0, false)
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

func TestUserCancelOrder_SucceedsForPending(t *testing.T) {
	svcCtx, cleanup := setupUserOrderTestContext(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	user := repository.User{
		Email:       "member@example.com",
		DisplayName: "Member",
		Roles:       []string{"user"},
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, svcCtx.DB.Create(&user).Error)

	plan := repository.Plan{
		Name:         "Trial",
		Slug:         "trial",
		PriceCents:   1000,
		Currency:     "CNY",
		DurationDays: 30,
		Status:       "active",
		Visible:      true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	require.NoError(t, svcCtx.DB.Create(&plan).Error)

	orderModel := repository.Order{
		Number:        repository.GenerateOrderNumber(),
		UserID:        user.ID,
		PlanID:        &plan.ID,
		Status:        repository.OrderStatusPending,
		PaymentMethod: repository.PaymentMethodBalance,
		TotalCents:    1000,
		Currency:      "CNY",
		Metadata:      map[string]any{},
		PlanSnapshot: map[string]any{
			"id":    plan.ID,
			"name":  plan.Name,
			"price": plan.PriceCents,
		},
		CreatedAt: now.Add(-1 * time.Hour),
		UpdatedAt: now.Add(-1 * time.Hour),
	}
	require.NoError(t, svcCtx.DB.Create(&orderModel).Error)

	item := repository.OrderItem{
		OrderID:        orderModel.ID,
		ItemType:       "plan",
		ItemID:         plan.ID,
		Name:           plan.Name,
		Quantity:       1,
		UnitPriceCents: 1000,
		Currency:       "CNY",
		SubtotalCents:  1000,
		Metadata:       map[string]any{},
		CreatedAt:      now.Add(-1 * time.Hour),
	}
	require.NoError(t, svcCtx.DB.Create(&item).Error)

	claims := security.UserClaims{ID: user.ID, Email: user.Email, Roles: []string{"user"}}
	ctx = security.WithUser(ctx, claims)

	logic := NewCancelLogic(ctx, svcCtx)
	resp, err := logic.Cancel(&types.UserCancelOrderRequest{OrderID: orderModel.ID, Reason: "changed mind"})
	require.NoError(t, err)
	require.Equal(t, repository.OrderStatusCancelled, resp.Order.Status)
	require.NotNil(t, resp.Order.CancelledAt)
	require.Equal(t, int64(0), resp.Order.RefundedCents)
	require.Contains(t, resp.Order.Metadata, "cancel_reason")
	require.Equal(t, "changed mind", resp.Order.Metadata["cancel_reason"])

	var updated repository.Order
	require.NoError(t, svcCtx.DB.First(&updated, orderModel.ID).Error)
	require.Equal(t, repository.OrderStatusCancelled, updated.Status)
	require.NotNil(t, updated.CancelledAt)
}

func TestUserCancelOrder_RejectsPaidOrder(t *testing.T) {
	svcCtx, cleanup := setupUserOrderTestContext(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	user := repository.User{
		Email:       "member2@example.com",
		DisplayName: "Member2",
		Roles:       []string{"user"},
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, svcCtx.DB.Create(&user).Error)

	orderModel := repository.Order{
		Number:        repository.GenerateOrderNumber(),
		UserID:        user.ID,
		Status:        repository.OrderStatusPaid,
		PaymentMethod: repository.PaymentMethodBalance,
		TotalCents:    800,
		Currency:      "CNY",
		RefundedCents: 0,
		CreatedAt:     now.Add(-2 * time.Hour),
		UpdatedAt:     now.Add(-2 * time.Hour),
	}
	paidAt := now.Add(-90 * time.Minute)
	orderModel.PaidAt = &paidAt
	require.NoError(t, svcCtx.DB.Create(&orderModel).Error)

	claims := security.UserClaims{ID: user.ID, Email: user.Email, Roles: []string{"user"}}
	ctx = security.WithUser(ctx, claims)

	logic := NewCancelLogic(ctx, svcCtx)
	_, err := logic.Cancel(&types.UserCancelOrderRequest{OrderID: orderModel.ID})
	require.ErrorIs(t, err, repository.ErrInvalidArgument)
}
