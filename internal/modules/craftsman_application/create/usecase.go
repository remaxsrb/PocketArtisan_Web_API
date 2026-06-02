package craftsman_application

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

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

func (uc *UseCase) Execute(ctx context.Context, req CraftsmanApplicationDTO) error {

	const maxAttemptsPerUser = 3
	const lockoutDays = 90

	uc.db = uc.db.WithContext(ctx)

	return uc.db.Transaction(func(tx *gorm.DB) error {

		var attempts int64

		if err := tx.Model(&CraftsmanApplication{}).
			Where("user_id = ?", req.ID).
			Count(&attempts).Error; err != nil {
			return fmt.Errorf("could not get attempts for user %d: %w", req.ID, err)
		}

		if attempts >= maxAttemptsPerUser {
			return fmt.Errorf("max attempts of %d exceeded", maxAttemptsPerUser)
		}

		var lastRejectedAttempt CraftsmanApplication
		err := tx.
			Where("user_id = ? AND status = ?", req.ID, StatusRejected).
			Order("created_at DESC").
			First(&lastRejectedAttempt).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("could not get last rejected attempt for user %d: %w", req.ID, err)
		}

		if err == nil {
			cutoff := time.Now().UTC().Add(-lockoutDays * 24 * time.Hour)
			if lastRejectedAttempt.CreatedAt.After(cutoff) {
				remaining := lastRejectedAttempt.CreatedAt.Add(lockoutDays * 24 * time.Hour).Sub(time.Now().UTC())

				return fmt.Errorf("you must wait %d days before re‑applying", int(math.Ceil(remaining.Hours()/24)))
			}
		}

		newCA := CraftsmanApplication{
			ID:     req.ID,
			Status: StatusPending,
		}
		if err := tx.Create(&newCA).Error; err != nil {
			return fmt.Errorf("failed to create application: %w", err)
		}

		return nil
	})
}
