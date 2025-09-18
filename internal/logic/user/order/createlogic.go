package order

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/zero-net-panel/zero-net-panel/internal/logic/orderutil"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/security"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// CreateLogic handles user order creation.
type CreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateLogic constructs CreateLogic.
func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Create issues an order for the given plan and settles payment via balance.
func (l *CreateLogic) Create(req *types.UserCreateOrderRequest) (*types.UserOrderResponse, error) {
	user, ok := security.UserFromContext(l.ctx)
	if !ok {
		return nil, repository.ErrUnauthorized
	}

	if req.PlanID == 0 {
		return nil, repository.ErrInvalidArgument
	}

	plan, err := l.svcCtx.Repositories.Plan.Get(l.ctx, req.PlanID)
	if err != nil {
		return nil, err
	}

	if !plan.Visible || !strings.EqualFold(plan.Status, "active") {
		return nil, repository.ErrInvalidArgument
	}

	quantity := req.Quantity
	if quantity <= 0 {
		quantity = 1
	}
	if quantity > 10 {
		quantity = 10
	}

	totalCents := plan.PriceCents * int64(quantity)
	orderNumber := repository.GenerateOrderNumber()

	var createdOrder repository.Order
	var createdItems []repository.OrderItem
	var balance repository.UserBalance
	var balanceTx repository.BalanceTransaction

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		orderRepo, err := repository.NewOrderRepository(tx)
		if err != nil {
			return err
		}
		balanceRepo, err := repository.NewBalanceRepository(tx)
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		currency := plan.Currency
		if strings.TrimSpace(currency) == "" {
			currency = balance.Currency
			if strings.TrimSpace(currency) == "" {
				currency = "CNY"
			}
		}

		snapshot := map[string]any{
			"id":                  plan.ID,
			"name":                plan.Name,
			"slug":                plan.Slug,
			"description":         plan.Description,
			"price_cents":         plan.PriceCents,
			"currency":            currency,
			"duration_days":       plan.DurationDays,
			"traffic_limit_bytes": plan.TrafficLimitBytes,
			"devices_limit":       plan.DevicesLimit,
			"features":            plan.Features,
			"tags":                plan.Tags,
		}

		orderModel := repository.Order{
			Number:        orderNumber,
			UserID:        user.ID,
			PlanID:        &plan.ID,
			Status:        repository.OrderStatusPending,
			PaymentMethod: repository.PaymentMethodBalance,
			TotalCents:    totalCents,
			Currency:      currency,
			Metadata: map[string]any{
				"quantity": quantity,
			},
			PlanSnapshot: snapshot,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if totalCents == 0 {
			orderModel.Status = repository.OrderStatusPaid
			paidAt := now
			orderModel.PaidAt = &paidAt
		}

		if totalCents > 0 {
			txRecord := repository.BalanceTransaction{
				Type:        "purchase",
				AmountCents: -totalCents,
				Currency:    currency,
				Reference:   fmt.Sprintf("order:%s", orderNumber),
				Description: fmt.Sprintf("购买套餐 %s", plan.Name),
				Metadata: map[string]any{
					"plan_id":      plan.ID,
					"quantity":     quantity,
					"order_number": orderNumber,
				},
			}
			createdTx, updatedBalance, err := balanceRepo.ApplyTransaction(l.ctx, user.ID, txRecord)
			if err != nil {
				return err
			}
			balanceTx = createdTx
			balance = updatedBalance
			paidAt := createdTx.CreatedAt
			orderModel.Status = repository.OrderStatusPaid
			orderModel.PaidAt = &paidAt
		} else {
			existingBalance, err := balanceRepo.GetBalance(l.ctx, user.ID)
			if err != nil {
				return err
			}
			balance = existingBalance
		}

		item := repository.OrderItem{
			ItemType:       "plan",
			ItemID:         plan.ID,
			Name:           plan.Name,
			Quantity:       quantity,
			UnitPriceCents: plan.PriceCents,
			Currency:       currency,
			SubtotalCents:  totalCents,
			Metadata: map[string]any{
				"duration_days":       plan.DurationDays,
				"traffic_limit_bytes": plan.TrafficLimitBytes,
				"devices_limit":       plan.DevicesLimit,
			},
			CreatedAt: now,
		}

		created, items, err := orderRepo.Create(l.ctx, orderModel, []repository.OrderItem{item})
		if err != nil {
			return err
		}
		createdOrder = created
		createdItems = items
		return nil
	})
	if err != nil {
		return nil, err
	}

	detail := orderutil.ToOrderDetail(createdOrder, createdItems)
	balanceView := orderutil.ToBalanceSnapshot(balance)

	var txView *types.BalanceTransactionSummary
	if balanceTx.ID != 0 {
		summary := orderutil.ToBalanceTransactionView(balanceTx)
		txView = &summary
	}

	resp := types.UserOrderResponse{
		Order:       detail,
		Balance:     balanceView,
		Transaction: txView,
	}

	return &resp, nil
}
