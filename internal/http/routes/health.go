package routes

import (
	"PocketArtisan/internal/modules/health/ping"

	"github.com/gin-gonic/gin"
)

func RegisterHealthRoute(router *gin.Engine) {
	healthPublic := router.Group("/api/health")
	ping.RegisterRoutes(healthPublic)
}
