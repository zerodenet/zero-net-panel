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
	"github.com/zero-net-panel/zero-net-panel/pkg/metrics"
)

// MarkPaidLogic handles manual payment confirmation by administrators.
type MarkPaidLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewMarkPaidLogic constructs MarkPaidLogic.
func NewMarkPaidLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkPaidLogic {
	return &MarkPaidLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// MarkPaid transitions the order into paid state optionally charging the user balance.
func (l *MarkPaidLogic) MarkPaid(req *types.AdminMarkOrderPaidRequest) (resp *types.AdminOrderResponse, err error) {
	start := time.Now()
	paymentMethod := strings.TrimSpace(strings.ToLower(req.PaymentMethod))
	defer func() {
		result := "success"
		if err != nil {
			result = "error"
		}
		metrics.ObserveOrderPayment(paymentMethod, "admin", result, time.Since(start))
	}()

	user, ok := security.UserFromContext(l.ctx)
	if !ok {
		return nil, repository.ErrUnauthorized
	}
	if !security.HasRole(user, "admin") {
		return nil, repository.ErrForbidden
	}
	if req.OrderID == 0 {
		return nil, repository.ErrInvalidArgument
	}

	if paymentMethod == "" {
		paymentMethod = repository.PaymentMethodManual
	}

	var updatedOrder repository.Order
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		orderRepo, repoErr := repository.NewOrderRepository(tx)
		if repoErr != nil {
			return repoErr
		}
		balanceRepo, repoErr := repository.NewBalanceRepository(tx)
		if repoErr != nil {
			return repoErr
		}

		order, repoErr := orderRepo.GetForUpdate(l.ctx, req.OrderID)
		if repoErr != nil {
			return repoErr
		}

		switch order.Status {
		case repository.OrderStatusPending:
			// allowed
		default:
			return repository.ErrInvalidState
		}

		now := time.Now().UTC()
		paidAt := now
		if req.PaidAt != nil && *req.PaidAt > 0 {
			paidAt = time.Unix(*req.PaidAt, 0).UTC()
		}

		if req.ChargeBalance && order.TotalCents > order.RefundedCents {
			amount := order.TotalCents - order.RefundedCents
			if amount <= 0 {
				return repository.ErrInvalidState
			}
			description := fmt.Sprintf("订单支付 %s", order.Number)
			if note := strings.TrimSpace(req.Note); note != "" {
				description = fmt.Sprintf("%s (%s)", description, note)
			}
			txRecord := repository.BalanceTransaction{
				Type:        "purchase",
				AmountCents: -amount,
				Currency:    order.Currency,
				Reference:   fmt.Sprintf("order:%s", order.Number),
				Description: description,
				Metadata: map[string]any{
					"order_id":     order.ID,
					"order_number": order.Number,
					"admin_id":     user.ID,
				},
			}
			if req.Reference != "" {
				txRecord.Metadata["reference"] = req.Reference
			}
			if _, _, repoErr = balanceRepo.ApplyTransaction(l.ctx, order.UserID, txRecord); repoErr != nil {
				return repoErr
			}
			paymentMethod = repository.PaymentMethodBalance
		}

		order.Status = repository.OrderStatusPaid
		order.PaymentMethod = paymentMethod
		order.PaidAt = &paidAt
		order.CancelledAt = nil
		if order.Metadata == nil {
			order.Metadata = make(map[string]any)
		}
		order.Metadata["payment_marked_at"] = now
		order.Metadata["payment_marked_by"] = fmt.Sprintf("admin:%d", user.ID)
		if note := strings.TrimSpace(req.Note); note != "" {
			order.Metadata["payment_note"] = note
		}
		if ref := strings.TrimSpace(req.Reference); ref != "" {
			order.Metadata["payment_reference"] = ref
		}

		saved, repoErr := orderRepo.Save(l.ctx, order)
		if repoErr != nil {
			return repoErr
		}
		updatedOrder = saved
		return nil
	})
	if err != nil {
		return nil, err
	}

	itemsMap, err := l.svcCtx.Repositories.Order.ListItems(l.ctx, []uint64{updatedOrder.ID})
	if err != nil {
		return nil, err
	}
	refundsMap, err := l.svcCtx.Repositories.Order.ListRefunds(l.ctx, []uint64{updatedOrder.ID})
	if err != nil {
		return nil, err
	}

	detail := orderutil.ToOrderDetail(updatedOrder, itemsMap[updatedOrder.ID], refundsMap[updatedOrder.ID])
	u, err := l.svcCtx.Repositories.User.Get(l.ctx, updatedOrder.UserID)
	if err != nil {
		return nil, err
	}

	resp = &types.AdminOrderResponse{
		Order: types.AdminOrderDetail{
			OrderDetail: detail,
			User: types.OrderUserSummary{
				ID:          u.ID,
				Email:       u.Email,
				DisplayName: u.DisplayName,
			},
		},
	}
	return resp, nil
}
