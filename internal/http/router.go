package http

import (
	"PocketArtisan/config"
	"PocketArtisan/internal/modules/files/delete"
	"PocketArtisan/internal/modules/files/serve"
	"PocketArtisan/internal/modules/files/storage"
	"PocketArtisan/internal/modules/files/upload"
	"PocketArtisan/internal/modules/users/change_password"
	"PocketArtisan/internal/modules/users/delete_account"
	"PocketArtisan/internal/modules/users/register"
	"PocketArtisan/internal/modules/users/set_profile_picture"
	"PocketArtisan/internal/modules/users/set_role"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// User routes
	userGroup := router.Group("/users")
	register.RegisterRoutes(userGroup, config.DB, config.RDB)
	set_role.RegisterRoutes(userGroup, config.DB, config.RDB)
	change_password.RegisterRoutes(userGroup, config.DB, config.RDB)
	set_profile_picture.RegisterRoutes(userGroup, config.DB, config.RDB)
	delete_account.RegisterRoutes(userGroup, config.DB, config.RDB)

	// File routes
	localStorage := storage.NewLocalStorage("./uploads", "http://localhost:8080/files")

	filesGroup := router.Group("/files")
	upload.RegisterRoutes(filesGroup, localStorage)
	serve.RegisterRoutes(filesGroup, localStorage)
	delete.RegisterRoutes(filesGroup, localStorage)

	return router
}
