package ship

import (
	ordermod "PocketArtisan/internal/modules/order"
	"PocketArtisan/internal/modules/payment"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, gw payment.Gateway) {
	svc := NewService(db, rdb, gw)
	errHandler := ordermod.NewErrorHandler()
	router.POST("/ship", func(c *gin.Context) {
		var req ShipOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		status, err := svc.Execute(c.Request.Context(), req)
		if err != nil {
			errHandler.HandleOrderOperationError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": status})
	})
}
