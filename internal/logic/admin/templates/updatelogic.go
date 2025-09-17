package templates

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// UpdateLogic 更新模板。
type UpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateLogic 构造函数。
func NewUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogic {
	return &UpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Update 执行更新操作。
func (l *UpdateLogic) Update(req *types.AdminUpdateSubscriptionTemplateRequest) (*types.SubscriptionTemplateSummary, error) {
	input := repository.UpdateSubscriptionTemplateInput{
		Name:        req.Name,
		Description: req.Description,
		Format:      req.Format,
		Content:     req.Content,
		Variables:   toRepositoryVariables(req.Variables),
		IsDefault:   req.IsDefault,
	}

	tpl, err := l.svcCtx.Repositories.SubscriptionTemplate.Update(l.ctx, req.TemplateID, input)
	if err != nil {
		return nil, err
	}

	summary := toTemplateSummary(tpl)
	return &summary, nil
}
