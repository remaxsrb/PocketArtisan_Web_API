package delete_account

import (
	"PocketArtisan/internal/modules/users"
	"context"
	"errors"

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

func (uc *UseCase) Execute(ctx context.Context, req DeleteAccountRequest) error {

	var existing users.User

	if err := uc.db.WithContext(ctx).Where("username = ?", req.Username).First(&existing).Error; err != nil {
		return errors.New("username not found")
	}

	if err := uc.db.Delete(&existing).Error; err != nil {
		return err
	}

	return nil
}
