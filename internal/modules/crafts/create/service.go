package create

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/utils"
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{db: db, cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req NewCraftRequest) error {

	var c entities.Craft
	if err := uc.db.WithContext(ctx).Where("name = ?", req.Name).First(&c).Error; err == nil {
		return errors.New("craft already exists")
	}

	searchKeywords := make([]string, 0, len(req.Keywords)+1)
	searchKeywords = append(searchKeywords, utils.NormalizeForSearch(req.Name))
	for _, kw := range req.Keywords {
		searchKeywords = append(searchKeywords, utils.NormalizeForSearch(kw))
	}

	c = entities.Craft{
		Name:           req.Name,
		Keywords:       req.Keywords,
		SearchKeywords: searchKeywords,
	}

	if err := uc.db.Create(&c).Error; err != nil {
		return err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "craftsmen", "crafts")

	return nil

}
