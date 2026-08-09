// Package textinput normalizes optional free-entry strings for API handlers.
package textinput

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	// ErrTooLong means the trimmed value exceeds maxRunes Unicode code points.
	ErrTooLong = errors.New("text too long")
	// ErrInvalidChars means the trimmed value contains disallowed runes
	// (C0 controls, DEL, or both ASCII '<' and '>' which can form HTML/markup).
	ErrInvalidChars = errors.New("text contains invalid characters")
)

// NormalizeOptional trims, rejects unsafe/control markup characters, and enforces
// maxRunes (Unicode code points). Empty after trim returns ("", nil) so callers
// can store NULL.
//
// Disallowed: C0 controls (U+0000–U+001F), DEL (U+007F), and the combination of
// both ASCII '<' and '>' in the same value (either alone is allowed, e.g. "A < B").
func NormalizeOptional(raw string, maxRunes int) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	hasLT, hasGT := false, false
	for _, r := range trimmed {
		if r < 0x20 || r == 0x7F {
			return "", ErrInvalidChars
		}
		if r == '<' {
			hasLT = true
		}
		if r == '>' {
			hasGT = true
		}
	}
	if hasLT && hasGT {
		return "", ErrInvalidChars
	}
	if utf8.RuneCountInString(trimmed) > maxRunes {
		return "", ErrTooLong
	}
	return trimmed, nil
}
