package set_profile_picture

import (
	"PocketArtisan/internal/http/response"
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
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		userID, ok := c.Request.Context().Value(middleware.ContextUserID).(uint64)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "user not resolved")
			return
		}
		req.UserID = userID

		err := r.Execute(c.Request.Context(), req)
		if err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Empty(c, http.StatusOK)
	})
}
