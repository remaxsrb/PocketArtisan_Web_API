package create

import (
	"PocketArtisan/internal/entities"
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

func (uc *Service) Execute(ctx context.Context, req NewCraftRequest) error {

	exists, err := uc.repo.ExistsByName(ctx, req.Name)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("craft already exists")
	}

	searchKeywords := make([]string, 0, len(req.Keywords)+1)
	searchKeywords = append(searchKeywords, utils.NormalizeForSearch(req.Name))
	for _, kw := range req.Keywords {
		searchKeywords = append(searchKeywords, utils.NormalizeForSearch(kw))
	}

	c := &entities.Craft{
		Name:           req.Name,
		Keywords:       req.Keywords,
		SearchKeywords: searchKeywords,
	}

	if err := uc.repo.Create(ctx, c); err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "craftsmen", "crafts")

	return nil

}
