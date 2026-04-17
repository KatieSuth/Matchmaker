package logger_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/logger"
	"github.com/KatieSuth/MatchmakerAPI/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureHandler is a fake slog.Handler that records the most recent record
// passed to Handle, so tests can assert on what attributes were added.
type captureHandler struct {
	enabled bool
	last    *slog.Record
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return h.enabled
}

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.last = &r
	return nil
}

func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(_ string) slog.Handler      { return h }

func newCapture() (*captureHandler, slog.Handler) {
	base := &captureHandler{enabled: true}
	return base, logger.New(base)
}

// attrsOf collects all attributes from a record into a map for easy lookup.
func attrsOf(r *slog.Record) map[string]string {
	out := make(map[string]string)
	r.Attrs(func(a slog.Attr) bool {
		out[a.Key] = a.Value.String()
		return true
	})
	return out
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

	ctx := middleware.ContextWithRequestID(context.Background(), "test-request-id")
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	require.NoError(t, h.Handle(ctx, r))

	attrs := attrsOf(base.last)
	assert.Equal(t, "test-request-id", attrs["request_id"])
}

func TestHandle_OmitsRequestIDWhenAbsent(t *testing.T) {
	base, h := newCapture()

	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	require.NoError(t, h.Handle(context.Background(), r))

	attrs := attrsOf(base.last)
	_, hasRequestID := attrs["request_id"]
	assert.False(t, hasRequestID, "request_id should not be added when context has none")
}

func TestHandle_DelegatesToBaseHandler(t *testing.T) {
	base, h := newCapture()

	r := slog.NewRecord(time.Now(), slog.LevelInfo, "delegated message", 0)
	require.NoError(t, h.Handle(context.Background(), r))

	require.NotNil(t, base.last)
	assert.Equal(t, "delegated message", base.last.Message)
}

func TestHandle_PreservesExistingAttributes(t *testing.T) {
	base, h := newCapture()

	ctx := middleware.ContextWithRequestID(context.Background(), "trace-abc")
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
	r.AddAttrs(slog.String("existing_key", "existing_value"))
	require.NoError(t, h.Handle(ctx, r))

	attrs := attrsOf(base.last)
	assert.Equal(t, "existing_value", attrs["existing_key"], "pre-existing attribute should be preserved")
	assert.Equal(t, "trace-abc", attrs["request_id"], "request_id should be added alongside existing attrs")
}
