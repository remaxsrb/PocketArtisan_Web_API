package routes

import (
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/order/accept"
	"PocketArtisan/internal/modules/order/create"
	"PocketArtisan/internal/modules/order/decline"
	"PocketArtisan/internal/modules/order/get_by_craftsman"
	"PocketArtisan/internal/modules/order/get_by_customer"
	"PocketArtisan/internal/modules/order/get_monthly_shipped"
	"PocketArtisan/internal/modules/order/get_monthly_shipped_by_category"
	"PocketArtisan/internal/modules/order/ship"

	"github.com/gin-gonic/gin"
)

func RegisterOrdertRoutes(router *gin.Engine, appContainer *container.AppContainer) {
	publicOrderRoutes := router.Group("/api/orders")
	get_monthly_shipped.RegisterRoutes(publicOrderRoutes, appContainer.DB, appContainer.RDB, appContainer.TimeService)

	customerOrderRoutes := router.Group("/api/orders")
	customerOrderRoutes.Use(middleware.JWT(appContainer.JWTService), middleware.RequireRoles("user"))
	create.RegisterRoutes(customerOrderRoutes, appContainer.DB, appContainer.RDB, appContainer.Storage, appContainer.Fonts, appContainer.BreakerGateway)
	get_by_customer.RegisterRoutes(customerOrderRoutes, appContainer.DB, appContainer.RDB)

	craftsmanOrderRoutes := router.Group("/api/orders")
	craftsmanOrderRoutes.Use(middleware.JWT(appContainer.JWTService), middleware.RequireRoles("craftsman"))
	get_by_craftsman.RegisterRoutes(craftsmanOrderRoutes, appContainer.DB, appContainer.RDB)
	get_monthly_shipped_by_category.RegisterRoutes(craftsmanOrderRoutes, appContainer.DB, appContainer.RDB, appContainer.TimeService)
	accept.RegisterRoutes(craftsmanOrderRoutes, appContainer.DB, appContainer.RDB)
	decline.RegisterRoutes(craftsmanOrderRoutes, appContainer.DB, appContainer.RDB, appContainer.BreakerGateway)
	ship.RegisterRoutes(craftsmanOrderRoutes, appContainer.DB, appContainer.RDB, appContainer.BreakerGateway)
}
