package auth

import (
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
