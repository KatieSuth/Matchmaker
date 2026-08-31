package discord

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// SetNowForTest overrides the clock used for access-token cache expiry.
func (c *Client) SetNowForTest(fn func() time.Time) {
	c.now = fn
}

// CachedAccessTokenForTest returns the in-memory access token for userID, if any.
func (c *Client) CachedAccessTokenForTest(userID uuid.UUID) (access string, expiry time.Time, ok bool) {
	s := c.slot(userID)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.token.access == "" {
		return "", time.Time{}, false
	}
	return s.token.access, s.token.expiry, true
}

// IsInvalidGrantForTest exposes isInvalidGrant for unit tests.
func IsInvalidGrantForTest(err error) bool {
	return isInvalidGrant(err)
}

// RetryAfterDurationForTest exposes retryAfterDuration for unit tests.
func RetryAfterDurationForTest(resp *http.Response) time.Duration {
	return retryAfterDuration(resp)
}

// WaitRetryAfterForTest exposes waitRetryAfter for unit tests.
func WaitRetryAfterForTest(ctx context.Context, delay time.Duration) error {
	return waitRetryAfter(ctx, delay)
}
