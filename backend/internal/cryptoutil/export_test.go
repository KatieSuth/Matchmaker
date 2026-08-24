package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// SetRandReaderForTest replaces the entropy source used to generate GCM nonces.
func SetRandReaderForTest(r io.Reader) {
	randReader = r
}

// ResetRandReaderForTest restores crypto/rand as the nonce entropy source.
func ResetRandReaderForTest() {
	randReader = rand.Reader
}

// SetAESNewCipherForTest replaces aes.NewCipher so tests can force constructor failures.
func SetAESNewCipherForTest(fn func([]byte) (cipher.Block, error)) {
	aesNewCipher = fn
}

// ResetAESNewCipherForTest restores aes.NewCipher as the block-cipher constructor.
func ResetAESNewCipherForTest() {
	aesNewCipher = aes.NewCipher
}

// SetCipherNewGCMForTest replaces cipher.NewGCM so tests can force AEAD init failures.
func SetCipherNewGCMForTest(fn func(cipher.Block) (cipher.AEAD, error)) {
	cipherNewGCM = fn
}

// ResetCipherNewGCMForTest restores cipher.NewGCM as the AEAD constructor.
func ResetCipherNewGCMForTest() {
	cipherNewGCM = cipher.NewGCM
}
