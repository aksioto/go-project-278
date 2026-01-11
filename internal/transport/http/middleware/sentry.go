package middleware

import (
	sentryinfra "code/internal/infra/sentry"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

func SentryMiddleware(client *sentryinfra.Client) gin.HandlerFunc {
	if client == nil || !client.Active() {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return sentrygin.New(sentrygin.Options{Repanic: true})
}
