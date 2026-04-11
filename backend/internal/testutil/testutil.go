package testutil

import (
	"context"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func LoadEnv(t *testing.T) {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "../..") // goes to backend/
	err := godotenv.Load(filepath.Join(dir, ".env"))
	if err != nil {
		t.Fatalf("failed to load .env: %v", err)
	}
}

func WithTestTx(t *testing.T, fn func(q *db.Queries, s *store.PostgresStore)) {
	LoadEnv(t)
	dsn := os.Getenv("DATABASE_URL_TESTS")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	t.Helper()
	tx, err := pool.Begin(t.Context())
	if err != nil {
		t.Fatalf("could not begin db transaction: %v", err)
	}
	defer tx.Rollback(context.Background())

	store := store.NewPostgresStoreFromTx(tx)
	fn(db.New(tx), store)
}

func GetJWTSecret(t *testing.T) ([]byte, error) {
	LoadEnv(t)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return []byte{}, errors.New("JWT_SECRET is required from .env")
	}
	jwtSecretBytes, err := hex.DecodeString(jwtSecret)
	if err != nil {
		return []byte{}, errors.New("invalid JWT_SECRET from .env")
	}

	return jwtSecretBytes, nil
}
