// Package middleware_test black-box tests JWT validation and the request ID hook.
package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newGinContext(method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)
	return c, w
}

func applyRequestID(c *gin.Context) {
	middleware.RequestID()(c)
}

// ============================================================
// RequestID middleware
// ============================================================

func TestRequestID_GeneratesIDWhenHeaderAbsent(t *testing.T) {
	c, w := newGinContext(http.MethodGet, "/")
	applyRequestID(c)

	id := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, id)
	// 8 random bytes encoded as hex = 16 characters.
	assert.Len(t, id, 16)
}

func TestRequestID_UsesProvidedHeader(t *testing.T) {
	c, w := newGinContext(http.MethodGet, "/")
	c.Request.Header.Set("X-Request-ID", "my-trace-id")
	applyRequestID(c)

	assert.Equal(t, "my-trace-id", w.Header().Get("X-Request-ID"))
}

func TestRequestID_SetsGinContextValue(t *testing.T) {
	c, _ := newGinContext(http.MethodGet, "/")
	c.Request.Header.Set("X-Request-ID", "gin-ctx-id")
	applyRequestID(c)

	val, exists := c.Get("request_id")
	require.True(t, exists)
	assert.Equal(t, "gin-ctx-id", val)
}

func TestRequestID_SetsGoContextValue(t *testing.T) {
	c, _ := newGinContext(http.MethodGet, "/")
	c.Request.Header.Set("X-Request-ID", "go-ctx-id")
	applyRequestID(c)

	id := middleware.RequestIDFromContext(c.Request.Context())
	assert.Equal(t, "go-ctx-id", id)
}

func TestRequestID_IDConsistentAcrossGinGoContextAndHeader(t *testing.T) {
	c, w := newGinContext(http.MethodGet, "/")
	applyRequestID(c)

	headerID := w.Header().Get("X-Request-ID")
	ginID, _ := c.Get("request_id")
	goCtxID := middleware.RequestIDFromContext(c.Request.Context())

	assert.Equal(t, headerID, ginID)
	assert.Equal(t, headerID, goCtxID)
}

func TestRequestID_EachRequestGetsUniqueID(t *testing.T) {
	c1, _ := newGinContext(http.MethodGet, "/")
	c2, _ := newGinContext(http.MethodGet, "/")
	applyRequestID(c1)
	applyRequestID(c2)

	id1 := middleware.RequestIDFromContext(c1.Request.Context())
	id2 := middleware.RequestIDFromContext(c2.Request.Context())
	assert.NotEqual(t, id1, id2)
}

// ============================================================
// ContextWithRequestID
// ============================================================

func TestContextWithRequestID_StoresAndRetrieves(t *testing.T) {
	ctx := middleware.ContextWithRequestID(context.Background(), "stored-id")
	assert.Equal(t, "stored-id", middleware.RequestIDFromContext(ctx))
}

func TestContextWithRequestID_DoesNotMutateParent(t *testing.T) {
	parent := context.Background()
	middleware.ContextWithRequestID(parent, "child-id")
	// Parent context should be unaffected.
	assert.Empty(t, middleware.RequestIDFromContext(parent))
}

// ============================================================
// RequestIDFromContext
// ============================================================

func TestRequestIDFromContext_ReturnsEmptyWhenNotSet(t *testing.T) {
	assert.Empty(t, middleware.RequestIDFromContext(context.Background()))
}
