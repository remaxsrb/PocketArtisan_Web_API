package product

import (
	"PocketArtisan/internal/entities"
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	// Craftsman lookups (used by product/service.go and product/create)
	FindCraftsmanByID(ctx context.Context, craftsmanID uint64) (*entities.Craftsman, error)
	FindCraftsmanByUsername(ctx context.Context, username string) (*entities.Craftsman, error)
	FindCraftsmanIDByUsername(ctx context.Context, username string) (uint64, error)

	// Category / craft-category link lookups (used by product/create)
	FindCategoryByName(ctx context.Context, name string) (*entities.ProductCategory, error)
	FindCraftCategoryLink(ctx context.Context, craftID, categoryID uint64) (*entities.CraftProductCategory, error)

	// Product writes
	ExistsByNameAndCraftsman(ctx context.Context, name string, craftsmanID uint64) (bool, error)
	Create(ctx context.Context, p *entities.Product) error
	FindByID(ctx context.Context, productID uint64) (*entities.Product, error)
	DeleteImages(ctx context.Context, productID uint64) error
	DeleteVideos(ctx context.Context, productID uint64) error
	Delete(ctx context.Context, p *entities.Product) error
	Save(ctx context.Context, p *entities.Product) error

	// Product reads
	CountByCraftsman(ctx context.Context, craftsmanID uint64) (int64, error)
	ListByCraftsman(ctx context.Context, craftsmanID uint64, skip, limit int) ([]*entities.Product, error)
	CountByCategory(ctx context.Context, normalizedSearch string) (int64, error)
	ListByCategory(ctx context.Context, normalizedSearch string, skip, limit int) ([]*entities.Product, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindCraftsmanByID(ctx context.Context, craftsmanID uint64) (*entities.Craftsman, error) {
	var c entities.Craftsman
	if err := r.db.WithContext(ctx).Where("id = ?", craftsmanID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *GormRepository) FindCraftsmanByUsername(ctx context.Context, username string) (*entities.Craftsman, error) {
	var c entities.Craftsman
	if err := r.db.WithContext(ctx).
		Joins("JOIN users ON users.id = craftsmen.user_id").
		Where("users.username = ?", username).
		First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("craftsman not found")
		}
		return nil, err
	}
	return &c, nil
}

func (r *GormRepository) FindCraftsmanIDByUsername(ctx context.Context, username string) (uint64, error) {
	c, err := r.FindCraftsmanByUsername(ctx, username)
	if err != nil {
		return 0, err
	}
	return c.ID, nil
}

func (r *GormRepository) FindCategoryByName(ctx context.Context, name string) (*entities.ProductCategory, error) {
	var pc entities.ProductCategory
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&pc).Error; err != nil {
		return nil, err
	}
	return &pc, nil
}

func (r *GormRepository) FindCraftCategoryLink(ctx context.Context, craftID, categoryID uint64) (*entities.CraftProductCategory, error) {
	var link entities.CraftProductCategory
	if err := r.db.WithContext(ctx).Where("craft_id = ? AND category_id = ?", craftID, categoryID).First(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *GormRepository) ExistsByNameAndCraftsman(ctx context.Context, name string, craftsmanID uint64) (bool, error) {
	var p entities.Product
	err := r.db.WithContext(ctx).Where("name = ? AND craftsman_id = ?", name, craftsmanID).First(&p).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func (r *GormRepository) Create(ctx context.Context, p *entities.Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *GormRepository) FindByID(ctx context.Context, productID uint64) (*entities.Product, error) {
	var p entities.Product
	if err := r.db.WithContext(ctx).Where("id = ?", productID).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *GormRepository) DeleteImages(ctx context.Context, productID uint64) error {
	return r.db.WithContext(ctx).Where("product_id = ?", productID).Delete(&entities.ProductImage{}).Error
}

func (r *GormRepository) DeleteVideos(ctx context.Context, productID uint64) error {
	return r.db.WithContext(ctx).Where("product_id = ?", productID).Delete(&entities.ProductVideo{}).Error
}

func (r *GormRepository) Delete(ctx context.Context, p *entities.Product) error {
	return r.db.WithContext(ctx).Delete(p).Error
}

func (r *GormRepository) Save(ctx context.Context, p *entities.Product) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *GormRepository) CountByCraftsman(ctx context.Context, craftsmanID uint64) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&entities.Product{}).Where("craftsman_id = ?", craftsmanID).Count(&total).Error
	return total, err
}

func (r *GormRepository) ListByCraftsman(ctx context.Context, craftsmanID uint64, skip, limit int) ([]*entities.Product, error) {
	raw := make([]*entities.Product, 0, limit)
	err := r.db.WithContext(ctx).
		Preload("Images").
		Preload("Videos").
		Preload("Category").
		Where("craftsman_id = ?", craftsmanID).
		Offset(skip).
		Limit(limit).
		Order("name asc").
		Find(&raw).Error
	return raw, err
}

func (r *GormRepository) CountByCategory(ctx context.Context, normalizedSearch string) (int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&entities.Product{})
	if normalizedSearch != "" {
		q = q.Joins("JOIN product_categories ON product_categories.id = products.category_id").
			Where("? = ANY(product_categories.search_keywords)", normalizedSearch)
	}
	return total, q.Count(&total).Error
}

func (r *GormRepository) ListByCategory(ctx context.Context, normalizedSearch string, skip, limit int) ([]*entities.Product, error) {
	raw := make([]*entities.Product, 0, limit)
	q := r.db.WithContext(ctx).
		Preload("Images").
		Preload("Videos").
		Preload("Category")
	if normalizedSearch != "" {
		q = q.Joins("JOIN product_categories ON product_categories.id = products.category_id").
			Where("? = ANY(product_categories.search_keywords)", normalizedSearch)
	}
	err := q.Offset(skip).Limit(limit).Order("name asc").Find(&raw).Error
	return raw, err
}
