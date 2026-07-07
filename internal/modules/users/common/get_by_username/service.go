package getbyusername

import (
	usersmod "PocketArtisan/internal/modules/users"
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

type Service struct {
	repo  usersmod.Repository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: usersmod.NewGormRepository(db), cache: cache}
}

type cacheEnvelope struct {
	Role string          `json:"role"`
	Data json.RawMessage `json:"data"`
}

func (uc *Service) Execute(ctx context.Context, username string) (interface{}, error) {
	cacheVersion := utils.GetCacheVersion(ctx, uc.cache, "users")
	cacheKey := fmt.Sprintf("user:username:v:%d:%s", cacheVersion, username)

	cached, err := uc.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var env cacheEnvelope
		if jsonErr := json.Unmarshal([]byte(cached), &env); jsonErr == nil {
			if env.Role == "craftsman" {
				var r usersmod.CraftsmanResponse
				if jsonErr := json.Unmarshal(env.Data, &r); jsonErr == nil {
					return &r, nil
				}
			} else {
				var r usersmod.RegularUserResponse
				if jsonErr := json.Unmarshal(env.Data, &r); jsonErr == nil {
					return &r, nil
				}
			}
		}
	} else if !errors.Is(err, redis.Nil) {
		fmt.Printf("Redis error on get_by_username: %v\n", err)
	}

	user, err := uc.repo.FindUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	var result interface{}
	if user.Role == "craftsman" {
		r, err := uc.repo.FindCraftsmanProfileByUsername(ctx, user.Username)
		if err != nil {
			return nil, err
		}
		result = r
	} else {
		result = &usersmod.RegularUserResponse{
			UserResponse: usersmod.UserResponse{
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
