package routes

import (
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/modules/files/delete"
	"PocketArtisan/internal/modules/files/serve"
	"PocketArtisan/internal/modules/files/upload"

	"github.com/gin-gonic/gin"
)

func RegisterFileRoutes(router *gin.Engine, appContainer *container.AppContainer) {
	files := router.Group("/files")
	upload.RegisterRoutes(files, appContainer.Storage)
	serve.RegisterRoutes(files, appContainer.Storage)
	delete.RegisterRoutes(files, appContainer.Storage)
}
