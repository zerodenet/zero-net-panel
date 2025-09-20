package order

import (
	"context"
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

// CancelLogic handles user initiated order cancellations.
type CancelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCancelLogic constructs CancelLogic.
func NewCancelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelLogic {
	return &CancelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Cancel transitions a pending order into the cancelled state.
func (l *CancelLogic) Cancel(req *types.UserCancelOrderRequest) (*types.UserOrderResponse, error) {
	user, ok := security.UserFromContext(l.ctx)
	if !ok {
		return nil, repository.ErrUnauthorized
	}

	order, items, err := l.svcCtx.Repositories.Order.Get(l.ctx, req.OrderID)
	if err != nil {
		return nil, err
	}
	if order.UserID != user.ID {
		return nil, repository.ErrForbidden
	}

	paymentsMap, err := l.svcCtx.Repositories.Order.ListPayments(l.ctx, []uint64{order.ID})
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(order.Status, repository.OrderStatusCancelled) {
		balance, err := l.svcCtx.Repositories.Balance.GetBalance(l.ctx, user.ID)
		if err != nil {
			return nil, err
		}
		detail := orderutil.ToOrderDetail(order, items, nil, paymentsMap[order.ID])
		resp := types.UserOrderResponse{
			Order:   detail,
			Balance: orderutil.ToBalanceSnapshot(balance),
		}
		return &resp, nil
	}

	if !strings.EqualFold(order.Status, repository.OrderStatusPending) {
		if !strings.EqualFold(order.Status, repository.OrderStatusPaid) || order.TotalCents != 0 {
			return nil, repository.ErrInvalidArgument
		}
	}

	var updated repository.Order
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		repo, err := repository.NewOrderRepository(tx)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		metadata := map[string]any{
			"cancelled_by": "user",
		}
		reason := strings.TrimSpace(req.Reason)
		if reason != "" {
			metadata["cancel_reason"] = reason
		}
		params := repository.UpdateOrderStatusParams{
			Status:        repository.OrderStatusCancelled,
			CancelledAt:   &now,
			MetadataPatch: metadata,
		}
		updatedOrder, err := repo.UpdateStatus(l.ctx, req.OrderID, params)
		if err != nil {
			return err
		}
		updated = updatedOrder
		return nil
	})
	if err != nil {
		return nil, err
	}

	balance, err := l.svcCtx.Repositories.Balance.GetBalance(l.ctx, user.ID)
	if err != nil {
		return nil, err
	}

	detail := orderutil.ToOrderDetail(updated, items, nil, paymentsMap[order.ID])
	resp := types.UserOrderResponse{
		Order:   detail,
		Balance: orderutil.ToBalanceSnapshot(balance),
	}
	return &resp, nil
}
