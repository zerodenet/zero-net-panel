package auth

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// RefreshLogic 处理刷新令牌请求。
type RefreshLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRefreshLogic 构造函数。
func NewRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshLogic {
	return &RefreshLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Refresh 使用刷新令牌换取新的访问凭证。
func (l *RefreshLogic) Refresh(req *types.AuthRefreshRequest) (*types.AuthRefreshResponse, error) {
	token := strings.TrimSpace(req.RefreshToken)
	if token == "" {
		return nil, repository.ErrInvalidArgument
	}

	claims, err := l.svcCtx.Auth.ParseRefreshToken(token)
	if err != nil {
		return nil, repository.ErrUnauthorized
	}

	userID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		return nil, repository.ErrUnauthorized
	}

	user, err := l.svcCtx.Repositories.User.Get(l.ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrUnauthorized
		}
		return nil, err
	}

	if !strings.EqualFold(user.Status, "active") {
		return nil, repository.ErrForbidden
	}

	audience := l.svcCtx.Config.Project.Name
	if audience == "" {
		audience = "znp"
	}

	pair, err := l.svcCtx.Auth.GenerateTokenPair(strconv.FormatUint(user.ID, 10), user.Roles, audience)
	if err != nil {
		return nil, err
	}

	resp := &types.AuthRefreshResponse{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        computeTTL(pair.AccessExpire),
		RefreshExpiresIn: computeTTL(pair.RefreshExpire),
		User:             toAuthenticatedUser(user),
	}

	return resp, nil
}
