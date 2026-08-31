// Package discord talks to Discord’s OAuth and REST APIs: token exchange, /users/@me,
// guild lists, and an in-process access-token cache minted from the encrypted refresh vault.
package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/apilink"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// accessTokenSkew is subtracted from Discord’s expires_in so we refresh before the token is dead.
const accessTokenSkew = 60 * time.Second

// guildListTTL is how long a successful /users/@me/guilds result is reused in-process.
// EventPage calls that endpoint twice in a row (/access then GET /events/:id); Discord 429s
// the second call within a few hundred milliseconds and we used to fail closed as a lock.
const guildListTTL = 60 * time.Second

const (
	discordGETMaxAttempts       = 2
	discordGETDefaultRetryAfter = time.Second
	discordGETMaxRetryAfter     = 2 * time.Second
)

// ErrMissingGrant means there is no Discord refresh token in the vault (user must re-login).
var ErrMissingGrant = errors.New("discord refresh token missing")

// ErrInvalidGrant means Discord rejected the refresh token; the vault row is deleted.
var ErrInvalidGrant = errors.New("discord refresh token invalid")

// ErrMalformedUser means Discord /users/@me returned a body we could not decode.
var ErrMalformedUser = errors.New("malformed discord user")

type cachedToken struct {
	access string
	expiry time.Time
}

type userSlot struct {
	mu           sync.Mutex
	token        cachedToken
	guilds       []model.DiscordGuild
	guildsExpiry time.Time
	guildsCached bool
}

// Client holds OAuth config, the vault, and a per-user in-memory access-token cache.
type Client struct {
	store   store.Store
	vault   *apilink.Vault
	oauth   *oauth2.Config
	apiBase string
	http    *http.Client
	now     func() time.Time
	slotsMu sync.Mutex
	slots   map[uuid.UUID]*userSlot
}

// New builds a Discord API client. httpClient may be nil (DefaultClient). apiBase is Discord’s API root.
func New(s store.Store, vault *apilink.Vault, oauth *oauth2.Config, apiBase string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if apiBase == "" {
		apiBase = "https://discord.com/api"
	}
	return &Client{
		store:   s,
		vault:   vault,
		oauth:   oauth,
		apiBase: strings.TrimRight(apiBase, "/"),
		http:    httpClient,
		now:     time.Now,
		slots:   make(map[uuid.UUID]*userSlot),
	}
}

// slot returns the per-user mutex + cached token entry, creating it if needed.
func (c *Client) slot(userID uuid.UUID) *userSlot {
	c.slotsMu.Lock()
	defer c.slotsMu.Unlock()
	s, ok := c.slots[userID]
	if !ok {
		s = &userSlot{}
		c.slots[userID] = s
	}
	return s
}

// SeedAccessToken stores a Discord access token in process memory until expiry (login callback).
func (c *Client) SeedAccessToken(userID uuid.UUID, accessToken string, expiry time.Time) {
	s := c.slot(userID)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = cachedToken{access: accessToken, expiry: expiry}
}

// evictLocked clears the cached access token and guild list. Caller must hold s.mu.
func evictLocked(s *userSlot) {
	s.token = cachedToken{}
	s.guilds = nil
	s.guildsExpiry = time.Time{}
	s.guildsCached = false
}

// isInvalidGrant reports whether err is Discord saying the refresh token is dead.
func isInvalidGrant(err error) bool {
	var retrieveErr *oauth2.RetrieveError
	if !errors.As(err, &retrieveErr) {
		return false
	}
	if retrieveErr.ErrorCode == "invalid_grant" {
		return true
	}
	if retrieveErr.Response == nil {
		return false
	}
	code := retrieveErr.Response.StatusCode
	if code != http.StatusBadRequest && code != http.StatusUnauthorized {
		return false
	}
	return strings.Contains(string(retrieveErr.Body), "invalid_grant")
}

// AccessToken returns a Discord bearer token for userID, using the memory cache or a vault refresh.
func (c *Client) AccessToken(ctx context.Context, userID uuid.UUID) (string, error) {
	s := c.slot(userID)
	s.mu.Lock()
	defer s.mu.Unlock()
	return c.accessTokenLocked(ctx, s, userID)
}

