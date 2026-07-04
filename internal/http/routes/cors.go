package routes

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var defaultAllowedOrigins = []string{
	"http://localhost:4200",
	"https://napravimi.com",
	"https://www.napravimi.com",
}

// allowedOrigins reads CORS_ALLOWED_ORIGINS (comma-separated) so origins can
// be configured per deployment (release, dev, feature-branch previews)
// without code changes. Falls back to defaultAllowedOrigins when unset.
func allowedOrigins() []string {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		return defaultAllowedOrigins
	}

	origins := make([]string, 0)
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins = append(origins, origin)
		}

	}
	return origins
}

func CorsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins(),
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "OPTIONS", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	})
}
