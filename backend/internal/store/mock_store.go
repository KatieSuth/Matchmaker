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

	// transactions
	WithTxFn func(ctx context.Context, fn func(Store) error) error

	// games
	GetSystemGamesFn func(ctx context.Context) ([]model.Game, error)
	GetGameRanksFn   func(ctx context.Context, gameID *uuid.UUID) ([]model.GameRank, error)
	GetUserGamesFn   func(ctx context.Context, ownerID *uuid.UUID) ([]model.Game, error)
	GetGameModesFn   func(ctx context.Context, gameID uuid.UUID) ([]model.GameMode, error)
	GetGameModeByIDFn func(ctx context.Context, gameModeID uuid.UUID) (model.GameMode, error)

	// user
	GetUserByDiscordIDFn  func(ctx context.Context, discordID string, errorOnNoRows bool) (model.User, error)
	CreateNewUserFn       func(ctx context.Context, discordUser model.DiscordUser) (model.User, error)
	UpdateUserFromLoginFn func(ctx context.Context, userID uuid.UUID, discordUser model.DiscordUser) (model.User, error)
	GetUserByUserIDFn     func(ctx context.Context, userID uuid.UUID) (model.User, error)
	UpdateUserFn          func(ctx context.Context, userID uuid.UUID, pronouns *string, showPronouns bool, region *string) (model.User, error)

	// user games
	GetUserGamesForUserFn func(ctx context.Context, userID uuid.UUID) ([]model.UserGame, error)
	UpsertGameForUserFn   func(ctx context.Context, userID uuid.UUID, ug model.UserGame) (model.UserGame, int, error)

	// refresh tokens
	CreateNewRefreshTokenFn func(ctx context.Context, refreshTokenHash string, userID uuid.UUID, expires time.Time) (model.RefreshToken, error)
	GetRefreshTokenFn       func(ctx context.Context, refreshTokenHash string) (model.RefreshToken, error)
	DeleteRefreshTokenFn    func(ctx context.Context, refreshTokenHash string) error

	// one-time codes
	CreateOneTimeCodeFn  func(ctx context.Context, code string, userID uuid.UUID) error
	ConsumeOneTimeCodeFn func(ctx context.Context, code string) (uuid.UUID, error)

	// events
	CreateEventGroupWithEventsFn func(ctx context.Context, userID, gameModeID uuid.UUID, subMin int32, registrationOpen bool, region string, startTime time.Time, gamesToRun int32) (uuid.UUID, error)
}

func (m *MockStore) WithTx(ctx context.Context, fn func(Store) error) error {
	return m.WithTxFn(ctx, fn)
}

func (m *MockStore) GetSystemGames(ctx context.Context) ([]model.Game, error) {
	return m.GetSystemGamesFn(ctx)
}

func (m *MockStore) GetGameRanks(ctx context.Context, gameID *uuid.UUID) ([]model.GameRank, error) {
	return m.GetGameRanksFn(ctx, gameID)
}

func (m *MockStore) GetUserGames(ctx context.Context, ownerID *uuid.UUID) ([]model.Game, error) {
	return m.GetUserGamesFn(ctx, ownerID)
}

func (m *MockStore) GetGameModes(ctx context.Context, gameID uuid.UUID) ([]model.GameMode, error) {
	return m.GetGameModesFn(ctx, gameID)
}

func (m *MockStore) GetGameModeByID(ctx context.Context, gameModeID uuid.UUID) (model.GameMode, error) {
	return m.GetGameModeByIDFn(ctx, gameModeID)
}

func (m *MockStore) GetUserByDiscordID(ctx context.Context, discordID string, errorOnNoRows bool) (model.User, error) {
	return m.GetUserByDiscordIDFn(ctx, discordID, errorOnNoRows)
}

func (m *MockStore) CreateNewUser(ctx context.Context, discordUser model.DiscordUser) (model.User, error) {
	return m.CreateNewUserFn(ctx, discordUser)
}

func (m *MockStore) UpdateUserFromLogin(ctx context.Context, userID uuid.UUID, discordUser model.DiscordUser) (model.User, error) {
	return m.UpdateUserFromLoginFn(ctx, userID, discordUser)
}

func (m *MockStore) GetUserByUserID(ctx context.Context, userID uuid.UUID) (model.User, error) {
	return m.GetUserByUserIDFn(ctx, userID)
}

func (m *MockStore) UpdateUser(ctx context.Context, userID uuid.UUID, pronouns *string, showPronouns bool, region *string) (model.User, error) {
	return m.UpdateUserFn(ctx, userID, pronouns, showPronouns, region)
}

func (m *MockStore) GetUserGamesForUser(ctx context.Context, userID uuid.UUID) ([]model.UserGame, error) {
	return m.GetUserGamesForUserFn(ctx, userID)
}

func (m *MockStore) UpsertGameForUser(ctx context.Context, userID uuid.UUID, ug model.UserGame) (model.UserGame, int, error) {
	return m.UpsertGameForUserFn(ctx, userID, ug)
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

func (m *MockStore) CreateEventGroupWithEvents(ctx context.Context, userID, gameModeID uuid.UUID, subMin int32, registrationOpen bool, region string, startTime time.Time, gamesToRun int32) (uuid.UUID, error) {
	return m.CreateEventGroupWithEventsFn(ctx, userID, gameModeID, subMin, registrationOpen, region, startTime, gamesToRun)
}
