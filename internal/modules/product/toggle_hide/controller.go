package toggle_hide

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
	router.PUT("/toggle_hide", func(c *gin.Context) {
		var req ToggleHideProduct
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

		err := r.Execute(c.Request.Context(), req)
		if err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Empty(c, http.StatusOK)
	})
}
