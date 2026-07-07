package getbycraftsman

import (
	"PocketArtisan/internal/http/middleware"
	"PocketArtisan/internal/http/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	uc := NewService(db, rdb)
	router.GET("/craftsman/me", func(c *gin.Context) {
		craftsmanID := c.GetUint64(middleware.ContextCraftsmanID)
		if craftsmanID == 0 {
			response.Error(c, http.StatusUnauthorized, "craftsman not resolved")
			return
		}

		categories, err := uc.Execute(c.Request.Context(), craftsmanID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		response.Data(c, http.StatusOK, categories)
	})
}
