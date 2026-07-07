package upload

import (
	"PocketArtisan/internal/http/response"
	"net/http"

	"PocketArtisan/internal/modules/files/storage"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, s storage.Storage) {
	uc := NewService(s)

	router.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		purpose := c.PostForm("purpose")
		if err != nil {
			response.Error(c, http.StatusBadRequest, "file required")
			return
		}
		url, err := uc.Execute(file, purpose)
		if err != nil {

			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Data(c, http.StatusOK, gin.H{"url": url})
	})
}
