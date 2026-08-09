package store

import (
	"context"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
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
	GetSystemGamesFn  func(ctx context.Context) ([]model.Game, error)
	GetGameRanksFn    func(ctx context.Context, gameID *uuid.UUID) ([]model.GameRank, error)
	GetUserGamesFn    func(ctx context.Context, ownerID *uuid.UUID) ([]model.Game, error)
	GetGameModesFn    func(ctx context.Context, gameID uuid.UUID) ([]model.GameMode, error)
	GetGameModeByIDFn func(ctx context.Context, gameModeID uuid.UUID) (model.GameMode, error)

	// user
	GetUserByDiscordIDFn  func(ctx context.Context, discordID string, errorOnNoRows bool) (model.User, error)
	CreateNewUserFn       func(ctx context.Context, discordUser model.DiscordUser) (model.User, error)
	UpdateUserFromLoginFn func(ctx context.Context, userID uuid.UUID, discordUser model.DiscordUser) (model.User, error)
	GetUserByUserIDFn     func(ctx context.Context, userID uuid.UUID) (model.User, error)
	UpdateUserFn          func(ctx context.Context, userID uuid.UUID, pronouns *string, showPronouns bool, region *string) (model.User, error)

	// user games
	GetUserGamesForUserFn func(ctx context.Context, userID uuid.UUID) ([]model.UserGame, error)
	UpsertGameForUserFn   func(ctx context.Context, userID uuid.UUID, ug model.UserGame) (model.UserGame, error)

	// refresh tokens
	CreateNewRefreshTokenFn func(ctx context.Context, refreshTokenHash string, userID uuid.UUID, expires time.Time) (model.RefreshToken, error)
	GetRefreshTokenFn       func(ctx context.Context, refreshTokenHash string) (model.RefreshToken, error)
	DeleteRefreshTokenFn    func(ctx context.Context, refreshTokenHash string) error

	// one-time codes
	CreateOneTimeCodeFn  func(ctx context.Context, code string, userID uuid.UUID) error
	ConsumeOneTimeCodeFn func(ctx context.Context, code string) (uuid.UUID, error)

	// events
	GetEventsForUserFn               func(ctx context.Context, userID uuid.UUID, hosting, past bool, from, to *time.Time, gameId, cursor, timezone string) ([]model.DashboardEvent, bool, string, error)
	CreateEventGroupWithEventsFn     func(ctx context.Context, userID, gameModeID uuid.UUID, subMin int32, registrationOpen bool, region string, sortLogic string, name string, startTime time.Time, gamesToRun int32) (uuid.UUID, error)
	GetEventGroupDetailFn            func(ctx context.Context, groupID, viewerID uuid.UUID) (model.EventGroupDetail, error)
	UpdateEventGroupSettingsFn       func(ctx context.Context, groupID, ownerID uuid.UUID, region string, subMin int32, sortLogic string, registrationOpen bool, name string, eventUpdates []GroupEventUpdate) error
	DeleteEventGroupFn               func(ctx context.Context, groupID, ownerID uuid.UUID) error
	SetEventGroupRegistrationOpenFn  func(ctx context.Context, groupID, ownerID uuid.UUID, open bool) error
	CreateTeamsForGroupFn            func(ctx context.Context, groupID, ownerID uuid.UUID, settings matchmaking.Settings) error
	DeleteTeamsForGroupFn            func(ctx context.Context, groupID, ownerID uuid.UUID) error
	SwapPlayersForEventFn            func(ctx context.Context, eventID, ownerID, userA, userB uuid.UUID, settings matchmaking.Settings) error
	MoveSubToUnplacedForEventFn      func(ctx context.Context, eventID, ownerID, userID uuid.UUID, settings matchmaking.Settings) error
	MoveUnplacedToSubsForEventFn     func(ctx context.Context, eventID, ownerID, userID, lobbyID uuid.UUID, settings matchmaking.Settings) error
	SetLobbyHostForEventFn           func(ctx context.Context, eventID, ownerID, userID uuid.UUID) error
	UpsertRegistrationForEventFn     func(ctx context.Context, eventID, userID uuid.UUID, canSubstitute, canLobbyHost bool, duoRequest *string) error
	UpsertRegistrationsForGroupFn    func(ctx context.Context, groupID, userID uuid.UUID, registrations []RegistrationUpsertItem, duoRequest *string) error
	DeleteRegistrationForEventFn     func(ctx context.Context, eventID, targetUserID, actorUserID uuid.UUID) error
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

func (m *MockStore) UpsertGameForUser(ctx context.Context, userID uuid.UUID, ug model.UserGame) (model.UserGame, error) {
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

func (m *MockStore) CreateEventGroupWithEvents(ctx context.Context, userID, gameModeID uuid.UUID, subMin int32, registrationOpen bool, region string, sortLogic string, name string, startTime time.Time, gamesToRun int32) (uuid.UUID, error) {
	return m.CreateEventGroupWithEventsFn(ctx, userID, gameModeID, subMin, registrationOpen, region, sortLogic, name, startTime, gamesToRun)
}

func (m *MockStore) GetEventsForUser(ctx context.Context, userID uuid.UUID, hosting, past bool, from, to *time.Time, gameId, cursor, timezone string) ([]model.DashboardEvent, bool, string, error) {
	return m.GetEventsForUserFn(ctx, userID, hosting, past, from, to, gameId, cursor, timezone)
}

func (m *MockStore) GetEventGroupDetail(ctx context.Context, groupID, viewerID uuid.UUID) (model.EventGroupDetail, error) {
	return m.GetEventGroupDetailFn(ctx, groupID, viewerID)
}

func (m *MockStore) UpdateEventGroupSettings(ctx context.Context, groupID, ownerID uuid.UUID, region string, subMin int32, sortLogic string, registrationOpen bool, name string, eventUpdates []GroupEventUpdate) error {
	return m.UpdateEventGroupSettingsFn(ctx, groupID, ownerID, region, subMin, sortLogic, registrationOpen, name, eventUpdates)
}

func (m *MockStore) DeleteEventGroup(ctx context.Context, groupID, ownerID uuid.UUID) error {
	return m.DeleteEventGroupFn(ctx, groupID, ownerID)
}

func (m *MockStore) SetEventGroupRegistrationOpen(ctx context.Context, groupID, ownerID uuid.UUID, open bool) error {
	return m.SetEventGroupRegistrationOpenFn(ctx, groupID, ownerID, open)
}

func (m *MockStore) CreateTeamsForGroup(ctx context.Context, groupID, ownerID uuid.UUID, settings matchmaking.Settings) error {
	return m.CreateTeamsForGroupFn(ctx, groupID, ownerID, settings)
}

func (m *MockStore) DeleteTeamsForGroup(ctx context.Context, groupID, ownerID uuid.UUID) error {
	return m.DeleteTeamsForGroupFn(ctx, groupID, ownerID)
}

func (m *MockStore) SwapPlayersForEvent(ctx context.Context, eventID, ownerID, userA, userB uuid.UUID, settings matchmaking.Settings) error {
	return m.SwapPlayersForEventFn(ctx, eventID, ownerID, userA, userB, settings)
}

func (m *MockStore) MoveSubToUnplacedForEvent(ctx context.Context, eventID, ownerID, userID uuid.UUID, settings matchmaking.Settings) error {
	return m.MoveSubToUnplacedForEventFn(ctx, eventID, ownerID, userID, settings)
}

func (m *MockStore) MoveUnplacedToSubsForEvent(ctx context.Context, eventID, ownerID, userID, lobbyID uuid.UUID, settings matchmaking.Settings) error {
	return m.MoveUnplacedToSubsForEventFn(ctx, eventID, ownerID, userID, lobbyID, settings)
}

func (m *MockStore) SetLobbyHostForEvent(ctx context.Context, eventID, ownerID, userID uuid.UUID) error {
	return m.SetLobbyHostForEventFn(ctx, eventID, ownerID, userID)
}

func (m *MockStore) UpsertRegistrationForEvent(ctx context.Context, eventID, userID uuid.UUID, canSubstitute, canLobbyHost bool, duoRequest *string) error {
	return m.UpsertRegistrationForEventFn(ctx, eventID, userID, canSubstitute, canLobbyHost, duoRequest)
}

func (m *MockStore) UpsertRegistrationsForGroup(ctx context.Context, groupID, userID uuid.UUID, registrations []RegistrationUpsertItem, duoRequest *string) error {
	return m.UpsertRegistrationsForGroupFn(ctx, groupID, userID, registrations, duoRequest)
}

func (m *MockStore) DeleteRegistrationForEvent(ctx context.Context, eventID, targetUserID, actorUserID uuid.UUID) error {
	return m.DeleteRegistrationForEventFn(ctx, eventID, targetUserID, actorUserID)
}
