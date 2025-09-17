package nodes

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// ListLogic 管理端节点列表。
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

// List 返回节点摘要列表。
func (l *ListLogic) List(req *types.AdminListNodesRequest) (*types.AdminNodeListResponse, error) {
	opts := repository.ListNodesOptions{
		Page:      req.Page,
		PerPage:   req.PerPage,
		Sort:      req.Sort,
		Direction: req.Direction,
		Query:     req.Query,
		Status:    req.Status,
		Protocol:  req.Protocol,
	}

	nodes, total, err := l.svcCtx.Repositories.Node.List(l.ctx, opts)
	if err != nil {
		return nil, err
	}

	summaries := make([]types.NodeSummary, 0, len(nodes))
	for _, node := range nodes {
		summaries = append(summaries, mapNodeSummary(node))
	}

	page, perPage := normalizePage(req.Page, req.PerPage)
	pagination := types.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		TotalCount: total,
		HasNext:    int64(page*perPage) < total,
		HasPrev:    page > 1,
	}

	return &types.AdminNodeListResponse{
		Nodes:      summaries,
		Pagination: pagination,
	}, nil
}

func mapNodeSummary(node repository.Node) types.NodeSummary {
	summary := types.NodeSummary{
		ID:           node.ID,
		Name:         node.Name,
		Region:       node.Region,
		Country:      node.Country,
		ISP:          node.ISP,
		Status:       node.Status,
		Tags:         append([]string(nil), node.Tags...),
		Protocols:    append([]string(nil), node.Protocols...),
		CapacityMbps: node.CapacityMbps,
		Description:  node.Description,
		LastSyncedAt: node.LastSyncedAt.Unix(),
		UpdatedAt:    node.UpdatedAt.Unix(),
	}
	return summary
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
