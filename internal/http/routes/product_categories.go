package routes

import (
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/product_categories/create"
	"PocketArtisan/internal/modules/product_categories/delete"
	"PocketArtisan/internal/modules/product_categories/get_all"
	getbycraftsman "PocketArtisan/internal/modules/product_categories/get_by_craftsman"

	"github.com/gin-gonic/gin"
)

func RegisterProductCategoryRoutes(router *gin.Engine, appContainer *container.AppContainer) {
	productCategoryModeration := router.Group("/api/product-categories")
	productCategoryModeration.Use(middleware.JWT(appContainer.JWTService), middleware.RequireRoles("admin"))
	create.RegisterRoutes(productCategoryModeration, appContainer.DB, appContainer.RDB)
	delete.RegisterRoutes(productCategoryModeration, appContainer.DB, appContainer.RDB)

	productCategoryPublic := router.Group("/api/product-categories")
	get_all.RegisterRoutes(productCategoryPublic, appContainer.DB, appContainer.RDB)

	productCategoryCraftsman := router.Group("/api/product-categories")
	productCategoryCraftsman.Use(middleware.JWT(appContainer.JWTService), middleware.RequireRoles("craftsman"), middleware.RequireCraftsman(appContainer.DB))
	getbycraftsman.RegisterRoutes(productCategoryCraftsman, appContainer.DB, appContainer.RDB)
}
