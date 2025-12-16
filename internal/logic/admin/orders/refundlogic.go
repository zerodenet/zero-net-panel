package orders

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/zero-net-panel/zero-net-panel/internal/logic/orderutil"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/security"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// RefundLogic handles order refund operations initiated by administrators.
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

// Refund credits balance back to the user and records refund metadata.
func (l *RefundLogic) Refund(req *types.AdminRefundOrderRequest) (*types.AdminOrderResponse, error) {
	actor, ok := security.UserFromContext(l.ctx)
	if !ok {
		return nil, repository.ErrUnauthorized
	}
	if !security.HasRole(actor, "admin") {
		return nil, repository.ErrForbidden
	}

	if req.AmountCents <= 0 {
		return nil, repository.ErrInvalidArgument
	}

	order, items, err := l.svcCtx.Repositories.Order.Get(l.ctx, req.OrderID)
	if err != nil {
		return nil, err
	}

	paymentsMap, err := l.svcCtx.Repositories.Order.ListPayments(l.ctx, []uint64{order.ID})
	if err != nil {
		return nil, err
	}

	if order.TotalCents <= 0 {
		return nil, repository.ErrInvalidArgument
	}
	if !strings.EqualFold(order.PaymentMethod, repository.PaymentMethodBalance) {
		return nil, repository.ErrInvalidArgument
	}
	remaining := order.TotalCents - order.RefundedCents
	if remaining <= 0 || req.AmountCents > remaining {
		return nil, repository.ErrInvalidArgument
	}

	var (
		updated repository.Order
	)
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		orderRepo, err := repository.NewOrderRepository(tx)
		if err != nil {
			return err
		}
		balanceRepo, err := repository.NewBalanceRepository(tx)
		if err != nil {
			return err
		}

		description := fmt.Sprintf("订单 %s 退款", order.Number)
		reason := strings.TrimSpace(req.Reason)
		if reason != "" {
			description = fmt.Sprintf("%s（%s）", description, reason)
		}

		metadata := map[string]any{
			"order_id":     order.ID,
			"order_number": order.Number,
			"operator":     actor.Email,
		}
		for k, v := range req.Metadata {
			metadata[k] = v
		}
		if reason != "" {
			metadata["reason"] = reason
		}

		txRecord := repository.BalanceTransaction{
			Type:        "refund",
			AmountCents: req.AmountCents,
			Currency:    order.Currency,
			Reference:   fmt.Sprintf("order:%s", order.Number),
			Description: description,
			Metadata:    metadata,
		}

		createdTx, _, err := balanceRepo.RecordRefund(l.ctx, order.UserID, txRecord)
		if err != nil {
			return err
		}

		refundEntryMetadata := map[string]any{
			"balance_tx_id": createdTx.ID,
			"operator":      actor.Email,
		}
		for k, v := range req.Metadata {
			refundEntryMetadata[k] = v
		}
		if reason != "" {
			refundEntryMetadata["reason"] = reason
		}

		refundRecord := repository.OrderRefund{
			OrderID:     order.ID,
			AmountCents: req.AmountCents,
			Reason:      reason,
			Reference:   txRecord.Reference,
			Metadata:    refundEntryMetadata,
		}
		if _, err := orderRepo.CreateRefund(l.ctx, refundRecord); err != nil {
			return err
		}

		metadataPatch := map[string]any{
			"last_refund_amount": req.AmountCents,
			"last_refund_tx_id":  createdTx.ID,
			"last_refund_by":     actor.Email,
		}
		if reason != "" {
			metadataPatch["last_refund_reason"] = reason
		}

		refundParams := repository.AddRefundParams{
			AmountCents:   req.AmountCents,
			RefundAt:      createdTx.CreatedAt,
			MetadataPatch: metadataPatch,
		}

		updatedOrder, err := orderRepo.AddRefund(l.ctx, req.OrderID, refundParams)
		if err != nil {
			return err
		}

		if updatedOrder.RefundedCents > 0 && updatedOrder.RefundedCents < order.TotalCents && !strings.EqualFold(updatedOrder.Status, repository.OrderStatusPartiallyRefunded) {
			partialStatus := repository.UpdateOrderStatusParams{Status: repository.OrderStatusPartiallyRefunded}
			partiallyUpdated, err := orderRepo.UpdateStatus(l.ctx, req.OrderID, partialStatus)
			if err != nil {
				return err
			}
			updatedOrder = partiallyUpdated
		}

		if updatedOrder.RefundedCents >= order.TotalCents && !strings.EqualFold(updatedOrder.Status, repository.OrderStatusRefunded) {
			cancelledMetadata := map[string]any{
				"cancelled_by":  actor.Email,
				"cancel_reason": "refund_completed",
			}
			cancelParams := repository.UpdateOrderStatusParams{
				Status:        repository.OrderStatusRefunded,
				MetadataPatch: cancelledMetadata,
			}
			cancelledOrder, err := orderRepo.UpdateStatus(l.ctx, req.OrderID, cancelParams)
			if err != nil {
				return err
			}
			updatedOrder = cancelledOrder
		}

		updated = updatedOrder
		return nil
	})
	if err != nil {
		return nil, err
	}

	refundsMap, err := l.svcCtx.Repositories.Order.ListRefunds(l.ctx, []uint64{updated.ID})
	if err != nil {
		return nil, err
	}

	detail := orderutil.ToOrderDetail(updated, items, refundsMap[updated.ID], paymentsMap[order.ID])
	u, err := l.svcCtx.Repositories.User.Get(l.ctx, updated.UserID)
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
