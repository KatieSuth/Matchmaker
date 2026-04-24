// Package middleware provides Gin middleware: request IDs for log correlation and JWT parsing.
package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs one structured entry per completed HTTP request.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		durationMs := time.Since(start).Milliseconds()
		status := c.Writer.Status()

		attrs := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"duration_ms", durationMs,
			"client_ip", c.ClientIP(),
		}
		if route := c.FullPath(); route != "" {
			attrs = append(attrs, "route", route)
		}

		switch {
		case status >= http.StatusInternalServerError:
			slog.ErrorContext(c.Request.Context(), "request completed", attrs...)
		case status >= http.StatusBadRequest:
			slog.WarnContext(c.Request.Context(), "request completed", attrs...)
		default:
			slog.InfoContext(c.Request.Context(), "request completed", attrs...)
		}
	}
}

// Recovery recovers panics and emits a structured error log before returning HTTP 500.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				attrs := []any{
					"panic", fmt.Sprint(rec),
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"client_ip", c.ClientIP(),
					"stack", string(debug.Stack()),
				}
				if route := c.FullPath(); route != "" {
					attrs = append(attrs, "route", route)
				}

				slog.ErrorContext(c.Request.Context(), "panic recovered", attrs...)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		c.Next()
	}
}
