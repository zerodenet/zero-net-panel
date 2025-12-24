package order

import (
	"context"
	"errors"
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
	"github.com/zero-net-panel/zero-net-panel/pkg/metrics"
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

// Create issues an order for the given plan and settles payment according to the selected method.
func (l *CreateLogic) Create(req *types.UserCreateOrderRequest) (resp *types.UserOrderResponse, err error) {
	start := time.Now()
	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	var idemPtr *string
	if idempotencyKey != "" {
		idemPtr = &idempotencyKey
	}
	method := strings.TrimSpace(strings.ToLower(req.PaymentMethod))
	if method == "" {
		method = repository.PaymentMethodBalance
	}
	paymentMethod := method
	defer func() {
		result := "success"
		if err != nil {
			result = "error"
		}
		metrics.ObserveOrderCreate(paymentMethod, result, time.Since(start))
	}()

	if method != repository.PaymentMethodBalance && method != repository.PaymentMethodExternal {
		return nil, repository.ErrInvalidArgument
	}

	user, ok := security.UserFromContext(l.ctx)
	if !ok {
		return nil, repository.ErrUnauthorized
	}

	if idempotencyKey != "" {
		if existing, items, payments, err := l.svcCtx.Repositories.Order.GetByIdempotencyKey(l.ctx, user.ID, idempotencyKey); err == nil {
			balance, balErr := l.svcCtx.Repositories.Balance.GetBalance(l.ctx, user.ID)
			if balErr != nil {
				return nil, balErr
			}
			detail := orderutil.ToOrderDetail(existing, items, nil, payments)
			resp := &types.UserOrderResponse{
				Order:   detail,
				Balance: orderutil.ToBalanceSnapshot(balance),
			}
			return resp, nil
		} else if !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
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

	channel := strings.TrimSpace(strings.ToLower(req.PaymentChannel))
	returnURL := strings.TrimSpace(req.PaymentReturnURL)

	totalCents := plan.PriceCents * int64(quantity)
	if method == repository.PaymentMethodExternal && totalCents > 0 && channel == "" {
		return nil, repository.ErrInvalidArgument
	}

	orderNumber := repository.GenerateOrderNumber()

	var createdOrder repository.Order
	var createdItems []repository.OrderItem
	var createdPayments []repository.OrderPayment
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
		existingBalance, err := balanceRepo.GetBalance(l.ctx, user.ID)
		if err != nil {
			return err
		}
		balance = existingBalance

		currency := strings.TrimSpace(plan.Currency)
		if currency == "" {
			currency = strings.TrimSpace(balance.Currency)
			if currency == "" {
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

		metadata := map[string]any{
			"quantity": quantity,
		}
		if channel != "" {
			metadata["payment_channel"] = channel
		}
		if returnURL != "" {
			metadata["payment_return_url"] = returnURL
		}

		orderModel := repository.Order{
			Number:         orderNumber,
			UserID:         user.ID,
			IdempotencyKey: idemPtr,
			PlanID:         &plan.ID,
			Status:         repository.OrderStatusPendingPayment,
			PaymentMethod:  method,
			PaymentStatus:  repository.OrderPaymentStatusPending,
			TotalCents:     totalCents,
			Currency:       currency,
			Metadata:       metadata,
			PlanSnapshot:   snapshot,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		if totalCents == 0 || method == repository.PaymentMethodBalance {
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
				paidAt := createdTx.CreatedAt.UTC()
				orderModel.Status = repository.OrderStatusPaid
				orderModel.PaymentStatus = repository.OrderPaymentStatusSucceeded
				orderModel.PaidAt = &paidAt
			} else {
				paidAt := now
				orderModel.Status = repository.OrderStatusPaid
				orderModel.PaymentStatus = repository.OrderPaymentStatusSucceeded
				orderModel.PaidAt = &paidAt
			}
		} else {
			intentID := fmt.Sprintf("%s-%s", channel, orderNumber)
			if channel == "" {
				intentID = orderNumber
			}
			orderModel.PaymentIntentID = intentID
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

		if method == repository.PaymentMethodExternal && totalCents > 0 {
			paymentMetadata := map[string]any{}
			if channel != "" {
				paymentMetadata["channel"] = channel
			}
			if returnURL != "" {
				paymentMetadata["return_url"] = returnURL
			}
			paymentRecord := repository.OrderPayment{
				OrderID:     created.ID,
				Provider:    channel,
				Method:      method,
				IntentID:    created.PaymentIntentID,
				Status:      repository.OrderPaymentStatusPending,
				AmountCents: totalCents,
				Currency:    currency,
				Metadata:    paymentMetadata,
			}
			payment, err := orderRepo.CreatePayment(l.ctx, paymentRecord)
			if err != nil {
				return err
			}
			createdPayments = append(createdPayments, payment)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, repository.ErrConflict) && idempotencyKey != "" {
			existing, items, payments, fetchErr := l.svcCtx.Repositories.Order.GetByIdempotencyKey(l.ctx, user.ID, idempotencyKey)
			if fetchErr == nil {
				balance, balErr := l.svcCtx.Repositories.Balance.GetBalance(l.ctx, user.ID)
				if balErr != nil {
					return nil, balErr
				}
				detail := orderutil.ToOrderDetail(existing, items, nil, payments)
				return &types.UserOrderResponse{
					Order:   detail,
					Balance: orderutil.ToBalanceSnapshot(balance),
				}, nil
			}
		}
		return nil, err
	}

	detail := orderutil.ToOrderDetail(createdOrder, createdItems, nil, createdPayments)
	balanceView := orderutil.ToBalanceSnapshot(balance)

	var txView *types.BalanceTransactionSummary
	if balanceTx.ID != 0 {
		summary := orderutil.ToBalanceTransactionView(balanceTx)
		txView = &summary
	}

	resp = &types.UserOrderResponse{
		Order:       detail,
		Balance:     balanceView,
		Transaction: txView,
	}

	return resp, nil
}
