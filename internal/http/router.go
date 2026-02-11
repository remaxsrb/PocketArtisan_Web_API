package http

import (
	"PocketArtisan/config"
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/auth"
	"PocketArtisan/internal/modules/files/delete"
	"PocketArtisan/internal/modules/files/serve"
	"PocketArtisan/internal/modules/files/storage"
	"PocketArtisan/internal/modules/files/upload"
	"PocketArtisan/internal/modules/users/admin/set_role"
	"PocketArtisan/internal/modules/users/common/change_password"
	"PocketArtisan/internal/modules/users/common/delete_account"
	"PocketArtisan/internal/modules/users/common/get_all"
	"PocketArtisan/internal/modules/users/common/login"
	"PocketArtisan/internal/modules/users/common/register"
	"PocketArtisan/internal/modules/users/common/set_profile_picture"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	jwtService := auth.GetJWTService()

	// User routes

	publicUserGroup := router.Group("/users")
	register.RegisterRoutes(publicUserGroup, config.DB, config.RDB)
	login.RegisterRoutes(publicUserGroup, config.DB, config.RDB, jwtService)

	protectedUserGroup := router.Group("/users")
	protectedUserGroup.Use(middleware.JWT())

	change_password.RegisterRoutes(protectedUserGroup, config.DB, config.RDB)
	set_profile_picture.RegisterRoutes(protectedUserGroup, config.DB, config.RDB)
	delete_account.RegisterRoutes(protectedUserGroup, config.DB, config.RDB)
	get_all.RegisterRoutes(protectedUserGroup, config.DB, config.RDB)

	// Admin-only routes

	adminUsers := protectedUserGroup.Group("/admin")
	adminUsers.Use(middleware.RequireRoles("admin"))

	set_role.RegisterRoutes(adminUsers, config.DB, config.RDB)

	// File routes
	localStorage := storage.NewLocalStorage("./uploads", "http://localhost:8080/files")

	filesGroup := router.Group("/files")
	upload.RegisterRoutes(filesGroup, localStorage)
	serve.RegisterRoutes(filesGroup, localStorage)
	delete.RegisterRoutes(filesGroup, localStorage)

	return router
}
