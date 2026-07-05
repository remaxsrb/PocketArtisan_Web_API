package setcategories

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/utils"
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{db: db, cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req SetCraftCategoriesRequest) error {
	var craft entities.Craft
	if err := uc.db.WithContext(ctx).Where("name = ?", req.Craft).First(&craft).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("craft not found")
		}
		return err
	}

	categories := make([]entities.ProductCategory, 0, len(req.Categories))
	for _, name := range req.Categories {
		var category entities.ProductCategory
		if err := uc.db.WithContext(ctx).Where("name = ?", name).First(&category).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("product category not found: " + name)
			}
			return err
		}
		categories = append(categories, category)
	}

	links := make([]entities.CraftProductCategory, 0, len(categories))
	for _, category := range categories {
		links = append(links, entities.CraftProductCategory{CraftID: craft.ID, CategoryID: category.ID})
	}

	err := uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("craft_id = ?", craft.ID).Delete(&entities.CraftProductCategory{}).Error; err != nil {
			return err
		}
		if len(links) == 0 {
			return nil
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&links).Error
	})
	if err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "product_categories")

	return nil
}