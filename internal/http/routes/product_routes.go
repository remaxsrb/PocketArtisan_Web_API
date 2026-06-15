package routes

import (
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/product/create"
	"PocketArtisan/internal/modules/product/delete"
	get_all "PocketArtisan/internal/modules/product/get_all_by_craftsman"
	"PocketArtisan/internal/modules/product/toggle_hide"

	"github.com/gin-gonic/gin"
)

func RegisterProductRoutes(router *gin.Engine, appContainer *container.AppContainer) {
	public := router.Group("/products")
	get_all.RegisterRoutes(public, appContainer.DB, appContainer.RDB)

	craftsman := router.Group("/products")
	craftsman.Use(middleware.JWT(appContainer.JWTService), middleware.RequireRoles("craftsman"))
	create.RegisterRoutes(craftsman, appContainer.DB, appContainer.RDB)
	delete.RegisterRoutes(craftsman, appContainer.DB, appContainer.RDB)
	toggle_hide.RegisterRoutes(craftsman, appContainer.DB, appContainer.RDB)

}
