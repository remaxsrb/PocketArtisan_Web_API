package ping

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {

	router.POST("/ping", func(c *gin.Context) {

		c.JSON(200, gin.H{"message": "pong"})

	})
}
