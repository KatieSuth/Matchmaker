// Package lobbyjoin normalizes and validates lobby join codes / Riot invite paths.
package lobbyjoin

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

const (
	plainCodeMaxLen = 64
	linkPathMaxLen  = 512
)

// ErrInvalidJoinCode is the sentinel for join-code validation failures.
var ErrInvalidJoinCode = errors.New("invalid lobby join code")

// ValidationError carries a client-facing message for join-code validation failures.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *ValidationError) Unwrap() error {
	return ErrInvalidJoinCode
}

var (
	plainCodePattern = regexp.MustCompile(`^[A-Za-z0-9-]+$`)
	// linkPathPattern allows a single path segment plus query (e.g. /LOL?joinCode=… or /VAL?joinCode=…).
	// Additional "/" characters after the leading slash are intentionally rejected for now;
	// widening this for nested paths is a future consideration after more invite-link research.
	linkPathPattern = regexp.MustCompile(`^/[A-Za-z0-9?&\-._=%]*$`)
)

// Normalize validates and normalizes a raw join-code input against the game's join_link_base.
// Empty/whitespace clears (returns nil, nil). Link-shaped values are parsed with hostname
// equality against joinLinkBase; plain codes are stored as-is.
func Normalize(raw string, joinLinkBase *string) (*string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	if isLinkShaped(trimmed) {
		return normalizeLink(trimmed, joinLinkBase)
	}

	if strings.HasPrefix(trimmed, "/") || strings.ContainsAny(trimmed, "?/") {
		return normalizePathSuffix(trimmed, joinLinkBase)
	}

	return normalizePlainCode(trimmed)
}

// isLinkShaped reports whether s looks like a URL (with or without a scheme).
// Schemeless values such as "gg.badactor.net/x" must still be treated as links
// so they are rejected rather than accepted as plain codes.
func isLinkShaped(s string) bool {
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return true
	}
	// Schemeless host/path or host?query where host contains a dot.
	slash := strings.IndexByte(s, '/')
	qmark := strings.IndexByte(s, '?')
	cut := -1
	switch {
	case slash >= 0 && qmark >= 0:
		cut = slash
		if qmark < slash {
			cut = qmark
		}
	case slash >= 0:
		cut = slash
	case qmark >= 0:
		cut = qmark
	}
	if cut <= 0 {
		return false
	}
	host := s[:cut]
	return strings.Contains(host, ".") && !strings.ContainsAny(host, " \t\r\n")
}

// normalizeLink parses a full or schemeless invite URL, requires hostname equality
// with joinLinkBase (never string prefix), and returns the path+query for storage.
func normalizeLink(raw string, joinLinkBase *string) (*string, error) {
	if joinLinkBase == nil || strings.TrimSpace(*joinLinkBase) == "" {
		return nil, &ValidationError{Message: "Links are not supported for this game; enter a lobby code instead"}
	}

	toParse := raw
	lower := strings.ToLower(raw)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		toParse = "https://" + raw
	}

	parsed, err := url.Parse(toParse)
	if err != nil || parsed.Host == "" {
		return nil, &ValidationError{Message: "Invalid lobby join link"}
	}
	if !strings.EqualFold(parsed.Scheme, "https") {
		return nil, &ValidationError{Message: "Lobby join links must use https"}
	}
	if parsed.User != nil {
		return nil, &ValidationError{Message: "Invalid lobby join link"}
	}

	base, err := parseJoinLinkBase(*joinLinkBase)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(parsed.Hostname(), base.Hostname()) {
		return nil, &ValidationError{Message: "Lobby join link must use the official game join host"}
	}
	if parsed.Port() != base.Port() {
		return nil, &ValidationError{Message: "Lobby join link must use the official game join host"}
	}

	path := parsed.EscapedPath()
	if path == "" || path == "/" {
		// url.Parse may give "/" for root; require a non-root path for invites.
		if path == "" || (path == "/" && parsed.RawQuery == "") {
			return nil, &ValidationError{Message: "Lobby join link is missing a path"}
		}
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	stored := path
	if parsed.RawQuery != "" {
		stored = path + "?" + parsed.RawQuery
	}
	// Drop fragments intentionally.

	return validateStoredPath(stored)
}

// normalizePathSuffix accepts an already-stripped invite path (e.g. "/LOL?joinCode=…")
// or a path without a leading slash, and stores it after allowlist checks.
func normalizePathSuffix(raw string, joinLinkBase *string) (*string, error) {
	if joinLinkBase == nil || strings.TrimSpace(*joinLinkBase) == "" {
		return nil, &ValidationError{Message: "Links are not supported for this game; enter a lobby code instead"}
	}
	path := raw
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return validateStoredPath(path)
}

// normalizePlainCode stores a short alphanumeric lobby code as-is (never combined with join_link_base).
func normalizePlainCode(raw string) (*string, error) {
	if len(raw) > plainCodeMaxLen {
		return nil, &ValidationError{Message: "Lobby code is too long"}
	}
	if !plainCodePattern.MatchString(raw) {
		return nil, &ValidationError{Message: "Lobby code must contain only letters, digits, and hyphens"}
	}
	out := raw
	return &out, nil
}

// validateStoredPath enforces length and charset rules on a path+query so rebuilt
// hrefs cannot smuggle credentials, schemes, or alternate hosts.
// Paths are game-agnostic (/LOL?…, /VAL?…, or other single-segment shapes), but only
// one leading "/" is allowed today—nested paths are deferred pending invite-link research.
func validateStoredPath(path string) (*string, error) {
	if len(path) > linkPathMaxLen {
		return nil, &ValidationError{Message: "Lobby join link path is too long"}
	}
	if !strings.HasPrefix(path, "/") {
		return nil, &ValidationError{Message: "Invalid lobby join link path"}
	}
	// Reject protocol-relative //evil.com and path smuggling.
	if strings.HasPrefix(path, "//") {
		return nil, &ValidationError{Message: "Invalid lobby join link path"}
	}
	for _, r := range path {
		if r == ':' || r == '@' || r == '\\' || unicode.IsSpace(r) || unicode.IsControl(r) {
			return nil, &ValidationError{Message: "Invalid lobby join link path"}
		}
	}
	if !linkPathPattern.MatchString(path) {
		return nil, &ValidationError{Message: "Invalid lobby join link path"}
	}
	out := path
	return &out, nil
}

// parseJoinLinkBase parses the game's configured HTTPS origin used for host equality checks.
func parseJoinLinkBase(base string) (*url.URL, error) {
	trimmed := strings.TrimSpace(base)
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host == "" {
		return nil, &ValidationError{Message: "Game join link base is misconfigured"}
	}
	if !strings.EqualFold(parsed.Scheme, "https") {
		return nil, &ValidationError{Message: "Game join link base is misconfigured"}
	}
	if parsed.User != nil {
		return nil, &ValidationError{Message: "Game join link base is misconfigured"}
	}
	path := parsed.EscapedPath()
	if path != "" && path != "/" {
		return nil, &ValidationError{Message: "Game join link base is misconfigured"}
	}
	return parsed, nil
}
