package subscription

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

const defaultUserID uint64 = 1

// ListLogic 查询用户订阅列表。
type ListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListLogic 构造函数。
func NewListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLogic {
	return &ListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// List 返回订阅列表。
func (l *ListLogic) List(req *types.UserListSubscriptionsRequest) (*types.UserSubscriptionListResponse, error) {
	userID := resolveUserID(l.ctx)

	opts := repository.ListSubscriptionsOptions{
		Page:      req.Page,
		PerPage:   req.PerPage,
		Sort:      req.Sort,
		Direction: req.Direction,
		Query:     req.Query,
		Status:    req.Status,
	}

	subs, total, err := l.svcCtx.Repositories.Subscription.ListByUser(l.ctx, userID, opts)
	if err != nil {
		return nil, err
	}

	items := make([]types.UserSubscriptionSummary, 0, len(subs))
	for _, sub := range subs {
		items = append(items, toUserSummary(sub))
	}

	page, perPage := normalizePage(req.Page, req.PerPage)
	pagination := types.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		TotalCount: total,
		HasNext:    int64(page*perPage) < total,
		HasPrev:    page > 1,
	}

	return &types.UserSubscriptionListResponse{
		Subscriptions: items,
		Pagination:    pagination,
	}, nil
}

func normalizePage(page, perPage int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage
}

func resolveUserID(ctx context.Context) uint64 {
	// TODO: 从鉴权上下文提取用户信息，目前使用默认用户便于联调。
	return defaultUserID
}
