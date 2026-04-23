// Package handler_test is a black-box test package for the HTTP API surface.
package handler_test

import (
	"net/http"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/stretchr/testify/assert"
)

// TestHealth checks that the health endpoint returns a 200 OK
// and the expected status message.
func TestHealth(t *testing.T) {
	// Initialize handler with a mock store
	h := newTestHandler(t, &store.MockStore{}, nil, "")

	// Create a Gin test context for a GET request
	c, w := test_util.NewGinContext(http.MethodGet, "/health")

	// Call the handler
	h.Health(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	body := test_util.DecodeJSON[map[string]string](t, w)
	assert.Equal(t, "ok", body["status"])
	assert.Equal(t, "API is running", body["message"])
	assert.NotEmpty(t, body["timestamp"], "timestamp should be present in health check")
}
