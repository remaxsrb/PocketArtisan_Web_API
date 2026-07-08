package routes

import (
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/craftsman_application/approve"
	"PocketArtisan/internal/modules/craftsman_application/create"
	"PocketArtisan/internal/modules/craftsman_application/get_all"
	"PocketArtisan/internal/modules/craftsman_application/reject"

	"github.com/gin-gonic/gin"
)

func RegisterCraftsmanApplicationRoutes(router *gin.Engine, appContainer *container.AppContainer) {

	public := router.Group("/api/craftsman-applications")
	create.RegisterRoutes(public, appContainer.DB, appContainer.RDB)

	admin := router.Group("/api/admin/craftsman-applications")
	admin.Use(middleware.JWT(appContainer.JWTService), middleware.RequireRoles("admin"))
	approve.RegisterRoutes(admin, appContainer.DB, appContainer.RDB, appContainer.MailService)
	reject.RegisterRoutes(admin, appContainer.DB, appContainer.RDB, appContainer.MailService)
	get_all.RegisterRoutes(admin, appContainer.DB, appContainer.RDB)
}
