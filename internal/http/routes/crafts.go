package routes

import (
	"PocketArtisan/config"
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/auth"
	"PocketArtisan/internal/modules/crafts/create"
	"PocketArtisan/internal/modules/crafts/delete"
	"PocketArtisan/internal/modules/crafts/get_all"

	"github.com/gin-gonic/gin"
)

func RegisterCraftRoutes(router *gin.Engine, jwtService auth.JWTService) {
	craftModeration := router.Group("/crafts")
	craftModeration.Use(middleware.JWT(), middleware.RequireRoles("admin"))
	create.RegisterRoutes(craftModeration, config.DB, config.RDB)
	delete.RegisterRoutes(craftModeration, config.DB, config.RDB)

	craftPublic := router.Group("/crafts")
	get_all.RegisterRoutes(craftPublic, config.DB, config.RDB)


}