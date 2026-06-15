package routes

import (
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/product_categories/create"
	"PocketArtisan/internal/modules/product_categories/delete"
	"PocketArtisan/internal/modules/product_categories/get_all"

	"github.com/gin-gonic/gin"
)

func RegisterProductCategoryRoutes(router *gin.Engine, appContainer *container.AppContainer) {
	productCategoryModeration := router.Group("/product-categories")
	productCategoryModeration.Use(middleware.JWT(appContainer.JWTService), middleware.RequireRoles("admin"))
	create.RegisterRoutes(productCategoryModeration, appContainer.DB, appContainer.RDB)
	delete.RegisterRoutes(productCategoryModeration, appContainer.DB, appContainer.RDB)

	productCategoryPublic := router.Group("/product-categories")
	get_all.RegisterRoutes(productCategoryPublic, appContainer.DB, appContainer.RDB)
}
