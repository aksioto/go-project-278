package main

import (
	"log"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

const (
	serverAddr = ":8080"
	sentryDSN  = "https://6a9ff8cb0669df014f75f2c761fc865c@o4510642877169664.ingest.de.sentry.io/4510642882150480"
)

func initSentry() {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDSN,
		TracesSampleRate: 1.0, // в проде обычно 0.1–0.2
	}); err != nil {
		log.Fatalf("sentry init failed: %v", err)
	}
}

func newRouter() *gin.Engine {
	router := gin.New()

	router.Use(
		gin.Logger(),
		gin.Recovery(),
		sentrygin.New(sentrygin.Options{
			Repanic: true, // чтобы gin.Recovery тоже отработал
		}),
	)

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	router.GET("/panic", func(c *gin.Context) {
		panic("test sentry panic")
	})

	return router
}

func main() {
	initSentry()
	defer sentry.Flush(2 * time.Second)

	router := newRouter()

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
