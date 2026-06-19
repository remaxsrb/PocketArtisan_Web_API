package routes

import (
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/crafts/create"
	"PocketArtisan/internal/modules/crafts/delete"
	"PocketArtisan/internal/modules/crafts/get_all"

	"github.com/gin-gonic/gin"
)

func RegisterCraftRoutes(router *gin.Engine, appContainer *container.AppContainer) {
	craftModeration := router.Group("/api/crafts")
	craftModeration.Use(middleware.JWT(appContainer.JWTService), middleware.RequireRoles("admin"))
	create.RegisterRoutes(craftModeration, appContainer.DB, appContainer.RDB)
	delete.RegisterRoutes(craftModeration, appContainer.DB, appContainer.RDB)

	craftPublic := router.Group("/api/crafts")
	get_all.RegisterRoutes(craftPublic, appContainer.DB, appContainer.RDB)

}
