package announcements

import (
	"context"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// CreateLogic 管理端公告创建。
type CreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateLogic 构造函数。
func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Create 创建公告。
func (l *CreateLogic) Create(req *types.AdminCreateAnnouncementRequest) (*types.AnnouncementSummary, error) {
	now := time.Now().UTC()
	announcement := repository.Announcement{
		Title:       strings.TrimSpace(req.Title),
		Content:     strings.TrimSpace(req.Content),
		Category:    strings.TrimSpace(req.Category),
		Audience:    defaultAudience(req.Audience),
		IsPinned:    req.IsPinned,
		Priority:    req.Priority,
		Status:      "draft",
		VisibleFrom: now,
		CreatedBy:   strings.TrimSpace(req.CreatedBy),
		UpdatedBy:   strings.TrimSpace(req.CreatedBy),
	}

	created, err := l.svcCtx.Repositories.Announcement.Create(l.ctx, announcement)
	if err != nil {
		return nil, err
	}

	summary := toAnnouncementSummary(created)
	return &summary, nil
}

func defaultAudience(audience string) string {
	audience = strings.TrimSpace(strings.ToLower(audience))
	if audience == "" {
		return "all"
	}
	return audience
}
