// internal/http/routes/user_routes.go
package routes

import (
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/users/admin/set_role"
	"PocketArtisan/internal/modules/users/common/change_password"
	"PocketArtisan/internal/modules/users/common/delete_account"
	"PocketArtisan/internal/modules/users/common/get_all"
	get_by_username "PocketArtisan/internal/modules/users/common/get_by_username"
	"PocketArtisan/internal/modules/users/common/login"
	"PocketArtisan/internal/modules/users/common/register"
	"PocketArtisan/internal/modules/users/common/set_profile_picture"
	"PocketArtisan/internal/modules/users/craftsman/rate"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(router *gin.Engine, appContainer *container.AppContainer) {
	public := router.Group("/users")
	register.RegisterRoutes(public, appContainer.DB, appContainer.RDB)
	login.RegisterRoutes(public, appContainer.DB, appContainer.RDB, appContainer.JWTService)
	change_password.RegisterRoutes(public, appContainer.DB, appContainer.RDB)
	get_by_username.RegisterRoutes(public, appContainer.DB, appContainer.RDB)

	protected := router.Group("/users")
	protected.Use(middleware.JWT(appContainer.JWTService))
	set_profile_picture.RegisterRoutes(protected, appContainer.DB, appContainer.RDB)
	delete_account.RegisterRoutes(protected, appContainer.DB, appContainer.RDB)
	rate.RegisterRoutes(protected, appContainer.DB, appContainer.RDB)

	admin := router.Group("/users")
	admin.Use(middleware.JWT(appContainer.JWTService), middleware.RequireRoles("admin"))
	get_all.RegisterRoutes(admin, appContainer.DB, appContainer.RDB)
	set_role.RegisterRoutes(admin, appContainer.DB, appContainer.RDB)
}
