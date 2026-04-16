package logger_test

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/logger"
	"github.com/KatieSuth/MatchmakerAPI/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureHandler is a fake slog.Handler that records the most recent record
// passed to Handle, so tests can assert on what attributes were added.
type captureHandler struct {
	enabled bool
	last    *slog.Record
	attrs   []slog.Attr
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return h.enabled
}

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.last = &r
	r.Attrs(func(a slog.Attr) bool {
		h.attrs = append(h.attrs, a)
		return true
	})
	return nil
}

func (h *captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *captureHandler) WithGroup(name string) slog.Handler {
	return h
}

// newCapture returns a captureHandler wrapped by the logger under test.
func newCapture() (*captureHandler, slog.Handler) {
	base := &captureHandler{enabled: true}
	return base, logger.New(base)
}

// contextWithRequestID plants a request ID directly into a Go context,
// mirroring what the middleware does.
func contextWithRequestID(id string) context.Context {
	c, _ := newGinContext()
	c.Request.Header.Set("X-Request-ID", id)
	middleware.RequestID()(c)
	return c.Request.Context()
}

// newGinContext creates a minimal Gin context for seeding the middleware.
// Kept local to this package to avoid importing test_util into a non-handler test.
func newGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	return c, w
}

// ============================================================
// New
// ============================================================

func TestNew_ReturnsHandler(t *testing.T) {
	_, h := newCapture()
	assert.NotNil(t, h)
}

// ============================================================
// Handle — request_id injection
// ============================================================

func TestHandle_AddsRequestIDWhenPresent(t *testing.T) {
	base, h := newCapture()

	ctx := contextWithRequestID("test-request-id")
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	err := h.Handle(ctx, r)

	require.NoError(t, err)
	require.NotNil(t, base.last)

	var found string
	base.last.Attrs(func(a slog.Attr) bool {
		if a.Key == "request_id" {
			found = a.Value.String()
		}
		return true
	})
	assert.Equal(t, "test-request-id", found)
}

func TestHandle_OmitsRequestIDWhenAbsent(t *testing.T) {
	base, h := newCapture()

	// Plain context with no request ID planted.
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	err := h.Handle(context.Background(), r)

	require.NoError(t, err)
	require.NotNil(t, base.last)

	base.last.Attrs(func(a slog.Attr) bool {
		assert.NotEqual(t, "request_id", a.Key, "request_id should not be added when context has none")
		return true
	})
}

func TestHandle_DelegatesToBaseHandler(t *testing.T) {
	base, h := newCapture()

	r := slog.NewRecord(time.Now(), slog.LevelInfo, "delegated message", 0)
	err := h.Handle(context.Background(), r)

	require.NoError(t, err)
	// base.last being set confirms Handle was delegated to the base handler.
	require.NotNil(t, base.last)
	assert.Equal(t, "delegated message", base.last.Message)
}

func TestHandle_PreservesExistingAttributes(t *testing.T) {
	base, h := newCapture()

	ctx := contextWithRequestID("trace-abc")
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
	r.AddAttrs(slog.String("existing_key", "existing_value"))

	require.NoError(t, h.Handle(ctx, r))

	var foundExisting, foundRequestID bool
	base.last.Attrs(func(a slog.Attr) bool {
		if a.Key == "existing_key" {
			foundExisting = true
		}
		if a.Key == "request_id" {
			foundRequestID = true
		}
		return true
	})
	assert.True(t, foundExisting, "pre-existing attribute should be preserved")
	assert.True(t, foundRequestID, "request_id should be added alongside existing attrs")
}
