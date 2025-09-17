package auth

import (
	"testing"
	"time"
)

func TestGeneratorTokenLifecycle(t *testing.T) {
	generator := NewGenerator("access-secret", "refresh-secret", time.Minute, 2*time.Minute)

	pair, err := generator.GenerateTokenPair("42", []string{"admin"}, "test")
	if err != nil {
		t.Fatalf("generate token pair: %v", err)
	}

	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("expected non-empty tokens: %+v", pair)
	}

	accessClaims, err := generator.ParseAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("parse access token: %v", err)
	}
	if accessClaims.UserID != "42" {
		t.Fatalf("unexpected subject: %s", accessClaims.UserID)
	}

	refreshClaims, err := generator.ParseRefreshToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("parse refresh token: %v", err)
	}
	if refreshClaims.Audience != "test" {
		t.Fatalf("unexpected audience: %s", refreshClaims.Audience)
	}

	if _, err := generator.ParseAccessToken("invalid.token"); err == nil {
		t.Fatalf("expected error for invalid token")
	}
}
