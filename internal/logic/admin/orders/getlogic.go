package orders

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/logic/orderutil"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/security"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// GetLogic retrieves detailed order information for administrators.
type GetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetLogic constructs the admin order get logic.
func NewGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLogic {
	return &GetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Get fetches a single order with user info.
func (l *GetLogic) Get(req *types.AdminGetOrderRequest) (*types.AdminOrderResponse, error) {
	user, ok := security.UserFromContext(l.ctx)
	if !ok {
		return nil, repository.ErrUnauthorized
	}
	if !security.HasRole(user, "admin") {
		return nil, repository.ErrForbidden
	}

	order, items, err := l.svcCtx.Repositories.Order.Get(l.ctx, req.OrderID)
	if err != nil {
		return nil, err
	}

	u, err := l.svcCtx.Repositories.User.Get(l.ctx, order.UserID)
	if err != nil {
		return nil, err
	}

	refundsMap, err := l.svcCtx.Repositories.Order.ListRefunds(l.ctx, []uint64{order.ID})
	if err != nil {
		return nil, err
	}

	paymentsMap, err := l.svcCtx.Repositories.Order.ListPayments(l.ctx, []uint64{order.ID})
	if err != nil {
		return nil, err
	}

	detail := orderutil.ToOrderDetail(order, items, refundsMap[order.ID], paymentsMap[order.ID])
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
