package crafts

import (
	"PocketArtisan/internal/entities"
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	ExistsByName(ctx context.Context, name string) (bool, error)
	FindByName(ctx context.Context, name string) (*entities.Craft, error)
	Create(ctx context.Context, craft *entities.Craft) error
	Delete(ctx context.Context, craft *entities.Craft) error
	FindAll(ctx context.Context) ([]entities.Craft, error)

	FindCategoryByName(ctx context.Context, name string) (*entities.ProductCategory, error)
	SetCategories(ctx context.Context, craftID uint64, links []entities.CraftProductCategory) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var c entities.Craft
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&c).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func (r *GormRepository) FindByName(ctx context.Context, name string) (*entities.Craft, error) {
	var c entities.Craft
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *GormRepository) Create(ctx context.Context, craft *entities.Craft) error {
	return r.db.WithContext(ctx).Create(craft).Error
}

func (r *GormRepository) Delete(ctx context.Context, craft *entities.Craft) error {
	return r.db.WithContext(ctx).Delete(craft).Error
}

func (r *GormRepository) FindAll(ctx context.Context) ([]entities.Craft, error) {
	var list []entities.Craft
	if err := r.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *GormRepository) FindCategoryByName(ctx context.Context, name string) (*entities.ProductCategory, error) {
	var pc entities.ProductCategory
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&pc).Error; err != nil {
		return nil, err
	}
	return &pc, nil
}

func (r *GormRepository) SetCategories(ctx context.Context, craftID uint64, links []entities.CraftProductCategory) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("craft_id = ?", craftID).Delete(&entities.CraftProductCategory{}).Error; err != nil {
			return err
		}
		if len(links) == 0 {
			return nil
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&links).Error
	})
}
