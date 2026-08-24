package apilink_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/apilink"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testKey() []byte {
	return bytes.Repeat([]byte{0x04}, 32)
}

func testKeyring(t *testing.T) *apilink.Keyring {
	t.Helper()
	kr, err := apilink.NewKeyring(apilink.DefaultKeyID, testKey(), nil)
	require.NoError(t, err)
	return kr
}

func memoryStore() (*store.MockStore, map[string]model.ApiLink) {
	links := map[string]model.ApiLink{}
	ms := &store.MockStore{
		UpsertApiLinkFn: func(_ context.Context, uid uuid.UUID, name, ciphertext, nonce, keyID string) (model.ApiLink, error) {
			link := model.ApiLink{
				UserID:         uid,
				Name:           name,
				RefreshToken:   ciphertext,
				RefreshTokenIv: nonce,
				KeyID:          keyID,
			}
			links[name] = link
			return link, nil
		},
		GetApiLinkByUserAndNameFn: func(_ context.Context, _ uuid.UUID, name string) (model.ApiLink, error) {
			link, ok := links[name]
			if !ok {
				return model.ApiLink{}, errors.New("not found")
			}
			return link, nil
		},
	}
	return ms, links
}

func TestVault_PutGetRoundTrip(t *testing.T) {
	userID := uuid.New()
	ms, links := memoryStore()

	v := apilink.New(testKeyring(t), ms)
	err := v.PutRefreshToken(context.Background(), userID, apilink.ProviderDiscord, "discord-refresh")
	require.NoError(t, err)

	got, err := v.GetRefreshToken(context.Background(), userID, apilink.ProviderDiscord)
	require.NoError(t, err)
	assert.Equal(t, "discord-refresh", got)
	assert.NotEqual(t, "discord-refresh", links[apilink.ProviderDiscord].RefreshToken)
	assert.Equal(t, apilink.DefaultKeyID, links[apilink.ProviderDiscord].KeyID)
}

func TestVault_ProvidersDoNotClobber(t *testing.T) {
	userID := uuid.New()
	ms, _ := memoryStore()

	v := apilink.New(testKeyring(t), ms)
	require.NoError(t, v.PutRefreshToken(context.Background(), userID, apilink.ProviderDiscord, "discord-secret"))
	require.NoError(t, v.PutRefreshToken(context.Background(), userID, "riot", "riot-secret"))

	discord, err := v.GetRefreshToken(context.Background(), userID, apilink.ProviderDiscord)
	require.NoError(t, err)
	assert.Equal(t, "discord-secret", discord)

	riot, err := v.GetRefreshToken(context.Background(), userID, "riot")
	require.NoError(t, err)
	assert.Equal(t, "riot-secret", riot)
}

func TestVault_CopiedBlobDoesNotDecryptForOtherUser(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()
	var stored model.ApiLink
	ms := &store.MockStore{
		UpsertApiLinkFn: func(_ context.Context, uid uuid.UUID, name, ciphertext, nonce, keyID string) (model.ApiLink, error) {
			stored = model.ApiLink{UserID: uid, Name: name, RefreshToken: ciphertext, RefreshTokenIv: nonce, KeyID: keyID}
			return stored, nil
		},
		GetApiLinkByUserAndNameFn: func(_ context.Context, uid uuid.UUID, name string) (model.ApiLink, error) {
			copied := stored
			copied.UserID = uid
			copied.Name = name
			return copied, nil
		},
	}

	v := apilink.New(testKeyring(t), ms)
	require.NoError(t, v.PutRefreshToken(context.Background(), userA, apilink.ProviderDiscord, "discord-refresh"))

	_, err := v.GetRefreshToken(context.Background(), userB, apilink.ProviderDiscord)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decrypt refresh token")
}

