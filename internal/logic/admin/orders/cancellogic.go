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

// CancelLogic handles administrative order cancellation.
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

// Cancel applies cancellation with optional override for already paid orders.
func (l *CancelLogic) Cancel(req *types.AdminCancelOrderRequest) (resp *types.AdminOrderResponse, err error) {
	start := time.Now()
	defer func() {
		result := "success"
		if err != nil {
			result = "error"
		}
		metrics.ObserveOrderCancel("admin", result, time.Since(start))
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

		if order.Status == repository.OrderStatusCancelled {
			updatedOrder = order
			return nil
		}
		if order.Status == repository.OrderStatusPaid && !req.AllowPaid {
			return repository.ErrInvalidState
		}

		now := time.Now().UTC()
		order.Status = repository.OrderStatusCancelled
		order.CancelledAt = &now
		if order.Metadata == nil {
			order.Metadata = make(map[string]any)
		}
		order.Metadata["cancelled_by"] = fmt.Sprintf("admin:%d", user.ID)
		order.Metadata["cancelled_at"] = now
		if reason := strings.TrimSpace(req.Reason); reason != "" {
			order.Metadata["cancel_reason"] = reason
		}
		if note := strings.TrimSpace(req.Note); note != "" {
			order.Metadata["cancel_note"] = note
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
