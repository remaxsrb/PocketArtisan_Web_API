package upload

import (
	"net/http"

	"PocketArtisan/internal/modules/files/storage"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, s storage.Storage) {
	uc := NewUseCase(s)

	router.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		purpose := c.PostForm("purpose")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
			return
		}
		url, err := uc.Execute(file, purpose)
		if err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"url": url})
	})
}
