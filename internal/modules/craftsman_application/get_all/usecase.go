package get_all

import (
	"PocketArtisan/internal/modules/craftsman_application"
	"context"

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

func (uc *UseCase) Execute(ctx context.Context, req GetAllRequest) (GetAllResponse, error) {

	const maxLimit = 100
	const defaultLimit = 20

	if req.Limit <= 0 {
		req.Limit = defaultLimit
	}
	if req.Limit > maxLimit {
		req.Limit = maxLimit
	}

	ca_list := make([]*craftsman_application.CraftsmanApplication, 0, req.Limit)

	var totalCAs int64
	uc.db.WithContext(ctx).Model(&craftsman_application.CraftsmanApplication{}).Count(&totalCAs)
	
	uc.db.WithContext(ctx).
		Model(&craftsman_application.CraftsmanApplication{}).
		Offset(req.Skip).
		Limit(req.Limit).
		Order("created_at desc, id asc").
		Find(&ca_list)

	resp := GetAllResponse{
		CraftsmanApplications: ca_list,
		Total: totalCAs,
		Page:  (req.Skip / req.Limit) + 1,
	}

	return resp, nil

}
