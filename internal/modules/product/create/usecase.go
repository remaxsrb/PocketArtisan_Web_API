package create

import (
	"PocketArtisan/internal/modules/product"
	"PocketArtisan/internal/modules/product_categories"
	"context"
	"errors"
	"strings"

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

func (uc *UseCase) Execute(ctx context.Context, req NewProductRequest) (*product.ProductResponse, error) {
	var existing product.Product

	CraftsmanID, err := product.GetCraftsmanIDByUsername(ctx, uc.db, req.Username)
	if err != nil {
		return nil, err
	}

	var pc product_categories.ProductCategory
	if err := uc.db.WithContext(ctx).Where("name = ?", req.Category).First(&pc).Error; err != nil {	
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product category not found")
		}
		return nil, err
	}

	if err := uc.db.WithContext(ctx).Where("name = ? AND craftsman_id = ?", req.Name, CraftsmanID).First(&existing).Error; err == nil {
		return nil, errors.New("product already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	new_product := &product.Product{
		Name:        req.Name,
		CraftsmanID: CraftsmanID,
		CategoryID:  pc.ID,
		Price:       req.Price,
		Hidden:      false,
		Description: req.Description,
	}

	for _, url := range req.Images {
		new_product.Images = append(new_product.Images, product.ProductImage{URL: url})
	}
	for _, url := range req.Videos {
		new_product.Videos = append(new_product.Videos, product.ProductVideo{URL: url})
	}

	if err := uc.db.WithContext(ctx).Create(new_product).Error; err != nil {
		if strings.Contains(err.Error(), "idx_craftsman_product") || strings.Contains(err.Error(), "duplicate key") {
			return nil, errors.New("product already exists")
		}
		return nil, err
	}

	var imageUrls []string
	for _, img := range new_product.Images {
		imageUrls = append(imageUrls, img.URL)
	}

	var videoUrls []string
	for _, vid := range new_product.Videos {
		videoUrls = append(videoUrls, vid.URL)
	}

	response := &product.ProductResponse{
		ID:          new_product.ID,
		Name:        new_product.Name,
		Hidden:      new_product.Hidden,
		Price:       new_product.Price,
		Description: new_product.Description,
		Images:      imageUrls,
		Videos:      videoUrls,
	}

	return response, nil
}
