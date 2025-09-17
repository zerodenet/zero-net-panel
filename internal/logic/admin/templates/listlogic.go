package templates

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// ListLogic 管理端模板列表。
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

// List 返回模板列表。
func (l *ListLogic) List(req *types.AdminListSubscriptionTemplatesRequest) (*types.AdminSubscriptionTemplateListResponse, error) {
	opts := repository.ListTemplatesOptions{
		Page:          req.Page,
		PerPage:       req.PerPage,
		Sort:          req.Sort,
		Direction:     req.Direction,
		Query:         req.Query,
		ClientType:    req.ClientType,
		Format:        req.Format,
		IncludeDrafts: req.IncludeDrafts,
	}

	templates, total, err := l.svcCtx.Repositories.SubscriptionTemplate.List(l.ctx, opts)
	if err != nil {
		return nil, err
	}

	result := make([]types.SubscriptionTemplateSummary, 0, len(templates))
	for _, tpl := range templates {
		result = append(result, toTemplateSummary(tpl))
	}

	page, perPage := normalizePage(req.Page, req.PerPage)
	pagination := types.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		TotalCount: total,
		HasNext:    int64(page*perPage) < total,
		HasPrev:    page > 1,
	}

	return &types.AdminSubscriptionTemplateListResponse{
		Templates:  result,
		Pagination: pagination,
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
