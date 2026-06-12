package routes

import (
	"PocketArtisan/config"
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/auth"
	"PocketArtisan/internal/modules/product_categories/create"
	"PocketArtisan/internal/modules/product_categories/delete"
	"PocketArtisan/internal/modules/product_categories/get_all"

	"github.com/gin-gonic/gin"
)


func RegisterProductCategoryRoutes(router *gin.Engine, jwtService auth.JWTService) {
	productCategoryModeration := router.Group("/product-categories")
	productCategoryModeration.Use(middleware.JWT(), middleware.RequireRoles("admin"))
	create.RegisterRoutes(productCategoryModeration, config.DB, config.RDB)
	delete.RegisterRoutes(productCategoryModeration, config.DB, config.RDB)

	productCategoryPublic := router.Group("/product-categories")
	get_all.RegisterRoutes(productCategoryPublic, config.DB, config.RDB)
}