package handler_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/apilink"
	"github.com/KatieSuth/MatchmakerAPI/internal/handler"
	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// testAPILinkKey is a fixed AES-256 key for handler tests (not a production secret).
var testAPILinkKey = bytes.Repeat([]byte{0x04}, 32)

// newTestHandler constructs a real *Handler with controlled dependencies.
// oauth2Cfg can be nil for tests that don't exercise OAuth code paths.
func newTestHandler(t *testing.T, s store.Store, oauth2Cfg *oauth2.Config, discordApiUrl string) *handler.Handler {
	t.Helper()

	return newTestHandlerWithCookieDomain(t, s, oauth2Cfg, discordApiUrl, "")
}

func newTestHandlerWithCookieDomain(t *testing.T, s store.Store, oauth2Cfg *oauth2.Config, discordApiUrl, cookieDomain string) *handler.Handler {
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
	return handler.New("test", s, sc, oauth2Cfg, cookieDomain, "http://localhost:3000", jwtSecret, int(7*24*time.Hour/time.Second), discordApiUrl, mmSettings, keyring)
}
