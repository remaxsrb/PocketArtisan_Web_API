package http

import (
	"PocketArtisan/config"
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/auth"
	"PocketArtisan/internal/modules/craftsman_application/approve"
	"PocketArtisan/internal/modules/craftsman_application/create"
	ca_get_all "PocketArtisan/internal/modules/craftsman_application/get_all"
	"PocketArtisan/internal/modules/craftsman_application/reject"
	"PocketArtisan/internal/modules/files/delete"
	"PocketArtisan/internal/modules/files/serve"
	"PocketArtisan/internal/modules/files/storage"
	"PocketArtisan/internal/modules/files/upload"
	"PocketArtisan/internal/modules/users/admin/set_role"
	"PocketArtisan/internal/modules/users/common/change_password"
	"PocketArtisan/internal/modules/users/common/delete_account"
	user_get_all "PocketArtisan/internal/modules/users/common/get_all"
	getbyusername "PocketArtisan/internal/modules/users/common/get_by_username"
	"PocketArtisan/internal/modules/users/common/login"
	"PocketArtisan/internal/modules/users/common/register"
	"PocketArtisan/internal/modules/users/common/set_profile_picture"
	craftsman_create "PocketArtisan/internal/modules/users/craftsman/create"
	"PocketArtisan/internal/modules/users/craftsman/rate"

	craftsman_all "PocketArtisan/internal/modules/users/craftsman/get_all"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	router.Use(
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:4200"},
			AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			AllowCredentials: true,
		}))

	jwtService := auth.GetJWTService()

	// User routes

	publicUserGroup := router.Group("/users")
	register.RegisterRoutes(publicUserGroup, config.DB, config.RDB)
	login.RegisterRoutes(publicUserGroup, config.DB, config.RDB, jwtService)
	change_password.RegisterRoutes(publicUserGroup, config.DB, config.RDB)
	create.RegisterRoutes(publicUserGroup, config.DB, config.RDB)
	getbyusername.RegisterRoutes(publicUserGroup, config.DB, config.RDB)

	publicCraftsmanGroup := router.Group("/craftsman")
	craftsman_all.RegisterRoutes(publicCraftsmanGroup, config.DB, config.RDB)

	protectedUserGroup := router.Group("/users")
	protectedUserGroup.Use(middleware.JWT())

	set_profile_picture.RegisterRoutes(protectedUserGroup, config.DB, config.RDB)
	delete_account.RegisterRoutes(protectedUserGroup, config.DB, config.RDB)
	rate.RegisterRoutes(protectedUserGroup, config.DB, config.RDB)

	adminLevelUsers := router.Group("/users")
	adminLevelUsers.Use(middleware.JWT())
	adminLevelUsers.Use(middleware.RequireRoles("admin"))

	adminLevelCraftsmen := router.Group("/craftsman")
	adminLevelCraftsmen.Use(middleware.JWT())
	adminLevelCraftsmen.Use(middleware.RequireRoles("admin"))
	craftsman_create.RegisterRoutes(adminLevelCraftsmen, config.DB, config.RDB)

	user_get_all.RegisterRoutes(adminLevelUsers, config.DB, config.RDB)
	set_role.RegisterRoutes(adminLevelUsers, config.DB, config.RDB)

	adminLevelCAs := router.Group("/craftsman-applications")
	adminLevelCAs.Use(middleware.JWT())
	adminLevelCAs.Use(middleware.RequireRoles("admin"))

	approve.RegisterRoutes(adminLevelCAs, config.DB, config.RDB)
	reject.RegisterRoutes(adminLevelCAs, config.DB, config.RDB)
	ca_get_all.RegisterRoutes(adminLevelCAs, config.DB, config.RDB)

	// File routes
	localStorage := storage.NewLocalStorage("./uploads", "http://localhost:8080/files")

	filesGroup := router.Group("/files")
	upload.RegisterRoutes(filesGroup, localStorage)
	serve.RegisterRoutes(filesGroup, localStorage)
	delete.RegisterRoutes(filesGroup, localStorage)

	return router
}
