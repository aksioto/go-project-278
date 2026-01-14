package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware(allowedOrigins ...string) gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     deduplicate(allowedOrigins),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Range"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	if len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
		config.AllowAllOrigins = true
		config.AllowOrigins = nil
	}

	return cors.New(config)
}

func deduplicate(origins []string) []string {
	seen := make(map[string]struct{}, len(origins))
	result := make([]string, 0, len(origins))
	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}
