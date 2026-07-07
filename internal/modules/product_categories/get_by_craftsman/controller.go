package getbycraftsman

import (
	"PocketArtisan/internal/http/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	uc := NewService(db, rdb)
	router.GET("/craftsman/:username", func(c *gin.Context) {
		username := c.Param("username")

		categories, err := uc.Execute(c.Request.Context(), username)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Data(c, http.StatusOK, categories)
	})
}
