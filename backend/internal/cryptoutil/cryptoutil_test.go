package cryptoutil_test

import (
	"bytes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"io"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/cryptoutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type errReader struct {
	err error
}

func (r errReader) Read([]byte) (int, error) {
	return 0, r.err
}

func testKey() []byte {
	return bytes.Repeat([]byte{0x04}, 32)
}

func testAAD() []byte {
	return []byte("user-id\x1fdiscord")
}

func TestParseAES256Key_Accepts64Hex(t *testing.T) {
	hexKey := hex.EncodeToString(testKey())
	key, err := cryptoutil.ParseAES256Key(hexKey)
	require.NoError(t, err)
	assert.Equal(t, testKey(), key)
}

func TestParseAES256Key_RejectsEmpty(t *testing.T) {
	_, err := cryptoutil.ParseAES256Key("")
	require.Error(t, err)
}

func TestParseAES256Key_RejectsOddLength(t *testing.T) {
	_, err := cryptoutil.ParseAES256Key("abc")
	require.Error(t, err)
}

func TestParseAES256Key_RejectsWrongSize(t *testing.T) {
	_, err := cryptoutil.ParseAES256Key("aa")
	require.Error(t, err)
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	cipherText, nonce, err := cryptoutil.Encrypt(testKey(), "discord-refresh-token", testAAD())
	require.NoError(t, err)
	assert.NotEmpty(t, cipherText)
	assert.NotEmpty(t, nonce)
	assert.NotEqual(t, "discord-refresh-token", cipherText)

	plain, err := cryptoutil.Decrypt(testKey(), cipherText, nonce, testAAD())
	require.NoError(t, err)
	assert.Equal(t, "discord-refresh-token", plain)
}

func TestEncrypt_RejectsWrongKeySize(t *testing.T) {
	_, _, err := cryptoutil.Encrypt([]byte("short"), "token", testAAD())
	require.Error(t, err)
}

func TestEncrypt_RejectsEmptyPlaintext(t *testing.T) {
	_, _, err := cryptoutil.Encrypt(testKey(), "", testAAD())
	require.Error(t, err)
}

func TestEncrypt_RejectsEmptyAAD(t *testing.T) {
	_, _, err := cryptoutil.Encrypt(testKey(), "token", nil)
	require.Error(t, err)
}

func TestDecrypt_RejectsEmptyAAD(t *testing.T) {
	cipherText, nonce, err := cryptoutil.Encrypt(testKey(), "secret", testAAD())
	require.NoError(t, err)
	_, err = cryptoutil.Decrypt(testKey(), cipherText, nonce, nil)
	require.Error(t, err)
}

func TestDecrypt_WrongKey(t *testing.T) {
	cipherText, nonce, err := cryptoutil.Encrypt(testKey(), "secret", testAAD())
	require.NoError(t, err)

	otherKey := bytes.Repeat([]byte{0x05}, 32)
	_, err = cryptoutil.Decrypt(otherKey, cipherText, nonce, testAAD())
	require.Error(t, err)
}

func TestDecrypt_WrongAAD(t *testing.T) {
	cipherText, nonce, err := cryptoutil.Encrypt(testKey(), "secret", testAAD())
	require.NoError(t, err)

	_, err = cryptoutil.Decrypt(testKey(), cipherText, nonce, []byte("other-user\x1fdiscord"))
	require.Error(t, err)
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	cipherText, nonce, err := cryptoutil.Encrypt(testKey(), "secret", testAAD())
	require.NoError(t, err)

	tampered := cipherText[:len(cipherText)-2] + "ff"
	_, err = cryptoutil.Decrypt(testKey(), tampered, nonce, testAAD())
	require.Error(t, err)
}

func TestDecrypt_InvalidHex(t *testing.T) {
	_, err := cryptoutil.Decrypt(testKey(), "not-hex", "00", testAAD())
	require.Error(t, err)
}

func TestDecrypt_InvalidNonceHex(t *testing.T) {
	_, err := cryptoutil.Decrypt(testKey(), "aa", "not-hex", testAAD())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonce hex")
}

func TestDecrypt_RejectsWrongKeySize(t *testing.T) {
	_, err := cryptoutil.Decrypt([]byte("short"), "aa", "bb", testAAD())
	require.Error(t, err)
}

func TestEncrypt_NonceReadFailure(t *testing.T) {
	cryptoutil.SetRandReaderForTest(errReader{err: io.ErrUnexpectedEOF})
	t.Cleanup(cryptoutil.ResetRandReaderForTest)

	_, _, err := cryptoutil.Encrypt(testKey(), "token", testAAD())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonce")
}

func TestEncrypt_AESNewCipherFailure(t *testing.T) {
	cryptoutil.SetAESNewCipherForTest(func([]byte) (cipher.Block, error) {
		return nil, errors.New("cipher boom")
	})
	t.Cleanup(cryptoutil.ResetAESNewCipherForTest)

	_, _, err := cryptoutil.Encrypt(testKey(), "token", testAAD())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "aes cipher")
}

func TestEncrypt_GCMInitFailure(t *testing.T) {
	cryptoutil.SetCipherNewGCMForTest(func(cipher.Block) (cipher.AEAD, error) {
		return nil, errors.New("gcm boom")
	})
	t.Cleanup(cryptoutil.ResetCipherNewGCMForTest)

	_, _, err := cryptoutil.Encrypt(testKey(), "token", testAAD())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gcm")
}
