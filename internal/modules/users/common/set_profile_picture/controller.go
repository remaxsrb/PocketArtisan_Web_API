package set_profile_picture

import (
	"net/http"

	"PocketArtisan/internal/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	r := NewService(db, rdb)
	router.PATCH("/set-profile-picture", func(c *gin.Context) {
		var req SetProfilePictureRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID, ok := c.Request.Context().Value(middleware.ContextUserID).(uint64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not resolved"})
			return
		}
		req.UserID = userID

		err := r.Execute(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, nil)
	})
}
