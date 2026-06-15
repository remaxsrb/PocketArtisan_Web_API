package routes

import (
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/craftsman_application/create"

	"github.com/gin-gonic/gin"
)

func RegisterOrdertRoutes(router *gin.Engine, appContainer *container.AppContainer) {
	customerOrderRoutes := router.Group("/orders")
	customerOrderRoutes.Use(middleware.JWT(appContainer.JWTService), middleware.RequireRoles("user"))
	create.RegisterRoutes(customerOrderRoutes, appContainer.DB, appContainer.RDB)
}
