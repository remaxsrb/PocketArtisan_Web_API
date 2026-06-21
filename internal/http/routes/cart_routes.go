package routes

import (
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http/middleware"
	addtocart "PocketArtisan/internal/modules/cart/add_to_cart"
	"PocketArtisan/internal/modules/cart/checkout"
	removefromcart "PocketArtisan/internal/modules/cart/remove_from_cart"

	"github.com/gin-gonic/gin"
)

func RegisterCartRoutes(router *gin.Engine, appContainer *container.AppContainer) {
	cartClosed := router.Group("/api/carts")
	cartClosed.Use(middleware.JWT(appContainer.JWTService), middleware.RequireRoles("user"))
	addtocart.RegisterRoutes(cartClosed, appContainer.DB, appContainer.RDB)
	removefromcart.RegisterRoutes(cartClosed, appContainer.DB, appContainer.RDB)
	checkout.RegisterRoutes(cartClosed, appContainer.DB, appContainer.RDB, appContainer.Storage, appContainer.Fonts, appContainer.BreakerGateway)
}
