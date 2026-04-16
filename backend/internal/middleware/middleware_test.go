package middleware_test

import (
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

	// Response header should be set.
	id := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, id)

	// Should be a 16-char hex string (8 random bytes).
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

func TestRequestID_GeneratedIDConsistentAcrossContexts(t *testing.T) {
	// The same ID should be in the Gin context, Go context, and response header.
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
// RequestIDFromContext
// ============================================================

func TestRequestIDFromContext_ReturnsEmptyWhenNotSet(t *testing.T) {
	id := middleware.RequestIDFromContext(t.Context())
	assert.Empty(t, id)
}
