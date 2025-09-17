package templates

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// CreateLogic 创建订阅模板。
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

// Create 执行创建。
func (l *CreateLogic) Create(req *types.AdminCreateSubscriptionTemplateRequest) (*types.SubscriptionTemplateSummary, error) {
	input := repository.CreateSubscriptionTemplateInput{
		Name:        req.Name,
		Description: req.Description,
		ClientType:  req.ClientType,
		Format:      req.Format,
		Content:     req.Content,
		Variables:   toRepositoryVariables(req.Variables),
		IsDefault:   req.IsDefault,
	}

	tpl, err := l.svcCtx.Repositories.SubscriptionTemplate.Create(l.ctx, input)
	if err != nil {
		return nil, err
	}

	summary := toTemplateSummary(tpl)
	return &summary, nil
}
