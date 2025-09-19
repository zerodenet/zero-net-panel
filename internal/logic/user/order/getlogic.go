package order

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/logic/orderutil"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/security"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// GetLogic handles retrieving a single user order.
type GetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetLogic constructs GetLogic.
func NewGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLogic {
	return &GetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Get returns detailed information of one order with balance snapshot.
func (l *GetLogic) Get(req *types.UserGetOrderRequest) (*types.UserOrderResponse, error) {
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

	refundsMap, err := l.svcCtx.Repositories.Order.ListRefunds(l.ctx, []uint64{order.ID})
	if err != nil {
		return nil, err
	}
	detail := orderutil.ToOrderDetail(order, items, refundsMap[order.ID])
	balance, err := l.svcCtx.Repositories.Balance.GetBalance(l.ctx, user.ID)
	if err != nil {
		return nil, err
	}

	resp := types.UserOrderResponse{
		Order:   detail,
		Balance: orderutil.ToBalanceSnapshot(balance),
	}
	return &resp, nil
}
