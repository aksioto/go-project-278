package middleware

import (
	"errors"
	"net/http"

	"code/internal/domain/link"
	"code/internal/transport/http/validation"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// ErrorsMiddleware is a middleware for centralized error handling.
func ErrorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		handleError(c, err)
	}
}

func handleError(c *gin.Context, err error) {
	var validationErr *validation.ValidationError
	if errors.As(err, &validationErr) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": validationErr.Fields})
		return
	}

	if errors.Is(err, link.ErrInvalidID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if errors.Is(err, link.ErrInvalidOriginalURL) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": map[string]string{"original_url": "cannot shorten own short URLs"},
		})
		return
	}

	if errors.Is(err, link.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}

	if errors.Is(err, link.ErrVisitNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "visit not found"})
		return
	}

	if errors.Is(err, link.ErrShortNameTaken) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"errors": map[string]string{"short_name": "short name already in use"},
		})
		return
	}

	captureError(c, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func captureError(c *gin.Context, err error) {
	hub := sentrygin.GetHubFromContext(c)
	if hub == nil {
		return
	}

	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelError)
		scope.SetRequest(c.Request)
		scope.SetTag("endpoint", c.FullPath())
		scope.SetTag("method", c.Request.Method)
		scope.SetFingerprint([]string{
			c.FullPath(),
			err.Error(),
		})
		hub.CaptureException(err)
	})
}
