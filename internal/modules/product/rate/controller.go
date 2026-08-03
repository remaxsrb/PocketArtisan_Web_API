package rate

import (
	"PocketArtisan/internal/http/response"
	prodmod "PocketArtisan/internal/modules/product"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	r := NewService(db, rdb)
	router.POST("/rate", func(c *gin.Context) {
		var req Request
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		resp, err := r.Execute(c.Request.Context(), req)
		if err != nil {
			response.Error(c, statusFor(err), err.Error())
			return
		}
		response.Data(c, http.StatusOK, resp)
	})
}

func statusFor(err error) int {
	switch {
	case prodmod.IsNotPurchasedError(err):
		return http.StatusForbidden
	case prodmod.IsAlreadyRatedError(err):
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
}
