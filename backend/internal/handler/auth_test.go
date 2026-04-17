package handler_test

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/handler"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGenerateState(t *testing.T) {
	//Test basic functionality and length
	state, err := handler.GenerateState()
	require.NoError(t, err)

	//Since we make([]byte, 16), the hex string should be 32 characters long (2 hex characters per byte)
	assert.Len(t, state, 32)

	//Test validity: ensure it is a valid hex string
	_, err = hex.DecodeString(state)
	assert.NoError(t, err, "generated state should be valid hex")

	//Test uniqueness (Statistical check)
	//Generating two states should (virtually) never result in the same string
	state2, err := handler.GenerateState()
	require.NoError(t, err)
	assert.NotEqual(t, state, state2, "generated states should be unique")
}

func TestLoginHandler_Success(t *testing.T) {
	//Create a dummy OAuth2 config (no real secrets needed for this test)
	dummyOauth := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}

	//Initialize handler with MockStore, dummy config, and secure cookie
	h := newTestHandler(t, &store.MockStore{}, dummyOauth, "")

	c, w := test_util.NewGinContext(http.MethodGet, "/auth/login")
	h.LoginHandler(c)

	//Assertions
	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)

	// Check that it's redirecting to Discord's AuthURL
	location := w.Header().Get("Location")
	assert.Contains(t, location, "discord.com/oauth2/authorize")
	assert.Contains(t, location, "state=") // Ensure state was generated and appended

	//Verify the oauth_state cookie was set
	cookies := w.Result().Cookies()
	var stateCookie *http.Cookie
	for _, ck := range cookies {
		if ck.Name == "oauth_state" {
			stateCookie = ck
		}
	}
	require.NotNil(t, stateCookie, "oauth_state cookie should be set")
}

func TestDiscordCallbackHandler_Success(t *testing.T) {
	// Set up Mock Discord Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"access_token": "fake-token", "token_type": "Bearer"}`))
			return
		}
		if r.URL.Path == "/api/users/@me" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id": "12345", "username": "testuser", "avatar": "hash"}`))
			return
		}
	}))
	defer ts.Close()

	// Configure Handler to use Mock Server
	sc, _ := test_util.GetSecureCookie(t)
	ms := &store.MockStore{
		GetUserByDiscordIDFn: func(ctx context.Context, id string, update bool) (model.User, error) {
			return model.User{ID: uuid.New(), NewUser: true}, nil
		},
		UpdateUserFromLoginFn: func(ctx context.Context, uid uuid.UUID, du model.DiscordUser) (model.User, error) {
			return model.User{ID: uid}, nil
		},
		CreateOneTimeCodeFn: func(ctx context.Context, otc string, uid uuid.UUID) error {
			return nil
		},
	}

	// Create a dummy OAuth2 config (no real secrets needed for this test) & point OAuth config to our test server
	dummyOauth := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/oauth2/authorize",
			TokenURL: ts.URL + "/token",
		},
	}

	h := newTestHandler(t, ms, dummyOauth, ts.URL+"/api")

	// Prepare Request with valid State Cookie
	state := "valid-state"
	encoded, _ := sc.Encode("oauth_state", state)

	c, w := test_util.NewGinContext(http.MethodGet, "/auth/discord_callback?state="+state+"&code=fake-code")
	test_util.SetCookie(c, "oauth_state", encoded)

	h.DiscordCallbackHandler(c)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/auth/callback?&otc=")
}

// ============================================================
// RefreshHandler
// ============================================================

