// internal/http/routes/user_routes.go
package routes

import (
    "PocketArtisan/config"
    "PocketArtisan/internal/http/middleware"
    "PocketArtisan/internal/modules/auth"
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

func RegisterUserRoutes(router *gin.Engine, jwtService auth.JWTService) {
	public := router.Group("/users")
	register.RegisterRoutes(public, config.DB, config.RDB)
	login.RegisterRoutes(public, config.DB, config.RDB, jwtService)
	change_password.RegisterRoutes(public, config.DB, config.RDB)
	get_by_username.RegisterRoutes(public, config.DB, config.RDB)

	protected := router.Group("/users")
	protected.Use(middleware.JWT())
	set_profile_picture.RegisterRoutes(protected, config.DB, config.RDB)
	delete_account.RegisterRoutes(protected, config.DB, config.RDB)
	rate.RegisterRoutes(protected, config.DB, config.RDB)

	admin := router.Group("/users")
	admin.Use(middleware.JWT(), middleware.RequireRoles("admin"))
	get_all.RegisterRoutes(admin, config.DB, config.RDB)
	set_role.RegisterRoutes(admin, config.DB, config.RDB)
}
