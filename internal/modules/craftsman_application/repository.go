package craftsman_application

import (
	"PocketArtisan/internal/entities"
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	FindByID(ctx context.Context, id uint64) (*entities.CraftsmanApplication, error)
	Save(ctx context.Context, ca *entities.CraftsmanApplication) error
	CreateWithRateLimit(ctx context.Context, ca *entities.CraftsmanApplication, maxAttempts, lockoutDays int) error
	CountTotal(ctx context.Context) (int64, error)
	ListPending(ctx context.Context, skip, limit int) ([]*entities.CraftsmanApplication, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindByID(ctx context.Context, id uint64) (*entities.CraftsmanApplication, error) {
	var ca entities.CraftsmanApplication
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&ca).Error; err != nil {
		return nil, err
	}
	return &ca, nil
}

func (r *GormRepository) Save(ctx context.Context, ca *entities.CraftsmanApplication) error {
	return r.db.WithContext(ctx).Save(ca).Error
}

func (r *GormRepository) CreateWithRateLimit(ctx context.Context, ca *entities.CraftsmanApplication, maxAttempts, lockoutDays int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var attempts int64
		if err := tx.Model(&entities.CraftsmanApplication{}).
			Where("email = ?", ca.Email).
			Count(&attempts).Error; err != nil {
			return fmt.Errorf("could not get attempts for user with email %s: %w", ca.Email, err)
		}

		if attempts >= int64(maxAttempts) {
			return fmt.Errorf("max attempts of %d exceeded", maxAttempts)
		}

		var lastRejected entities.CraftsmanApplication
		err := tx.Where("email = ? AND status = ?", ca.Email, entities.StatusRejected).
			Order("created_at DESC").
			First(&lastRejected).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("could not get last rejected attempt for user with email %s: %w", ca.Email, err)
		}

		if err == nil {
			cutoff := time.Now().UTC().Add(-time.Duration(lockoutDays) * 24 * time.Hour)
			if lastRejected.CreatedAt.After(cutoff) {
				remaining := lastRejected.CreatedAt.Add(time.Duration(lockoutDays) * 24 * time.Hour).Sub(time.Now().UTC())
				return fmt.Errorf("you must wait %d days before re‑applying", int(math.Ceil(remaining.Hours()/24)))
			}
		}

		return tx.Create(ca).Error
	})
}

func (r *GormRepository) CountTotal(ctx context.Context) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&entities.CraftsmanApplication{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *GormRepository) ListPending(ctx context.Context, skip, limit int) ([]*entities.CraftsmanApplication, error) {
	list := make([]*entities.CraftsmanApplication, 0, limit)
	err := r.db.WithContext(ctx).
		Model(&entities.CraftsmanApplication{}).
		Where("status = ?", "pending").
		Offset(skip).
		Limit(limit).
		Order("created_at desc, id asc").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}
