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
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

func setupPaymentCallbackTest(t *testing.T) (*svc.ServiceContext, func()) {
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

func TestPaymentCallbackLogic_Success(t *testing.T) {
	svcCtx, cleanup := setupPaymentCallbackTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	customer := repository.User{
		Email:       "customer@test.dev",
		DisplayName: "Customer",
		Roles:       []string{"user"},
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, svcCtx.DB.Create(&customer).Error)

	plan := repository.Plan{
		Name:              "Premium",
		Slug:              "premium",
		Description:       "Premium plan",
		PriceCents:        3200,
		Currency:          "CNY",
		DurationDays:      30,
		TrafficLimitBytes: 4096,
		DevicesLimit:      5,
		Status:            "active",
		Visible:           true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	require.NoError(t, svcCtx.DB.Create(&plan).Error)

	orderRepo := svcCtx.Repositories.Order

	order, items, err := orderRepo.Create(ctx, repository.Order{
		UserID:        customer.ID,
		PlanID:        &plan.ID,
		Status:        repository.OrderStatusPendingPayment,
		PaymentMethod: repository.PaymentMethodExternal,
		PaymentStatus: repository.OrderPaymentStatusPending,
		TotalCents:    plan.PriceCents,
		Currency:      plan.Currency,
	}, []repository.OrderItem{{
		ItemType:       "plan",
		ItemID:         plan.ID,
		Name:           plan.Name,
		Quantity:       1,
		UnitPriceCents: plan.PriceCents,
		Currency:       plan.Currency,
		SubtotalCents:  plan.PriceCents,
		CreatedAt:      now,
	}})
	require.NoError(t, err)
	require.Len(t, items, 1)

	payment, err := orderRepo.CreatePayment(ctx, repository.OrderPayment{
		OrderID:     order.ID,
		Provider:    "stripe",
		Method:      repository.PaymentMethodExternal,
		Status:      repository.OrderPaymentStatusPending,
		AmountCents: plan.PriceCents,
		Currency:    plan.Currency,
	})
	require.NoError(t, err)

	logic := NewPaymentCallbackLogic(ctx, svcCtx)
	paidAt := now.Add(2 * time.Minute).Unix()
	resp, err := logic.Process(&types.AdminPaymentCallbackRequest{
		OrderID:   order.ID,
		PaymentID: payment.ID,
		Status:    repository.OrderPaymentStatusSucceeded,
		Reference: "gateway-ref",
		PaidAt:    &paidAt,
	})
	require.NoError(t, err)

	require.Equal(t, repository.OrderStatusPaid, resp.Order.OrderDetail.Status)
	require.Equal(t, repository.OrderPaymentStatusSucceeded, resp.Order.OrderDetail.PaymentStatus)
	require.Equal(t, "gateway-ref", resp.Order.OrderDetail.PaymentReference)
	require.NotNil(t, resp.Order.OrderDetail.PaidAt)
	require.Len(t, resp.Order.OrderDetail.Payments, 1)
	require.Equal(t, repository.OrderPaymentStatusSucceeded, resp.Order.OrderDetail.Payments[0].Status)

	storedOrder, _, err := orderRepo.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, repository.OrderStatusPaid, storedOrder.Status)
	require.Equal(t, repository.OrderPaymentStatusSucceeded, storedOrder.PaymentStatus)
	require.Equal(t, "gateway-ref", storedOrder.PaymentReference)

	paymentsMap, err := orderRepo.ListPayments(ctx, []uint64{order.ID})
	require.NoError(t, err)
	require.Len(t, paymentsMap[order.ID], 1)
	require.Equal(t, repository.OrderPaymentStatusSucceeded, paymentsMap[order.ID][0].Status)
}

func TestPaymentCallbackLogic_Failed(t *testing.T) {
	svcCtx, cleanup := setupPaymentCallbackTest(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().UTC()

	customer := repository.User{
		Email:       "customer2@test.dev",
		DisplayName: "Customer 2",
		Roles:       []string{"user"},
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, svcCtx.DB.Create(&customer).Error)

	orderRepo := svcCtx.Repositories.Order

	order, _, err := orderRepo.Create(ctx, repository.Order{
		UserID:        customer.ID,
		Status:        repository.OrderStatusPendingPayment,
		PaymentMethod: repository.PaymentMethodExternal,
		PaymentStatus: repository.OrderPaymentStatusPending,
		TotalCents:    4200,
		Currency:      "CNY",
	}, []repository.OrderItem{})
	require.NoError(t, err)

	payment, err := orderRepo.CreatePayment(ctx, repository.OrderPayment{
		OrderID:     order.ID,
		Provider:    "alipay",
		Method:      repository.PaymentMethodExternal,
		Status:      repository.OrderPaymentStatusPending,
		AmountCents: 4200,
		Currency:    "CNY",
	})
	require.NoError(t, err)

	logic := NewPaymentCallbackLogic(ctx, svcCtx)
	failureCode := "timeout"
	failureMsg := "payment timeout"
	resp, err := logic.Process(&types.AdminPaymentCallbackRequest{
		OrderID:        order.ID,
		PaymentID:      payment.ID,
		Status:         repository.OrderPaymentStatusFailed,
		FailureCode:    failureCode,
		FailureMessage: failureMsg,
	})
	require.NoError(t, err)

	require.Equal(t, repository.OrderStatusPaymentFailed, resp.Order.OrderDetail.Status)
	require.Equal(t, repository.OrderPaymentStatusFailed, resp.Order.OrderDetail.PaymentStatus)
	require.Equal(t, failureCode, resp.Order.OrderDetail.PaymentFailureCode)
	require.Equal(t, failureMsg, resp.Order.OrderDetail.PaymentFailureMessage)
	require.Len(t, resp.Order.OrderDetail.Payments, 1)
	require.Equal(t, repository.OrderPaymentStatusFailed, resp.Order.OrderDetail.Payments[0].Status)

	storedOrder, _, err := orderRepo.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, repository.OrderStatusPaymentFailed, storedOrder.Status)
	require.Equal(t, repository.OrderPaymentStatusFailed, storedOrder.PaymentStatus)
	require.Equal(t, failureCode, storedOrder.PaymentFailureCode)
	require.Equal(t, failureMsg, storedOrder.PaymentFailureReason)

	paymentsMap, err := orderRepo.ListPayments(ctx, []uint64{order.ID})
	require.NoError(t, err)
	require.Equal(t, repository.OrderPaymentStatusFailed, paymentsMap[order.ID][0].Status)
}
