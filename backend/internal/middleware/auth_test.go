package middleware_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/middleware"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateAccessToken creates a signed JWT for the given userID using the
// provided secret, with a standard future expiry.
func generateAccessToken(t *testing.T, userID string, secret []byte) string {
	t.Helper()
	accessToken, _, err := model.GenerateTokens(userID, secret)
	require.NoError(t, err, "failed to generate token")
	return accessToken
}

// generateExpiredToken creates a signed JWT that is already expired.
func generateExpiredToken(t *testing.T, userID string, secret []byte) string {
	t.Helper()
	claims := model.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-15 * time.Minute)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	require.NoError(t, err, "failed to generate expired token")
	return token
}

// ============================================================
// ValidateAuth
// ============================================================

func TestValidateAuth_EmptyHeader(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	_, status, err := middleware.ValidateAuth(jwtSecret, "")

	assert.Equal(t, http.StatusUnauthorized, status)
	assert.Error(t, err)
}

func TestValidateAuth_MissingBearerPrefix(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	userID := uuid.New().String()
	token := generateAccessToken(t, userID, jwtSecret)

	// Token provided without "Bearer " prefix.
	_, status, err := middleware.ValidateAuth(jwtSecret, token)

	assert.Equal(t, http.StatusUnauthorized, status)
	assert.Error(t, err)
}

func TestValidateAuth_WrongPrefix(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	userID := uuid.New().String()
	token := generateAccessToken(t, userID, jwtSecret)

	_, status, err := middleware.ValidateAuth(jwtSecret, fmt.Sprintf("Token %s", token))

	assert.Equal(t, http.StatusUnauthorized, status)
	assert.Error(t, err)
}

func TestValidateAuth_InvalidSecret(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	userID := uuid.New().String()
	token := generateAccessToken(t, userID, jwtSecret)

	_, status, err := middleware.ValidateAuth([]byte("wrong-secret"), fmt.Sprintf("Bearer %s", token))

	assert.Equal(t, http.StatusUnauthorized, status)
	assert.Error(t, err)
}

func TestValidateAuth_ExpiredToken(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	userID := uuid.New().String()
	token := generateExpiredToken(t, userID, jwtSecret)

	_, status, err := middleware.ValidateAuth(jwtSecret, fmt.Sprintf("Bearer %s", token))

	assert.Equal(t, http.StatusUnauthorized, status)
	assert.Error(t, err)
}

func TestValidateAuth_MalformedToken(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	_, status, err := middleware.ValidateAuth(jwtSecret, "Bearer this.is.notavalidjwt")

	assert.Equal(t, http.StatusUnauthorized, status)
	assert.Error(t, err)
}

func TestValidateAuth_Success(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	userID := uuid.New().String()
	token := generateAccessToken(t, userID, jwtSecret)

	gotUserID, status, err := middleware.ValidateAuth(jwtSecret, fmt.Sprintf("Bearer %s", token))

	assert.NoError(t, err)
	assert.Equal(t, 0, status)
	assert.Equal(t, userID, gotUserID)
}

// ============================================================
// Auth middleware handler
// ============================================================

func TestAuth_MissingHeader(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	c, w := newGinContext(http.MethodGet, "/")
	middleware.Auth(jwtSecret)(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	_, exists := c.Get("userID")
	assert.False(t, exists, "userID should not be set on failed auth")
}

func TestAuth_InvalidToken(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	c, w := newGinContext(http.MethodGet, "/")
	c.Request.Header.Set("Authorization", "Bearer invalid.token.here")
	middleware.Auth(jwtSecret)(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	_, exists := c.Get("userID")
	assert.False(t, exists, "userID should not be set on failed auth")
}

func TestAuth_Success(t *testing.T) {
	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	userID := uuid.New().String()
	token := generateAccessToken(t, userID, jwtSecret)

	c, w := newGinContext(http.MethodGet, "/")
	c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	middleware.Auth(jwtSecret)(c)

	// Auth middleware does not abort on success, so the recorder
	// will show 200 (default) rather than an explicit status.
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)

	gotUserID, exists := c.Get("userID")
	require.True(t, exists, "userID should be set in context after successful auth")
	assert.Equal(t, userID, gotUserID)
}
