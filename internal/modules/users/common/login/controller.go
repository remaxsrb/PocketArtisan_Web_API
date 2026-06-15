package login

import (
	"PocketArtisan/internal/http/response"
	"PocketArtisan/internal/modules/auth"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, rdb *redis.Client, jwtService auth.JWTService) {
	r := NewUseCase(db, rdb)
	router.POST("/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		result, err := r.Execute(c.Request.Context(), req)
		if err != nil {
			if errors.Is(err, ErrUsernameNotFound) {
				response.Error(c, http.StatusNotFound, err.Error())
				return
			}
			if errors.Is(err, ErrInvalidPassword) {
				response.Error(c, http.StatusUnauthorized, err.Error())
				return
			}
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		identity := auth.Identity{
			ID:   strconv.FormatInt(int64(result.ID), 10),
			Role: result.Role,
		}
		if result.CraftsmanID != 0 {
			identity.CraftsmanID = strconv.FormatUint(result.CraftsmanID, 10)
		}

		token, err := jwtService.Generate(identity)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{"access_token": token, "user": result.Response})
	})
}
