package discord_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/apilink"
	"github.com/KatieSuth/MatchmakerAPI/internal/discord"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func testKeyring(t *testing.T) *apilink.Keyring {
	t.Helper()
	kr, err := apilink.NewKeyring(apilink.DefaultKeyID, bytes.Repeat([]byte{0x04}, 32), nil)
	require.NoError(t, err)
	return kr
}

func memoryVaultStore() (*store.MockStore, map[string]model.ApiLink) {
	links := map[string]model.ApiLink{}
	ms := &store.MockStore{
		UpsertApiLinkFn: func(_ context.Context, uid uuid.UUID, name, ciphertext, nonce, keyID string) (model.ApiLink, error) {
			link := model.ApiLink{
				UserID:         uid,
				Name:           name,
				RefreshToken:   ciphertext,
				RefreshTokenIv: nonce,
				KeyID:          keyID,
			}
			links[name] = link
			return link, nil
		},
		GetApiLinkByUserAndNameFn: func(_ context.Context, _ uuid.UUID, name string) (model.ApiLink, error) {
			link, ok := links[name]
			if !ok {
				return model.ApiLink{}, errors.New("not found")
			}
			return link, nil
		},
		DeleteApiLinkByUserAndNameFn: func(_ context.Context, _ uuid.UUID, name string) error {
			delete(links, name)
			return nil
		},
	}
	return ms, links
}

func newTestDiscord(t *testing.T, tokenStatus int, tokenBody string, apiStatus int, apiBody string) (*discord.Client, *store.MockStore, map[string]model.ApiLink, *atomic.Int32) {
	t.Helper()
	var tokenHits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		tokenHits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(tokenStatus)
		_, _ = w.Write([]byte(tokenBody))
	})
	mux.HandleFunc("/users/@me/guilds", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(apiStatus)
		_, _ = w.Write([]byte(apiBody))
	})
	mux.HandleFunc("/users/@me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(apiStatus)
		_, _ = w.Write([]byte(apiBody))
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	ms, links := memoryVaultStore()
	oauth := &oauth2.Config{
		ClientID:     "id",
		ClientSecret: "secret",
		Endpoint:     oauth2.Endpoint{TokenURL: ts.URL + "/oauth2/token"},
	}
	client := discord.New(ms, apilink.New(testKeyring(t), ms), oauth, ts.URL, ts.Client())
	return client, ms, links, &tokenHits
}

func putRefresh(t *testing.T, ms *store.MockStore, userID uuid.UUID, plaintext string) {
	t.Helper()
	v := apilink.New(testKeyring(t), ms)
	require.NoError(t, v.PutRefreshToken(context.Background(), userID, apilink.ProviderDiscord, plaintext))
}

func TestAccessToken_CacheHitSkipsRefresh(t *testing.T) {
	client, _, _, hits := newTestDiscord(t, http.StatusOK, `{"access_token":"refreshed","token_type":"Bearer","expires_in":3600}`, http.StatusOK, `[{"id":"g1","name":"One"}]`)
	userID := uuid.New()
	client.SeedAccessToken(userID, "cached", time.Now().Add(time.Hour))

	guilds, err := client.ListUserGuilds(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, guilds, 1)
	assert.Equal(t, "g1", guilds[0].ID)
	assert.Equal(t, int32(0), hits.Load())
}

func TestAccessToken_RefreshPersistsRotatedToken(t *testing.T) {
	client, ms, links, hits := newTestDiscord(t, http.StatusOK, `{"access_token":"new-access","token_type":"Bearer","expires_in":3600,"refresh_token":"rotated"}`, http.StatusOK, `[]`)
	userID := uuid.New()
	putRefresh(t, ms, userID, "original")

	tok, err := client.AccessToken(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, "new-access", tok)
	assert.Equal(t, int32(1), hits.Load())

	v := apilink.New(testKeyring(t), ms)
	got, err := v.GetRefreshToken(context.Background(), userID, apilink.ProviderDiscord)
	require.NoError(t, err)
	assert.Equal(t, "rotated", got)
	assert.NotEqual(t, "rotated", links[apilink.ProviderDiscord].RefreshToken)
}

func TestAccessToken_MissingGrant(t *testing.T) {
	client, _, _, _ := newTestDiscord(t, http.StatusOK, `{}`, http.StatusOK, `[]`)
	_, err := client.AccessToken(context.Background(), uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, discord.ErrMissingGrant)
}

