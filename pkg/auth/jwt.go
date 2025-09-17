package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   string   `json:"userId"`
	Roles    []string `json:"roles"`
	Audience string   `json:"audience"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken   string
	RefreshToken  string
	AccessExpire  time.Time
	RefreshExpire time.Time
}

type Generator struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// ErrInvalidToken 用于指示令牌解析失败或已失效。
var ErrInvalidToken = errors.New("auth: invalid token")

func NewGenerator(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *Generator {
	return &Generator{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (g *Generator) GenerateTokenPair(userID string, roles []string, audience string) (*TokenPair, error) {
	now := time.Now()

	accessClaims := &Claims{
		UserID:   userID,
		Roles:    roles,
		Audience: audience,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{audience},
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(g.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	refreshClaims := &Claims{
		UserID:   userID,
		Roles:    roles,
		Audience: audience,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{audience},
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(g.refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(g.accessSecret)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(g.refreshSecret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		AccessExpire:  now.Add(g.accessTTL),
		RefreshExpire: now.Add(g.refreshTTL),
	}, nil
}

// ParseAccessToken 校验并解析访问令牌。
func (g *Generator) ParseAccessToken(token string) (*Claims, error) {
	return g.parseToken(token, g.accessSecret)
}

// ParseRefreshToken 校验并解析刷新令牌。
func (g *Generator) ParseRefreshToken(token string) (*Claims, error) {
	return g.parseToken(token, g.refreshSecret)
}

func (g *Generator) parseToken(token string, secret []byte) (*Claims, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}

	claims := &Claims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !parsedToken.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
