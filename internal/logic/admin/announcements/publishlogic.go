package announcements

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// PublishLogic 发布公告。
type PublishLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewPublishLogic 构造函数。
func NewPublishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishLogic {
	return &PublishLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Publish 更新公告状态为已发布。
func (l *PublishLogic) Publish(req *types.AdminPublishAnnouncementRequest) (*types.AnnouncementSummary, error) {
	publishAt := time.Now().UTC()
	var visibleTo *time.Time
	if req.VisibleTo > 0 {
		ts := time.Unix(req.VisibleTo, 0).UTC()
		visibleTo = &ts
	}

	updated, err := l.svcCtx.Repositories.Announcement.Publish(l.ctx, req.AnnouncementID, publishAt, visibleTo, req.Operator)
	if err != nil {
		return nil, err
	}

	summary := toAnnouncementSummary(updated)
	return &summary, nil
}
