package by_rating

import (
	"PocketArtisan/internal/custom_types"
	"PocketArtisan/internal/http/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	uc := NewService(db, rdb)

	router.GET("/sort/:direction", func(c *gin.Context) {
		param := c.Param("direction")

		direction := custom_types.SortDirection(param)
		if !direction.IsValid() {
			direction = custom_types.Descending
		} else {
			direction = direction.Normalized()
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

		req := SortDtoRequest{Skip: skip, Limit: limit}

		resp, err := uc.Execute(c.Request.Context(), direction, req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Data(c, http.StatusOK, resp)
	})
}
