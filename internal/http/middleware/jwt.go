package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"PocketArtisan/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID      = "user_id"
	ContextRole        = "role"
	ContextCraftsmanID = "craftsman_id"
)

func JWT(jwtService auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		identity, err := jwtService.Validate(parts[1])
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set(ContextUserID, identity.ID)

		idInt, err := strconv.Atoi(identity.ID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
			return
		}

		ctx := context.WithValue(c.Request.Context(), "user_id", uint64(idInt))
		c.Request = c.Request.WithContext(ctx)
		c.Set(ContextRole, identity.Role)

		c.Next()
	}
}
