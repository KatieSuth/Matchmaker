// Command server is the HTTP API for Matchmaker: loads configuration from the environment,
// optionally runs database migrations, wires PostgreSQL and Discord OAuth, and serves Gin routes
// (public auth, JWT-protected REST handlers, CORS, and structured request logging).
//
// Subcommands:
//
//	health       — local HTTP probe for container HEALTHCHECK
//	migrate      — run goose migrations and exit (used in production via Cloud Run Job)
//	db-bootstrap — create least-privilege DB roles (postgres admin; production cutover)
package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // embed IANA zones; scratch image has no /usr/share/zoneinfo

	"github.com/KatieSuth/MatchmakerAPI/internal/apilink"
	"github.com/KatieSuth/MatchmakerAPI/internal/cryptoutil"
	"github.com/KatieSuth/MatchmakerAPI/internal/discord"
	"github.com/KatieSuth/MatchmakerAPI/internal/handler"
	"github.com/KatieSuth/MatchmakerAPI/internal/logger"
	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/KatieSuth/MatchmakerAPI/internal/middleware"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"

	"golang.org/x/oauth2"
)

func fatalExit(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}

// runHealthCheck performs an in-process HTTP probe of the local /health endpoint
// and exits 0 on success or 1 on failure.
func runHealthCheck() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + "/health")
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}

// autoMigrateEnabled is true unless AUTO_MIGRATE is explicitly false/0/no/off.
// Local Compose leaves it unset (migrate on API start). Production Cloud Run sets AUTO_MIGRATE=false
// and runs migrations via `server migrate` (Cloud Run Job) instead.
func autoMigrateEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("AUTO_MIGRATE")))
	switch v {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func runGooseUp(sqlDB *sql.DB) {
	if err := goose.SetDialect("postgres"); err != nil {
		fatalExit("failed to set goose dialect", "error", err)
	}
	goose.SetBaseFS(nil) // use OS filesystem
	if err := goose.Up(sqlDB, "sql/migrations"); err != nil {
		fatalExit("failed to run migrations", "error", err)
	}
}

// runMigrateOnly connects, applies migrations, and exits (production deploy path).
func runMigrateOnly() {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fatalExit("DATABASE_URL is required")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		fatalExit("db connect failed", "error", err)
	}
	defer pool.Close()

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	runGooseUp(sqlDB)
	slog.Info("migrations applied successfully")
}

// configureLogger sets process-wide slog defaults from GIN_MODE.
// release mode uses production-safe verbosity; all other modes use dev/test verbosity.
func configureLogger(ginEnv string) {
	options := &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}
	if ginEnv == gin.ReleaseMode {
		options = &slog.HandlerOptions{Level: slog.LevelInfo}
	}

	base := slog.NewJSONHandler(os.Stdout, options)
	slog.SetDefault(slog.New(logger.New(base)))
}

