package create

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/utils"
	"context"
	"errors"

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

func (uc *UseCase) Execute(ctx context.Context, req NewProductCategoryRequest) error {

	var pc entities.ProductCategory
	if err := uc.db.WithContext(ctx).Where("name = ?", req.Name).First(&pc).Error; err == nil {
		return errors.New("product category already exists")
	}

	searchKeywords := make([]string, 0, len(req.Keywords)+1)
	searchKeywords = append(searchKeywords, utils.NormalizeForSearch(req.Name))
	for _, kw := range req.Keywords {
		searchKeywords = append(searchKeywords, utils.NormalizeForSearch(kw))
	}

	pc = entities.ProductCategory{
		Name:           req.Name,
		Keywords:       req.Keywords,
		SearchKeywords: searchKeywords,
	}

	if err := uc.db.Create(&pc).Error; err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "products", "product_categories")

	return nil

}
