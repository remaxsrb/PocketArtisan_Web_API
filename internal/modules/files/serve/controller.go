package serve

import (
	"net/http"

	"PocketArtisan/internal/modules/files/storage"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, s storage.Storage) {
	uc := NewService(s)

	router.GET("/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		if len(filepath) > 0 && filepath[0] == '/' {
			filepath = filepath[1:]
		}

		url, err := uc.Execute(filepath)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}

		// If local storage, serve file directly
		if local, ok := s.(*storage.LocalStorage); ok {
			c.File(local.BasePath + "/" + filepath)
			return
		}

		// If cloud storage, redirect to the object URL
		c.Redirect(http.StatusTemporaryRedirect, url)
	})
}
