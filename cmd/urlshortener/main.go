package main

import (
	"code/internal/config"
	"code/internal/db/repository"
	"code/internal/infra/postgres"
	transportHTTP "code/internal/transport/http"
	"code/internal/transport/http/middleware"
	linkusecase "code/internal/usecase/link"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"

	sentryinfra "code/internal/infra/sentry"

	"github.com/gin-gonic/gin"
)

func main() {
	// Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err = cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Sentry
	sentryClient, err := sentryinfra.Init(
		sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			TracesSampleRate: cfg.SentrySampleRate,
			Environment:      cfg.Env,
		},
		time.Duration(cfg.SentryFlushTimeoutMs)*time.Millisecond,
	)
	if err != nil {
		log.Fatalf("sentry init failed: %v", err)
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
		log.Fatalf("failed to create db pool: %v", err)
	}
	defer dbPool.Close()

	// Init repository + usecase
	linkRepo := repository.NewLinkPostgres(dbPool)
	linkService := linkusecase.NewService(linkRepo)

	// Setup Gin router + middleware
	if cfg.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(
		gin.Logger(),
		gin.Recovery(),
		middleware.SentryMiddleware(sentryClient),
		middleware.CORSMiddleware(cfg.AllowedOrigins...),
	)

	// Healthcheck
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	transportHTTP.SetupRoutes(router, linkService, cfg.BaseURL)

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	if err = router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
