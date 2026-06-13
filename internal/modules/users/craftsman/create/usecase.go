package create

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/utils"
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

	//find userID based on email

	var user entities.User
	if err := uc.db.WithContext(ctx).Where("email = ?", req.Email).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	var craft entities.Craft
	if err := uc.db.WithContext(ctx).Where("name = ?", req.Craft).First(&craft).Error; err != nil {
		return errors.New("craft not found")
	}

	user.Role = "craftsman"

	if err := uc.db.Save(&user).Error; err != nil {
		return err
	}

	craftsman := entities.Craftsman{
		UserID:  user.ID,
		CraftID: craft.ID,
	}

	if err := uc.db.Create(&craftsman).Error; err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "users", "craftsmen")

	return nil

}
