package get_all

import (
	camod "PocketArtisan/internal/modules/craftsman_application"
	"context"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo  camod.Repository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: camod.NewGormRepository(db), cache: cache}
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

	totalCAs, err := uc.repo.CountTotal(ctx)
	if err != nil {
		return GetAllResponse{}, err
	}

	ca_list, err := uc.repo.ListPending(ctx, req.Skip, req.Limit)
	if err != nil {
		return GetAllResponse{}, err
	}

	resp := GetAllResponse{
		CraftsmanApplications: ca_list,
		Total:                 totalCAs,
		Page:                  (req.Skip / req.Limit) + 1,
	}

	return resp, nil

}
