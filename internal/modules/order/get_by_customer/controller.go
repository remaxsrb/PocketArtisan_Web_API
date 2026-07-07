package get_by_customer

import (
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/http/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	uc := NewService(db, rdb)

	router.GET("/customers/me", func(c *gin.Context) {
		userID, ok := c.Request.Context().Value(middleware.ContextUserID).(uint64)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "user not resolved")
			return
		}

		skip, err := strconv.Atoi(c.DefaultQuery("skip", "0"))
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid skip")
			return
		}
		limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid limit")
			return
		}

		req := GetAllRequest{CustomerID: userID, Skip: skip, Limit: limit}

		resp, err := uc.Execute(c.Request.Context(), req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Data(c, http.StatusOK, resp)
	})
}
