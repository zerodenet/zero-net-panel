package security

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// GetLogic 查询第三方安全配置。
type GetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetLogic 构造函数。
func NewGetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLogic {
	return &GetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Get 返回当前安全配置。
func (l *GetLogic) Get() (*types.AdminSecuritySettingResponse, error) {
	setting, err := l.svcCtx.Repositories.Security.GetThirdPartyAPIConfig(l.ctx)
	if err != nil {
		return nil, err
	}

	resp := &types.AdminSecuritySettingResponse{
		Setting: toSecuritySetting(setting),
	}
	return resp, nil
}
