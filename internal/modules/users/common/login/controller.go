package login

import (
	"PocketArtisan/internal/modules/auth"
	"PocketArtisan/internal/modules/users"
	"net/http"
	"strconv"

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

		var id int64
		var role string

		switch r := resp.(type) {
		case *users.RegularUserResponse:
			id = r.ID
			role = r.Role
		case *users.CraftsmanResponse:
			id = r.ID
			role = r.Role
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected response type"})
			return
		}
		token, err := jwtService.Generate(auth.Identity{
			ID:   strconv.FormatInt(id, 10),
			Role: role,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"access_token": token, "user": resp})
	})
}
