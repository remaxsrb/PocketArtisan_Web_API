package getbyusername

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db interface{}, rdb interface{}) {
	uc := NewUseCase(db.(*gorm.DB), rdb.(*redis.Client))

	router.GET("username/:username", func(c *gin.Context) {
		username := c.Param("username")

		resp, err := uc.Execute(c.Request.Context(), username)
		if err != nil {
			if err.Error() == "user not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": resp})
	})
}
