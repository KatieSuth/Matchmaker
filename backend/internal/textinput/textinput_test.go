package textinput_test

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/KatieSuth/MatchmakerAPI/internal/textinput"
)

func TestNormalizeOptional_EmptyAndWhitespace(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"", "   ", "\t\n"} {
		got, err := textinput.NormalizeOptional(raw, 50)
		if err != nil {
			t.Fatalf("NormalizeOptional(%q): unexpected err %v", raw, err)
		}
		if got != "" {
			t.Fatalf("NormalizeOptional(%q) = %q, want empty", raw, got)
		}
	}
}

func TestNormalizeOptional_Trims(t *testing.T) {
	t.Parallel()

	got, err := textinput.NormalizeOptional("  hi  ", 50)
	if err != nil {
		t.Fatal(err)
	}
	if got != "hi" {
		t.Fatalf("got %q, want hi", got)
	}
}

func TestNormalizeOptional_TooLong(t *testing.T) {
	t.Parallel()

	raw := strings.Repeat("a", 51)
	_, err := textinput.NormalizeOptional(raw, 50)
	if !errors.Is(err, textinput.ErrTooLong) {
		t.Fatalf("got err %v, want ErrTooLong", err)
	}
}

func TestNormalizeOptional_ExactMax(t *testing.T) {
	t.Parallel()

	raw := strings.Repeat("a", 50)
	got, err := textinput.NormalizeOptional(raw, 50)
	if err != nil {
		t.Fatal(err)
	}
	if got != raw {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizeOptional_ControlChars(t *testing.T) {
	t.Parallel()

	cases := []string{
		"a\nb",
		"a\x00b",
		"a\x1fb",
		"a\x7fb",
	}
	for _, raw := range cases {
		_, err := textinput.NormalizeOptional(raw, 50)
		if !errors.Is(err, textinput.ErrInvalidChars) {
			t.Fatalf("NormalizeOptional(%q): got %v, want ErrInvalidChars", raw, err)
		}
	}
}

func TestNormalizeOptional_RejectsPairedAngleBrackets(t *testing.T) {
	t.Parallel()

	reject := []string{
		"<script>alert(1)</script>",
		"hello<script>",
		"<img src=x onerror=alert(1)>",
		"A < B >",
		"<>",
		"><",
	}
	for _, raw := range reject {
		_, err := textinput.NormalizeOptional(raw, 50)
		if !errors.Is(err, textinput.ErrInvalidChars) {
			t.Fatalf("NormalizeOptional(%q): got %v, want ErrInvalidChars", raw, err)
		}
	}

	allow := []string{
		"A < B",
		"A > B",
		"<solo",
		"solo>",
	}
	for _, raw := range allow {
		got, err := textinput.NormalizeOptional(raw, 50)
		if err != nil {
			t.Fatalf("NormalizeOptional(%q): unexpected err %v", raw, err)
		}
		if got != raw {
			t.Fatalf("NormalizeOptional(%q) = %q", raw, got)
		}
	}
}

func TestNormalizeOptional_NonASCIIAndEmoji(t *testing.T) {
	t.Parallel()

	// "a👨b" is 3 Unicode code points (matches Go utf8.RuneCountInString).
	sample := "a👨b"
	if utf8.RuneCountInString(sample) != 3 {
		t.Fatalf("fixture rune count = %d, want 3", utf8.RuneCountInString(sample))
	}
	got, err := textinput.NormalizeOptional(sample, 50)
	if err != nil {
		t.Fatal(err)
	}
	if got != sample {
		t.Fatalf("got %q", got)
	}

	// ZWJ sequences count as multiple code points, not one grapheme.
	zwj := "👨‍👩‍👧‍👦"
	n := utf8.RuneCountInString(zwj)
	if n <= 1 {
		t.Fatalf("expected multi-code-point emoji, got %d", n)
	}
	got, err = textinput.NormalizeOptional(zwj, n)
	if err != nil {
		t.Fatal(err)
	}
	if got != zwj {
		t.Fatalf("got %q", got)
	}
	_, err = textinput.NormalizeOptional(zwj, n-1)
	if !errors.Is(err, textinput.ErrTooLong) {
		t.Fatalf("got err %v, want ErrTooLong", err)
	}

	cjk := "你好世界"
	got, err = textinput.NormalizeOptional(cjk, 50)
	if err != nil {
		t.Fatal(err)
	}
	if got != cjk {
		t.Fatalf("got %q", got)
	}
}
