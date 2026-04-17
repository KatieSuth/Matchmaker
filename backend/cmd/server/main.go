package main

import (
	"context"
	"encoding/hex"
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/KatieSuth/MatchmakerAPI/internal/handler"
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

func main() {
	/*** ENVIRONMENT VARIABLES ***/
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ginEnv := os.Getenv("GIN_MODE")
	if ginEnv == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Database
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	// Run migrations

	// wrap the existing pgxpool in a *sql.DB interface without opening a second connection
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose dialect: %v", err)
	}

	goose.SetBaseFS(nil) // use OS filesystem
	if err := goose.Up(sqlDB, "sql/migrations"); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	s := store.NewPostgresStore(pool)

	// discord OAuth2 setup
	cookieHashKey := os.Getenv("COOKIE_HASH_KEY")
	cookieEncryptKey := os.Getenv("COOKIE_ENCRYPT_KEY")

	if cookieHashKey == "" {
		log.Fatal("COOKIE_HASH_KEY is required")
	}

	hashKeyBytes, err := hex.DecodeString(cookieHashKey)
	if err != nil {
		log.Fatal("invalid COOKIE_HASH_KEY")
	}

	if cookieEncryptKey == "" {
		log.Fatal("COOKIE_ENCRYPT_KEY is required")
	}
	encryptKeyBytes, err := hex.DecodeString(cookieEncryptKey)
	if err != nil {
		log.Fatal("invalid COOKIE_ENCRYPT_KEY")
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
		log.Fatal("DISCORD_CLIENT_ID is required")
	}

	if discordClientSecret == "" {
		log.Fatal("DISCORD_CLIENT_SECRET is required")
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
		log.Fatal("JWT_SECRET is required")
	}
	jwtSecretBytes, err := hex.DecodeString(jwtSecret)
	if err != nil {
		log.Fatal("invalid JWT_SECRET")
	}

	refreshExpiration := os.Getenv("REFRESH_EXPIRE_LIMIT")
	var refreshInt int
	if refreshExpiration == "" {
		refreshInt = 604800 //7 days (7 * 24 * 60 * 60)
	} else {
		refreshInt, err = strconv.Atoi(refreshExpiration)
		if err != nil {
			log.Fatal("invalid refresh expiration--must be integer value")
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

	// Handlers
	h := handler.New(ginEnv, s, sc, discordOauth, cookieDomain, frontendURL, jwtSecretBytes, refreshInt, discordAPIURL)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.SetTrustedProxies([]string{"172.20.0.0/16"})

	// CORS — allow the Next.js origin
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Request ID
	r.Use(middleware.RequestID())

	// logging
	base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(logger.New(base)))

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
		}

		games := protected.Group("/games")
		{
			games.GET("", h.GetSystemGamesHandler)
			games.GET("/:gameId/ranks", h.GetGameRanksByGame)
		}
	}

	log.Printf("🚀  API listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
