package product_categories

import (
	"PocketArtisan/internal/entities"
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	ExistsByName(ctx context.Context, name string) (bool, error)
	FindByName(ctx context.Context, name string) (*entities.ProductCategory, error)
	Create(ctx context.Context, pc *entities.ProductCategory) error
	Delete(ctx context.Context, pc *entities.ProductCategory) error
	FindAll(ctx context.Context) ([]entities.ProductCategory, error)
	FindByCraftID(ctx context.Context, craftID uint64) ([]entities.ProductCategory, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var pc entities.ProductCategory
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&pc).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func (r *GormRepository) FindByName(ctx context.Context, name string) (*entities.ProductCategory, error) {
	var pc entities.ProductCategory
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&pc).Error; err != nil {
		return nil, err
	}
	return &pc, nil
}

func (r *GormRepository) Create(ctx context.Context, pc *entities.ProductCategory) error {
	return r.db.WithContext(ctx).Create(pc).Error
}

func (r *GormRepository) Delete(ctx context.Context, pc *entities.ProductCategory) error {
	return r.db.WithContext(ctx).Delete(pc).Error
}

func (r *GormRepository) FindAll(ctx context.Context) ([]entities.ProductCategory, error) {
	var list []entities.ProductCategory
	if err := r.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GormRepository) FindByCraftID(ctx context.Context, craftID uint64) ([]entities.ProductCategory, error) {
	var list []entities.ProductCategory
	err := r.db.WithContext(ctx).
		Table("product_categories").
		Joins("INNER JOIN craft_product_categories ON craft_product_categories.category_id = product_categories.id").
		Where("craft_product_categories.craft_id = ?", craftID).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}
