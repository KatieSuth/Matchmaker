// Package middleware_test verifies request/recovery middleware emit structured slog fields.
package middleware_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/logger"
	"github.com/KatieSuth/MatchmakerAPI/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type capturedRecord struct {
	level slog.Level
	msg   string
	attrs map[string]any
}

type captureHandler struct {
	mu      sync.Mutex
	records []capturedRecord
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, capturedRecord{
		level: r.Level,
		msg:   r.Message,
		attrs: attrs,
	})
	return nil
}

func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(_ string) slog.Handler      { return h }

func (h *captureHandler) byMessage(msg string) []capturedRecord {
	h.mu.Lock()
	defer h.mu.Unlock()

	out := make([]capturedRecord, 0)
	for _, r := range h.records {
		if r.msg == msg {
			out = append(out, r)
		}
	}
	return out
}

func setupCapturedDefaultLogger(t *testing.T) *captureHandler {
	t.Helper()

	capture := &captureHandler{}
	prev := slog.Default()
	slog.SetDefault(slog.New(logger.New(capture)))
	t.Cleanup(func() {
		slog.SetDefault(prev)
	})

	return capture
}

func newRouterWithStructuredMiddleware() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.Recovery(), middleware.RequestLogger())
	return r
}

func TestRequestLogger_InfoIncludesCoreFieldsAndRequestID(t *testing.T) {
	capture := setupCapturedDefaultLogger(t)
	r := newRouterWithStructuredMiddleware()
	r.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req.Header.Set("X-Request-ID", "req-abc")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	records := capture.byMessage("request completed")
	require.Len(t, records, 1)
	rec := records[0]

	assert.Equal(t, slog.LevelInfo, rec.level)
	assert.Equal(t, http.MethodGet, rec.attrs["method"])
	assert.Equal(t, "/ok", rec.attrs["path"])
	assert.Equal(t, int64(http.StatusOK), rec.attrs["status"])
	assert.Equal(t, "/ok", rec.attrs["route"])
	assert.Equal(t, "req-abc", rec.attrs["request_id"])
	_, hasDuration := rec.attrs["duration_ms"]
	assert.True(t, hasDuration)
	_, hasClientIP := rec.attrs["client_ip"]
	assert.True(t, hasClientIP)
}

func TestRequestLogger_WarnsOn4xx(t *testing.T) {
	capture := setupCapturedDefaultLogger(t)
	r := newRouterWithStructuredMiddleware()
	r.GET("/bad", func(c *gin.Context) { c.Status(http.StatusBadRequest) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/bad", nil))

	require.Equal(t, http.StatusBadRequest, w.Code)

	records := capture.byMessage("request completed")
	require.Len(t, records, 1)
	assert.Equal(t, slog.LevelWarn, records[0].level)
	assert.Equal(t, int64(http.StatusBadRequest), records[0].attrs["status"])
}

func TestRequestLogger_ErrorsOn5xx(t *testing.T) {
	capture := setupCapturedDefaultLogger(t)
	r := newRouterWithStructuredMiddleware()
	r.GET("/error", func(c *gin.Context) { c.Status(http.StatusServiceUnavailable) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/error", nil))

	require.Equal(t, http.StatusServiceUnavailable, w.Code)

	records := capture.byMessage("request completed")
	require.Len(t, records, 1)
	assert.Equal(t, slog.LevelError, records[0].level)
	assert.Equal(t, int64(http.StatusServiceUnavailable), records[0].attrs["status"])
}

func TestRecovery_RecoversPanicAndLogsStructuredError(t *testing.T) {
	capture := setupCapturedDefaultLogger(t)
	r := newRouterWithStructuredMiddleware()
	r.GET("/panic", func(_ *gin.Context) { panic("boom") })

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set("X-Request-ID", "panic-req-id")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	panicRecords := capture.byMessage("panic recovered")
	require.Len(t, panicRecords, 1)
	assert.Equal(t, slog.LevelError, panicRecords[0].level)
	assert.Equal(t, "boom", panicRecords[0].attrs["panic"])
	assert.Equal(t, "panic-req-id", panicRecords[0].attrs["request_id"])
	_, hasStack := panicRecords[0].attrs["stack"]
	assert.True(t, hasStack)
}
