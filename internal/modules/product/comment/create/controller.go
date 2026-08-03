package create

import (
	"PocketArtisan/internal/http/response"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, mongoDB *mongo.Database) {
	uc := NewService(db, mongoDB)

	router.POST("/comment", func(c *gin.Context) {
		var req Request
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		resp, err := uc.Execute(c.Request.Context(), req)
		if err != nil {
			response.Error(c, statusFor(err), err.Error())
			return
		}
		response.Data(c, http.StatusCreated, resp)
	})
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, ErrProductNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrNotPurchased):
		return http.StatusForbidden
	default:
		return http.StatusBadRequest
	}
}
