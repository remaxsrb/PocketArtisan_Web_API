package create

import (
	"PocketArtisan/internal/http/response"
	"PocketArtisan/internal/modules/files/storage"
	ordermod "PocketArtisan/internal/modules/order"
	"PocketArtisan/internal/modules/payment"
	"PocketArtisan/internal/modules/utils/fonts"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, s storage.Storage, f *fonts.Service, gw payment.Gateway) {
	svc := NewService(db, rdb, s, f, gw)
	errHandler := ordermod.NewErrorHandler()
	router.POST("/create", func(c *gin.Context) {
		var req NewOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		result, err := svc.Execute(c.Request.Context(), req)
		if err != nil {
			errHandler.HandleOrderCreationError(c, err)
			return
		}
		response.Data(c, http.StatusCreated, gin.H{"message": "order created successfully", "url": result.PDFURL})
	})
}
