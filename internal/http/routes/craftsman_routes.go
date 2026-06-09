package routes

import (
    "PocketArtisan/config"
    "PocketArtisan/internal/http/middleware"
    craftsman_create "PocketArtisan/internal/modules/users/craftsman/create"
    craftsman_get_all "PocketArtisan/internal/modules/users/craftsman/get_all"
    get_by_craft "PocketArtisan/internal/modules/users/craftsman/get_by_craft"

    "github.com/gin-gonic/gin"
)

func RegisterCraftsmanRoutes(router *gin.Engine) {
	public := router.Group("/craftsman")
	craftsman_get_all.RegisterRoutes(public, config.DB, config.RDB)
	get_by_craft.RegisterRoutes(public, config.DB, config.RDB)

	admin := router.Group("/craftsman")
	admin.Use(middleware.JWT(), middleware.RequireRoles("admin"))
	craftsman_create.RegisterRoutes(admin, config.DB, config.RDB)
}
