package subscription

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/security"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// UpdateTemplateLogic 用户更新订阅模板。
type UpdateTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateTemplateLogic 构造函数。
func NewUpdateTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTemplateLogic {
	return &UpdateTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateTemplate 调整用户订阅模板。
func (l *UpdateTemplateLogic) UpdateTemplate(req *types.UserUpdateSubscriptionTemplateRequest) (*types.UserUpdateSubscriptionTemplateResponse, error) {
	user, ok := security.UserFromContext(l.ctx)
	if !ok {
		return nil, repository.ErrForbidden
	}

	sub, err := l.svcCtx.Repositories.Subscription.UpdateTemplate(l.ctx, req.SubscriptionID, req.TemplateID, user.ID)
	if err != nil {
		return nil, err
	}

	return &types.UserUpdateSubscriptionTemplateResponse{
		SubscriptionID: sub.ID,
		TemplateID:     sub.TemplateID,
		UpdatedAt:      time.Now().Unix(),
	}, nil
}
