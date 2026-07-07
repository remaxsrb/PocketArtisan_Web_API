package setcategories

import (
	"PocketArtisan/internal/entities"
	craftsmod "PocketArtisan/internal/modules/crafts"
	"PocketArtisan/internal/modules/utils"
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo  craftsmod.Repository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: craftsmod.NewGormRepository(db), cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req SetCraftCategoriesRequest) error {
	craft, err := uc.repo.FindByName(ctx, req.Craft)
	if err != nil {
		return errors.New("craft not found")
	}

	links := make([]entities.CraftProductCategory, 0, len(req.Categories))
	for _, name := range req.Categories {
		category, err := uc.repo.FindCategoryByName(ctx, name)
		if err != nil {
			return errors.New("product category not found: " + name)
		}
		links = append(links, entities.CraftProductCategory{CraftID: craft.ID, CategoryID: category.ID})
	}

	if err := uc.repo.SetCategories(ctx, craft.ID, links); err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "product_categories")

	return nil
}
