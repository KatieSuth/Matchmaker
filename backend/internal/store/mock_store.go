package store

import (
	"context"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
)

// mockStore implements store.Store, allowing individual methods to be overridden
// per test. Any method not overridden will panic, making accidental calls obvious.
type MockStore struct {
	Store // embed to satisfy the interface; unset methods will panic

	// games
	GetSystemGamesFn func(ctx context.Context) ([]model.Game, error)
	GetGameRanksFn   func(ctx context.Context, gameID *uuid.UUID) ([]model.GameRank, error)

	// users
	GetUserByDiscordIDFn  func(ctx context.Context, discordID string, errorOnNoRows bool) (model.User, error)
	CreateNewUserFn       func(ctx context.Context, discordUser model.DiscordUser) (model.User, error)
	UpdateUserFromLoginFn func(ctx context.Context, userID uuid.UUID, discordUser model.DiscordUser) (model.User, error)

	// refresh tokens
	CreateNewRefreshTokenFn func(ctx context.Context, refreshTokenHash string, userID uuid.UUID, expires time.Time) (model.RefreshToken, error)
	GetRefreshTokenFn       func(ctx context.Context, refreshTokenHash string) (model.RefreshToken, error)
	DeleteRefreshTokenFn    func(ctx context.Context, refreshTokenHash string) error

	// one-time codes
	CreateOneTimeCodeFn  func(ctx context.Context, code string, userID uuid.UUID) error
	ConsumeOneTimeCodeFn func(ctx context.Context, code string) (uuid.UUID, error)
}

func (m *MockStore) GetSystemGames(ctx context.Context) ([]model.Game, error) {
	return m.GetSystemGamesFn(ctx)
}

func (m *MockStore) GetGameRanks(ctx context.Context, gameID *uuid.UUID) ([]model.GameRank, error) {
	return m.GetGameRanksFn(ctx, gameID)
}

func (m *MockStore) CreateNewUser(ctx context.Context, discordUser model.DiscordUser) (model.User, error) {
	return m.CreateNewUserFn(ctx, discordUser)
}

func (m *MockStore) UpdateUserFromLogin(ctx context.Context, userID uuid.UUID, discordUser model.DiscordUser) (model.User, error) {
	return m.UpdateUserFromLoginFn(ctx, userID, discordUser)
}

func (m *MockStore) CreateNewRefreshToken(ctx context.Context, refreshTokenHash string, userID uuid.UUID, expires time.Time) (model.RefreshToken, error) {
	return m.CreateNewRefreshTokenFn(ctx, refreshTokenHash, userID, expires)
}

func (m *MockStore) GetRefreshToken(ctx context.Context, refreshTokenHash string) (model.RefreshToken, error) {
	return m.GetRefreshTokenFn(ctx, refreshTokenHash)
}

func (m *MockStore) DeleteRefreshToken(ctx context.Context, refreshTokenHash string) error {
	return m.DeleteRefreshTokenFn(ctx, refreshTokenHash)
}

func (m *MockStore) CreateOneTimeCode(ctx context.Context, code string, userID uuid.UUID) error {
	return m.CreateOneTimeCodeFn(ctx, code, userID)
}

func (m *MockStore) ConsumeOneTimeCode(ctx context.Context, code string) (uuid.UUID, error) {
	return m.ConsumeOneTimeCodeFn(ctx, code)
}
