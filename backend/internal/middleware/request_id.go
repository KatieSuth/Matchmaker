package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// RequestID propagates X-Request-ID (or generates one) on the Gin context and the Go context for slog.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			b := make([]byte, 8)
			rand.Read(b)
			id = hex.EncodeToString(b)
		}

		c.Set("request_id", id)
		c.Header("X-Request-ID", id)

		// Also store on the Go context so slog can pick it up
		ctx := context.WithValue(c.Request.Context(), requestIDKey, id)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// ContextWithRequestID stores a request ID in a context. Useful for
// seeding request IDs outside of the Gin middleware chain, such as in tests.
func ContextWithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}
