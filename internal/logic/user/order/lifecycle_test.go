package order

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/zero-net-panel/zero-net-panel/internal/bootstrap/migrations"
	adminorders "github.com/zero-net-panel/zero-net-panel/internal/logic/admin/orders"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/security"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

func setupTestServiceContext(t *testing.T) (*svc.ServiceContext, context.Context) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	ctx := context.Background()
	if _, err := migrations.Apply(ctx, db, 0, false, false); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	repos, err := repository.NewRepositories(db)
	if err != nil {
		t.Fatalf("create repositories: %v", err)
	}

	svcCtx := &svc.ServiceContext{
		DB:           db,
		Repositories: repos,
	}

	return svcCtx, ctx
}

func TestOrderLifecycle(t *testing.T) {
	svcCtx, ctx := setupTestServiceContext(t)

	now := time.Now().UTC()

	customer := repository.User{
		Email:       "user@example.com",
		DisplayName: "Test User",
		Roles:       []string{"user"},
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := svcCtx.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}

	admin := repository.User{
		Email:       "admin@example.com",
		DisplayName: "Admin",
		Roles:       []string{"admin"},
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := svcCtx.DB.Create(&admin).Error; err != nil {
		t.Fatalf("create admin: %v", err)
	}

	orderRepo := svcCtx.Repositories.Order

	// Prepare order for user cancellation.
	cancelOrder, _, err := orderRepo.Create(ctx, repository.Order{
		UserID:        customer.ID,
		Status:        repository.OrderStatusPending,
		PaymentMethod: repository.PaymentMethodBalance,
		TotalCents:    1200,
		Currency:      "CNY",
		Metadata:      map[string]any{},
	}, []repository.OrderItem{{
		ItemType:       "plan",
		ItemID:         1,
		Name:           "Plan A",
		Quantity:       1,
		UnitPriceCents: 1200,
		Currency:       "CNY",
		SubtotalCents:  1200,
	}})
	if err != nil {
		t.Fatalf("create cancel order: %v", err)
	}

	// User cancels pending order.
	userCtx := security.WithUser(context.Background(), security.UserClaims{ID: customer.ID, Roles: []string{"user"}})
	cancelLogic := NewCancelLogic(userCtx, svcCtx)
	cancelResp, err := cancelLogic.Cancel(&types.UserCancelOrderRequest{OrderID: cancelOrder.ID, Reason: "no longer needed"})
	if err != nil {
		t.Fatalf("cancel order: %v", err)
	}
	if cancelResp.Order.Status != repository.OrderStatusCancelled {
		t.Fatalf("expected order status cancelled, got %s", cancelResp.Order.Status)
	}
	if cancelResp.Order.Metadata["cancel_reason"] != "no longer needed" {
		t.Fatalf("expected cancel reason metadata, got %v", cancelResp.Order.Metadata["cancel_reason"])
	}

	// Prepare second order for manual payment and refunds.
	payOrder, _, err := orderRepo.Create(ctx, repository.Order{
		UserID:        customer.ID,
		Status:        repository.OrderStatusPending,
		PaymentMethod: repository.PaymentMethodBalance,
		TotalCents:    2000,
		Currency:      "CNY",
		Metadata:      map[string]any{},
	}, []repository.OrderItem{{
		ItemType:       "plan",
		ItemID:         2,
		Name:           "Plan B",
		Quantity:       1,
		UnitPriceCents: 2000,
		Currency:       "CNY",
		SubtotalCents:  2000,
	}})
	if err != nil {
		t.Fatalf("create pay order: %v", err)
	}

	balanceRepo := svcCtx.Repositories.Balance
	if _, _, err := balanceRepo.ApplyTransaction(ctx, customer.ID, repository.BalanceTransaction{
		Type:        "recharge",
		AmountCents: 5000,
		Currency:    "CNY",
		Reference:   "test-topup",
		Description: "seed balance",
	}); err != nil {
		t.Fatalf("seed balance: %v", err)
	}

	adminCtx := security.WithUser(context.Background(), security.UserClaims{ID: admin.ID, Roles: []string{"admin"}})

	markLogic := adminorders.NewMarkPaidLogic(adminCtx, svcCtx)
	markResp, err := markLogic.MarkPaid(&types.AdminMarkOrderPaidRequest{
		OrderID:       payOrder.ID,
		PaymentMethod: "manual",
		Note:          "manual charge",
		Reference:     "manual-ref",
		ChargeBalance: true,
	})
	if err != nil {
		t.Fatalf("mark order paid: %v", err)
	}
	if markResp.Order.OrderDetail.Status != repository.OrderStatusPaid {
		t.Fatalf("expected paid status, got %s", markResp.Order.OrderDetail.Status)
	}
	if markResp.Order.OrderDetail.PaymentMethod != repository.PaymentMethodBalance {
		t.Fatalf("expected payment method balance, got %s", markResp.Order.OrderDetail.PaymentMethod)
	}

	balanceAfterPay, err := balanceRepo.GetBalance(ctx, customer.ID)
	if err != nil {
		t.Fatalf("get balance after pay: %v", err)
	}
	if balanceAfterPay.BalanceCents != 3000 {
		t.Fatalf("expected balance 3000 after charge, got %d", balanceAfterPay.BalanceCents)
	}

	refundLogic := adminorders.NewRefundLogic(adminCtx, svcCtx)
	half := payOrder.TotalCents / 2
	refundResp1, err := refundLogic.Refund(&types.AdminRefundOrderRequest{
		OrderID:       payOrder.ID,
		AmountCents:   half,
		Reason:        "partial",
		CreditBalance: true,
	})
	if err != nil {
		t.Fatalf("partial refund: %v", err)
	}
	if refundResp1.Order.OrderDetail.Status != repository.OrderStatusPartiallyRefunded {
		t.Fatalf("expected partially_refunded status, got %s", refundResp1.Order.OrderDetail.Status)
	}
	if refundResp1.Order.OrderDetail.RefundedCents != half {
		t.Fatalf("expected refunded cents %d, got %d", half, refundResp1.Order.OrderDetail.RefundedCents)
	}

	balanceAfterPartial, err := balanceRepo.GetBalance(ctx, customer.ID)
	if err != nil {
		t.Fatalf("get balance after partial refund: %v", err)
	}
	if balanceAfterPartial.BalanceCents != 3000+half {
		t.Fatalf("expected balance %d after partial refund, got %d", 3000+half, balanceAfterPartial.BalanceCents)
	}

	finalAmount := payOrder.TotalCents - half
	refundResp2, err := refundLogic.Refund(&types.AdminRefundOrderRequest{
		OrderID:       payOrder.ID,
		AmountCents:   finalAmount,
		Reason:        "final",
		CreditBalance: true,
	})
	if err != nil {
		t.Fatalf("final refund: %v", err)
	}
	if refundResp2.Order.OrderDetail.Status != repository.OrderStatusRefunded {
		t.Fatalf("expected refunded status, got %s", refundResp2.Order.OrderDetail.Status)
	}
	if refundResp2.Order.OrderDetail.RefundedCents != payOrder.TotalCents {
		t.Fatalf("expected refunded cents %d, got %d", payOrder.TotalCents, refundResp2.Order.OrderDetail.RefundedCents)
	}
	if len(refundResp2.Order.OrderDetail.Refunds) != 2 {
		t.Fatalf("expected 2 refund records, got %d", len(refundResp2.Order.OrderDetail.Refunds))
	}

	balanceAfterFull, err := balanceRepo.GetBalance(ctx, customer.ID)
	if err != nil {
		t.Fatalf("get balance after full refund: %v", err)
	}
	if balanceAfterFull.BalanceCents != 5000 {
		t.Fatalf("expected balance restored to 5000, got %d", balanceAfterFull.BalanceCents)
	}

	refundsMap, err := orderRepo.ListRefunds(ctx, []uint64{payOrder.ID})
	if err != nil {
		t.Fatalf("list refunds: %v", err)
	}
	if len(refundsMap[payOrder.ID]) != 2 {
		t.Fatalf("expected repository to return 2 refunds, got %d", len(refundsMap[payOrder.ID]))
	}

	// Admin cancel another pending order.
	cancelAdminOrder, _, err := orderRepo.Create(ctx, repository.Order{
		UserID:        customer.ID,
		Status:        repository.OrderStatusPending,
		PaymentMethod: repository.PaymentMethodBalance,
		TotalCents:    1500,
		Currency:      "CNY",
		Metadata:      map[string]any{},
	}, []repository.OrderItem{{
		ItemType:       "plan",
		ItemID:         3,
		Name:           "Plan C",
		Quantity:       1,
		UnitPriceCents: 1500,
		Currency:       "CNY",
		SubtotalCents:  1500,
	}})
	if err != nil {
		t.Fatalf("create admin cancel order: %v", err)
	}

	adminCancelLogic := adminorders.NewCancelLogic(adminCtx, svcCtx)
	cancelRespAdmin, err := adminCancelLogic.Cancel(&types.AdminCancelOrderRequest{OrderID: cancelAdminOrder.ID, Reason: "fraud"})
	if err != nil {
		t.Fatalf("admin cancel: %v", err)
	}
	if cancelRespAdmin.Order.OrderDetail.Status != repository.OrderStatusCancelled {
		t.Fatalf("expected admin cancelled status, got %s", cancelRespAdmin.Order.OrderDetail.Status)
	}
}
