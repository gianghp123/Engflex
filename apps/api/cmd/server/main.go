package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"engflex-api/config"
	"engflex-api/internal/logger"
	"engflex-api/internal/observability"
	"engflex-api/internal/server"
)

// @title			EngFlex API
// @version		1.0
// @description	English learning control plane. Success responses use the {message, data[, pagination]} envelope; failures use {message} with the HTTP status code.
// @host			localhost:8000
// @BasePath		/api/v1
// @securitydefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Clerk session token: "Bearer <token>". Paste any token into the Authorize dialog (remembered across reloads); apps/api/.env carries a long-lived SWAGGER_TESTING_JWT for manual curl calls.
func main() {
	// Load .env (when present) + environment into a typed config.
	// Run from apps/api so the default ".env" resolves.
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Install the structured logger before anything else logs.
	slog.SetDefault(logger.New(cfg.Log.Level, cfg.Log.Format))

	sentryEnabled, err := observability.Init(cfg.Sentry)
	if err != nil {
		slog.Warn("sentry disabled: init failed", "error", err)
	} else if sentryEnabled {
		defer sentry.Flush(2 * time.Second)
		slog.Info("sentry enabled", "environment", cfg.Sentry.Environment)
	} else {
		slog.Info("sentry disabled: no DSN")
	}

	// Gin's debug output is development noise.
	if cfg.Server.Env != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	if cfg.Auth.ClerkSecretKey == "" {
		slog.Error("missing CLERK_SECRET_KEY")
		os.Exit(1)
	}
	clerk.SetKey(cfg.Auth.ClerkSecretKey)

	slog.Info("starting engflex api",
		"env", cfg.Server.Env,
		"addr", cfg.Server.Addr(),
		"database", cfg.Database.LogValue(),
	)

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	router := server.NewRouter(db, cfg)

	slog.Info("swagger docs available", "url", "http://localhost:"+cfg.Server.Port+"/swagger/index.html")

	if err := router.Run(cfg.Server.Addr()); err != nil {
		slog.Error("server exited unexpectedly", "error", err)
		os.Exit(1)
	}
}
