package routes

import (
	"PocketArtisan/config"
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/auth"
	"PocketArtisan/internal/modules/craftsman_application/approve"
	"PocketArtisan/internal/modules/craftsman_application/create"
	"PocketArtisan/internal/modules/craftsman_application/get_all"
	"PocketArtisan/internal/modules/craftsman_application/reject"

	"github.com/gin-gonic/gin"
)

func RegisterCraftsmanApplicationRoutes(router *gin.Engine, jwtService auth.JWTService) {

	public := router.Group("/craftsman-applications")
	create.RegisterRoutes(public, config.DB, config.RDB)

	admin := router.Group("/craftsman-applications")
	admin.Use(middleware.JWT(), middleware.RequireRoles("admin"))
	approve.RegisterRoutes(admin, config.DB, config.RDB)
	reject.RegisterRoutes(admin, config.DB, config.RDB)
	get_all.RegisterRoutes(admin, config.DB, config.RDB)
}
