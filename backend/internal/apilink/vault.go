// Package apilink encrypts and persists third-party OAuth refresh tokens keyed by provider name.
package apilink

import (
	"context"
	"errors"
	"fmt"

	"github.com/KatieSuth/MatchmakerAPI/internal/cryptoutil"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
)

// ProviderDiscord is the api_links.name value for Discord OAuth refresh tokens.
const ProviderDiscord = "discord"

// Store is the persistence subset the vault needs. Ciphertext in, ciphertext out.
type Store interface {
	// UpsertApiLink inserts or replaces the encrypted blob for this user and provider.
	UpsertApiLink(ctx context.Context, userID uuid.UUID, name, ciphertext, nonce, keyID string) (model.ApiLink, error)
	// GetApiLinkByUserAndName loads the encrypted blob for this user and provider.
	GetApiLinkByUserAndName(ctx context.Context, userID uuid.UUID, name string) (model.ApiLink, error)
	// GetApiLinkByUserAndNameForUpdate loads the blob and holds a row lock until the tx ends.
	GetApiLinkByUserAndNameForUpdate(ctx context.Context, userID uuid.UUID, name string) (model.ApiLink, error)
	// DeleteApiLinkByUserAndName removes the blob for this user and provider.
	DeleteApiLinkByUserAndName(ctx context.Context, userID uuid.UUID, name string) error
}

// Vault encrypts refresh tokens before writing them and decrypts on read.
type Vault struct {
	keys *Keyring
	s    Store
}

// New returns a Vault that encrypts tokens with keys and persists them via s.
func New(keys *Keyring, s Store) *Vault {
	return &Vault{keys: keys, s: s}
}

// refreshTokenAAD binds ciphertext to this user and provider so a copied blob will not Open.
func refreshTokenAAD(userID uuid.UUID, provider string) []byte {
	return []byte(userID.String() + "\x1f" + provider)
}

// PutRefreshToken encrypts plaintext with the current keyring key and upserts api_links
// for this user and provider. Empty provider is rejected.
func (v *Vault) PutRefreshToken(ctx context.Context, userID uuid.UUID, provider, plaintext string) error {
	if provider == "" {
		return errors.New("provider name must not be empty")
	}
	ciphertext, nonce, err := cryptoutil.Encrypt(v.keys.currentKey(), plaintext, refreshTokenAAD(userID, provider))
	if err != nil {
		return fmt.Errorf("encrypt refresh token: %w", err)
	}
	_, err = v.s.UpsertApiLink(ctx, userID, provider, ciphertext, nonce, v.keys.CurrentID())
	if err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}
	return nil
}

// GetRefreshToken loads the stored blob for this user and provider and opens it with
// the keyring key named by api_links.key_id. Empty provider is rejected.
func (v *Vault) GetRefreshToken(ctx context.Context, userID uuid.UUID, provider string) (string, error) {
	if provider == "" {
		return "", errors.New("provider name must not be empty")
	}
	link, err := v.s.GetApiLinkByUserAndName(ctx, userID, provider)
	if err != nil {
		return "", fmt.Errorf("load refresh token: %w", err)
	}
	key, err := v.keys.key(link.KeyID)
	if err != nil {
		return "", fmt.Errorf("decrypt refresh token: %w", err)
	}
	plain, err := cryptoutil.Decrypt(key, link.RefreshToken, link.RefreshTokenIv, refreshTokenAAD(userID, provider))
	if err != nil {
		return "", fmt.Errorf("decrypt refresh token: %w", err)
	}
	return plain, nil
}

// ForStore returns a Vault that reads and writes through s (typically a transaction-scoped store).
func (v *Vault) ForStore(s Store) *Vault {
	return &Vault{keys: v.keys, s: s}
}

// GetRefreshTokenForUpdate is GetRefreshToken but uses SELECT … FOR UPDATE on the api_links row.
func (v *Vault) GetRefreshTokenForUpdate(ctx context.Context, userID uuid.UUID, provider string) (string, error) {
	if provider == "" {
		return "", errors.New("provider name must not be empty")
	}
	link, err := v.s.GetApiLinkByUserAndNameForUpdate(ctx, userID, provider)
	if err != nil {
		return "", fmt.Errorf("load refresh token: %w", err)
	}
	key, err := v.keys.key(link.KeyID)
	if err != nil {
		return "", fmt.Errorf("decrypt refresh token: %w", err)
	}
	plain, err := cryptoutil.Decrypt(key, link.RefreshToken, link.RefreshTokenIv, refreshTokenAAD(userID, provider))
	if err != nil {
		return "", fmt.Errorf("decrypt refresh token: %w", err)
	}
	return plain, nil
}

// DeleteRefreshToken removes the stored blob for this user and provider. Empty provider is rejected.
func (v *Vault) DeleteRefreshToken(ctx context.Context, userID uuid.UUID, provider string) error {
	if provider == "" {
		return errors.New("provider name must not be empty")
	}
	if err := v.s.DeleteApiLinkByUserAndName(ctx, userID, provider); err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}
