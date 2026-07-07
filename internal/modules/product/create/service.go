package create

import (
	"PocketArtisan/internal/entities"
	prodmod "PocketArtisan/internal/modules/product"
	"PocketArtisan/internal/modules/utils"
	"context"
	"errors"
	"strings"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo  prodmod.Repository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: prodmod.NewGormRepository(db), cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req NewProductRequest) error {
	craftsman, err := uc.repo.FindCraftsmanByID(ctx, req.CraftsmanID)
	if err != nil {
		return errors.New("craftsman not found")
	}

	pc, err := uc.repo.FindCategoryByName(ctx, req.Category)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("product category not found")
		}
		return err
	}

	_, err = uc.repo.FindCraftCategoryLink(ctx, craftsman.CraftID, pc.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("product category is not related to your craft")
		}
		return err
	}

	exists, err := uc.repo.ExistsByNameAndCraftsman(ctx, req.Name, craftsman.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("product already exists")
	}

	new_product := &entities.Product{
		Name:        req.Name,
		CraftsmanID: craftsman.ID,
		CategoryID:  pc.ID,
		Price:       req.Price,
		Hidden:      false,
		Description: req.Description,
	}

	for _, url := range req.Images {
		new_product.Images = append(new_product.Images, entities.ProductImage{URL: url})
	}
	for _, url := range req.Videos {
		new_product.Videos = append(new_product.Videos, entities.ProductVideo{URL: url})
	}

	if err := uc.repo.Create(ctx, new_product); err != nil {
		if strings.Contains(err.Error(), "idx_craftsman_product") || strings.Contains(err.Error(), "duplicate key") {
			return errors.New("product already exists")
		}
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "products")

	return nil
}