// accessTokenLocked is AccessToken with s.mu already held (so ListUserGuilds can share the slot lock).
func (c *Client) accessTokenLocked(ctx context.Context, s *userSlot, userID uuid.UUID) (string, error) {
	now := c.now()
	if s.token.access != "" && now.Before(s.token.expiry.Add(-accessTokenSkew)) {
		return s.token.access, nil
	}

	var access string
	var expiry time.Time
	grantDead := false
	err := c.store.WithTx(ctx, func(tx store.Store) error {
		txVault := c.vault.ForStore(tx)
		refresh, err := txVault.GetRefreshTokenForUpdate(ctx, userID, apilink.ProviderDiscord)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrMissingGrant, err)
		}

		tokenCtx := context.WithValue(ctx, oauth2.HTTPClient, c.http)
		tok, err := c.oauth.TokenSource(tokenCtx, &oauth2.Token{RefreshToken: refresh}).Token()
		if err != nil {
			if isInvalidGrant(err) {
				grantDead = true
				// Commit the delete (return nil from WithTx). Callers still see ErrInvalidGrant.
				return txVault.DeleteRefreshToken(ctx, userID, apilink.ProviderDiscord)
			}
			return err
		}
		if tok.RefreshToken != "" && tok.RefreshToken != refresh {
			if err := txVault.PutRefreshToken(ctx, userID, apilink.ProviderDiscord, tok.RefreshToken); err != nil {
				return err
			}
		}
		access = tok.AccessToken
		expiry = tok.Expiry
		return nil
	})
	if grantDead {
		evictLocked(s)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrInvalidGrant, err)
		}
		return "", ErrInvalidGrant
	}
	if err != nil {
		evictLocked(s)
		return "", err
	}
	s.token = cachedToken{access: access, expiry: expiry}
	return access, nil
}

// Exchange trades an OAuth authorization code for Discord tokens.
func (c *Client) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	tokenCtx := context.WithValue(ctx, oauth2.HTTPClient, c.http)
	return c.oauth.Exchange(tokenCtx, code)
}

// FetchMe loads Discord /users/@me using a bearer access token.
func (c *Client) FetchMe(ctx context.Context, accessToken string) (model.DiscordUser, error) {
	var user model.DiscordUser
	body, err := c.discordGET(ctx, "/users/@me", accessToken)
	if err != nil {
		return user, err
	}
	if err := json.Unmarshal(body, &user); err != nil {
		return user, fmt.Errorf("%w: %v", ErrMalformedUser, err)
	}
	return user, nil
}

type discordGuildJSON struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListUserGuilds returns the Discord servers the Matchmaker user currently belongs to.
// Successful results are cached on the per-user slot so /access and GET /events share one Discord call.
func (c *Client) ListUserGuilds(ctx context.Context, userID uuid.UUID) ([]model.DiscordGuild, error) {
	s := c.slot(userID)
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.guildsCached && c.now().Before(s.guildsExpiry) {
		return cloneGuilds(s.guilds), nil
	}

	access, err := c.accessTokenLocked(ctx, s, userID)
	if err != nil {
		return nil, err
	}
	body, err := c.discordGET(ctx, "/users/@me/guilds", access)
	if err != nil {
		return nil, err
	}
	var raw []discordGuildJSON
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode discord guilds: %w", err)
	}
	out := make([]model.DiscordGuild, 0, len(raw))
	for _, g := range raw {
		out = append(out, model.DiscordGuild{ID: g.ID, Name: g.Name})
	}
	s.guilds = cloneGuilds(out)
	s.guildsExpiry = c.now().Add(guildListTTL)
	s.guildsCached = true
	return out, nil
}

// cloneGuilds copies guilds so cached slices are not mutated by callers.
func cloneGuilds(in []model.DiscordGuild) []model.DiscordGuild {
	out := make([]model.DiscordGuild, len(in))
	copy(out, in)
	return out
}

// retryAfterDuration reads Discord’s Retry-After header (seconds) and caps how long we wait.
func retryAfterDuration(resp *http.Response) time.Duration {
	raw := strings.TrimSpace(resp.Header.Get("Retry-After"))
	if raw == "" {
		return discordGETDefaultRetryAfter
	}
	secs, err := strconv.ParseFloat(raw, 64)
	if err != nil || secs < 0 {
		return discordGETDefaultRetryAfter
	}
	d := time.Duration(secs * float64(time.Second))
	if d > discordGETMaxRetryAfter {
		return discordGETMaxRetryAfter
	}
	return d
}

// waitRetryAfter blocks until delay elapses or ctx is done.
func waitRetryAfter(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// discordGET performs an authenticated GET against Discord’s API and fails closed on non-2xx.
// A single 429 is retried after Retry-After (capped) so a burst of EventPage checks can recover.
func (c *Client) discordGET(ctx context.Context, path, accessToken string) ([]byte, error) {
	var lastStatus int
	for attempt := 1; attempt <= discordGETMaxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBase+path, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return body, nil
		}
		lastStatus = resp.StatusCode
		if resp.StatusCode != http.StatusTooManyRequests || attempt == discordGETMaxAttempts {
			return nil, fmt.Errorf("discord GET %s: status %d", path, resp.StatusCode)
		}
		delay := retryAfterDuration(resp)
		slog.WarnContext(ctx, "discord rate limited, retrying", "path", path, "retry_after_ms", delay.Milliseconds())
		if err := waitRetryAfter(ctx, delay); err != nil {
			return nil, err
		}
	}
	return nil, fmt.Errorf("discord GET %s: status %d", path, lastStatus)
}
