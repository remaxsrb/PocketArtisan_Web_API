package checkout

import (
	"PocketArtisan/internal/modules/files/storage"
	"PocketArtisan/internal/modules/order/create"
	"PocketArtisan/internal/modules/payment"
	"PocketArtisan/internal/modules/utils/fonts"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, s storage.Storage, f *fonts.Service, gw payment.Gateway) {
	orderCreate := create.NewService(db, rdb, s, f, gw)
	svc := NewService(db, rdb, orderCreate)

	router.POST("/checkout", func(c *gin.Context) {
		var req CheckoutRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		results, err := svc.Execute(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"orders": results})
	})
}