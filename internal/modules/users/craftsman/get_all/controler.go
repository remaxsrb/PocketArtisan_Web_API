package getall

import (
	"PocketArtisan/internal/http/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	uc := NewService(db, rdb)

	router.GET("/all", func(c *gin.Context) {
		var req GetAllRequest

		if err := c.ShouldBindQuery(&req); err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := uc.Execute(c.Request.Context(), req)
		if err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		response.Data(c, http.StatusOK, resp)
	})
}
