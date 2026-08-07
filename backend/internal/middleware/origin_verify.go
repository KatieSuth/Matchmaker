package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

const originVerifyHeader = "X-Origin-Verify"

// OriginVerify rejects requests missing or not matching the expected secret when
// secret is non-empty. When secret is empty the middleware is a no-op (local/dev).
// GET /health is always allowed so Cloud Run / Docker probes work without the header.
func OriginVerify(secret string) gin.HandlerFunc {
	if secret == "" {
		return func(c *gin.Context) { c.Next() }
	}
	expected := []byte(secret)
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet && c.Request.URL.Path == "/health" {
			c.Next()
			return
		}
		got := []byte(c.GetHeader(originVerifyHeader))
		if subtle.ConstantTimeCompare(got, expected) != 1 {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
