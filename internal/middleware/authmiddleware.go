package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/security"
	"github.com/zero-net-panel/zero-net-panel/pkg/auth"
)

// AuthMiddleware 负责解析并验证访问令牌。
type AuthMiddleware struct {
	generator *auth.Generator
	users     repository.UserRepository
}

// NewAuthMiddleware 构造函数。
func NewAuthMiddleware(generator *auth.Generator, users repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{generator: generator, users: users}
}

// RequireRoles 返回中间件，确保当前用户具备指定角色。若未指定角色，则仅校验登录态。
func (m *AuthMiddleware) RequireRoles(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r.Header.Get("Authorization"))
			if token == "" {
				writeAuthError(w, r, http.StatusUnauthorized, "missing authorization header")
				return
			}

			claims, err := m.generator.ParseAccessToken(token)
			if err != nil {
				writeAuthError(w, r, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			userID, err := strconv.ParseUint(claims.UserID, 10, 64)
			if err != nil {
				writeAuthError(w, r, http.StatusUnauthorized, "invalid subject in token")
				return
			}

			user, err := m.users.Get(r.Context(), userID)
			if err != nil {
				writeAuthError(w, r, http.StatusUnauthorized, "user not found")
				return
			}

			if !strings.EqualFold(user.Status, "active") {
				writeAuthError(w, r, http.StatusForbidden, "user is disabled")
				return
			}

			if len(roles) > 0 {
				allowed := false
				for _, userRole := range user.Roles {
					for _, required := range roles {
						if strings.EqualFold(userRole, required) {
							allowed = true
							break
						}
					}
					if allowed {
						break
					}
				}
				if !allowed {
					writeAuthError(w, r, http.StatusForbidden, "insufficient permissions")
					return
				}
			}

			ctxWithUser := security.WithUser(r.Context(), security.UserClaims{
				ID:          user.ID,
				Email:       user.Email,
				DisplayName: user.DisplayName,
				Roles:       user.Roles,
			})

			next(w, r.WithContext(ctxWithUser))
		}
	}
}

func extractBearerToken(header string) string {
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

func writeAuthError(w http.ResponseWriter, r *http.Request, status int, message string) {
	httpx.WriteJsonCtx(r.Context(), w, status, map[string]any{
		"message": message,
	})
}
