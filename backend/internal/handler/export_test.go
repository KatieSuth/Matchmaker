package handler

import (
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SetGenerateAccessTokenForRefreshTest overrides access-token generation during refresh tests.
func SetGenerateAccessTokenForRefreshTest(fn func(string, []byte) (string, error)) {
	generateAccessTokenForRefresh = fn
}

// ResetGenerateAccessTokenForRefreshTest restores the default refresh access-token generator.
func ResetGenerateAccessTokenForRefreshTest() {
	generateAccessTokenForRefresh = model.GenerateAccessToken
}

// SetLoginTestHooks overrides LoginHandler internals for deterministic tests.
func SetLoginTestHooks(
	h *Handler,
	generateState func() (string, error),
	encodeStateCookie func(name string, value interface{}) (string, error),
) {
	if generateState != nil {
		h.generateState = generateState
	}
	if encodeStateCookie != nil {
		h.encodeStateCookie = encodeStateCookie
	}
}

// SessionCookieDomainsForTest exposes session cookie domain variants for tests.
func SessionCookieDomainsForTest(h *Handler) []string {
	return h.sessionCookieDomains()
}

// WriteSessionCookieForTest exposes session cookie writing for tests.
func WriteSessionCookieForTest(
	h *Handler,
	c *gin.Context,
	name, value string,
	maxAge int,
	httpOnly bool,
	domain string,
) {
	h.writeSessionCookie(c, name, value, maxAge, httpOnly, domain)
}

// SetDiscordAPIForTest swaps the Discord client (including nil) for handler tests.
func SetDiscordAPIForTest(h *Handler, d DiscordAPI) {
	h.discord = d
}

// DiscordRestrictionErrorForTest exercises the Discord lock error String/Unwrap paths.
func DiscordRestrictionErrorForTest(cause error) error {
	return &discordGuildRestriction{eventTitle: "t", cause: cause}
}

// WriteDiscordGuildRestrictionForTest writes a 403 with nil guilds to cover the empty-slice branch.
func WriteDiscordGuildRestrictionForTest(h *Handler, c *gin.Context, userID, groupID uuid.UUID) {
	h.writeDiscordGuildRestriction(c, userID, groupID, &discordGuildRestriction{eventTitle: "t", guilds: nil})
}
