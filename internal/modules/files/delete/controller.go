package delete

import (
	"PocketArtisan/internal/http/response"
	"net/http"

	"PocketArtisan/internal/modules/files/storage"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, s storage.Storage) {
	uc := NewService(s)

	router.DELETE("/delete/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		if filename == "" {
			response.Error(c, http.StatusBadRequest, "filename required")
			return
		}

		if err := uc.Execute(filename); err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Data(c, http.StatusOK, gin.H{"message": "file deleted"})
	})
}
