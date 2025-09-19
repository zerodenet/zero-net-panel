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

// RefundLogic handles administrative order refund operations.
type RefundLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRefundLogic constructs RefundLogic.
func NewRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundLogic {
	return &RefundLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Refund records a refund entry and optionally credits the user balance.
func (l *RefundLogic) Refund(req *types.AdminRefundOrderRequest) (resp *types.AdminOrderResponse, err error) {
	start := time.Now()
	defer func() {
		result := "success"
		if err != nil {
			result = "error"
		}
		metrics.ObserveOrderRefund("admin", float64(req.AmountCents)/100.0, result, time.Since(start))
	}()

	if req.AmountCents <= 0 {
		return nil, repository.ErrInvalidArgument
	}

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
		case repository.OrderStatusPaid, repository.OrderStatusPartiallyRefunded, repository.OrderStatusRefunded:
			// allow
		default:
			return repository.ErrInvalidState
		}

		available := order.TotalCents - order.RefundedCents
		if available <= 0 || req.AmountCents > available {
			return repository.ErrInvalidArgument
		}

		now := time.Now().UTC()

		refundRecord := repository.OrderRefund{
			OrderID:     order.ID,
			AmountCents: req.AmountCents,
			Reason:      strings.TrimSpace(req.Reason),
			Reference:   strings.TrimSpace(req.Reference),
			Metadata:    req.Metadata,
		}
		if refundRecord.Metadata == nil {
			refundRecord.Metadata = make(map[string]any)
		}
		refundRecord.Metadata["admin_id"] = user.ID

		if req.CreditBalance {
			description := fmt.Sprintf("订单退款 %s", order.Number)
			if refundRecord.Reason != "" {
				description = fmt.Sprintf("%s (%s)", description, refundRecord.Reason)
			}
			txRecord := repository.BalanceTransaction{
				Type:        "refund",
				AmountCents: req.AmountCents,
				Currency:    order.Currency,
				Reference:   fmt.Sprintf("order:%s", order.Number),
				Description: description,
				Metadata: map[string]any{
					"order_id":     order.ID,
					"order_number": order.Number,
					"admin_id":     user.ID,
				},
			}
			if refundRecord.Reference != "" {
				txRecord.Metadata["reference"] = refundRecord.Reference
			}
			if _, _, repoErr = balanceRepo.ApplyTransaction(l.ctx, order.UserID, txRecord); repoErr != nil {
				return repoErr
			}
		}

		if _, repoErr = orderRepo.CreateRefund(l.ctx, refundRecord); repoErr != nil {
			return repoErr
		}

		order.RefundedCents += req.AmountCents
		if order.RefundedCents >= order.TotalCents {
			order.RefundedCents = order.TotalCents
			order.Status = repository.OrderStatusRefunded
		} else {
			order.Status = repository.OrderStatusPartiallyRefunded
		}
		order.RefundedAt = &now
		if order.Metadata == nil {
			order.Metadata = make(map[string]any)
		}
		order.Metadata["last_refund_by"] = fmt.Sprintf("admin:%d", user.ID)
		order.Metadata["last_refund_at"] = now
		order.Metadata["last_refund_amount"] = req.AmountCents
		if refundRecord.Reason != "" {
			order.Metadata["last_refund_reason"] = refundRecord.Reason
		}
		if refundRecord.Reference != "" {
			order.Metadata["last_refund_reference"] = refundRecord.Reference
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
