package get_all

import (
	usersmod "PocketArtisan/internal/modules/users"
	"context"

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

func (uc *Service) Execute(ctx context.Context, req GetAllRequest) (GetAllResponse, error) {

	const maxLimit = 100
	const defaultLimit = 20

	if req.Limit <= 0 {
		req.Limit = defaultLimit
	}
	if req.Limit > maxLimit {
		req.Limit = maxLimit
	}

	totalUsers, err := uc.repo.CountNonAdminUsers(ctx, nil, nil)
	if err != nil {
		return GetAllResponse{}, err
	}

	rawUsers, err := uc.repo.ListAllUsers(ctx, req.Skip, req.Limit)
	if err != nil {
		return GetAllResponse{}, err
	}

	user_list := make([]*usersmod.RegularUserResponse, 0, len(rawUsers))
	for _, u := range rawUsers {
		user_list = append(user_list, &usersmod.RegularUserResponse{
			UserResponse: usersmod.UserResponse{
				Username:       u.Username,
				Firstname:      u.Firstname,
				Lastname:       u.Lastname,
				Email:          u.Email,
				ProfilePicture: u.ProfilePicture,
				Gender:         u.Gender,
				Role:           u.Role,
			},
		})
	}

	resp := GetAllResponse{
		Users: user_list,
		Total: totalUsers,
		Page:  (req.Skip / req.Limit) + 1,
	}

	return resp, nil

}
