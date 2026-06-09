package getallbycraftsman

import (
	"PocketArtisan/internal/modules/product"
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

		var cachedResp GetAllResponse
		if err := json.Unmarshal([]byte(cachedData), &cachedResp); err == nil {
			return cachedResp, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		fmt.Printf("Redis error: %v\n", err)
	}

	product_list := make([]*product.ProductResponse, 0, req.Limit)

	var totalProducts int64
	uc.db.WithContext(ctx).Model(&product.Product{}).Count(&totalProducts)

	uc.db.WithContext(ctx).
		Table("products").
		Select(`
        products.name, 
        products.craftsman_id, 
		products.hidden,
        products.picture, 
        products.material_price + products.labor_price as totalPrice, 
        products.description`).
		Where("products.craftsman_id = ?", req.CraftsmanID).
		Offset(req.Skip).
		Limit(req.Limit).
		Order("products.name asc").
		Scan(&product_list)

	resp := GetAllResponse{
		Products: product_list,
		Total:    totalProducts,
		Page:     (req.Skip / req.Limit) + 1,
	}

	jsonData, err := json.Marshal(resp)
	if err == nil {
		_ = uc.cache.Set(ctx, cacheKey, jsonData, cacheTTL).Err()
	}

	return resp, nil

}
  