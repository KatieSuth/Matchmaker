// Package store is the data access layer: PostgreSQL via sqlc-generated queries, transactions,
// and store-level validation errors that handlers map to HTTP status codes.
package store

import (
	"context"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	q    *db.Queries
	pool *pgxpool.Pool // store the pool directly so we can begin transactions
}

// Store abstracts persistence for handlers and allows transactional tests to swap in a
// transaction-backed store via WithTx and NewPostgresStoreFromTx.
type Store interface {
	WithTx(ctx context.Context, fn func(Store) error) error

	//users
	GetUserByDiscordID(ctx context.Context, discordId string, errorOnNoRows bool) (model.User, error)
	CreateNewUser(ctx context.Context, discordUser model.DiscordUser) (model.User, error)
	UpdateUserFromLogin(ctx context.Context, userId uuid.UUID, discordUser model.DiscordUser) (model.User, error)
	GetUserByUserID(ctx context.Context, userID uuid.UUID) (model.User, error)
	UpdateUser(ctx context.Context, userId uuid.UUID, pronouns *string, showPronous bool, region *string) (model.User, error)

	//refresh tokens
	CreateNewRefreshToken(ctx context.Context, refreshTokenHash string, userID uuid.UUID, expires time.Time) (model.RefreshToken, error)
	GetRefreshToken(ctx context.Context, refreshTokenHash string) (model.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, refreshTokenHash string) error

	//one-time codes
	CreateOneTimeCode(ctx context.Context, code string, userID uuid.UUID) error
	ConsumeOneTimeCode(ctx context.Context, code string) (uuid.UUID, error)

	//user's games
	GetUserGamesForUser(ctx context.Context, userID uuid.UUID) ([]model.UserGame, error)
	UpsertGameForUser(ctx context.Context, userID uuid.UUID, ug model.UserGame) (model.UserGame, error)

	//games
	GetSystemGames(ctx context.Context) ([]model.Game, error)
	GetUserGames(ctx context.Context, ownerID *uuid.UUID) ([]model.Game, error)
	GetGameModes(ctx context.Context, gameID uuid.UUID) ([]model.GameMode, error)
	GetGameModeByID(ctx context.Context, gameModeID uuid.UUID) (model.GameMode, error)

	//game ranks
	GetGameRanks(ctx context.Context, gameID *uuid.UUID) ([]model.GameRank, error)

	//events
	GetEventsForUser(ctx context.Context, userID uuid.UUID, hosting, past bool, from, to *time.Time, gameId, cursor, timezone string) ([]model.DashboardEvent, bool, string, error)
	CreateEventGroupWithEvents(ctx context.Context, userID, gameModeID uuid.UUID, subMin int32, registrationOpen bool, region string, sortLogic string, startTime time.Time, gamesToRun int32) (uuid.UUID, error)
	GetEventGroupDetail(ctx context.Context, groupID, viewerID uuid.UUID) (model.EventGroupDetail, error)
	UpdateEventGroupSettings(ctx context.Context, groupID, ownerID uuid.UUID, region string, subMin int32, sortLogic string, registrationOpen bool) error
	DeleteEventGroup(ctx context.Context, groupID, ownerID uuid.UUID) error
	SetEventGroupRegistrationOpen(ctx context.Context, groupID, ownerID uuid.UUID, open bool) error
	CreateTeamsForGroup(ctx context.Context, groupID, ownerID uuid.UUID) error
	DeleteTeamsForGroup(ctx context.Context, groupID, ownerID uuid.UUID) error
	UpsertRegistrationForEvent(ctx context.Context, eventID, userID uuid.UUID, canSubstitute, canLobbyHost bool, duoRequest *string) error
	UpsertRegistrationsForGroup(ctx context.Context, groupID, userID uuid.UUID, registrations []RegistrationUpsertItem, duoRequest *string) error
	DeleteRegistrationForEvent(ctx context.Context, eventID, targetUserID, actorUserID uuid.UUID) error
}

// NewPostgresStore wires a connection pool. Callers are responsible for pool lifecycle.
func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{
		q:    db.New(pool),
		pool: pool,
	}
}

func (s *PostgresStore) WithTx(ctx context.Context, fn func(Store) error) error {
	// If this store is already transaction-backed (tests commonly construct it via
	// NewPostgresStoreFromTx), execute inline so outer transaction controls commit/rollback.
	if s.pool == nil {
		return fn(s)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // no-op if already committed

	txStore := &PostgresStore{
		q:    db.New(tx),
		pool: s.pool,
	}

	if err := fn(txStore); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// For use in tests only — no pool, so WithTx executes inline against this transaction.
func NewPostgresStoreFromTx(tx pgx.Tx) *PostgresStore {
	return &PostgresStore{
		q: db.New(tx),
	}
}
