package orders

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

// CancelLogic handles administrative order cancellations.
type CancelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCancelLogic constructs the admin cancellation logic.
func NewCancelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelLogic {
	return &CancelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Cancel updates order status to cancelled when eligible.
func (l *CancelLogic) Cancel(req *types.AdminCancelOrderRequest) (*types.AdminOrderResponse, error) {
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

	if strings.EqualFold(order.Status, repository.OrderStatusCancelled) {
		return l.buildResponse(order, items, paymentsMap[order.ID])
	}

	if strings.EqualFold(order.Status, repository.OrderStatusPaid) {
		if order.TotalCents > 0 && order.RefundedCents < order.TotalCents {
			return nil, repository.ErrInvalidArgument
		}
	} else if !strings.EqualFold(order.Status, repository.OrderStatusPending) {
		return nil, repository.ErrInvalidArgument
	}

	var updated repository.Order
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		repo, err := repository.NewOrderRepository(tx)
		if err != nil {
			return err
		}

		cancelledAt := time.Now().UTC()
		if req.CancelledAt != nil && *req.CancelledAt > 0 {
			cancelledAt = time.Unix(*req.CancelledAt, 0).UTC()
		}

		metadata := map[string]any{
			"cancelled_by": actor.Email,
		}
		reason := strings.TrimSpace(req.Reason)
		if reason != "" {
			metadata["cancel_reason"] = reason
		}

		params := repository.UpdateOrderStatusParams{
			Status:        repository.OrderStatusCancelled,
			CancelledAt:   &cancelledAt,
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

	return l.buildResponse(updated, items, paymentsMap[order.ID])
}

func (l *CancelLogic) buildResponse(order repository.Order, items []repository.OrderItem, payments []repository.OrderPayment) (*types.AdminOrderResponse, error) {
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
