package auth

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/svc"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

// LoginLogic 处理登录请求。
type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewLoginLogic 构造函数。
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Login 执行登录。
func (l *LoginLogic) Login(req *types.AuthLoginRequest) (*types.AuthLoginResponse, error) {
	email := strings.TrimSpace(req.Email)
	if email == "" || strings.TrimSpace(req.Password) == "" {
		return nil, repository.ErrInvalidArgument
	}

	user, err := l.svcCtx.Repositories.User.GetByEmail(l.ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrUnauthorized
		}
		return nil, err
	}

	if !strings.EqualFold(user.Status, "active") {
		return nil, repository.ErrForbidden
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, repository.ErrUnauthorized
	}

	audience := l.svcCtx.Config.Project.Name
	if audience == "" {
		audience = "znp"
	}

	subject := strconv.FormatUint(user.ID, 10)
	pair, err := l.svcCtx.Auth.GenerateTokenPair(subject, user.Roles, audience)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	_ = l.svcCtx.Repositories.User.UpdateLastLogin(l.ctx, user.ID, now)

	resp := &types.AuthLoginResponse{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        computeTTL(pair.AccessExpire),
		RefreshExpiresIn: computeTTL(pair.RefreshExpire),
		User:             toAuthenticatedUser(user),
	}

	return resp, nil
}

func computeTTL(expire time.Time) int64 {
	ttl := int64(time.Until(expire).Seconds())
	if ttl < 0 {
		return 0
	}
	return ttl
}

func toAuthenticatedUser(user repository.User) types.AuthenticatedUser {
	return types.AuthenticatedUser{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Roles:       append([]string(nil), user.Roles...),
		CreatedAt:   user.CreatedAt.Unix(),
		UpdatedAt:   user.UpdatedAt.Unix(),
	}
}
