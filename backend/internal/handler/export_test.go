package handler

import "github.com/gin-gonic/gin"

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
