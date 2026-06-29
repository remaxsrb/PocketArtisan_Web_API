package by_rating

import (
	"PocketArtisan/internal/custom_types"
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

type Service struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{db: db, cache: cache}
}

func (uc *Service) Execute(ctx context.Context, direction custom_types.SortDirection, req SortDtoRequest) (SortCraftsmenResponse, error) {
	const maxLimit = 100
	const defaultLimit = 20
	if req.Limit <= 0 {
		req.Limit = defaultLimit
	}
	if req.Limit > maxLimit {
		req.Limit = maxLimit
	}

	cacheVersion := utils.GetCacheVersion(ctx, uc.cache, "craftsmen")
	cacheKey := fmt.Sprintf("craftsmen:sort:rating:direction:v:%d:%s:skip:%d:limit:%d", cacheVersion, direction, req.Skip, req.Limit)
	cachedData, err := uc.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedResp SortCraftsmenResponse
		if err := json.Unmarshal([]byte(cachedData), &cachedResp); err == nil {
			return cachedResp, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		fmt.Printf("Redis error: %v\n", err)
	}

	baseQuery := uc.db.WithContext(ctx).
		Table("users").
		Joins("INNER JOIN craftsmen ON craftsmen.user_id = users.id").
		Joins("INNER JOIN crafts ON crafts.id = craftsmen.craft_id").
		Where("users.role = ?", "craftsman")

	var total int64
	if err := baseQuery.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return SortCraftsmenResponse{}, err
	}

	craftsmanList := make([]*users.CraftsmanResponse, 0, req.Limit)
	if err := baseQuery.Session(&gorm.Session{}).
		Select(`
        users.firstname, 
        users.lastname, 
        users.username,
        users.email, 
        users.profile_picture, 
        craftsmen.id as craftsman_id,
        crafts.name as craft, 
        craftsmen.rating, 
        craftsmen.number_of_ratings
    `).
		Offset(req.Skip).
		Limit(req.Limit).
		Order("rating " + direction).
		Scan(&craftsmanList).Error; err != nil {
		return SortCraftsmenResponse{}, err
	}

	resp := SortCraftsmenResponse{
		Craftsmen: craftsmanList,
		Total:     total,
		Page:      (req.Skip / req.Limit) + 1,
	}

	if jsonData, err := json.Marshal(resp); err == nil {
		_ = uc.cache.Set(ctx, cacheKey, jsonData, cacheTTL).Err()
	}

	return resp, nil
}
