package security

import "context"

type contextKey string

const userContextKey contextKey = "znp.security.user"

// UserClaims 代表鉴权用户信息。
type UserClaims struct {
	ID          uint64
	Email       string
	DisplayName string
	Roles       []string
}

// WithUser 将用户信息写入上下文。
func WithUser(ctx context.Context, user UserClaims) context.Context {
	roles := append([]string(nil), user.Roles...)
	copyUser := user
	copyUser.Roles = roles
	return context.WithValue(ctx, userContextKey, copyUser)
}

// UserFromContext 从上下文中读取鉴权用户。
func UserFromContext(ctx context.Context) (UserClaims, bool) {
	if ctx == nil {
		return UserClaims{}, false
	}
	value := ctx.Value(userContextKey)
	if value == nil {
		return UserClaims{}, false
	}
	user, ok := value.(UserClaims)
	if !ok {
		return UserClaims{}, false
	}
	roles := append([]string(nil), user.Roles...)
	user.Roles = roles
	return user, true
}

// HasRole 判断用户是否具有指定角色。
func HasRole(user UserClaims, target string) bool {
	for _, role := range user.Roles {
		if role == target {
			return true
		}
	}
	return false
}