func TestAccessToken_InvalidGrantDeletesVaultRow(t *testing.T) {
	client, ms, links, _ := newTestDiscord(t, http.StatusBadRequest, `{"error":"invalid_grant"}`, http.StatusOK, `[]`)
	userID := uuid.New()
	putRefresh(t, ms, userID, "dead-refresh")
	require.Contains(t, links, apilink.ProviderDiscord)

	_, err := client.AccessToken(context.Background(), userID)
	require.Error(t, err)
	assert.ErrorIs(t, err, discord.ErrInvalidGrant)
	_, ok := links[apilink.ProviderDiscord]
	assert.False(t, ok)
}

func TestAccessToken_TransientKeepsVaultRowAndEvictsCache(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{name: "429", status: http.StatusTooManyRequests, body: `{"message":"rate limited"}`},
		{name: "502", status: http.StatusBadGateway, body: `{"message":"upstream"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, ms, links, _ := newTestDiscord(t, tc.status, tc.body, http.StatusOK, `[]`)
			userID := uuid.New()
			putRefresh(t, ms, userID, "keep-me")
			client.SetNowForTest(func() time.Time { return now })
			// Near-expiry token would be reused if we did not evict after the failed refresh.
			client.SeedAccessToken(userID, "stale-near-expiry", now.Add(30*time.Second))

			_, err := client.AccessToken(context.Background(), userID)
			require.Error(t, err)
			assert.NotErrorIs(t, err, discord.ErrInvalidGrant)
			_, ok := links[apilink.ProviderDiscord]
			assert.True(t, ok)
			_, _, cached := client.CachedAccessTokenForTest(userID)
			assert.False(t, cached)

			got, err := apilink.New(testKeyring(t), ms).GetRefreshToken(context.Background(), userID, apilink.ProviderDiscord)
			require.NoError(t, err)
			assert.Equal(t, "keep-me", got)
		})
	}
}

func TestAccessToken_TransientNetworkEvictsCache(t *testing.T) {
	now := time.Now()
	ms, links := memoryVaultStore()
	userID := uuid.New()
	putRefresh(t, ms, userID, "keep-me")
	oauth := &oauth2.Config{
		ClientID:     "id",
		ClientSecret: "secret",
		Endpoint:     oauth2.Endpoint{TokenURL: "http://127.0.0.1:1/oauth2/token"},
	}
	client := discord.New(ms, apilink.New(testKeyring(t), ms), oauth, "http://127.0.0.1:1", &http.Client{Timeout: 50 * time.Millisecond})
	client.SetNowForTest(func() time.Time { return now })
	client.SeedAccessToken(userID, "stale-near-expiry", now.Add(30*time.Second))

	_, err := client.AccessToken(context.Background(), userID)
	require.Error(t, err)
	assert.NotErrorIs(t, err, discord.ErrInvalidGrant)
	_, ok := links[apilink.ProviderDiscord]
	assert.True(t, ok)
	_, _, cached := client.CachedAccessTokenForTest(userID)
	assert.False(t, cached)
}

func TestListUserGuilds_DecodeError(t *testing.T) {
	client, _, _, _ := newTestDiscord(t, http.StatusOK, `{"access_token":"a","token_type":"Bearer","expires_in":3600}`, http.StatusOK, `not-json`)
	userID := uuid.New()
	client.SeedAccessToken(userID, "cached", time.Now().Add(time.Hour))
	_, err := client.ListUserGuilds(context.Background(), userID)
	require.Error(t, err)
}

func TestFetchMe_SuccessAndMalformed(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/@me", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"123","username":"pat","avatar":"h"}`))
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	ms, _ := memoryVaultStore()
	client := discord.New(ms, apilink.New(testKeyring(t), ms), &oauth2.Config{}, ts.URL, ts.Client())

	user, err := client.FetchMe(context.Background(), "tok")
	require.NoError(t, err)
	assert.Equal(t, "123", user.ID)
	assert.Equal(t, "pat", user.Username)
}

func TestFetchMe_Malformed(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/@me", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	ms, _ := memoryVaultStore()
	client := discord.New(ms, apilink.New(testKeyring(t), ms), &oauth2.Config{}, ts.URL, ts.Client())
	_, err := client.FetchMe(context.Background(), "tok")
	require.Error(t, err)
	assert.ErrorIs(t, err, discord.ErrMalformedUser)
}

