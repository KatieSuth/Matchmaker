package handler_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/apilink"
	"github.com/KatieSuth/MatchmakerAPI/internal/discord"
	"github.com/KatieSuth/MatchmakerAPI/internal/handler"
	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// testAPILinkKey is a fixed AES-256 key for handler tests (not a production secret).
var testAPILinkKey = bytes.Repeat([]byte{0x04}, 32)

// stubDiscordAPI is a no-op DiscordAPI: empty guilds (unrestricted events) and unused OAuth methods.
type stubDiscordAPI struct{}

func (stubDiscordAPI) Exchange(context.Context, string) (*oauth2.Token, error) {
	return nil, errDiscordStubUnused
}

func (stubDiscordAPI) FetchMe(context.Context, string) (model.DiscordUser, error) {
	return model.DiscordUser{}, errDiscordStubUnused
}

func (stubDiscordAPI) ListUserGuilds(context.Context, uuid.UUID) ([]model.DiscordGuild, error) {
	return []model.DiscordGuild{}, nil
}

func (stubDiscordAPI) SeedAccessToken(uuid.UUID, string, time.Time) {}

var errDiscordStubUnused = errString("discord stub: oauth not configured")

type errString string

func (e errString) Error() string { return string(e) }

// fakeDiscordAPI is a test double for guild membership and optional OAuth hooks.
type fakeDiscordAPI struct {
	guilds       []model.DiscordGuild
	guildsByUser map[uuid.UUID][]model.DiscordGuild
	listErr      error
	exchangeTok  *oauth2.Token
	exchangeErr  error
	me           model.DiscordUser
	meErr        error
	seeded       map[uuid.UUID]string
}

func (f *fakeDiscordAPI) Exchange(context.Context, string) (*oauth2.Token, error) {
	if f.exchangeErr != nil {
		return nil, f.exchangeErr
	}
	if f.exchangeTok != nil {
		return f.exchangeTok, nil
	}
	return nil, errDiscordStubUnused
}

func (f *fakeDiscordAPI) FetchMe(context.Context, string) (model.DiscordUser, error) {
	if f.meErr != nil {
		return model.DiscordUser{}, f.meErr
	}
	return f.me, nil
}

func (f *fakeDiscordAPI) ListUserGuilds(_ context.Context, userID uuid.UUID) ([]model.DiscordGuild, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.guildsByUser != nil {
		guilds := f.guildsByUser[userID]
		if guilds == nil {
			return []model.DiscordGuild{}, nil
		}
		return guilds, nil
	}
	if f.guilds == nil {
		return []model.DiscordGuild{}, nil
	}
	return f.guilds, nil
}

func (f *fakeDiscordAPI) SeedAccessToken(userID uuid.UUID, accessToken string, _ time.Time) {
	if f.seeded == nil {
		f.seeded = map[uuid.UUID]string{}
	}
	f.seeded[userID] = accessToken
}

// newTestHandler constructs a real *Handler with controlled dependencies.
// oauth2Cfg can be nil for tests that don't exercise OAuth code paths.
func newTestHandler(t *testing.T, s store.Store, oauth2Cfg *oauth2.Config, discordApiUrl string) *handler.Handler {
	t.Helper()

	return newTestHandlerWithCookieDomain(t, s, oauth2Cfg, discordApiUrl, "")
}

func newTestHandlerWithCookieDomain(t *testing.T, s store.Store, oauth2Cfg *oauth2.Config, discordApiUrl, cookieDomain string) *handler.Handler {
	t.Helper()
	return newTestHandlerWithDiscord(t, s, oauth2Cfg, discordApiUrl, cookieDomain, nil)
}

func newTestHandlerWithDiscord(t *testing.T, s store.Store, oauth2Cfg *oauth2.Config, discordApiUrl, cookieDomain string, discordAPI handler.DiscordAPI) *handler.Handler {
	t.Helper()

	sc, err := test_util.GetSecureCookie(t)
	require.NoError(t, err)

	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	mmSettings := matchmaking.Settings{
		FairnessOutlierGap:         6,
		FairnessTeamSeparation:     3,
		FairnessReferenceTierCount: 25,
	}
	keyring, err := apilink.NewKeyring(apilink.DefaultKeyID, testAPILinkKey, nil)
	require.NoError(t, err)

	if discordAPI == nil {
		if oauth2Cfg != nil {
			discordAPI = discord.New(s, apilink.New(keyring, s), oauth2Cfg, discordApiUrl, nil)
		} else {
			discordAPI = stubDiscordAPI{}
		}
	}

	return handler.New("test", s, sc, oauth2Cfg, cookieDomain, "http://localhost:3000", jwtSecret, int(7*24*time.Hour/time.Second), discordApiUrl, mmSettings, keyring, discordAPI)
}
