package delete

import (
	craftsmod "PocketArtisan/internal/modules/crafts"
	"PocketArtisan/internal/modules/utils"
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo  craftsmod.Repository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: craftsmod.NewGormRepository(db), cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req DeleteCraftRequest) error {

	c, err := uc.repo.FindByName(ctx, req.Name)
	if err != nil {
		return errors.New("craft does not exist")
	}

	if err := uc.repo.Delete(ctx, c); err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "craftsmen", "crafts")

	return nil

}
