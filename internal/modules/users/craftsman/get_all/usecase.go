package getall

import (
	"PocketArtisan/internal/modules/users"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type UseCase struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewUseCase(db *gorm.DB, cache *redis.Client) *UseCase {
	return &UseCase{db: db, cache: cache}
}

func (uc *UseCase) Execute(ctx context.Context, req GetAllRequest) (GetAllResponse, error) {

	const maxLimit = 100
	const defaultLimit = 20

	const cacheTTL = 5 * time.Minute

	if req.Limit <= 0 {
		req.Limit = defaultLimit
	}
	if req.Limit > maxLimit {
		req.Limit = maxLimit
	}

	cacheKey := fmt.Sprintf("craftsmen:all:skip:%d:limit:%d", req.Skip, req.Limit)

	cachedData, err := uc.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache Hit! Unmarshal JSON back into your response struct
		var cachedResp GetAllResponse
		if err := json.Unmarshal([]byte(cachedData), &cachedResp); err == nil {
			return cachedResp, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		fmt.Printf("Redis error: %v\n", err)
	}

	craftsman_list := make([]*users.CraftsmanResponse, 0, req.Limit)

	var totalCraftsmen int64
	uc.db.WithContext(ctx).Model(&users.Craftsman{}).Count(&totalCraftsmen)

	uc.db.WithContext(ctx).
		Table("users").
		Select(`
        users.firstname, 
        users.lastname, 
        users.username,
        users.email, 
        users.profile_picture, 
        craftsmen.id as craftsman_id,
        craftsmen.craft, 
        craftsmen.rating, 
        craftsmen.number_of_ratings
    `).
		Joins("INNER JOIN craftsmen ON craftsmen.user_id = users.id").
		Where("users.role = ?", "craftsman").
		Offset(req.Skip).
		Limit(req.Limit).
		Order("users.created_at desc, users.id asc").
		Scan(&craftsman_list)
		
	resp := GetAllResponse{
		Craftsmen: craftsman_list,
		Total:     totalCraftsmen,
		Page:      (req.Skip / req.Limit) + 1,
	}

	jsonData, err := json.Marshal(resp)
	if err == nil {
		// Set the data with your defined TTL
		_ = uc.cache.Set(ctx, cacheKey, jsonData, cacheTTL).Err()
	}

	return resp, nil

}
