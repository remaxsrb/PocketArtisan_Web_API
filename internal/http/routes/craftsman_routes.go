package routes

import (
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/users/admin/get_approved_craftsmen"
	"PocketArtisan/internal/modules/users/admin/get_approved_craftsmen_by_month"
	craftsman_create "PocketArtisan/internal/modules/users/craftsman/create"
	craftsman_get_all "PocketArtisan/internal/modules/users/craftsman/get_all"
	get_by_craft "PocketArtisan/internal/modules/users/craftsman/get_by_craft"
	"PocketArtisan/internal/modules/users/craftsman/set_biography"
	"PocketArtisan/internal/modules/users/craftsman/sort/by_rating"

	"github.com/gin-gonic/gin"
)

func RegisterCraftsmanRoutes(router *gin.Engine, appContainer *container.AppContainer) {
	public := router.Group("/api/craftsmen")
	craftsman_get_all.RegisterRoutes(public, appContainer.DB, appContainer.RDB)
	get_by_craft.RegisterRoutes(public, appContainer.DB, appContainer.RDB)
	by_rating.RegisterRoutes(public, appContainer.DB, appContainer.RDB)

	craftsman := router.Group("/api/craftsmen")
	craftsman.Use(middleware.JWT(appContainer.JWTService), middleware.RequireRoles("craftsman"), middleware.RequireCraftsman(appContainer.DB))
	set_biography.RegisterRoutes(craftsman, appContainer.DB, appContainer.RDB)

	admin := router.Group("/api/admin/craftsmen")
	admin.Use(middleware.JWT(appContainer.JWTService), middleware.RequireRoles("admin"))
	craftsman_create.RegisterRoutes(admin, appContainer.DB, appContainer.RDB)
	get_approved_craftsmen.RegisterRoutes(admin, appContainer.DB, appContainer.RDB, appContainer.TimeService)
	get_approved_craftsmen_by_month.RegisterRoutes(admin, appContainer.DB, appContainer.RDB, appContainer.TimeService)
}
