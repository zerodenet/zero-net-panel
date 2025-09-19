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

	if strings.EqualFold(order.Status, repository.OrderStatusPaid) {
		return l.buildResponse(order, items)
	}
	if !strings.EqualFold(order.Status, repository.OrderStatusPending) {
		return nil, repository.ErrInvalidArgument
	}

	var updated repository.Order
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		repo, err := repository.NewOrderRepository(tx)
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

		params := repository.UpdateOrderStatusParams{
			Status:        repository.OrderStatusPaid,
			PaidAt:        &paidAt,
			MetadataPatch: metadata,
		}
		method := strings.TrimSpace(req.PaymentMethod)
		if method != "" {
			params.PaymentMethod = &method
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

	return l.buildResponse(updated, items)
}

func (l *MarkPaidLogic) buildResponse(order repository.Order, items []repository.OrderItem) (*types.AdminOrderResponse, error) {
	detail := orderutil.ToOrderDetail(order, items)
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
