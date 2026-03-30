package store

import (
	"context"

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
	GetUserByDiscordID(ctx context.Context, discordId string, errorOnNoRows bool) (model.User, error)
	CreateNewUser(ctx context.Context, discordUser model.DiscordUser) (model.User, error)
	UpdateUserFromLogin(ctx context.Context, userId uuid.UUID, discordUser model.DiscordUser) (model.User, error)
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{q: db.New(pool)}
}
