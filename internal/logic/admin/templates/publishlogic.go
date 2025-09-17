package templates

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/security"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// PublishLogic 发布模板。
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

// Publish 执行发布。
func (l *PublishLogic) Publish(req *types.AdminPublishSubscriptionTemplateRequest) (*types.AdminPublishSubscriptionTemplateResponse, error) {
	operator := strings.TrimSpace(req.Operator)
	if operator == "" {
		if user, ok := security.UserFromContext(l.ctx); ok {
			operator = strings.TrimSpace(user.DisplayName)
			if operator == "" {
				operator = strings.TrimSpace(user.Email)
			}
		}
	}
	if operator == "" {
		operator = "system"
	}

	input := repository.PublishSubscriptionTemplateInput{
		Changelog: strings.TrimSpace(req.Changelog),
		Operator:  operator,
	}

	tpl, history, err := l.svcCtx.Repositories.SubscriptionTemplate.Publish(l.ctx, req.TemplateID, input)
	if err != nil {
		return nil, err
	}

	summary := toTemplateSummary(tpl)
	historyEntry := toHistoryEntry(history)

	return &types.AdminPublishSubscriptionTemplateResponse{
		Template: summary,
		History:  historyEntry,
	}, nil
}