func TestVault_CopiedBlobDoesNotDecryptForOtherProvider(t *testing.T) {
	userID := uuid.New()
	var stored model.ApiLink
	ms := &store.MockStore{
		UpsertApiLinkFn: func(_ context.Context, uid uuid.UUID, name, ciphertext, nonce, keyID string) (model.ApiLink, error) {
			stored = model.ApiLink{UserID: uid, Name: name, RefreshToken: ciphertext, RefreshTokenIv: nonce, KeyID: keyID}
			return stored, nil
		},
		GetApiLinkByUserAndNameFn: func(_ context.Context, uid uuid.UUID, name string) (model.ApiLink, error) {
			copied := stored
			copied.UserID = uid
			copied.Name = name
			return copied, nil
		},
	}

	v := apilink.New(testKeyring(t), ms)
	require.NoError(t, v.PutRefreshToken(context.Background(), userID, apilink.ProviderDiscord, "discord-refresh"))

	_, err := v.GetRefreshToken(context.Background(), userID, "riot")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decrypt refresh token")
}

func TestVault_DecryptsWithPreviousKey(t *testing.T) {
	userID := uuid.New()
	oldKey := bytes.Repeat([]byte{0x0a}, 32)
	newKey := bytes.Repeat([]byte{0x0b}, 32)
	oldRing, err := apilink.NewKeyring("1", oldKey, nil)
	require.NoError(t, err)
	newRing, err := apilink.NewKeyring("2", newKey, map[string][]byte{"1": oldKey})
	require.NoError(t, err)

	var stored model.ApiLink
	ms := &store.MockStore{
		UpsertApiLinkFn: func(_ context.Context, uid uuid.UUID, name, ciphertext, nonce, keyID string) (model.ApiLink, error) {
			stored = model.ApiLink{UserID: uid, Name: name, RefreshToken: ciphertext, RefreshTokenIv: nonce, KeyID: keyID}
			return stored, nil
		},
		GetApiLinkByUserAndNameFn: func(context.Context, uuid.UUID, string) (model.ApiLink, error) {
			return stored, nil
		},
	}

	oldVault := apilink.New(oldRing, ms)
	require.NoError(t, oldVault.PutRefreshToken(context.Background(), userID, apilink.ProviderDiscord, "legacy-token"))
	assert.Equal(t, "1", stored.KeyID)

	newVault := apilink.New(newRing, ms)
	got, err := newVault.GetRefreshToken(context.Background(), userID, apilink.ProviderDiscord)
	require.NoError(t, err)
	assert.Equal(t, "legacy-token", got)

	require.NoError(t, newVault.PutRefreshToken(context.Background(), userID, apilink.ProviderDiscord, "rotated-token"))
	assert.Equal(t, "2", stored.KeyID)
	got, err = newVault.GetRefreshToken(context.Background(), userID, apilink.ProviderDiscord)
	require.NoError(t, err)
	assert.Equal(t, "rotated-token", got)
}

func TestVault_GetUnknownKeyID(t *testing.T) {
	ms := &store.MockStore{
		GetApiLinkByUserAndNameFn: func(context.Context, uuid.UUID, string) (model.ApiLink, error) {
			return model.ApiLink{
				RefreshToken:   "aa",
				RefreshTokenIv: "bb",
				KeyID:          "missing",
			}, nil
		},
	}
	v := apilink.New(testKeyring(t), ms)
	_, err := v.GetRefreshToken(context.Background(), uuid.New(), apilink.ProviderDiscord)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown encryption key id")
}

func TestVault_PutRejectsEmptyProvider(t *testing.T) {
	v := apilink.New(testKeyring(t), &store.MockStore{})
	err := v.PutRefreshToken(context.Background(), uuid.New(), "", "token")
	require.Error(t, err)
}

func TestVault_GetRejectsEmptyProvider(t *testing.T) {
	v := apilink.New(testKeyring(t), &store.MockStore{})
	_, err := v.GetRefreshToken(context.Background(), uuid.New(), "")
	require.Error(t, err)
}

func TestVault_PutStoreError(t *testing.T) {
	ms := &store.MockStore{
		UpsertApiLinkFn: func(context.Context, uuid.UUID, string, string, string, string) (model.ApiLink, error) {
			return model.ApiLink{}, errors.New("db down")
		},
	}
	v := apilink.New(testKeyring(t), ms)
	err := v.PutRefreshToken(context.Background(), uuid.New(), apilink.ProviderDiscord, "token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "store refresh token")
}
