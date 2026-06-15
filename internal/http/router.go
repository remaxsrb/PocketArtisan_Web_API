package http

import (
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http/routes"

	"github.com/gin-gonic/gin"
)

func SetupRouter(appContainer *container.AppContainer) *gin.Engine {
	router := gin.Default()
	router.Use(routes.CorsMiddleware())

	router.Static("/assets", "./assets")

	routes.RegisterUserRoutes(router, appContainer)
	routes.RegisterCraftsmanRoutes(router, appContainer)
	routes.RegisterProductRoutes(router, appContainer)
	routes.RegisterCraftsmanApplicationRoutes(router, appContainer)
	routes.RegisterFileRoutes(router, appContainer)
	routes.RegisterCartRoutes(router, appContainer)
	routes.RegisterProductCategoryRoutes(router, appContainer)
	routes.RegisterCraftRoutes(router, appContainer)
	routes.RegisterOrdertRoutes(router, appContainer)

	return router
}
