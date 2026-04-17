package test_util

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

var migrationOnce sync.Once

func runMigrations(pool *pgxpool.Pool) {
	migrationOnce.Do(func() {
		sqlDB := stdlib.OpenDBFromPool(pool)
		// Don't defer sqlDB.Close() here because it would close the underlying pool that the tests need.

		if err := goose.SetDialect("postgres"); err != nil {
			log.Fatalf("goose dialect: %v", err)
		}

		_, filename, _, _ := runtime.Caller(0)
		if err := goose.Up(sqlDB, filepath.Join(filepath.Dir(filename), "../../sql/migrations")); err != nil {
			log.Fatalf("migrations up: %v", err)
		}
	})
}

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

	// Run migrations once per suite execution
	runMigrations(pool)

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

func GetSecureCookie(t *testing.T) (*securecookie.SecureCookie, error) {
	LoadEnv(t)

	var sc *securecookie.SecureCookie

	cookieHashKey := os.Getenv("COOKIE_HASH_KEY")
	cookieEncryptKey := os.Getenv("COOKIE_ENCRYPT_KEY")

	if cookieHashKey == "" {
		return sc, errors.New("COOKIE_HASH_KEY is required from .env")
	}
	hashKeyBytes, err := hex.DecodeString(cookieHashKey)
	if err != nil {
		return sc, errors.New("invalid COOKIE_HASH_KEY from .env")
	}

	if cookieEncryptKey == "" {
		return sc, errors.New("COOKIE_ENCRYPT_KEY is required from .env")
	}
	encryptKeyBytes, err := hex.DecodeString(cookieEncryptKey)
	if err != nil {
		return sc, errors.New("invalid COOKIE_ENCRYPT_KEY from .env")
	}

	sc = securecookie.New(
		[]byte(hashKeyBytes),
		[]byte(encryptKeyBytes),
	)

	return sc, nil
}

// NewGinContext creates a minimal Gin context backed by a ResponseRecorder.
func NewGinContext(method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)
	return c, w
}

// WithUserID stamps a userID into the Gin context, simulating a passing auth middleware.
func WithUserID(c *gin.Context, id uuid.UUID) {
	c.Set("userID", id)
}

// WithUserIDString stamps a userID into the Gin context as a string.
// Use this for handlers that cast userID with userID.(string).
func WithUserIDString(c *gin.Context, id uuid.UUID) {
	c.Set("userID", id.String())
}

// SetCookie plants a named cookie on the Gin test request.
func SetCookie(c *gin.Context, name, value string) {
	c.Request.AddCookie(&http.Cookie{Name: name, Value: value})
}

// decodeJSON is a small helper to unmarshal a response body in tests.
func DecodeJSON[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var out T
	require.NoError(t, json.NewDecoder(w.Body).Decode(&out))
	return out
}
