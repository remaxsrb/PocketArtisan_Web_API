package middleware

import (
	"net/http"

	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/http/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RequireCraftsman resolves the logged-in user's own craftsman record from
// their JWT-authenticated user_id and stores its ID in the gin context under
// ContextCraftsmanID. Handlers should read the craftsman ID from there
// instead of trusting a client-supplied craftsman_id/username, so a
// craftsman can only ever act on or view their own data.
//
// Must run after JWT() and RequireRoles("craftsman").
func RequireCraftsman(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Request.Context().Value(ContextUserID).(uint64)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "user not resolved")
			c.Abort()
			return
		}

		var craftsman entities.Craftsman
		if err := db.WithContext(c.Request.Context()).Where("user_id = ?", userID).First(&craftsman).Error; err != nil {
			response.Error(c, http.StatusForbidden, "craftsman profile not found")
			c.Abort()
			return
		}

		c.Set(ContextCraftsmanID, craftsman.ID)
		c.Next()
	}
}
