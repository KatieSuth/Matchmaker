package apilink_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/apilink"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func keyA() []byte { return bytes.Repeat([]byte{0x04}, 32) }
func keyB() []byte { return bytes.Repeat([]byte{0x05}, 32) }

func TestNewKeyring_StoresCurrent(t *testing.T) {
	kr, err := apilink.NewKeyring("1", keyA(), nil)
	require.NoError(t, err)
	assert.Equal(t, "1", kr.CurrentID())
}

func TestNewKeyring_RejectsEmptyCurrentID(t *testing.T) {
	_, err := apilink.NewKeyring("", keyA(), nil)
	require.Error(t, err)
}

func TestNewKeyring_RejectsInvalidCurrentID(t *testing.T) {
	_, err := apilink.NewKeyring("1:sneaky", keyA(), nil)
	require.Error(t, err)
}

func TestNewKeyring_RejectsWrongKeySize(t *testing.T) {
	_, err := apilink.NewKeyring("1", []byte("short"), nil)
	require.Error(t, err)
}

func TestNewKeyring_RejectsPreviousDuplicateOfCurrent(t *testing.T) {
	_, err := apilink.NewKeyring("2", keyB(), map[string][]byte{"2": keyA()})
	require.Error(t, err)
}

func TestNewKeyring_RejectsPreviousWrongSize(t *testing.T) {
	_, err := apilink.NewKeyring("2", keyB(), map[string][]byte{"1": []byte("short")})
	require.Error(t, err)
}

func TestParsePreviousKeys_Empty(t *testing.T) {
	keys, err := apilink.ParsePreviousKeys("")
	require.NoError(t, err)
	assert.Empty(t, keys)

	keys, err = apilink.ParsePreviousKeys("   ")
	require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestParsePreviousKeys_Single(t *testing.T) {
	hexKey := hex.EncodeToString(keyA())
	keys, err := apilink.ParsePreviousKeys("1:" + hexKey)
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, keyA(), keys["1"])
}

func TestParsePreviousKeys_Multiple(t *testing.T) {
	keys, err := apilink.ParsePreviousKeys("1:" + hex.EncodeToString(keyA()) + ",old:" + hex.EncodeToString(keyB()))
	require.NoError(t, err)
	require.Len(t, keys, 2)
	assert.Equal(t, keyA(), keys["1"])
	assert.Equal(t, keyB(), keys["old"])
}

func TestParsePreviousKeys_TrimsWhitespace(t *testing.T) {
	keys, err := apilink.ParsePreviousKeys(" 1:" + hex.EncodeToString(keyA()) + " , 2:" + hex.EncodeToString(keyB()) + " ")
	require.NoError(t, err)
	require.Len(t, keys, 2)
}

func TestParsePreviousKeys_RejectsBadEntry(t *testing.T) {
	_, err := apilink.ParsePreviousKeys("nocolon")
	require.Error(t, err)
}

func TestParsePreviousKeys_RejectsDuplicateIDs(t *testing.T) {
	hexKey := hex.EncodeToString(keyA())
	_, err := apilink.ParsePreviousKeys("1:" + hexKey + ",1:" + hexKey)
	require.Error(t, err)
}

func TestParsePreviousKeys_RejectsInvalidHex(t *testing.T) {
	_, err := apilink.ParsePreviousKeys("1:zzzz")
	require.Error(t, err)
}
