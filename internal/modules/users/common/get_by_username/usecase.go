package getbyusername

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/users"
	"PocketArtisan/internal/modules/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

const cacheTTL = 5 * time.Minute

type UseCase struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewUseCase(db *gorm.DB, cache *redis.Client) *UseCase {
	return &UseCase{db: db, cache: cache}
}

type cacheEnvelope struct {
	Role string          `json:"role"`
	Data json.RawMessage `json:"data"`
}

func (uc *UseCase) Execute(ctx context.Context, username string) (interface{}, error) {
	cacheVersion := utils.GetCacheVersion(ctx, uc.cache, "users")
	cacheKey := fmt.Sprintf("user:username:v:%d:%s", cacheVersion, username)

	// cache hit

	cached, err := uc.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var env cacheEnvelope
		if jsonErr := json.Unmarshal([]byte(cached), &env); jsonErr == nil {
			if env.Role == "craftsman" {
				var r users.CraftsmanResponse
				if jsonErr := json.Unmarshal(env.Data, &r); jsonErr == nil {
					return &r, nil
				}
			} else {
				var r users.RegularUserResponse
				if jsonErr := json.Unmarshal(env.Data, &r); jsonErr == nil {
					return &r, nil
				}
			}
		}
	} else if !errors.Is(err, redis.Nil) {
		fmt.Printf("Redis error on get_by_username: %v\n", err)
	}

	// cache miss

	var user entities.User
	if err := uc.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	var result interface{}
	if user.Role == "craftsman" {
		var r users.CraftsmanResponse
		uc.db.WithContext(ctx).
			Table("users").
			Select(`
                users.username,
                users.firstname,
                users.lastname,
                users.email,
                users.profile_picture,
                users.gender,
                craftsmen.craft,
                craftsmen.rating,
                craftsmen.number_of_ratings
            `).
			Joins("INNER JOIN craftsmen ON craftsmen.user_id = users.id").
			Where("users.username = ?", user.Username).
			Scan(&r)
		result = &r
	} else {
		result = &users.RegularUserResponse{
			UserResponse: users.UserResponse{
				Username:       user.Username,
				Firstname:      user.Firstname,
				Lastname:       user.Lastname,
				Email:          user.Email,
				ProfilePicture: user.ProfilePicture,
				Gender:         user.Gender,
			},
		}
	}

	dataJSON, err := json.Marshal(result)
	if err == nil {
		env := cacheEnvelope{Role: user.Role, Data: json.RawMessage(dataJSON)}
		envJSON, err := json.Marshal(env)
		if err == nil {
			uc.cache.Set(ctx, cacheKey, envJSON, cacheTTL)
		}
	}

	return result, nil
}
