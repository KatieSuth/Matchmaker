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
		// Attribute flags only on CREATE — Cloud SQL rejects many of them on ALTER.
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
	// Existing role (e.g. after a partial bootstrap): password + LOGIN only.
	_, err := pool.Exec(ctx, fmt.Sprintf(
		`ALTER ROLE %s WITH LOGIN PASSWORD %s`,
		role, pw,
	))
	if err != nil {
		return fmt.Errorf("alter role %s: %w", role, err)
	}
	slog.Info("updated login role password", "role", role)
	return nil
}

func roleExists(ctx context.Context, pool *pgxpool.Pool, role string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)`, role).Scan(&exists)
	return exists, err
}

func execOrFatal(ctx context.Context, pool *pgxpool.Pool, msg, sql string) {
	if _, err := pool.Exec(ctx, sql); err != nil {
		fatalExit(msg, "sql", sql, "error", err)
	}
}

// poolAsUser clones the admin DSN with a different login (same host/db).
func poolAsUser(adminDSN, user, password string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(adminDSN)
	if err != nil {
		return nil, err
	}
	cfg.ConnConfig.User = user
	cfg.ConnConfig.Password = password
	return pgxpool.NewWithConfig(context.Background(), cfg)
}

// runDBBootstrap creates least-privilege matchmaker_app / matchmaker_migrator roles,
// grants, and optionally reassigns ownership from the legacy matchmaker user.
// Requires DATABASE_URL as postgres (admin), plus DB_APP_PASSWORD and DB_MIGRATOR_PASSWORD.
//
// Cloud SQL notes:
//   - postgres is cloudsqlsuperuser, not a true SUPERUSER.
//   - REASSIGN OWNED requires membership in both old and new roles; run it as legacy
//     matchmaker after GRANT matchmaker_migrator TO matchmaker.
//   - Table GRANTs run as matchmaker_migrator (object owner after reassign).
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

	// Database / schema grants (do not require table ownership).
	execOrFatal(ctx, pool, "grant failed",
		fmt.Sprintf(`GRANT CONNECT, TEMPORARY ON DATABASE %s TO %s, %s`,
			dbBootstrapDatabase, dbBootstrapMigratorRole, dbBootstrapAppRole))
	execOrFatal(ctx, pool, "grant failed",
		fmt.Sprintf(`GRANT USAGE, CREATE ON SCHEMA public TO %s`, dbBootstrapMigratorRole))
	execOrFatal(ctx, pool, "grant failed",
		fmt.Sprintf(`GRANT USAGE ON SCHEMA public TO %s`, dbBootstrapAppRole))

	legacyExists, err := roleExists(ctx, pool, dbBootstrapLegacyRole)
	if err != nil {
		fatalExit("check legacy role", "error", err)
	}
	if legacyExists {
		// Let legacy session reassign: must be a member of the destination role.
		execOrFatal(ctx, pool, "grant migrator to legacy failed",
			fmt.Sprintf(`GRANT %s TO %s`, dbBootstrapMigratorRole, dbBootstrapLegacyRole))

		legacyPool, err := poolAsUser(dsn, dbBootstrapLegacyRole, appPassword)
		if err != nil {
			fatalExit("legacy db connect failed", "error", err)
		}
		execOrFatal(ctx, legacyPool, "reassign owned failed",
			fmt.Sprintf(`REASSIGN OWNED BY %s TO %s`, dbBootstrapLegacyRole, dbBootstrapMigratorRole))
		legacyPool.Close()
		slog.Info("reassigned owned objects from legacy role",
			"from", dbBootstrapLegacyRole, "to", dbBootstrapMigratorRole)
	}

	// Table/sequence grants and default privileges as the owning role.
	migratorPool, err := poolAsUser(dsn, dbBootstrapMigratorRole, migratorPassword)
	if err != nil {
		fatalExit("migrator db connect failed", "error", err)
	}
	defer migratorPool.Close()

	tableGrants := []string{
		fmt.Sprintf(`GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO %s`, dbBootstrapMigratorRole),
		fmt.Sprintf(`GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO %s`, dbBootstrapMigratorRole),
		fmt.Sprintf(`GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO %s`, dbBootstrapAppRole),
		fmt.Sprintf(`GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO %s`, dbBootstrapAppRole),
		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES FOR ROLE %s IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO %s`,
			dbBootstrapMigratorRole, dbBootstrapAppRole),
		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES FOR ROLE %s IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO %s`,
			dbBootstrapMigratorRole, dbBootstrapAppRole),
	}
	for _, stmt := range tableGrants {
		execOrFatal(ctx, migratorPool, "grant failed", stmt)
	}

	if legacyExists {
		// Do not DROP ROLE here — Terraform removes the Cloud SQL user on Apply B.
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
