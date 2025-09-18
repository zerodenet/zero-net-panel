package security

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// UpdateLogic 更新第三方安全配置。
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

// Update 根据请求更新安全配置。
func (l *UpdateLogic) Update(req *types.AdminUpdateSecuritySettingRequest) (*types.AdminSecuritySettingResponse, error) {
	setting, err := l.svcCtx.Repositories.Security.GetThirdPartyAPIConfig(l.ctx)
	if err != nil {
		return nil, err
	}

	if req.ThirdPartyAPIEnabled != nil {
		setting.ThirdPartyAPIEnabled = *req.ThirdPartyAPIEnabled
	}
	if req.APIKey != nil {
		setting.APIKey = strings.TrimSpace(*req.APIKey)
	}
	if req.APISecret != nil {
		setting.APISecret = strings.TrimSpace(*req.APISecret)
	}
	if req.EncryptionAlgorithm != nil {
		algo := strings.TrimSpace(*req.EncryptionAlgorithm)
		if algo != "" {
			setting.EncryptionAlgorithm = strings.ToLower(algo)
		}
	}
	if req.NonceTTLSeconds != nil {
		ttl := *req.NonceTTLSeconds
		if ttl < 0 {
			ttl = 0
		}
		setting.NonceTTLSeconds = ttl
	}

	updated, err := l.svcCtx.Repositories.Security.UpsertThirdPartyAPIConfig(l.ctx, setting)
	if err != nil {
		return nil, err
	}

	resp := &types.AdminSecuritySettingResponse{
		Setting: toSecuritySetting(updated),
	}
	return resp, nil
}
