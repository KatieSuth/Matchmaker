package lobbyjoin_test

import (
	"strings"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/lobbyjoin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr(s string) *string { return &s }

func TestValidationError_ErrorAndUnwrap(t *testing.T) {
	t.Parallel()
	err := &lobbyjoin.ValidationError{Message: "bad join code"}
	assert.Equal(t, "bad join code", err.Error())
	assert.ErrorIs(t, err, lobbyjoin.ErrInvalidJoinCode)
}

func TestNormalize_EmptyClears(t *testing.T) {
	t.Parallel()
	got, err := lobbyjoin.Normalize("  ", ptr("https://gg.riotgames.com"))
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestNormalize_PlainCode(t *testing.T) {
	t.Parallel()
	got, err := lobbyjoin.Normalize("JHL829", ptr("https://gg.riotgames.com"))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "JHL829", *got)
}

func TestNormalize_PlainCodeTooLong(t *testing.T) {
	t.Parallel()
	_, err := lobbyjoin.Normalize(strings.Repeat("A", 65), ptr("https://gg.riotgames.com"))
	require.Error(t, err)
	assert.ErrorIs(t, err, lobbyjoin.ErrInvalidJoinCode)
	var valErr *lobbyjoin.ValidationError
	require.ErrorAs(t, err, &valErr)
	assert.Equal(t, "Lobby code is too long", valErr.Message)
}

func TestNormalize_PlainCodeInvalidChars(t *testing.T) {
	t.Parallel()
	_, err := lobbyjoin.Normalize("BAD CODE!", ptr("https://gg.riotgames.com"))
	require.Error(t, err)
	assert.ErrorIs(t, err, lobbyjoin.ErrInvalidJoinCode)
}

func TestNormalize_FullRiotURL(t *testing.T) {
	t.Parallel()
	got, err := lobbyjoin.Normalize(
		"https://gg.riotgames.com/LOL?joinCode=FWXy-7C7m-3KQN",
		ptr("https://gg.riotgames.com"),
	)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "/LOL?joinCode=FWXy-7C7m-3KQN", *got)
}

func TestNormalize_DropsFragment(t *testing.T) {
	t.Parallel()
	got, err := lobbyjoin.Normalize(
		"https://gg.riotgames.com/LOL?joinCode=abc#section",
		ptr("https://gg.riotgames.com"),
	)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "/LOL?joinCode=abc", *got)
}

func TestNormalize_RootPathWithQuery(t *testing.T) {
	t.Parallel()
	got, err := lobbyjoin.Normalize(
		"https://gg.riotgames.com/?joinCode=abc",
		ptr("https://gg.riotgames.com"),
	)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "/?joinCode=abc", *got)
}

func TestNormalize_MissingPath(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		"https://gg.riotgames.com",
		"https://gg.riotgames.com/",
	} {
		_, err := lobbyjoin.Normalize(raw, ptr("https://gg.riotgames.com"))
		require.Error(t, err, raw)
		assert.ErrorIs(t, err, lobbyjoin.ErrInvalidJoinCode)
	}
}

func TestNormalize_PortMismatch(t *testing.T) {
	t.Parallel()
	_, err := lobbyjoin.Normalize(
		"https://gg.riotgames.com:8443/LOL?joinCode=x",
		ptr("https://gg.riotgames.com"),
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, lobbyjoin.ErrInvalidJoinCode)
}

func TestNormalize_EmptyHostURL(t *testing.T) {
	t.Parallel()
	_, err := lobbyjoin.Normalize("https:///LOL", ptr("https://gg.riotgames.com"))
	require.Error(t, err)
	assert.ErrorIs(t, err, lobbyjoin.ErrInvalidJoinCode)
}

func TestNormalize_EmptyJoinLinkBaseString(t *testing.T) {
	t.Parallel()
	_, err := lobbyjoin.Normalize("https://gg.riotgames.com/LOL", ptr("   "))
	require.Error(t, err)
	var valErr *lobbyjoin.ValidationError
	require.ErrorAs(t, err, &valErr)
	assert.Equal(t, "Links are not supported for this game; enter a lobby code instead", valErr.Message)
}

func TestNormalize_MisconfiguredBase(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		base string
	}{
		{"http scheme", "http://gg.riotgames.com"},
		{"with path", "https://gg.riotgames.com/extra"},
		{"with userinfo", "https://user@gg.riotgames.com"},
		{"no host", "https://"},
		{"not a url", "not-a-url"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := lobbyjoin.Normalize("https://gg.riotgames.com/LOL?joinCode=x", ptr(tc.base))
			require.Error(t, err)
			assert.ErrorIs(t, err, lobbyjoin.ErrInvalidJoinCode)
		})
	}
}

func TestNormalize_SchemelessOfficial(t *testing.T) {
	t.Parallel()
	got, err := lobbyjoin.Normalize(
		"gg.riotgames.com/LOL?joinCode=FWXy-7C7m-3KQN",
		ptr("https://gg.riotgames.com"),
	)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "/LOL?joinCode=FWXy-7C7m-3KQN", *got)
}

func TestNormalize_SchemelessQueryOnly(t *testing.T) {
	t.Parallel()
	// host?query with no slash — still link-shaped; missing path fails after parse.
	_, err := lobbyjoin.Normalize("gg.riotgames.com?joinCode=x", ptr("https://gg.riotgames.com"))
	require.Error(t, err)
}

func TestNormalize_RejectBadActorHost(t *testing.T) {
	t.Parallel()
	_, err := lobbyjoin.Normalize("gg.badactor.net/somecode", ptr("https://gg.riotgames.com"))
	require.Error(t, err)
	assert.ErrorIs(t, err, lobbyjoin.ErrInvalidJoinCode)
}

