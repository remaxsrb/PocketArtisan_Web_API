package ship

import (
	"PocketArtisan/internal/http/response"
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
	router.POST("/ship", func(c *gin.Context) {
		var req ShipOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		craftsmanID := c.GetUint64(middleware.ContextCraftsmanID)
		if craftsmanID == 0 {
			response.Error(c, http.StatusUnauthorized, "craftsman not resolved")
			return
		}
		req.CraftsmanID = craftsmanID

		status, err := svc.Execute(c.Request.Context(), req)
		if err != nil {
			errHandler.HandleOrderOperationError(c, err)
			return
		}
		response.Data(c, http.StatusOK, gin.H{"status": status})
	})
}