func TestExchange_UsesTokenURL(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "code=abc")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "ex",
			"token_type":    "Bearer",
			"refresh_token": "rt",
			"expires_in":    60,
		})
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	ms, _ := memoryVaultStore()
	oauth := &oauth2.Config{
		ClientID:     "id",
		ClientSecret: "secret",
		Endpoint:     oauth2.Endpoint{TokenURL: ts.URL + "/oauth2/token"},
	}
	client := discord.New(ms, apilink.New(testKeyring(t), ms), oauth, ts.URL, ts.Client())
	tok, err := client.Exchange(context.Background(), "abc")
	require.NoError(t, err)
	assert.Equal(t, "ex", tok.AccessToken)
}

func TestNew_Defaults(t *testing.T) {
	ms, _ := memoryVaultStore()
	client := discord.New(ms, apilink.New(testKeyring(t), ms), &oauth2.Config{}, "", nil)
	require.NotNil(t, client)
}

func TestIsInvalidGrant(t *testing.T) {
	assert.False(t, discord.IsInvalidGrantForTest(errors.New("nope")))
	assert.True(t, discord.IsInvalidGrantForTest(&oauth2.RetrieveError{ErrorCode: "invalid_grant"}))
	assert.False(t, discord.IsInvalidGrantForTest(&oauth2.RetrieveError{ErrorCode: "other"}))
	assert.True(t, discord.IsInvalidGrantForTest(&oauth2.RetrieveError{
		Response: &http.Response{StatusCode: http.StatusUnauthorized},
		Body:     []byte(`{"error":"invalid_grant"}`),
	}))
	assert.False(t, discord.IsInvalidGrantForTest(&oauth2.RetrieveError{
		Response: &http.Response{StatusCode: http.StatusInternalServerError},
		Body:     []byte(`{"error":"invalid_grant"}`),
	}))
	assert.False(t, discord.IsInvalidGrantForTest(&oauth2.RetrieveError{
		Response: &http.Response{StatusCode: http.StatusBadRequest},
		Body:     []byte(`{"error":"other"}`),
	}))
}

func TestMemberOfAny_AllBlankRequired(t *testing.T) {
	assert.False(t, discord.MemberOfAny([]model.DiscordGuild{{ID: "1"}}, []string{"", ""}))
}

func TestFetchMe_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(ts.Close)
	ms, _ := memoryVaultStore()
	client := discord.New(ms, apilink.New(testKeyring(t), ms), &oauth2.Config{}, ts.URL, ts.Client())
	_, err := client.FetchMe(context.Background(), "tok")
	require.Error(t, err)
}

func TestDiscordGET_CanceledContext(t *testing.T) {
	client, _, _, _ := newTestDiscord(t, http.StatusOK, `{"access_token":"a","token_type":"Bearer","expires_in":3600}`, http.StatusOK, `[]`)
	userID := uuid.New()
	client.SeedAccessToken(userID, "cached", time.Now().Add(time.Hour))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.ListUserGuilds(ctx, userID)
	require.Error(t, err)
}

func TestDiscordGET_InvalidURL(t *testing.T) {
	ms, _ := memoryVaultStore()
	client := discord.New(ms, apilink.New(testKeyring(t), ms), &oauth2.Config{}, "http://[", http.DefaultClient)
	userID := uuid.New()
	client.SeedAccessToken(userID, "cached", time.Now().Add(time.Hour))
	_, err := client.ListUserGuilds(context.Background(), userID)
	require.Error(t, err)
}

func TestAccessToken_DeleteGrantError(t *testing.T) {
	client, ms, _, _ := newTestDiscord(t, http.StatusBadRequest, `{"error":"invalid_grant"}`, http.StatusOK, `[]`)
	userID := uuid.New()
	putRefresh(t, ms, userID, "dead")
	ms.DeleteApiLinkByUserAndNameFn = func(context.Context, uuid.UUID, string) error {
		return errors.New("delete failed")
	}
	_, err := client.AccessToken(context.Background(), userID)
	require.Error(t, err)
	assert.ErrorIs(t, err, discord.ErrInvalidGrant)
}

func TestListUserGuilds_CachesSecondCall(t *testing.T) {
	var hits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/users/@me/guilds", func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"g1","name":"One"}]`))
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	ms, _ := memoryVaultStore()
	client := discord.New(ms, apilink.New(testKeyring(t), ms), &oauth2.Config{}, ts.URL, ts.Client())
	userID := uuid.New()
	client.SeedAccessToken(userID, "cached", time.Now().Add(time.Hour))

	first, err := client.ListUserGuilds(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, first, 1)
	second, err := client.ListUserGuilds(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, first, second)
	assert.Equal(t, int32(1), hits.Load())
}

