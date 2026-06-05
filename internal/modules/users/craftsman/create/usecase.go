package create

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
	var existing_craftsman users.Craftsman

	if err := uc.db.WithContext(ctx).Where("email = ?", req.Email).First(&existing_craftsman).Error; err == nil {
		return errors.New("craftsman already exists with this email")
	}

	//find userID based on email

	var user users.User
	if err := uc.db.WithContext(ctx).Where("email = ?", req.Email).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	craftsman := users.Craftsman{
		UserID: user.ID,
		Craft:  req.Craft,
	}

	if err := uc.db.Create(&craftsman).Error; err != nil {
		return err
	}

	return nil

}
