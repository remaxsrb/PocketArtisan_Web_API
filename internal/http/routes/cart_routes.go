package routes

import (
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http/middleware"

	"github.com/gin-gonic/gin"

	addtocart "PocketArtisan/internal/modules/cart/add_to_cart"
	removefromcart "PocketArtisan/internal/modules/cart/remove_from_cart"
)

func RegisterCartRoutes(router *gin.Engine, appContainer *container.AppContainer) {
	cartClosed := router.Group("/cart")
	cartClosed.Use(middleware.JWT(appContainer.JWTService), middleware.RequireRoles("user"))
	addtocart.RegisterRoutes(cartClosed, appContainer.DB, appContainer.RDB)
	removefromcart.RegisterRoutes(cartClosed, appContainer.DB, appContainer.RDB)
}
