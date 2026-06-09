package routes

import (
	"PocketArtisan/internal/modules/files/delete"
	"PocketArtisan/internal/modules/files/serve"
	"PocketArtisan/internal/modules/files/storage"
	"PocketArtisan/internal/modules/files/upload"

	"github.com/gin-gonic/gin"
)

func RegisterFileRoutes(router *gin.Engine) {
	localStorage := storage.NewLocalStorage("./uploads", "http://localhost:8080/files")

	files := router.Group("/files")
	upload.RegisterRoutes(files, localStorage)
	serve.RegisterRoutes(files, localStorage)
	delete.RegisterRoutes(files, localStorage)
}
