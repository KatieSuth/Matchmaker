// Command server is the HTTP API for Matchmaker: loads configuration from the environment,
// runs database migrations, wires PostgreSQL and Discord OAuth, and serves Gin routes
// (public auth, JWT-protected REST handlers, CORS, and structured request logging).
package main

import (
	"context"
	"encoding/hex"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/handler"
	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/KatieSuth/MatchmakerAPI/internal/logger"
	"github.com/KatieSuth/MatchmakerAPI/internal/middleware"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
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

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		fatalExit("db connect failed", "error", err)
	}
	defer pool.Close()

	// Run migrations

	// wrap the existing pgxpool in a *sql.DB interface without opening a second connection
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		fatalExit("failed to set goose dialect", "error", err)
	}

	goose.SetBaseFS(nil) // use OS filesystem
	if err := goose.Up(sqlDB, "sql/migrations"); err != nil {
		fatalExit("failed to run migrations", "error", err)
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

	// Handlers
	h := handler.New(ginEnv, s, sc, discordOauth, cookieDomain, frontendURL, jwtSecretBytes, refreshInt, discordAPIURL, mmSettings)

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.Recovery(), middleware.RequestLogger())
	r.SetTrustedProxies([]string{"172.20.0.0/16"})

	// CORS — allow the Next.js origin
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
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
			users.GET("/me/events", h.UsersMeEventsHandler)
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
	}

	slog.Info("API listening", "port", port)
	if err := r.Run(":" + port); err != nil {
		fatalExit("server error", "error", err)
	}
}
