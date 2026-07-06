package decline

import (
	ordermod "PocketArtisan/internal/modules/order"
	"PocketArtisan/internal/modules/payment"
	"net/http"

	"PocketArtisan/internal/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, gw payment.Gateway) {
	svc := NewService(db, rdb, gw)
	errHandler := ordermod.NewErrorHandler()
	router.POST("/decline", func(c *gin.Context) {
		var req DeclineOrderRequest
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

		status, err := svc.Execute(c.Request.Context(), req)
		if err != nil {
			errHandler.HandleOrderOperationError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": status})
	})
}
