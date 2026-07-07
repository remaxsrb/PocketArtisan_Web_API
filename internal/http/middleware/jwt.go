package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"PocketArtisan/internal/http/response"
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
			response.Error(c, http.StatusUnauthorized, "authorization header missing")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, "invalid authorization header")
			c.Abort()
			return
		}

		identity, err := jwtService.Validate(parts[1])
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(ContextUserID, identity.ID)

		idInt, err := strconv.Atoi(identity.ID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid user ID format")
			c.Abort()
			return
		}

		ctx := context.WithValue(c.Request.Context(), "user_id", uint64(idInt))
		c.Request = c.Request.WithContext(ctx)
		c.Set(ContextRole, identity.Role)

		c.Next()
	}
}
