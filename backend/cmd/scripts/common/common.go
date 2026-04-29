package common

import (
	"context"
	"log/slog"
	"os"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type SeedContext struct {
	Ctx     context.Context
	Pool    *pgxpool.Pool
	Queries *db.Queries
}

func NewSeedContext() *SeedContext {
	ctx := context.Background()
	// Best-effort local env loading for script convenience.
	// Existing process environment variables still take precedence.
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env.local")

	primaryDSN := os.Getenv("DATABASE_URL")
	fallbackDSN := os.Getenv("DATABASE_URL_TESTS")
	if primaryDSN == "" && fallbackDSN == "" {
		Fatal("DATABASE_URL or DATABASE_URL_TESTS is required")
	}

	var (
		pool   *pgxpool.Pool
		err    error
		source string
	)

	if primaryDSN != "" {
		pool, err = pgxpool.New(ctx, primaryDSN)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				source = "DATABASE_URL"
			} else {
				pool.Close()
				pool = nil
				err = pingErr
			}
		}
	}

	if pool == nil && fallbackDSN != "" {
		pool, err = pgxpool.New(ctx, fallbackDSN)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				source = "DATABASE_URL_TESTS"
			} else {
				pool.Close()
				pool = nil
				err = pingErr
			}
		}
	}

	if pool == nil {
		Fatal("db connect failed using DATABASE_URL and DATABASE_URL_TESTS", "error", err)
	}

	slog.Info("seed db connection established", "source_env", source)

	return &SeedContext{
		Ctx:     ctx,
		Pool:    pool,
		Queries: db.New(pool),
	}
}

func (s *SeedContext) Close() {
	if s != nil && s.Pool != nil {
		s.Pool.Close()
	}
}

func Fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
