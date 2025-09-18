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

// ListLogic handles listing user orders.
type ListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListLogic constructs ListLogic.
func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// List returns paginated user orders.
func (l *ListLogic) List(req *types.UserOrderListRequest) (*types.UserOrderListResponse, error) {
	user, ok := security.UserFromContext(l.ctx)
	if !ok {
		return nil, repository.ErrUnauthorized
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
	userID := user.ID
	opts.UserID = &userID

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

	entries := make([]types.OrderDetail, 0, len(orders))
	for _, order := range orders {
		items := itemsMap[order.ID]
		entries = append(entries, orderutil.ToOrderDetail(order, items))
	}

	pagination := types.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		TotalCount: total,
		HasNext:    int64(page*perPage) < total,
		HasPrev:    page > 1,
	}

	resp := types.UserOrderListResponse{
		Orders:     entries,
		Pagination: pagination,
	}
	return &resp, nil
}
