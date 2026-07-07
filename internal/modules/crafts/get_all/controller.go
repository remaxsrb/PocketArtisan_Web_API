package get_all

import (
	"PocketArtisan/internal/http/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	r := NewService(db, rdb)
	router.GET("/all", func(c *gin.Context) {
		categories, err := r.Execute(c.Request.Context())
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		response.Data(c, http.StatusOK, categories)
	})
}
