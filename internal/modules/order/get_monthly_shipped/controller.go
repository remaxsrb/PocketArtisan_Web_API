package get_monthly_shipped

import (
	"net/http"

	"PocketArtisan/internal/modules/utils/timeutil"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, timeService timeutil.Service) {
	uc := NewService(db, rdb, timeService)

	router.GET("/stats/monthly", func(c *gin.Context) {
		req := MonthlyShippedRequest{
			From: c.Query("from"),
			To:   c.Query("to"),
		}

		resp, err := uc.Execute(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": resp})
	})
}