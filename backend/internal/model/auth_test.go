package model_test

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseAccessToken is a helper that parses and validates a JWT, returning its claims.
func parseAccessToken(t *testing.T, token string, secret []byte) *model.Claims {
	t.Helper()
	claims := &model.Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(_ *jwt.Token) (any, error) {
		return secret, nil
	})
	require.NoError(t, err)
	require.True(t, parsed.Valid)
	return claims
}

// ============================================================
// GenerateTokens
// ============================================================

func TestGenerateTokens_EmptyUserID(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	_, _, err = model.GenerateTokens("", jwtSecret)
	assert.Error(t, err)
}

func TestGenerateTokens_ReturnsNoError(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	_, _, err = model.GenerateTokens(uuid.New().String(), jwtSecret)
	assert.NoError(t, err)
}

func TestGenerateTokens_AccessTokenIsValidJWT(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	accessToken, _, err := model.GenerateTokens(uuid.New().String(), jwtSecret)
	require.NoError(t, err)

	parseAccessToken(t, accessToken, jwtSecret) // fails the test internally if invalid
}

func TestGenerateTokens_AccessTokenContainsCorrectUserID(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	userID := uuid.New().String()
	accessToken, _, err := model.GenerateTokens(userID, jwtSecret)
	require.NoError(t, err)

	claims := parseAccessToken(t, accessToken, jwtSecret)
	assert.Equal(t, userID, claims.UserID)
}

func TestGenerateTokens_AccessTokenExpiresInFifteenMinutes(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	before := time.Now()
	accessToken, _, err := model.GenerateTokens(uuid.New().String(), jwtSecret)
	require.NoError(t, err)

	claims := parseAccessToken(t, accessToken, jwtSecret)

	expectedExpiry := before.Add(15 * time.Minute)
	assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, 5*time.Second)
}

func TestGenerateTokens_AccessTokenUsesHS256(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	accessToken, _, err := model.GenerateTokens(uuid.New().String(), jwtSecret)
	require.NoError(t, err)

	token, _, err := new(jwt.Parser).ParseUnverified(accessToken, &model.Claims{})
	require.NoError(t, err)
	assert.Equal(t, jwt.SigningMethodHS256, token.Method)
}

func TestGenerateTokens_AccessTokenInvalidWithWrongSecret(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	accessToken, _, err := model.GenerateTokens(uuid.New().String(), jwtSecret)
	require.NoError(t, err)

	_, err = jwt.ParseWithClaims(accessToken, &model.Claims{}, func(_ *jwt.Token) (any, error) {
		return []byte("wrong-secret"), nil
	})
	assert.Error(t, err)
}

func TestGenerateTokens_RefreshTokenIsValidHex(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	_, refreshToken, err := model.GenerateTokens(uuid.New().String(), jwtSecret)
	require.NoError(t, err)

	// 32 random bytes encoded as hex = 64 characters.
	assert.Len(t, refreshToken, 64)
	decoded, err := hex.DecodeString(refreshToken)
	require.NoError(t, err)
	assert.Len(t, decoded, 32)
}

func TestGenerateTokens_RefreshTokensAreUnique(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	userID := uuid.New().String()
	_, refresh1, err := model.GenerateTokens(userID, jwtSecret)
	require.NoError(t, err)
	_, refresh2, err := model.GenerateTokens(userID, jwtSecret)
	require.NoError(t, err)

	assert.NotEqual(t, refresh1, refresh2)
}
