package announcement

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// ListLogic 用户公告列表。
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

// List 返回当前公告。
func (l *ListLogic) List(req *types.UserAnnouncementListRequest) (*types.UserAnnouncementListResponse, error) {
	audience := strings.TrimSpace(req.Audience)
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	items, err := l.svcCtx.Repositories.Announcement.ListActive(l.ctx, audience, limit)
	if err != nil {
		return nil, err
	}

	list := make([]types.UserAnnouncementSummary, 0, len(items))
	for _, ann := range items {
		list = append(list, toUserAnnouncementSummary(ann))
	}

	return &types.UserAnnouncementListResponse{Announcements: list}, nil
}
