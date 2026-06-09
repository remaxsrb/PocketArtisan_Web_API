package routes

import (
	"PocketArtisan/config"
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/auth"
	"PocketArtisan/internal/modules/product/create"
	get_all "PocketArtisan/internal/modules/product/get_all_by_craftsman"

	"github.com/gin-gonic/gin"
)

func RegisterProductRoutes(router *gin.Engine, jwtService auth.JWTService) {
	// public := router.Group("/products")  -- when needed

	craftsman := router.Group("/products")
	craftsman.Use(middleware.JWT(), middleware.RequireRoles("craftsman"))
	create.RegisterRoutes(craftsman, config.DB, config.RDB)
	get_all.RegisterRoutes(craftsman, config.DB, config.RDB)
}
