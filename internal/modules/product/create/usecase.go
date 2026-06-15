package create

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/product"
	"PocketArtisan/internal/modules/utils"
	"context"
	"errors"
	"strings"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type UseCase struct {
	db             *gorm.DB
	cache          *redis.Client
	productService product.Service
}

func NewUseCase(db *gorm.DB, cache *redis.Client) *UseCase {
	return &UseCase{db: db, cache: cache, productService: product.NewService(db)}
}

func (uc *UseCase) Execute(ctx context.Context, req NewProductRequest) error {
	var existing entities.Product

	CraftsmanID, err := uc.productService.GetCraftsmanIDByUsername(ctx, req.Username)
	if err != nil {
		return err
	}

	var pc entities.ProductCategory
	if err := uc.db.WithContext(ctx).Where("name = ?", req.Category).First(&pc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("product category not found")
		}
		return err
	}

	if err := uc.db.WithContext(ctx).Where("name = ? AND craftsman_id = ?", req.Name, CraftsmanID).First(&existing).Error; err == nil {
		return errors.New("product already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	new_product := &entities.Product{
		Name:        req.Name,
		CraftsmanID: CraftsmanID,
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

	if err := uc.db.WithContext(ctx).Create(new_product).Error; err != nil {
		if strings.Contains(err.Error(), "idx_craftsman_product") || strings.Contains(err.Error(), "duplicate key") {
			return errors.New("product already exists")
		}
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "products")

	var imageUrls []string
	for _, img := range new_product.Images {
		imageUrls = append(imageUrls, img.URL)
	}

	var videoUrls []string
	for _, vid := range new_product.Videos {
		videoUrls = append(videoUrls, vid.URL)
	}

	return nil
}
