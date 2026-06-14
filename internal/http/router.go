package http

import (
	"PocketArtisan/internal/http/routes"
	"PocketArtisan/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(routes.CorsMiddleware())

	router.Static("/assets", "./assets")

	jwtService := auth.GetJWTService()

	routes.RegisterUserRoutes(router, jwtService)
	routes.RegisterCraftsmanRoutes(router)
	routes.RegisterProductRoutes(router, jwtService)
	routes.RegisterCraftsmanApplicationRoutes(router, jwtService)
	routes.RegisterFileRoutes(router)
	routes.RegisterCartRoutes(router, jwtService)
	routes.RegisterProductCategoryRoutes(router, jwtService)
	routes.RegisterCraftRoutes(router, jwtService)

	return router
}
