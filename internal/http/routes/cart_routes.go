package routes

import (
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

func RegisterCartRoutes(router *gin.Engine, jwtService auth.JWTService) {
	cart := router.Group("/cart")
	cart.Use(middleware.JWT(), middleware.RequireRoles("user"))

}