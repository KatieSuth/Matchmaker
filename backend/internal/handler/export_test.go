package handler

import (
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/gin-gonic/gin"
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
