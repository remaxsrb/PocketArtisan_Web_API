package register

import (
	"PocketArtisan/internal/http/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	r := NewService(db, rdb)
	router.POST("/register", func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		user, err := r.Execute(c.Request.Context(), req, c.ClientIP())
		if err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Data(c, http.StatusCreated, gin.H{
			"username": user.Username,
		})
	})
}
