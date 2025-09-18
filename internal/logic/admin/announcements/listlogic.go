package announcements

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// ListLogic 管理端公告列表。
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

// List 返回公告列表。
func (l *ListLogic) List(req *types.AdminListAnnouncementsRequest) (*types.AdminAnnouncementListResponse, error) {
	opts := repository.ListAnnouncementsOptions{
		Page:      req.Page,
		PerPage:   req.PerPage,
		Status:    req.Status,
		Category:  req.Category,
		Audience:  req.Audience,
		Query:     req.Query,
		Sort:      req.Sort,
		Direction: req.Direction,
	}

	items, total, err := l.svcCtx.Repositories.Announcement.List(l.ctx, opts)
	if err != nil {
		return nil, err
	}

	list := make([]types.AnnouncementSummary, 0, len(items))
	for _, ann := range items {
		list = append(list, toAnnouncementSummary(ann))
	}

	page, perPage := normalizePage(req.Page, req.PerPage)
	pagination := types.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		TotalCount: total,
		HasNext:    int64(page*perPage) < total,
		HasPrev:    page > 1,
	}

	return &types.AdminAnnouncementListResponse{
		Announcements: list,
		Pagination:    pagination,
	}, nil
}