func TestRefreshHandler_NoCookie(t *testing.T) {
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	c, w := test_util.NewGinContext(http.MethodPost, "/auth/refresh")
	h.RefreshHandler(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefreshHandler_TokenNotFound(t *testing.T) {
	ms := &store.MockStore{
		GetRefreshTokenFn: func(_ context.Context, _ string) (model.RefreshToken, error) {
			return model.RefreshToken{}, errors.New("not found")
		},
	}
	h := newTestHandler(t, ms, nil, "")

	c, w := test_util.NewGinContext(http.MethodPost, "/auth/refresh")
	test_util.SetCookie(c, "refresh_token", "some-token-value")
	h.RefreshHandler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefreshHandler_TokenExpired(t *testing.T) {
	userID := uuid.New()
	ms := &store.MockStore{
		GetRefreshTokenFn: func(_ context.Context, _ string) (model.RefreshToken, error) {
			return model.RefreshToken{
				UserID:    userID,
				ExpiresAt: time.Now().Add(-time.Hour), // already expired
			}, nil
		},
		DeleteRefreshTokenFn: func(_ context.Context, _ string) error {
			return nil
		},
	}
	h := newTestHandler(t, ms, nil, "")

	c, w := test_util.NewGinContext(http.MethodPost, "/auth/refresh")
	test_util.SetCookie(c, "refresh_token", "expired-token")
	h.RefreshHandler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefreshHandler_Success(t *testing.T) {
	userID := uuid.New()
	var deletedHash string

	ms := &store.MockStore{
		GetRefreshTokenFn: func(_ context.Context, _ string) (model.RefreshToken, error) {
			return model.RefreshToken{
				UserID:    userID,
				ExpiresAt: time.Now().Add(time.Hour), // still valid
			}, nil
		},
		CreateNewRefreshTokenFn: func(_ context.Context, _ string, _ uuid.UUID, _ time.Time) (model.RefreshToken, error) {
			return model.RefreshToken{}, nil
		},
		DeleteRefreshTokenFn: func(_ context.Context, hash string) error {
			deletedHash = hash
			return nil
		},
	}
	h := newTestHandler(t, ms, nil, "")

	c, w := test_util.NewGinContext(http.MethodPost, "/auth/refresh")
	test_util.SetCookie(c, "refresh_token", "valid-token")
	h.RefreshHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, deletedHash, "old refresh token should have been deleted")

	body := test_util.DecodeJSON[map[string]string](t, w)
	assert.NotEmpty(t, body["access_token"])
}

func TestRefreshHandler_DeleteOldTokenFails(t *testing.T) {
	userID := uuid.New()
	ms := &store.MockStore{
		GetRefreshTokenFn: func(_ context.Context, _ string) (model.RefreshToken, error) {
			return model.RefreshToken{UserID: userID, ExpiresAt: time.Now().Add(time.Hour)}, nil
		},
		CreateNewRefreshTokenFn: func(_ context.Context, _ string, _ uuid.UUID, _ time.Time) (model.RefreshToken, error) {
			return model.RefreshToken{}, nil
		},
		DeleteRefreshTokenFn: func(_ context.Context, _ string) error {
			return errors.New("db error")
		},
	}
	h := newTestHandler(t, ms, nil, "")

	c, w := test_util.NewGinContext(http.MethodPost, "/auth/refresh")
	test_util.SetCookie(c, "refresh_token", "valid-token")
	h.RefreshHandler(c)

	// Deletion failure of the old token aborts with 500 per the handler.
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ============================================================
// CompleteAuthHandler
// ============================================================

func TestCompleteAuthHandler_MissingOTC(t *testing.T) {
	h := newTestHandler(t, &store.MockStore{}, nil, "")

	c, w := test_util.NewGinContext(http.MethodPost, "/auth/complete")
	c.Request.Body = http.NoBody
	h.CompleteAuthHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCompleteAuthHandler_EmptyOTC(t *testing.T) {
	h := newTestHandler(t, &store.MockStore{}, nil, "")

	c, w := test_util.NewGinContext(http.MethodPost, "/auth/complete")
	c.Request.Body = io.NopCloser(strings.NewReader(`{"otc":""}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.CompleteAuthHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCompleteAuthHandler_InvalidOTC(t *testing.T) {
	ms := &store.MockStore{
		ConsumeOneTimeCodeFn: func(_ context.Context, _ string) (uuid.UUID, error) {
			return uuid.UUID{}, errors.New("not found")
		},
	}
	h := newTestHandler(t, ms, nil, "")

	c, w := test_util.NewGinContext(http.MethodPost, "/auth/complete")
	c.Request.Body = io.NopCloser(strings.NewReader(`{"otc":"bad-code"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.CompleteAuthHandler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCompleteAuthHandler_Success(t *testing.T) {
	userID := uuid.New()
	ms := &store.MockStore{
		ConsumeOneTimeCodeFn: func(_ context.Context, _ string) (uuid.UUID, error) {
			return userID, nil
		},
		CreateNewRefreshTokenFn: func(_ context.Context, _ string, _ uuid.UUID, _ time.Time) (model.RefreshToken, error) {
			return model.RefreshToken{}, nil
		},
	}
	h := newTestHandler(t, ms, nil, "")

	c, w := test_util.NewGinContext(http.MethodPost, "/auth/complete")
	c.Request.Body = io.NopCloser(strings.NewReader(`{"otc":"valid-code"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.CompleteAuthHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	body := test_util.DecodeJSON[map[string]string](t, w)
	assert.NotEmpty(t, body["access_token"])

	// refresh_token cookie must be set
	cookies := w.Result().Cookies()
	var refreshCookie *http.Cookie
	for _, ck := range cookies {
		if ck.Name == "refresh_token" {
			refreshCookie = ck
		}
	}
	require.NotNil(t, refreshCookie, "expected refresh_token cookie to be set")
	assert.NotEmpty(t, refreshCookie.Value)
}

// ============================================================
// LogoutHandler
// ============================================================

func TestLogoutHandler_NoCookie(t *testing.T) {
	// No cookie — handler clears cookies and returns 204 anyway.
	h := newTestHandler(t, &store.MockStore{}, nil, "")
	c, w := test_util.NewGinContext(http.MethodPost, "/auth/logout")
	h.LogoutHandler(c)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestLogoutHandler_DeleteFails(t *testing.T) {
	ms := &store.MockStore{
		DeleteRefreshTokenFn: func(_ context.Context, _ string) error {
			return errors.New("db error")
		},
	}
	h := newTestHandler(t, ms, nil, "")

	c, w := test_util.NewGinContext(http.MethodPost, "/auth/logout")
	test_util.SetCookie(c, "refresh_token", "some-token")
	h.LogoutHandler(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLogoutHandler_Success(t *testing.T) {
	var deletedHash string
	ms := &store.MockStore{
		DeleteRefreshTokenFn: func(_ context.Context, hash string) error {
			deletedHash = hash
			return nil
		},
	}
	h := newTestHandler(t, ms, nil, "")

	c, w := test_util.NewGinContext(http.MethodPost, "/auth/logout")
	test_util.SetCookie(c, "refresh_token", "some-token")
	h.LogoutHandler(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.NotEmpty(t, deletedHash, "token should have been deleted from the store")

	// Both auth cookies should be cleared (MaxAge = -1).
	cookies := w.Result().Cookies()
	for _, ck := range cookies {
		if ck.Name == "refresh_token" || ck.Name == "auth_session" {
			assert.Equal(t, -1, ck.MaxAge, "cookie %q should be cleared", ck.Name)
		}
	}
}