func TestNormalize_RejectPrefixBypass(t *testing.T) {
	t.Parallel()
	_, err := lobbyjoin.Normalize(
		"https://gg.riotgames.com.evil.com/LOL?joinCode=x",
		ptr("https://gg.riotgames.com"),
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, lobbyjoin.ErrInvalidJoinCode)
}

func TestNormalize_RejectCredentialedURL(t *testing.T) {
	t.Parallel()
	_, err := lobbyjoin.Normalize(
		"https://user@gg.riotgames.com/LOL",
		ptr("https://gg.riotgames.com"),
	)
	require.Error(t, err)
}

func TestNormalize_RejectPathSmuggling(t *testing.T) {
	t.Parallel()
	cases := []string{
		"//evil.com",
		"/@evil.com",
		"/LOL://evil",
		`/\evil`,
	}
	for _, c := range cases {
		_, err := lobbyjoin.Normalize(c, ptr("https://gg.riotgames.com"))
		require.Error(t, err, c)
	}
}

func TestNormalize_PathTooLong(t *testing.T) {
	t.Parallel()
	path := "/" + strings.Repeat("a", 512)
	_, err := lobbyjoin.Normalize(path, ptr("https://gg.riotgames.com"))
	require.Error(t, err)
	var valErr *lobbyjoin.ValidationError
	require.ErrorAs(t, err, &valErr)
	assert.Equal(t, "Lobby join link path is too long", valErr.Message)
}

func TestNormalize_PathWithWhitespace(t *testing.T) {
	t.Parallel()
	_, err := lobbyjoin.Normalize("/LOL?join Code=x", ptr("https://gg.riotgames.com"))
	require.Error(t, err)
}

func TestNormalize_PathSuffixRejectedWhenNoBase(t *testing.T) {
	t.Parallel()
	_, err := lobbyjoin.Normalize("/LOL?joinCode=x", nil)
	require.Error(t, err)
	_, err = lobbyjoin.Normalize("/LOL?joinCode=x", ptr(""))
	require.Error(t, err)
}

func TestNormalize_AlreadyStrippedPath(t *testing.T) {
	t.Parallel()
	got, err := lobbyjoin.Normalize("/LOL?joinCode=abc", ptr("https://gg.riotgames.com"))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "/LOL?joinCode=abc", *got)
}

func TestNormalize_PrependSlashToSuffix(t *testing.T) {
	t.Parallel()
	got, err := lobbyjoin.Normalize("LOL?joinCode=abc", ptr("https://gg.riotgames.com"))
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "/LOL?joinCode=abc", *got)
}

func TestNormalize_RejectHTTP(t *testing.T) {
	t.Parallel()
	_, err := lobbyjoin.Normalize(
		"http://gg.riotgames.com/LOL?joinCode=x",
		ptr("https://gg.riotgames.com"),
	)
	require.Error(t, err)
}

func TestNormalize_LinksRejectedWhenNoBase(t *testing.T) {
	t.Parallel()
	_, err := lobbyjoin.Normalize("https://gg.riotgames.com/LOL?joinCode=x", nil)
	require.Error(t, err)
}

func TestIsLinkShaped(t *testing.T) {
	t.Parallel()
	assert.True(t, lobbyjoin.ExportedIsLinkShaped("http://example.com/x"))
	assert.True(t, lobbyjoin.ExportedIsLinkShaped("HTTPS://example.com/x"))
	assert.True(t, lobbyjoin.ExportedIsLinkShaped("example.com/path"))
	assert.True(t, lobbyjoin.ExportedIsLinkShaped("example.com?q=1"))
	// ? appears before / → cut at ?
	assert.True(t, lobbyjoin.ExportedIsLinkShaped("example.com?q=/path"))
	assert.False(t, lobbyjoin.ExportedIsLinkShaped("nopath"))
	assert.False(t, lobbyjoin.ExportedIsLinkShaped("/onlypath"))
	assert.False(t, lobbyjoin.ExportedIsLinkShaped("nodothost/path"))
	assert.False(t, lobbyjoin.ExportedIsLinkShaped("bad host.com/path"))
	assert.False(t, lobbyjoin.ExportedIsLinkShaped(""))
}

func TestValidateStoredPath_Direct(t *testing.T) {
	t.Parallel()
	got, err := lobbyjoin.ExportedValidateStoredPath("/ok")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "/ok", *got)

	_, err = lobbyjoin.ExportedValidateStoredPath("missing-slash")
	require.Error(t, err)

	_, err = lobbyjoin.ExportedValidateStoredPath("//evil")
	require.Error(t, err)

	_, err = lobbyjoin.ExportedValidateStoredPath("/has\x00null")
	require.Error(t, err)

	_, err = lobbyjoin.ExportedValidateStoredPath("/bad!char")
	require.Error(t, err)
}

func TestParseJoinLinkBase_Direct(t *testing.T) {
	t.Parallel()
	base, err := lobbyjoin.ExportedParseJoinLinkBase("https://gg.riotgames.com")
	require.NoError(t, err)
	assert.Equal(t, "gg.riotgames.com", base.Hostname())

	base, err = lobbyjoin.ExportedParseJoinLinkBase("https://gg.riotgames.com/")
	require.NoError(t, err)
	assert.Equal(t, "gg.riotgames.com", base.Hostname())

	_, err = lobbyjoin.ExportedParseJoinLinkBase("http://gg.riotgames.com")
	require.Error(t, err)

	_, err = lobbyjoin.ExportedParseJoinLinkBase("https://user@gg.riotgames.com")
	require.Error(t, err)

	_, err = lobbyjoin.ExportedParseJoinLinkBase("https://gg.riotgames.com/path")
	require.Error(t, err)

	_, err = lobbyjoin.ExportedParseJoinLinkBase("://")
	require.Error(t, err)
}
