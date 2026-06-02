package approve_craftsman_application

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

func (uc *UseCase) Execute(ctx context.Context, req Request) error {
	var ca users.User

	if err := uc.db.WithContext(ctx).Where("username = ?", req.Username).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	user.Role = req.Role

	if err := uc.db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}
