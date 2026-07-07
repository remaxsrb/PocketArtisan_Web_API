package set_biography

import (
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

func (uc *Service) Execute(ctx context.Context, req SetBiographyRequest) error {

	craftsman, err := uc.repo.FindCraftsmanByID(ctx, req.CraftsmanID)
	if err != nil {
		return errors.New("craftsman not found")
	}

	craftsman.Biography = req.Biography

	if err := uc.repo.SaveCraftsman(ctx, craftsman); err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "craftsmen")

	return nil
}
