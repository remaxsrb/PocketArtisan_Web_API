package routes

import (
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http/middleware"
	craftsman_create "PocketArtisan/internal/modules/users/craftsman/create"
	craftsman_get_all "PocketArtisan/internal/modules/users/craftsman/get_all"
	get_by_craft "PocketArtisan/internal/modules/users/craftsman/get_by_craft"

	"github.com/gin-gonic/gin"
)

func RegisterCraftsmanRoutes(router *gin.Engine, appContainer *container.AppContainer) {
	public := router.Group("/craftsman")
	craftsman_get_all.RegisterRoutes(public, appContainer.DB, appContainer.RDB)
	get_by_craft.RegisterRoutes(public, appContainer.DB, appContainer.RDB)

	admin := router.Group("/craftsman")
	admin.Use(middleware.JWT(appContainer.JWTService), middleware.RequireRoles("admin"))
	craftsman_create.RegisterRoutes(admin, appContainer.DB, appContainer.RDB)
}
