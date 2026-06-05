package login

import (
	"PocketArtisan/internal/modules/auth"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db interface{}, rdb interface{}, jwtService auth.JWTService) {
	r := NewUseCase(db.(*gorm.DB), rdb.(*redis.Client))
	router.POST("/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		resp, err := r.Execute(c.Request.Context(), req)
		if err != nil {

			if err.Error() == "username not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			if err.Error() == "invalid password" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}

		}

		token, err := jwtService.Generate(auth.Identity{
			ID:   resp.ID,
			Role: resp.Role,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"access_token": token, "user": resp})
	})
}
