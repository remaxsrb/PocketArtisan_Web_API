package getbycategory

import (
	"PocketArtisan/internal/modules/product"
	"PocketArtisan/internal/modules/utils"
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

func (uc *UseCase) Execute(ctx context.Context, req GetByCategoryRequest) (GetByCategoryResponse, error) {

	const maxLimit = 100
	const defaultLimit = 20
	const cacheTTL = 3 * time.Second

	if req.Limit <= 0 {
		req.Limit = defaultLimit
	}
	if req.Limit > maxLimit {
		req.Limit = maxLimit
	}

	cacheVersion := utils.GetCacheVersion(ctx, uc.cache, "products")
	cacheKey := fmt.Sprintf("products:category:search:v:%d:%s:skip:%d:limit:%d", cacheVersion, req.Search, req.Skip, req.Limit)

	cachedData, err := uc.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedResp GetByCategoryResponse
		if err := json.Unmarshal([]byte(cachedData), &cachedResp); err == nil {
			return cachedResp, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		fmt.Printf("Redis error: %v\n", err)
	}

	normalizedSearch := utils.NormalizeForSearch(req.Search)

	var totalProducts int64
	countQuery := uc.db.WithContext(ctx).Model(&product.Product{})
	if normalizedSearch != "" {
		countQuery = countQuery.
			Joins("JOIN product_categories ON product_categories.id = products.category_id").
			Where("? = ANY(product_categories.search_keywords)", normalizedSearch)
	}
	countQuery.Count(&totalProducts)

	listQuery := uc.db.WithContext(ctx).
		Preload("Images").
		Preload("Videos").
		Preload("Category")
	if normalizedSearch != "" {
		listQuery = listQuery.
			Joins("JOIN product_categories ON product_categories.id = products.category_id").
			Where("? = ANY(product_categories.search_keywords)", normalizedSearch)
	}

	raw := make([]*product.Product, 0, req.Limit)
	listQuery.
		Offset(req.Skip).
		Limit(req.Limit).
		Order("name asc").
		Find(&raw)

	product_list := make([]*product.ProductResponse, 0, len(raw))
	for _, p := range raw {
		images := make([]string, 0, len(p.Images))
		for _, img := range p.Images {
			images = append(images, img.URL)
		}
		videos := make([]string, 0, len(p.Videos))
		for _, vid := range p.Videos {
			videos = append(videos, vid.URL)
		}
		categoryName := ""
		if p.Category != nil {
			categoryName = p.Category.Name
		}
		product_list = append(product_list, &product.ProductResponse{
			ID:              p.ID,
			CraftsmanID:     p.CraftsmanID,
			Name:            p.Name,
			Hidden:          p.Hidden,
			Price:           p.Price,
			Description:     p.Description,
			Rating:          p.Rating,
			NumberOfRatings: p.NumberOfRatings,
			Available:       p.Available,
			Images:          images,
			Videos:          videos,
			Category:        categoryName,
		})
	}

	resp := GetByCategoryResponse{
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