func TestListUserGuilds_CacheExpires(t *testing.T) {
	var hits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/users/@me/guilds", func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"g1","name":"One"}]`))
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	ms, _ := memoryVaultStore()
	client := discord.New(ms, apilink.New(testKeyring(t), ms), &oauth2.Config{}, ts.URL, ts.Client())
	now := time.Now()
	client.SetNowForTest(func() time.Time { return now })
	userID := uuid.New()
	client.SeedAccessToken(userID, "cached", now.Add(time.Hour))

	_, err := client.ListUserGuilds(context.Background(), userID)
	require.NoError(t, err)
	now = now.Add(61 * time.Second)
	_, err = client.ListUserGuilds(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, int32(2), hits.Load())
}

func TestListUserGuilds_Retries429ThenSucceeds(t *testing.T) {
	var hits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/users/@me/guilds", func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"g1","name":"One"}]`))
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	ms, _ := memoryVaultStore()
	client := discord.New(ms, apilink.New(testKeyring(t), ms), &oauth2.Config{}, ts.URL, ts.Client())
	userID := uuid.New()
	client.SeedAccessToken(userID, "cached", time.Now().Add(time.Hour))

	guilds, err := client.ListUserGuilds(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, guilds, 1)
	assert.Equal(t, int32(2), hits.Load())
}

func TestListUserGuilds_DoesNotCacheError(t *testing.T) {
	var hits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/users/@me/guilds", func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"g1","name":"One"}]`))
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	ms, _ := memoryVaultStore()
	client := discord.New(ms, apilink.New(testKeyring(t), ms), &oauth2.Config{}, ts.URL, ts.Client())
	userID := uuid.New()
	client.SeedAccessToken(userID, "cached", time.Now().Add(time.Hour))

	_, err := client.ListUserGuilds(context.Background(), userID)
	require.Error(t, err)
	guilds, err := client.ListUserGuilds(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, guilds, 1)
	assert.Equal(t, int32(2), hits.Load())
}

func TestRetryAfterDuration(t *testing.T) {
	assert.Equal(t, time.Second, discord.RetryAfterDurationForTest(&http.Response{Header: http.Header{}}))
	assert.Equal(t, time.Duration(0), discord.RetryAfterDurationForTest(&http.Response{Header: http.Header{"Retry-After": []string{"0"}}}))
	assert.Equal(t, 500*time.Millisecond, discord.RetryAfterDurationForTest(&http.Response{Header: http.Header{"Retry-After": []string{"0.5"}}}))
	assert.Equal(t, 2*time.Second, discord.RetryAfterDurationForTest(&http.Response{Header: http.Header{"Retry-After": []string{"30"}}}))
	assert.Equal(t, time.Second, discord.RetryAfterDurationForTest(&http.Response{Header: http.Header{"Retry-After": []string{"nope"}}}))
}

func TestWaitRetryAfter(t *testing.T) {
	require.NoError(t, discord.WaitRetryAfterForTest(context.Background(), 0))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := discord.WaitRetryAfterForTest(ctx, time.Second)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestListUserGuilds_Persistent429Fails(t *testing.T) {
	var hits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/users/@me/guilds", func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	ms, _ := memoryVaultStore()
	client := discord.New(ms, apilink.New(testKeyring(t), ms), &oauth2.Config{}, ts.URL, ts.Client())
	userID := uuid.New()
	client.SeedAccessToken(userID, "cached", time.Now().Add(time.Hour))

	_, err := client.ListUserGuilds(context.Background(), userID)
	require.Error(t, err)
	assert.Equal(t, int32(2), hits.Load())
}

func TestListUserGuilds_CachesEmptyList(t *testing.T) {
	var hits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/users/@me/guilds", func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	ms, _ := memoryVaultStore()
	client := discord.New(ms, apilink.New(testKeyring(t), ms), &oauth2.Config{}, ts.URL, ts.Client())
	userID := uuid.New()
	client.SeedAccessToken(userID, "cached", time.Now().Add(time.Hour))

	first, err := client.ListUserGuilds(context.Background(), userID)
	require.NoError(t, err)
	assert.Empty(t, first)
	second, err := client.ListUserGuilds(context.Background(), userID)
	require.NoError(t, err)
	assert.Empty(t, second)
	assert.Equal(t, int32(1), hits.Load())
}
