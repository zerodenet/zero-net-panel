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

// ListLogic aggregates admin order listing.
type ListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListLogic constructs admin order list logic.
func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// List returns admin scoped order listings with user details.
func (l *ListLogic) List(req *types.AdminListOrdersRequest) (*types.AdminOrderListResponse, error) {
	user, ok := security.UserFromContext(l.ctx)
	if !ok {
		return nil, repository.ErrUnauthorized
	}
	if !security.HasRole(user, "admin") {
		return nil, repository.ErrForbidden
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	perPage := req.PerPage
	if perPage <= 0 || perPage > 100 {
		perPage = 20
	}

	opts := repository.ListOrdersOptions{
		Page:          page,
		PerPage:       perPage,
		Status:        req.Status,
		PaymentMethod: req.PaymentMethod,
		Number:        req.Number,
		Sort:          req.Sort,
		Direction:     req.Direction,
	}
	if req.UserID != 0 {
		userID := req.UserID
		opts.UserID = &userID
	}

	orders, total, err := l.svcCtx.Repositories.Order.List(l.ctx, opts)
	if err != nil {
		return nil, err
	}

	ids := make([]uint64, 0, len(orders))
	for _, order := range orders {
		ids = append(ids, order.ID)
	}

	itemsMap, err := l.svcCtx.Repositories.Order.ListItems(l.ctx, ids)
	if err != nil {
		return nil, err
	}

	userCache := make(map[uint64]types.OrderUserSummary)
	entries := make([]types.AdminOrderDetail, 0, len(orders))
	for _, order := range orders {
		summary, ok := userCache[order.UserID]
		if !ok {
			u, err := l.svcCtx.Repositories.User.Get(l.ctx, order.UserID)
			if err != nil {
				return nil, err
			}
			summary = types.OrderUserSummary{
				ID:          u.ID,
				Email:       u.Email,
				DisplayName: u.DisplayName,
			}
			userCache[order.UserID] = summary
		}
		detail := orderutil.ToOrderDetail(order, itemsMap[order.ID])
		entries = append(entries, types.AdminOrderDetail{
			OrderDetail: detail,
			User:        summary,
		})
	}

	pagination := types.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		TotalCount: total,
		HasNext:    int64(page*perPage) < total,
		HasPrev:    page > 1,
	}

	resp := types.AdminOrderListResponse{
		Orders:     entries,
		Pagination: pagination,
	}
	return &resp, nil
}
