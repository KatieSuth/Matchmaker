package model_test

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/testutil"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestGenerateTokens(t *testing.T) {
	jwtSecret, err := testutil.GetJWTSecret(t)
	if err != nil {
		t.Error(err)
	}

	userID := uuid.New().String()

	t.Run("returns no error with valid inputs", func(t *testing.T) {
		_, _, err := model.GenerateTokens(userID, jwtSecret)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("access token is a valid JWT", func(t *testing.T) {
		accessToken, _, err := model.GenerateTokens(userID, jwtSecret)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		claims := &model.Claims{}
		token, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil {
			t.Fatalf("failed to parse access token: %v", err)
		}
		if !token.Valid {
			t.Error("expected token to be valid")
		}
	})

	t.Run("access token contains correct user ID", func(t *testing.T) {
		accessToken, _, err := model.GenerateTokens(userID, jwtSecret)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		claims := &model.Claims{}
		_, err = jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil {
			t.Fatalf("failed to parse token: %v", err)
		}

		if claims.UserID != userID {
			t.Errorf("expected UserID %q, got %q", userID, claims.UserID)
		}
	})

	t.Run("access token expires in ~15 minutes", func(t *testing.T) {
		before := time.Now()
		accessToken, _, err := model.GenerateTokens(userID, jwtSecret)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		claims := &model.Claims{}
		_, err = jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil {
			t.Fatalf("failed to parse token: %v", err)
		}

		expectedExpiry := before.Add(15 * time.Minute)
		expiry := claims.ExpiresAt.Time

		tolerance := 5 * time.Second
		diff := expiry.Sub(expectedExpiry)
		if diff < -tolerance || diff > tolerance {
			t.Errorf("expected expiry ~%v, got %v (diff: %v)", expectedExpiry, expiry, diff)
		}
	})

	t.Run("access token uses HS256 signing method", func(t *testing.T) {
		accessToken, _, err := model.GenerateTokens(userID, jwtSecret)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		token, _, err := new(jwt.Parser).ParseUnverified(accessToken, &model.Claims{})
		if err != nil {
			t.Fatalf("failed to parse token: %v", err)
		}

		if token.Method != jwt.SigningMethodHS256 {
			t.Errorf("expected HS256 signing method, got %v", token.Method.Alg())
		}
	})

	t.Run("access token is invalid with wrong secret", func(t *testing.T) {
		accessToken, _, err := model.GenerateTokens(userID, jwtSecret)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = jwt.ParseWithClaims(accessToken, &model.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte("wrong-secret"), nil
		})
		if err == nil {
			t.Error("expected error when verifying with wrong secret, got nil")
		}
	})

	t.Run("refresh token is valid 64-char hex string", func(t *testing.T) {
		_, refreshToken, err := model.GenerateTokens(userID, jwtSecret)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(refreshToken) != 64 {
			t.Errorf("expected refresh token length 64, got %d", len(refreshToken))
		}

		decoded, err := hex.DecodeString(refreshToken)
		if err != nil {
			t.Errorf("refresh token is not valid hex: %v", err)
		}
		if len(decoded) != 32 {
			t.Errorf("expected 32 decoded bytes, got %d", len(decoded))
		}
	})

	t.Run("refresh tokens are unique across calls", func(t *testing.T) {
		_, refresh1, err := model.GenerateTokens(userID, jwtSecret)
		if err != nil {
			t.Fatalf("unexpected error on first call: %v", err)
		}
		_, refresh2, err := model.GenerateTokens(userID, jwtSecret)
		if err != nil {
			t.Fatalf("unexpected error on second call: %v", err)
		}

		if refresh1 == refresh2 {
			t.Error("expected unique refresh tokens, got identical values")
		}
	})

	t.Run("works with empty user ID", func(t *testing.T) {
		_, _, err := model.GenerateTokens("", jwtSecret)
		if err == nil {
			t.Errorf("expected error fo empty userID, got nil")
		}
	})
}
