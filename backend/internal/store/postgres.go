package store

import (
	"context"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	q *db.Queries
}

type Store interface {
	//users
	GetUserByDiscordID(ctx context.Context, discordId string, errorOnNoRows bool) (model.User, error)
	CreateNewUser(ctx context.Context, discordUser model.DiscordUser) (model.User, error)
	UpdateUserFromLogin(ctx context.Context, userId uuid.UUID, discordUser model.DiscordUser) (model.User, error)
	GetUserByUserID(ctx context.Context, userID uuid.UUID) (model.User, error)

	//refresh tokens
	CreateNewRefreshToken(ctx context.Context, refreshTokenHash string, userID uuid.UUID, expires time.Time) (model.RefreshToken, error)
	GetRefreshToken(ctx context.Context, refreshTokenHash string) (model.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, refreshTokenHash string) error

	//one-time codes
	CreateOneTimeCode(ctx context.Context, code string, userID uuid.UUID) error
	ConsumeOneTimeCode(ctx context.Context, code string) (uuid.UUID, error)
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{q: db.New(pool)}
}
