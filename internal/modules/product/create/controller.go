package create

import (
	"net/http"

	"PocketArtisan/internal/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	r := NewService(db, rdb)
	router.POST("/create", func(c *gin.Context) {
		var req NewProductRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		craftsmanID := c.GetUint64(middleware.ContextCraftsmanID)
		if craftsmanID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "craftsman not resolved"})
			return
		}
		req.CraftsmanID = craftsmanID

		err := r.Execute(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "product created successfully"})
	})
}
