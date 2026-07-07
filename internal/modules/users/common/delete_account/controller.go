package delete_account

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
	handler := func(c *gin.Context) {
		userID, ok := c.Request.Context().Value(middleware.ContextUserID).(uint64)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "user not resolved")
			return
		}

		err := r.Execute(c.Request.Context(), DeleteAccountRequest{UserID: userID})
		if err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Empty(c, http.StatusOK)
	}

	router.DELETE("/delete", handler)
	router.DELETE("/delete/", handler)
}
