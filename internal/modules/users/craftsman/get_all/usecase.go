package getall

import (
	"PocketArtisan/internal/modules/users"
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

	var craftsman_list []users.User

	var totalCraftsmen int64
	uc.db.WithContext(ctx).Model(&users.Craftsman{}).Count(&totalCraftsmen)

	uc.db.WithContext(ctx).
		Preload("Craftsman").Where("role = ?", "craftsman").
		Model(&users.User{}).
		Offset(req.Skip).
		Limit(req.Limit).
		Order("created_at desc, id asc").
		Find(&craftsman_list)

	resp := GetAllResponse{
		Craftsmen: craftsman_list,
		Total:     totalCraftsmen,
		Page:      (req.Skip / req.Limit) + 1,
	}

	return resp, nil

}
