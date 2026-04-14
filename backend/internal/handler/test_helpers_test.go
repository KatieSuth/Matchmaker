package handler_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/handler"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/gorilla/securecookie"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// newTestHandler constructs a real *Handler with controlled dependencies.
// oauth2Cfg can be nil for tests that don't exercise OAuth code paths.
func newTestHandler(t *testing.T, s store.Store, oauth2Cfg *oauth2.Config) *handler.Handler {
	t.Helper()

	hashKey := []byte("test-hash-key-32-bytes-padding!!")
	blockKey := []byte("test-block-key-16")
	sc := securecookie.New(hashKey, blockKey)

	jwtSecret, err := test_util.GetJWTSecret(t)
	require.NoError(t, err)

	return handler.New("test", s, sc, oauth2Cfg, "", "http://localhost:3000", jwtSecret, int(7*24*time.Hour/time.Second))
}
