package http

import (
	"PocketArtisan/internal/http/routes"
	"PocketArtisan/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(routes.CorsMiddleware())

	jwtService := auth.GetJWTService()

	routes.RegisterUserRoutes(router, jwtService)
	routes.RegisterCraftsmanRoutes(router)
	routes.RegisterProductRoutes(router, jwtService)
	routes.RegisterCraftsmanApplicationRoutes(router, jwtService)
	routes.RegisterFileRoutes(router)

	return router
}
