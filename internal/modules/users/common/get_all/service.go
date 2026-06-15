package get_all

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/users"
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

	user_list := make([]*users.RegularUserResponse, 0, req.Limit)

	var totalUsers int64
	uc.db.WithContext(ctx).Model(&entities.User{}).Count(&totalUsers)

	uc.db.WithContext(ctx).
		Model(&entities.User{}).
		Offset(req.Skip).
		Limit(req.Limit).
		Order("created_at desc, id asc").
		Find(&user_list)

	resp := GetAllResponse{
		Users: user_list,
		Total: totalUsers,
		Page:  (req.Skip / req.Limit) + 1,
	}

	return resp, nil

}
