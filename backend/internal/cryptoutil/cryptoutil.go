// Package cryptoutil provides AES-256-GCM helpers for reversible at-rest secrets.
package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

const aes256KeySize = 32
const gcmNonceSize = 12

// Overridable in tests so constructor / entropy failures can be exercised.
var (
	randReader   io.Reader                               = rand.Reader
	aesNewCipher func([]byte) (cipher.Block, error)      = aes.NewCipher
	cipherNewGCM func(cipher.Block) (cipher.AEAD, error) = cipher.NewGCM
)

// ParseAES256Key decodes a 64-character hex string into a 32-byte AES-256 key.
func ParseAES256Key(hexKey string) ([]byte, error) {
	if hexKey == "" {
		return nil, errors.New("encryption key must not be empty")
	}
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("encryption key hex: %w", err)
	}
	if len(key) != aes256KeySize {
		return nil, fmt.Errorf("encryption key must be %d bytes", aes256KeySize)
	}
	return key, nil
}

// newGCM builds an AES-256-GCM AEAD from key. Key length must be 32 bytes.
func newGCM(key []byte) (cipher.AEAD, error) {
	if len(key) != aes256KeySize {
		return nil, fmt.Errorf("encryption key must be %d bytes", aes256KeySize)
	}
	block, err := aesNewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipherNewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	return gcm, nil
}

// Encrypt seals plaintext with AES-256-GCM. aad is additional authenticated data bound
// into the tag (not encrypted); Open fails if it does not match. ciphertextHex is the
// sealed blob; nonceHex is the 12-byte nonce. Both are hex-encoded for storage in text columns.
func Encrypt(key []byte, plaintext string, aad []byte) (ciphertextHex, nonceHex string, err error) {
	if plaintext == "" {
		return "", "", errors.New("plaintext must not be empty")
	}
	if len(aad) == 0 {
		return "", "", errors.New("additional authenticated data must not be empty")
	}

	gcm, err := newGCM(key)
	if err != nil {
		return "", "", err
	}

	nonce := make([]byte, gcmNonceSize)
	if _, err := io.ReadFull(randReader, nonce); err != nil {
		return "", "", fmt.Errorf("nonce: %w", err)
	}

	sealed := gcm.Seal(nil, nonce, []byte(plaintext), aad)
	return hex.EncodeToString(sealed), hex.EncodeToString(nonce), nil
}

// Decrypt opens AES-256-GCM ciphertextHex using nonceHex and the same aad used at Seal.
// Both hex values must be hex-encoded.
func Decrypt(key []byte, ciphertextHex, nonceHex string, aad []byte) (string, error) {
	if len(aad) == 0 {
		return "", errors.New("additional authenticated data must not be empty")
	}

	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}

	sealed, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("ciphertext hex: %w", err)
	}
	nonce, err := hex.DecodeString(nonceHex)
	if err != nil {
		return "", fmt.Errorf("nonce hex: %w", err)
	}

	plain, err := gcm.Open(nil, nonce, sealed, aad)
	if err != nil {
		return "", fmt.Errorf("open: %w", err)
	}
	return string(plain), nil
}
