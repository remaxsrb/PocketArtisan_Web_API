package ping

import (
	"PocketArtisan/internal/http/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {

	router.POST("/ping", func(c *gin.Context) {

		response.Data(c, http.StatusOK, gin.H{"message": "pong"})

	})
}
