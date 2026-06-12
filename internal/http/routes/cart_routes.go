package routes

import (
	"PocketArtisan/config"
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/auth"

	"github.com/gin-gonic/gin"

	"PocketArtisan/internal/modules/cart/add_to_cart"
	"PocketArtisan/internal/modules/cart/remove_from_cart"		

)

func RegisterCartRoutes(router *gin.Engine, jwtService auth.JWTService) {
	cartClosed := router.Group("/cart")
	cartClosed.Use(middleware.JWT(), middleware.RequireRoles("user"))
	addtocart.RegisterRoutes(cartClosed, config.DB, config.RDB)
	removefromcart.RegisterRoutes(cartClosed, config.DB, config.RDB)
}