package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	dbBootstrapAppRole      = "matchmaker_app"
	dbBootstrapMigratorRole = "matchmaker_migrator"
	dbBootstrapLegacyRole   = "matchmaker"
	dbBootstrapDatabase     = "matchmaker"
)

// quoteLiteral escapes a string for use as a PostgreSQL string literal.
func quoteLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func ensureLoginRole(ctx context.Context, pool *pgxpool.Pool, role, password string) error {
	var exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)`, role).Scan(&exists); err != nil {
		return fmt.Errorf("check role %s: %w", role, err)
	}
	pw := quoteLiteral(password)
	if !exists {
		_, err := pool.Exec(ctx, fmt.Sprintf(
			`CREATE ROLE %s WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION PASSWORD %s`,
			role, pw,
		))
		if err != nil {
			return fmt.Errorf("create role %s: %w", role, err)
		}
		slog.Info("created login role", "role", role)
		return nil
	}
	_, err := pool.Exec(ctx, fmt.Sprintf(
		`ALTER ROLE %s WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION PASSWORD %s`,
		role, pw,
	))
	if err != nil {
		return fmt.Errorf("alter role %s: %w", role, err)
	}
	slog.Info("updated login role", "role", role)
	return nil
}

func roleExists(ctx context.Context, pool *pgxpool.Pool, role string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)`, role).Scan(&exists)
	return exists, err
}

// runDBBootstrap creates least-privilege matchmaker_app / matchmaker_migrator roles,
// grants, and optionally reassigns ownership from the legacy matchmaker user.
// Requires DATABASE_URL as postgres (admin), plus DB_APP_PASSWORD and DB_MIGRATOR_PASSWORD.
func runDBBootstrap() {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fatalExit("DATABASE_URL is required (postgres admin DSN)")
	}
	appPassword := os.Getenv("DB_APP_PASSWORD")
	if appPassword == "" {
		fatalExit("DB_APP_PASSWORD is required")
	}
	migratorPassword := os.Getenv("DB_MIGRATOR_PASSWORD")
	if migratorPassword == "" {
		fatalExit("DB_MIGRATOR_PASSWORD is required")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		fatalExit("db connect failed", "error", err)
	}
	defer pool.Close()

	ctx := context.Background()

	if err := ensureLoginRole(ctx, pool, dbBootstrapMigratorRole, migratorPassword); err != nil {
		fatalExit("migrator role", "error", err)
	}
	if err := ensureLoginRole(ctx, pool, dbBootstrapAppRole, appPassword); err != nil {
		fatalExit("app role", "error", err)
	}

	stmts := []string{
		fmt.Sprintf(`GRANT CONNECT, TEMPORARY ON DATABASE %s TO %s, %s`, dbBootstrapDatabase, dbBootstrapMigratorRole, dbBootstrapAppRole),
		fmt.Sprintf(`GRANT USAGE, CREATE ON SCHEMA public TO %s`, dbBootstrapMigratorRole),
		fmt.Sprintf(`GRANT USAGE ON SCHEMA public TO %s`, dbBootstrapAppRole),
		fmt.Sprintf(`GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO %s`, dbBootstrapMigratorRole),
		fmt.Sprintf(`GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO %s`, dbBootstrapMigratorRole),
		fmt.Sprintf(`GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO %s`, dbBootstrapAppRole),
		fmt.Sprintf(`GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO %s`, dbBootstrapAppRole),
		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES FOR ROLE %s IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO %s`, dbBootstrapMigratorRole, dbBootstrapAppRole),
		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES FOR ROLE %s IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO %s`, dbBootstrapMigratorRole, dbBootstrapAppRole),
	}

	for _, stmt := range stmts {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			fatalExit("grant failed", "sql", stmt, "error", err)
		}
	}

	legacyExists, err := roleExists(ctx, pool, dbBootstrapLegacyRole)
	if err != nil {
		fatalExit("check legacy role", "error", err)
	}
	if legacyExists {
		// Reassign table ownership so goose (migrator) can ALTER/DROP.
		if _, err := pool.Exec(ctx, fmt.Sprintf(
			`REASSIGN OWNED BY %s TO %s`, dbBootstrapLegacyRole, dbBootstrapMigratorRole,
		)); err != nil {
			fatalExit("reassign owned failed", "error", err)
		}
		slog.Info("reassigned owned objects from legacy role", "from", dbBootstrapLegacyRole, "to", dbBootstrapMigratorRole)

		// Drop leftover privileges; do not DROP ROLE here — Terraform removes the Cloud SQL user on Apply B.
		if _, err := pool.Exec(ctx, fmt.Sprintf(`REVOKE ALL ON DATABASE %s FROM %s`, dbBootstrapDatabase, dbBootstrapLegacyRole)); err != nil {
			slog.Warn("revoke database from legacy (may be ok)", "error", err)
		}
		if _, err := pool.Exec(ctx, fmt.Sprintf(`REVOKE ALL ON SCHEMA public FROM %s`, dbBootstrapLegacyRole)); err != nil {
			slog.Warn("revoke schema from legacy (may be ok)", "error", err)
		}
	}

	slog.Info("db-bootstrap completed",
		"app_role", dbBootstrapAppRole,
		"migrator_role", dbBootstrapMigratorRole,
	)
}
