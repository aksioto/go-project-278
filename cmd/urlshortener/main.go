package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"code/internal/config"
	"code/internal/db/repository"
	"code/internal/infra/postgres"
	sentryinfra "code/internal/infra/sentry"
	transportHTTP "code/internal/transport/http"
	"code/internal/transport/http/handler"
	"code/internal/transport/http/middleware"
	"code/internal/transport/http/validation"
	linkusecase "code/internal/usecase/link"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Config
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err = cfg.Validate(); err != nil {
		logger.Error("configuration validation failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("configuration loaded",
		slog.String("env", cfg.Env),
		slog.String("port", cfg.AppPort),
	)

	// Sentry
	sentryClient, err := sentryinfra.Init(
		sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			TracesSampleRate: cfg.SentrySampleRate,
			Environment:      cfg.Env,
			Release:          fmt.Sprintf("%s@%s", cfg.ServiceName, cfg.AppVersion),
		},
		time.Duration(cfg.SentryFlushTimeoutMs)*time.Millisecond,
	)
	if err != nil {
		logger.Error("sentry init failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer sentryClient.Close()

	// Database
	ctx := context.Background()
	dbPool, err := postgres.NewPool(ctx, postgres.Config{
		DatabaseURL: cfg.DatabaseURL,
		MaxConns:    cfg.MaxConns,
		IdleTimeMs:  cfg.IdleTimeMs,
	})
	if err != nil {
		logger.Error("failed to create db pool", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbPool.Close()

	logger.Info("database connected")

	// Init repositories + usecase
	linkRepo := repository.NewLinkPostgres(dbPool)
	visitRepo := repository.NewVisitPostgres(dbPool)
	linkService := linkusecase.NewService(linkRepo, visitRepo, logger)

	// Setup Gin router + middleware
	if cfg.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	validation.Init()

	router := gin.New()
	router.TrustedPlatform = gin.PlatformCloudflare
	router.Use(
		gin.Recovery(),
		middleware.SentryMiddleware(sentryClient),
		gin.Logger(),
		middleware.CORSMiddleware(cfg.AllowedOrigins...),
		middleware.ErrorsMiddleware(),
	)

	// Healthcheck
	router.GET("/ping", handler.PingHandler)

	transportHTTP.SetupRoutes(router, linkService, cfg.BaseURL)

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	logger.Info("starting server", slog.String("addr", addr))

	if err = router.Run(addr); err != nil {
		logger.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
