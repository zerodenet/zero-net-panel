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
	"github.com/zero-net-panel/zero-net-panel/pkg/metrics"
)

// CancelLogic handles user initiated order cancellation.
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

// Cancel transitions a pending order into cancelled status and returns the updated snapshot.
func (l *CancelLogic) Cancel(req *types.UserCancelOrderRequest) (resp *types.UserOrderResponse, err error) {
	start := time.Now()
	defer func() {
		result := "success"
		if err != nil {
			result = "error"
		}
		metrics.ObserveOrderCancel("user", result, time.Since(start))
	}()

	user, ok := security.UserFromContext(l.ctx)
	if !ok {
		return nil, repository.ErrUnauthorized
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

		order, repoErr := orderRepo.GetForUpdate(l.ctx, req.OrderID)
		if repoErr != nil {
			return repoErr
		}
		if order.UserID != user.ID {
			return repository.ErrForbidden
		}

		if order.Status == repository.OrderStatusCancelled {
			updatedOrder = order
			return nil
		}
		if order.Status != repository.OrderStatusPending {
			return repository.ErrInvalidState
		}

		now := time.Now().UTC()
		order.Status = repository.OrderStatusCancelled
		order.CancelledAt = &now
		if order.Metadata == nil {
			order.Metadata = make(map[string]any)
		}
		if reason := strings.TrimSpace(req.Reason); reason != "" {
			order.Metadata["cancel_reason"] = reason
		}
		order.Metadata["cancelled_by"] = "user"
		order.Metadata["cancelled_at"] = now

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
	balance, err := l.svcCtx.Repositories.Balance.GetBalance(l.ctx, user.ID)
	if err != nil {
		return nil, err
	}

	resp = &types.UserOrderResponse{
		Order:   detail,
		Balance: orderutil.ToBalanceSnapshot(balance),
	}
	return resp, nil
}
