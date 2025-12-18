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

func setupCreateLogicTest(t *testing.T) (*svc.ServiceContext, func()) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	_, err = migrations.Apply(context.Background(), db, 0, false)
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

func TestCreateOrderWithBalancePayment(t *testing.T) {
	svcCtx, cleanup := setupCreateLogicTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	user := repository.User{
		Email:       "buyer@test.dev",
		DisplayName: "Buyer",
		Roles:       []string{"user"},
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, svcCtx.DB.Create(&user).Error)

	plan := repository.Plan{
		Name:              "Standard",
		Slug:              "standard",
		Description:       "Standard plan",
		PriceCents:        1500,
		Currency:          "CNY",
		DurationDays:      30,
		TrafficLimitBytes: 1024,
		DevicesLimit:      2,
		Status:            "active",
		Visible:           true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	require.NoError(t, svcCtx.DB.Create(&plan).Error)

	balanceRepo := svcCtx.Repositories.Balance
	_, _, err := balanceRepo.ApplyTransaction(ctx, user.ID, repository.BalanceTransaction{
		Type:        "recharge",
		AmountCents: 5000,
		Currency:    "CNY",
		Reference:   "seed",
		Description: "seed balance",
	})
	require.NoError(t, err)

	claims := security.UserClaims{ID: user.ID, Email: user.Email, Roles: []string{"user"}}
	reqCtx := security.WithUser(ctx, claims)

	logic := NewCreateLogic(reqCtx, svcCtx)
	resp, err := logic.Create(&types.UserCreateOrderRequest{
		PlanID: plan.ID,
	})
	require.NoError(t, err)

	require.Equal(t, repository.OrderStatusPaid, resp.Order.Status)
	require.Equal(t, repository.OrderPaymentStatusSucceeded, resp.Order.PaymentStatus)
	require.Equal(t, repository.PaymentMethodBalance, resp.Order.PaymentMethod)
	require.NotNil(t, resp.Order.PaidAt)
	require.NotNil(t, resp.Transaction)
	require.Equal(t, int64(-1500), resp.Transaction.AmountCents)
	require.Empty(t, resp.Order.Payments)

	balanceSnapshot := resp.Balance
	require.Equal(t, user.ID, balanceSnapshot.UserID)
	require.Equal(t, int64(3500), balanceSnapshot.BalanceCents)

	storedOrder, _, err := svcCtx.Repositories.Order.Get(ctx, resp.Order.ID)
	require.NoError(t, err)
	require.Equal(t, repository.OrderStatusPaid, storedOrder.Status)
	require.Equal(t, repository.OrderPaymentStatusSucceeded, storedOrder.PaymentStatus)
}

func TestCreateOrderWithExternalPayment(t *testing.T) {
	svcCtx, cleanup := setupCreateLogicTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	user := repository.User{
		Email:       "buyer2@test.dev",
		DisplayName: "Buyer 2",
		Roles:       []string{"user"},
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, svcCtx.DB.Create(&user).Error)

	plan := repository.Plan{
		Name:              "Premium",
		Slug:              "premium",
		Description:       "Premium plan",
		PriceCents:        2600,
		Currency:          "CNY",
		DurationDays:      30,
		TrafficLimitBytes: 2048,
		DevicesLimit:      3,
		Status:            "active",
		Visible:           true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	require.NoError(t, svcCtx.DB.Create(&plan).Error)

	claims := security.UserClaims{ID: user.ID, Email: user.Email, Roles: []string{"user"}}
	reqCtx := security.WithUser(ctx, claims)

	logic := NewCreateLogic(reqCtx, svcCtx)
	resp, err := logic.Create(&types.UserCreateOrderRequest{
		PlanID:           plan.ID,
		PaymentMethod:    repository.PaymentMethodExternal,
		PaymentChannel:   "stripe",
		PaymentReturnURL: "https://example.com/return",
	})
	require.NoError(t, err)

	require.Equal(t, repository.OrderStatusPendingPayment, resp.Order.Status)
	require.Equal(t, repository.OrderPaymentStatusPending, resp.Order.PaymentStatus)
	require.Equal(t, repository.PaymentMethodExternal, resp.Order.PaymentMethod)
	require.NotEmpty(t, resp.Order.PaymentIntentID)
	require.Nil(t, resp.Transaction)
	require.Equal(t, int64(0), resp.Balance.BalanceCents)
	require.Len(t, resp.Order.Payments, 1)

	payment := resp.Order.Payments[0]
	require.Equal(t, repository.OrderPaymentStatusPending, payment.Status)
	require.Equal(t, plan.PriceCents, payment.AmountCents)
	require.Equal(t, plan.Currency, payment.Currency)
	require.Equal(t, "stripe", payment.Provider)

	storedOrder, _, err := svcCtx.Repositories.Order.Get(ctx, resp.Order.ID)
	require.NoError(t, err)
	require.Equal(t, repository.OrderStatusPendingPayment, storedOrder.Status)
	require.Equal(t, repository.OrderPaymentStatusPending, storedOrder.PaymentStatus)
	require.Equal(t, repository.PaymentMethodExternal, storedOrder.PaymentMethod)

	paymentsMap, err := svcCtx.Repositories.Order.ListPayments(ctx, []uint64{storedOrder.ID})
	require.NoError(t, err)
	require.Len(t, paymentsMap[storedOrder.ID], 1)
	require.Equal(t, repository.OrderPaymentStatusPending, paymentsMap[storedOrder.ID][0].Status)
}

func TestCreateOrderIdempotent(t *testing.T) {
	svcCtx, cleanup := setupCreateLogicTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	user := repository.User{
		Email:       "buyer3@test.dev",
		DisplayName: "Buyer 3",
		Roles:       []string{"user"},
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, svcCtx.DB.Create(&user).Error)

	plan := repository.Plan{
		Name:              "Standard",
		Slug:              "standard",
		Description:       "Standard plan",
		PriceCents:        2000,
		Currency:          "CNY",
		DurationDays:      30,
		TrafficLimitBytes: 1024,
		DevicesLimit:      2,
		Status:            "active",
		Visible:           true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	require.NoError(t, svcCtx.DB.Create(&plan).Error)

	balanceRepo := svcCtx.Repositories.Balance
	_, _, err := balanceRepo.ApplyTransaction(ctx, user.ID, repository.BalanceTransaction{
		Type:        "recharge",
		AmountCents: 10000,
		Currency:    "CNY",
		Reference:   "seed",
		Description: "seed balance",
	})
	require.NoError(t, err)

	claims := security.UserClaims{ID: user.ID, Email: user.Email, Roles: []string{"user"}}
	reqCtx := security.WithUser(ctx, claims)

	logic := NewCreateLogic(reqCtx, svcCtx)
	idemKey := "order-123"

	first, err := logic.Create(&types.UserCreateOrderRequest{
		PlanID:         plan.ID,
		Quantity:       2,
		IdempotencyKey: idemKey,
	})
	require.NoError(t, err)
	require.NotNil(t, first.Transaction)
	require.Equal(t, int64(6000), first.Balance.BalanceCents)

	second, err := logic.Create(&types.UserCreateOrderRequest{
		PlanID:         plan.ID,
		Quantity:       5, // should be ignored due to idempotency
		IdempotencyKey: idemKey,
	})
	require.NoError(t, err)
	require.Nil(t, second.Transaction)
	require.Equal(t, first.Order.ID, second.Order.ID)
	require.Equal(t, first.Order.Number, second.Order.Number)
	require.Equal(t, int64(6000), second.Balance.BalanceCents)
	require.Equal(t, 2, second.Order.Items[0].Quantity)

	var orderCount int64
	require.NoError(t, svcCtx.DB.Model(&repository.Order{}).Count(&orderCount).Error)
	require.EqualValues(t, 1, orderCount)

	txList, _, err := balanceRepo.ListTransactions(ctx, user.ID, repository.ListBalanceTransactionsOptions{Type: "purchase"})
	require.NoError(t, err)
	require.Len(t, txList, 1)
}
