package get_registered_users

import (
	"PocketArtisan/internal/http/response"
	"net/http"

	"PocketArtisan/internal/modules/utils/timeutil"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, timeService timeutil.Service) {
	uc := NewService(db, rdb, timeService)

	router.GET("/registered", func(c *gin.Context) {
		var req Request
		if err := c.ShouldBindQuery(&req); err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := uc.Execute(c.Request.Context(), req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Data(c, http.StatusOK, resp)
	})
}
