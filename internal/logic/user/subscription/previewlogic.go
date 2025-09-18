package subscription

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/security"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
	subtemplate "github.com/zero-net-panel/zero-net-panel/pkg/subscription/template"
)

// PreviewLogic 渲染订阅预览。
type PreviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewPreviewLogic 构造函数。
func NewPreviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreviewLogic {
	return &PreviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Preview 生成订阅预览。
func (l *PreviewLogic) Preview(req *types.UserSubscriptionPreviewRequest) (*types.UserSubscriptionPreviewResponse, error) {
	user, ok := security.UserFromContext(l.ctx)
	if !ok {
		return nil, repository.ErrForbidden
	}

	sub, err := l.svcCtx.Repositories.Subscription.Get(l.ctx, req.SubscriptionID)
	if err != nil {
		return nil, err
	}

	if sub.UserID != user.ID {
		return nil, repository.ErrForbidden
	}

	templateID := req.TemplateID
	if templateID == 0 {
		templateID = sub.TemplateID
	} else {
		if !isTemplateAllowed(sub, templateID) {
			return nil, repository.ErrForbidden
		}
	}

	tpl, err := l.svcCtx.Repositories.SubscriptionTemplate.Get(l.ctx, templateID)
	if err != nil {
		return nil, err
	}

	nodes, _, err := l.svcCtx.Repositories.Node.List(l.ctx, repository.ListNodesOptions{PerPage: 5, Sort: "updated_at"})
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	data := map[string]any{
		"subscription": map[string]any{
			"id":                      sub.ID,
			"name":                    sub.Name,
			"plan":                    sub.PlanName,
			"status":                  sub.Status,
			"token":                   sub.Token,
			"expires_at":              sub.ExpiresAt.Format(time.RFC3339),
			"traffic_total_bytes":     sub.TrafficTotalBytes,
			"traffic_used_bytes":      sub.TrafficUsedBytes,
			"traffic_remaining_bytes": maxInt64(sub.TrafficTotalBytes-sub.TrafficUsedBytes, 0),
			"devices_limit":           sub.DevicesLimit,
			"available_template_ids":  sub.AvailableTemplateIDs,
		},
		"nodes": normalizeNodeContext(nodes),
		"template": map[string]any{
			"id":      tpl.ID,
			"name":    tpl.Name,
			"format":  tpl.Format,
			"version": tpl.Version,
		},
		"generated_at": now.Format(time.RFC3339),
	}

	content, err := subtemplate.Render(tpl.Format, tpl.Content, data)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256([]byte(content))
	etag := hex.EncodeToString(hash[:])

	contentType := "text/plain; charset=utf-8"
	switch tpl.Format {
	case "json":
		contentType = "application/json"
	}

	return &types.UserSubscriptionPreviewResponse{
		SubscriptionID: sub.ID,
		TemplateID:     templateID,
		Content:        content,
		ContentType:    contentType,
		ETag:           etag,
		GeneratedAt:    now.Unix(),
	}, nil
}

func normalizeNodeContext(nodes []repository.Node) []map[string]any {
	result := make([]map[string]any, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, map[string]any{
			"id":         node.ID,
			"name":       node.Name,
			"region":     node.Region,
			"country":    node.Country,
			"protocols":  node.Protocols,
			"status":     node.Status,
			"updated_at": node.UpdatedAt.Format(time.RFC3339),
		})
	}
	return result
}

func isTemplateAllowed(sub repository.Subscription, templateID uint64) bool {
	if templateID == sub.TemplateID {
		return true
	}
	for _, id := range sub.AvailableTemplateIDs {
		if id == templateID {
			return true
		}
	}
	return false
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
