package create

import (
	"PocketArtisan/internal/entities"
	pcmod "PocketArtisan/internal/modules/product_categories"
	"PocketArtisan/internal/modules/utils"
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo  pcmod.Repository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: pcmod.NewGormRepository(db), cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req NewProductCategoryRequest) error {

	exists, err := uc.repo.ExistsByName(ctx, req.Name)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("product category already exists")
	}

	searchKeywords := make([]string, 0, len(req.Keywords)+1)
	searchKeywords = append(searchKeywords, utils.NormalizeForSearch(req.Name))
	for _, kw := range req.Keywords {
		searchKeywords = append(searchKeywords, utils.NormalizeForSearch(kw))
	}

	pc := &entities.ProductCategory{
		Name:           req.Name,
		Keywords:       req.Keywords,
		SearchKeywords: searchKeywords,
	}

	if err := uc.repo.Create(ctx, pc); err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "products", "product_categories")

	return nil

}
