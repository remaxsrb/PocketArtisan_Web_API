package serve

import (
	"PocketArtisan/internal/http/response"
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
			response.Error(c, http.StatusNotFound, "file not found")
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
