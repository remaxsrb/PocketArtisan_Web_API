package get_monthly_shipped_by_category

import (
	"PocketArtisan/internal/http/response"
	"net/http"

	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/modules/utils/timeutil"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, timeService timeutil.Service) {
	uc := NewService(db, rdb, timeService)

	router.GET("/stats/monthly-by-category", func(c *gin.Context) {
		craftsmanID := c.GetUint64(middleware.ContextCraftsmanID)
		if craftsmanID == 0 {
			response.Error(c, http.StatusUnauthorized, "craftsman not resolved")
			return
		}

		req := MonthlyShippedByCategoryRequest{
			CraftsmanID: craftsmanID,
			From:        c.Query("from"),
			To:          c.Query("to"),
		}

		resp, err := uc.Execute(c.Request.Context(), req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Data(c, http.StatusOK, resp)
	})
}
