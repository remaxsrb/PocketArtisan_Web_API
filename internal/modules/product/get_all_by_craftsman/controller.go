package getallbycraftsman

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db interface{}, rdb interface{}) {
	uc := NewUseCase(db.(*gorm.DB), rdb.(*redis.Client))

	router.GET("/all", func(c *gin.Context) {

		craftsmanID, err := strconv.ParseUint(c.DefaultQuery("craftsmanId", "0"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid craftsmanId"})
			return
		}

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

		req := GetAllRequest{CraftsmanID: craftsmanID, Skip: skip, Limit: limit}

		resp, err := uc.Execute(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": resp})
	})
}
