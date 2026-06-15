package delete

import (
	"net/http"

	"PocketArtisan/internal/modules/files/storage"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, s storage.Storage) {
	uc := NewService(s)

	router.DELETE("/delete/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		if filename == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "filename required"})
			return
		}

		if err := uc.Execute(filename); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "file deleted"})
	})
}
