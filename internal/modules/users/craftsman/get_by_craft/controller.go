package getbycraft

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	uc := NewUseCase(db, rdb)

	router.GET("/craft/:craft", func(c *gin.Context) {
		craft := c.Param("craft")

		skip, err := strconv.Atoi(c.DefaultQuery("skip", "0"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid skip"})
			return
		}
		limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}

		req := GetByCraftRequest{Skip: skip, Limit: limit}

		resp, err := uc.Execute(c.Request.Context(), craft, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": resp})
	})
}
