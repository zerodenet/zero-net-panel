package orders

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

// MarkPaidLogic handles manual payment confirmation for orders.
type MarkPaidLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewMarkPaidLogic constructs the logic handler.
func NewMarkPaidLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkPaidLogic {
	return &MarkPaidLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// MarkPaid transitions an order into paid status and attaches manual metadata.
func (l *MarkPaidLogic) MarkPaid(req *types.AdminMarkOrderPaidRequest) (*types.AdminOrderResponse, error) {
	actor, ok := security.UserFromContext(l.ctx)
	if !ok {
		return nil, repository.ErrUnauthorized
	}
	if !security.HasRole(actor, "admin") {
		return nil, repository.ErrForbidden
	}

	order, items, err := l.svcCtx.Repositories.Order.Get(l.ctx, req.OrderID)
	if err != nil {
		return nil, err
	}

	paymentsMap, err := l.svcCtx.Repositories.Order.ListPayments(l.ctx, []uint64{order.ID})
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(order.Status, repository.OrderStatusPaid) {
		return l.buildResponse(order, items, paymentsMap[order.ID])
	}
	if !strings.EqualFold(order.Status, repository.OrderStatusPendingPayment) {
		return nil, repository.ErrInvalidArgument
	}

	var updated repository.Order
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		repo, err := repository.NewOrderRepository(tx)
		if err != nil {
			return err
		}
		balanceRepo, err := repository.NewBalanceRepository(tx)
		if err != nil {
			return err
		}

		paidAt := time.Now().UTC()
		if req.PaidAt != nil && *req.PaidAt > 0 {
			paidAt = time.Unix(*req.PaidAt, 0).UTC()
		}

		metadata := map[string]any{
			"manual_paid_by": actor.Email,
		}
		note := strings.TrimSpace(req.Note)
		if note != "" {
			metadata["manual_payment_note"] = note
		}
		providedMethod := strings.TrimSpace(req.PaymentMethod)
		if providedMethod != "" {
			metadata["manual_payment_method"] = providedMethod
		}
		method := providedMethod
		if req.ChargeBalance {
			method = repository.PaymentMethodBalance
		}

		if req.ChargeBalance && order.TotalCents > 0 {
			currency := order.Currency
			if strings.TrimSpace(currency) == "" {
				currency = "CNY"
			}
			txRecord := repository.BalanceTransaction{
				Type:        "purchase",
				AmountCents: -order.TotalCents,
				Currency:    currency,
				Reference:   fmt.Sprintf("order:%s", order.Number),
				Description: fmt.Sprintf("Manual charge for order %s", order.Number),
				Metadata: map[string]any{
					"order_id":      order.ID,
					"manual_charge": true,
				},
			}
			if _, _, err := balanceRepo.ApplyTransaction(l.ctx, order.UserID, txRecord); err != nil {
				return err
			}
			metadata["balance_charged"] = true
		}

		stateParams := repository.UpdateOrderPaymentStateParams{
			PaymentStatus: repository.OrderPaymentStatusSucceeded,
			OrderStatus:   pointerOf(repository.OrderStatusPaid),
			PaidAt:        &paidAt,
			MetadataPatch: metadata,
		}
		if ref := strings.TrimSpace(req.Reference); ref != "" {
			stateParams.PaymentReference = &ref
		}

		updatedOrder, err := repo.UpdatePaymentState(l.ctx, req.OrderID, stateParams)
		if err != nil {
			return err
		}
		if method != "" && !strings.EqualFold(updatedOrder.PaymentMethod, method) {
			statusParams := repository.UpdateOrderStatusParams{
				Status:        repository.OrderStatusPaid,
				PaymentMethod: &method,
				PaidAt:        &paidAt,
			}
			refreshed, err := repo.UpdateStatus(l.ctx, req.OrderID, statusParams)
			if err != nil {
				return err
			}
			updatedOrder = refreshed
		}
		updated = updatedOrder
		return nil
	})
	if err != nil {
		return nil, err
	}

	return l.buildResponse(updated, items, paymentsMap[order.ID])
}

func (l *MarkPaidLogic) buildResponse(order repository.Order, items []repository.OrderItem, payments []repository.OrderPayment) (*types.AdminOrderResponse, error) {
	detail := orderutil.ToOrderDetail(order, items, nil, payments)
	u, err := l.svcCtx.Repositories.User.Get(l.ctx, order.UserID)
	if err != nil {
		return nil, err
	}
	resp := types.AdminOrderResponse{
		Order: types.AdminOrderDetail{
			OrderDetail: detail,
			User: types.OrderUserSummary{
				ID:          u.ID,
				Email:       u.Email,
				DisplayName: u.DisplayName,
			},
		},
	}
	return &resp, nil
}

func pointerOf[T any](v T) *T {
	return &v
}