func main() {
	// Lightweight health probe used by the container HEALTHCHECK
	if len(os.Args) > 1 && os.Args[1] == "health" {
		runHealthCheck()
		return
	}

	ginEnv := os.Getenv("GIN_MODE")
	configureLogger(ginEnv)

	// Best-effort local env loading for development convenience.
	// Existing process environment variables still take precedence.
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrateOnly()
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "db-bootstrap" {
		runDBBootstrap()
		return
	}

	switch ginEnv {
	case gin.ReleaseMode:
		gin.SetMode(gin.ReleaseMode)
	case gin.TestMode:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	/*** ENVIRONMENT VARIABLES ***/
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Database
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fatalExit("DATABASE_URL is required")
	}

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		fatalExit("invalid DATABASE_URL", "error", err)
	}
	if v := os.Getenv("DB_MAX_CONNS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			fatalExit("invalid DB_MAX_CONNS", "error", err)
		}
		poolCfg.MaxConns = int32(n)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		fatalExit("db connect failed", "error", err)
	}
	defer pool.Close()

	// wrap the existing pgxpool in a *sql.DB interface without opening a second connection
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	// Local/dev: migrate on start. Production: AUTO_MIGRATE=false; use `server migrate` Job.
	if autoMigrateEnabled() {
		runGooseUp(sqlDB)
	} else {
		slog.Info("skipping auto-migrate (AUTO_MIGRATE disabled)")
	}

	s := store.NewPostgresStore(pool)

	// discord OAuth2 setup
	cookieHashKey := os.Getenv("COOKIE_HASH_KEY")
	cookieEncryptKey := os.Getenv("COOKIE_ENCRYPT_KEY")

	if cookieHashKey == "" {
		fatalExit("COOKIE_HASH_KEY is required")
	}

	hashKeyBytes, err := hex.DecodeString(cookieHashKey)
	if err != nil {
		fatalExit("invalid COOKIE_HASH_KEY", "error", err)
	}

	if cookieEncryptKey == "" {
		fatalExit("COOKIE_ENCRYPT_KEY is required")
	}
	encryptKeyBytes, err := hex.DecodeString(cookieEncryptKey)
	if err != nil {
		fatalExit("invalid COOKIE_ENCRYPT_KEY", "error", err)
	}

	sc := securecookie.New(
		[]byte(hashKeyBytes),
		[]byte(encryptKeyBytes),
	)

	discordClientID := os.Getenv("DISCORD_CLIENT_ID")
	discordClientSecret := os.Getenv("DISCORD_CLIENT_SECRET")
	discordRedirectURI := os.Getenv("DISCORD_REDIRECT_URI")
	discordAPIURL := os.Getenv("DISCORD_API_URL")

	if discordClientID == "" {
		fatalExit("DISCORD_CLIENT_ID is required")
	}

	if discordClientSecret == "" {
		fatalExit("DISCORD_CLIENT_SECRET is required")
	}

	if discordRedirectURI == "" {
		discordRedirectURI = "https://matchmaker.localhost/api/auth/discord_redirect"
	}

	if discordAPIURL == "" {
		discordAPIURL = "https://discord.com/api"
	}

	var discordOauth = &oauth2.Config{
		ClientID:     discordClientID,
		ClientSecret: discordClientSecret,
		RedirectURL:  discordRedirectURI,
		Scopes:       []string{"identify", "guilds"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}

	// jwt setup
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		fatalExit("JWT_SECRET is required")
	}
	jwtSecretBytes, err := hex.DecodeString(jwtSecret)
	if err != nil {
		fatalExit("invalid JWT_SECRET", "error", err)
	}

	refreshExpiration := os.Getenv("REFRESH_EXPIRE_LIMIT")
	var refreshInt int
	if refreshExpiration == "" {
		refreshInt = 604800 //7 days (7 * 24 * 60 * 60)
	} else {
		refreshInt, err = strconv.Atoi(refreshExpiration)
		if err != nil {
			fatalExit("invalid refresh expiration--must be integer value", "error", err)
		}
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://matchmaker.localhost"
	}

	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	if cookieDomain == "" {
		cookieDomain = "matchmaker.localhost"
	}

	fairnessOutlierGap := 6
	if v := os.Getenv("FAIRNESS_OUTLIER_GAP"); v != "" {
		fairnessOutlierGap, err = strconv.Atoi(v)
		if err != nil || fairnessOutlierGap <= 0 {
			fatalExit("invalid FAIRNESS_OUTLIER_GAP", "error", err)
		}
	}

	fairnessTeamSeparation := 3.0
	if v := os.Getenv("FAIRNESS_TEAM_SEPARATION"); v != "" {
		fairnessTeamSeparation, err = strconv.ParseFloat(v, 64)
		if err != nil || fairnessTeamSeparation <= 0 {
			fatalExit("invalid FAIRNESS_TEAM_SEPARATION", "error", err)
		}
	}

	fairnessReferenceTierCount := 25
	if v := os.Getenv("FAIRNESS_REFERENCE_TIER_COUNT"); v != "" {
		fairnessReferenceTierCount, err = strconv.Atoi(v)
		if err != nil || fairnessReferenceTierCount <= 0 {
			fatalExit("invalid FAIRNESS_REFERENCE_TIER_COUNT", "error", err)
		}
	}

	mmSettings := matchmaking.Settings{
		FairnessOutlierGap:         fairnessOutlierGap,
		FairnessTeamSeparation:     fairnessTeamSeparation,
		FairnessReferenceTierCount: fairnessReferenceTierCount,
	}

	apiLinkEncryptionKey := os.Getenv("API_LINK_ENCRYPTION_KEY")
	if apiLinkEncryptionKey == "" {
		fatalExit("API_LINK_ENCRYPTION_KEY is required")
	}
	apiLinkKeyBytes, err := cryptoutil.ParseAES256Key(apiLinkEncryptionKey)
	if err != nil {
		fatalExit("invalid API_LINK_ENCRYPTION_KEY", "error", err)
	}
	apiLinkKeyID := os.Getenv("API_LINK_ENCRYPTION_KEY_ID")
	if apiLinkKeyID == "" {
		apiLinkKeyID = apilink.DefaultKeyID
	}
	previousKeys, err := apilink.ParsePreviousKeys(os.Getenv("API_LINK_ENCRYPTION_PREVIOUS_KEYS"))
	if err != nil {
		fatalExit("invalid API_LINK_ENCRYPTION_PREVIOUS_KEYS", "error", err)
	}
	apiLinkKeys, err := apilink.NewKeyring(apiLinkKeyID, apiLinkKeyBytes, previousKeys)
	if err != nil {
		fatalExit("invalid API_LINK encryption keyring", "error", err)
	}

	// Optional: reject direct origin traffic when set (Cloudflare Worker injects the header).
	originVerifySecret := os.Getenv("ORIGIN_VERIFY_SECRET")

	trustedProxies := []string{"172.20.0.0/16"}
	if v := os.Getenv("TRUSTED_PROXIES"); v != "" {
		parts := strings.Split(v, ",")
		trustedProxies = make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				trustedProxies = append(trustedProxies, p)
			}
		}
	}

	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              dsn,
			Environment:      ginEnv,
			TracesSampleRate: 0,
		}); err != nil {
			fatalExit("sentry init failed", "error", err)
		}
		defer sentry.Flush(2 * time.Second)
	}

	// Handlers
	discordClient := discord.New(s, apilink.New(apiLinkKeys, s), discordOauth, discordAPIURL, nil)
	h := handler.New(ginEnv, s, sc, discordOauth, cookieDomain, frontendURL, jwtSecretBytes, refreshInt, discordAPIURL, mmSettings, apiLinkKeys, discordClient)

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.Recovery(), middleware.RequestLogger())
	r.Use(middleware.OriginVerify(originVerifySecret))
	if os.Getenv("SENTRY_DSN") != "" {
		r.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
	}
	if err := r.SetTrustedProxies(trustedProxies); err != nil {
		fatalExit("invalid TRUSTED_PROXIES", "error", err)
	}

	// CORS — allow the Next.js origin
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Origin-Verify"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// ── Routes ─────────────────────────────────────────────────
	r.GET("/health", h.Health)

	//public routes
	auth := r.Group("/auth")
	{
		auth.POST("/complete", h.CompleteAuthHandler)
		auth.GET("/discord_redirect", h.DiscordCallbackHandler)
		auth.GET("/login", h.LoginHandler)
		auth.POST("/refresh", h.RefreshHandler)
		auth.POST("/logout", h.LogoutHandler)
	}

	protected := r.Group("/")
	protected.Use(middleware.Auth(jwtSecretBytes))
	{

		users := protected.Group("/users")
		{
			users.GET("/me", h.UsersMeHandler)
			users.PUT("/me", h.UpdateUsersMeHandler)
			users.GET("/me/games", h.UsersMeGamesHandler)
			users.PUT("/me/games/:gameId", h.UpsertUsersMeGameHandler)
			users.DELETE("/me/games/:gameId", h.DeleteUsersMeGameHandler)
			users.GET("/me/events", h.UsersMeEventsHandler)
			users.GET("/me/discord/guilds", h.ListMyDiscordGuildsHandler)
		}

		games := protected.Group("/games")
		{
			games.GET("", h.GetSystemGamesHandler)
			games.GET("/users/:ownerId", h.GetUserGamesHandler)
			games.GET("/:gameId/modes", h.GetGameModesByGame)
			games.GET("/:gameId/ranks", h.GetGameRanksByGame)
		}

		events := protected.Group("/events")
		{
			events.POST("", h.CreateEventHandler)
			events.GET("/:groupId/access", h.GetEventGroupAccessHandler)
			events.GET("/:groupId", h.GetEventGroupHandler)
			events.PATCH("/:groupId", h.UpdateEventGroupSettingsHandler)
			events.DELETE("/:groupId", h.DeleteEventGroupHandler)
			events.PATCH("/:groupId/registration", h.UpdateEventGroupRegistrationStatusHandler)
			events.POST("/:groupId/teams", h.CreateTeamsHandler)
			events.DELETE("/:groupId/teams", h.DeleteTeamsHandler)
		}

		registrations := protected.Group("/registrations")
		{
			registrations.PUT("/group/:groupId/me", h.UpsertMyGroupRegistrationsHandler)
			registrations.PUT("/:eventId/me", h.UpsertMyRegistrationHandler)
			registrations.POST("/:eventId/player-swap", h.SwapPlayersHandler)
			registrations.POST("/:eventId/sub-to-unplaced", h.MoveSubToUnplacedHandler)
			registrations.POST("/:eventId/unplaced-to-subs", h.MoveUnplacedToSubsHandler)
			registrations.POST("/:eventId/lobby-host", h.SetLobbyHostHandler)
			registrations.DELETE("/:eventId/:userId", h.DeleteRegistrationHandler)
			registrations.DELETE("/:eventId/me", h.DeleteRegistrationHandler)
		}

		lobbies := protected.Group("/lobbies")
		{
			lobbies.PATCH("/:lobbyId/join-code", h.UpdateLobbyJoinCodeHandler)
		}
	}

	slog.Info("API listening", "port", port)
	if err := r.Run(":" + port); err != nil {
		fatalExit("server error", "error", err)
	}
}
