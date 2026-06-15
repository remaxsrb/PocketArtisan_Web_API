package get_all

import (
	"PocketArtisan/internal/entities"
	"context"

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

func (uc *Service) Execute(ctx context.Context, req GetAllRequest) (GetAllResponse, error) {

	const maxLimit = 100
	const defaultLimit = 20

	if req.Limit <= 0 {
		req.Limit = defaultLimit
	}
	if req.Limit > maxLimit {
		req.Limit = maxLimit
	}

	ca_list := make([]*entities.CraftsmanApplication, 0, req.Limit)

	var totalCAs int64
	uc.db.WithContext(ctx).Model(&entities.CraftsmanApplication{}).Count(&totalCAs)

	uc.db.WithContext(ctx).
		Model(&entities.CraftsmanApplication{}).
		Where("status = ?", "pending").
		Offset(req.Skip).
		Limit(req.Limit).
		Order("created_at desc, id asc").
		Find(&ca_list)

	resp := GetAllResponse{
		CraftsmanApplications: ca_list,
		Total:                 totalCAs,
		Page:                  (req.Skip / req.Limit) + 1,
	}

	return resp, nil

}
