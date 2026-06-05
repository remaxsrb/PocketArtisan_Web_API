package create

import (
	"PocketArtisan/internal/modules/craftsman_application"
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

func (uc *UseCase) Execute(ctx context.Context, req CraftsmanApplicationRequest) error {

	const maxAttemptsPerUser = 3
	const lockoutDays = 90

	uc.db = uc.db.WithContext(ctx)

	return uc.db.Transaction(func(tx *gorm.DB) error {

		var attempts int64

		if err := tx.Model(&craftsman_application.CraftsmanApplication{}).
			Where("email = ?", req.Email).
			Count(&attempts).Error; err != nil {
			return fmt.Errorf("could not get attempts for user with email %s: %w", req.Email, err)
		}

		if attempts >= maxAttemptsPerUser {
			return fmt.Errorf("max attempts of %d exceeded", maxAttemptsPerUser)
		}

		var lastRejectedAttempt craftsman_application.CraftsmanApplication
		err := tx.
			Where("email = ? AND status = ?", req.Email, craftsman_application.StatusRejected).
			Order("created_at DESC").
			First(&lastRejectedAttempt).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("could not get last rejected attempt for user with email %s: %w", req.Email, err)
		}

		if err == nil {
			cutoff := time.Now().UTC().Add(-lockoutDays * 24 * time.Hour)
			if lastRejectedAttempt.CreatedAt.After(cutoff) {
				remaining := lastRejectedAttempt.CreatedAt.Add(lockoutDays * 24 * time.Hour).Sub(time.Now().UTC())

				return fmt.Errorf("you must wait %d days before re‑applying", int(math.Ceil(remaining.Hours()/24)))
			}
		}

		newCA := craftsman_application.CraftsmanApplication{
			Email:  req.Email,
			Status: craftsman_application.StatusPending,
			Craft:  req.Craft,
		}
		if err := tx.Create(&newCA).Error; err != nil {
			return fmt.Errorf("failed to create application: %w", err)
		}

		return nil
	})
}
