package create

import (
	"PocketArtisan/internal/entities"
	usersmod "PocketArtisan/internal/modules/users"
	"PocketArtisan/internal/modules/utils"
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo  usersmod.Repository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: usersmod.NewGormRepository(db), cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req Request) error {

	user, err := uc.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return errors.New("user not found")
	}

	craft, err := uc.repo.FindCraftByName(ctx, req.Craft)
	if err != nil {
		return errors.New("craft not found")
	}

	user.Role = "craftsman"

	if err := uc.repo.SaveUser(ctx, user); err != nil {
		return err
	}

	craftsman := &entities.Craftsman{
		UserID:  user.ID,
		CraftID: craft.ID,
	}

	if err := uc.repo.CreateCraftsman(ctx, craftsman); err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "users", "craftsmen")

	return nil

}
