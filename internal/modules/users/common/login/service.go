package login

import (
	cartmod "PocketArtisan/internal/modules/cart"
	usersmod "PocketArtisan/internal/modules/users"
	"PocketArtisan/internal/modules/utils"
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo       usersmod.Repository
	cache      *redis.Client
	cartReader cartmod.CartReader
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{
		repo:       usersmod.NewGormRepository(db),
		cache:      cache,
		cartReader: cartmod.NewCartReader(db),
	}
}

func (uc *Service) Execute(ctx context.Context, req LoginRequest) (LoginResult, error) {

	existing, err := uc.repo.FindUserByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return LoginResult{}, ErrUsernameNotFound
		}
		return LoginResult{}, err
	}

	if !existing.CheckPassword(req.Password) {
		return LoginResult{}, ErrInvalidPassword
	}

	existing.LastLoginAt = time.Now()
	if err := uc.repo.SaveUser(ctx, existing); err != nil {
		return LoginResult{}, err
	}
	utils.BumpCacheVersion(ctx, uc.cache, "users")

	if existing.Role == "craftsman" {
		r, err := uc.repo.FindCraftsmanProfileByUsername(ctx, existing.Username)
		if err != nil {
			return LoginResult{}, err
		}

		craftsman, err := uc.repo.FindCraftsmanByUserID(ctx, existing.ID)
		if err != nil {
			return LoginResult{}, err
		}

		return LoginResult{ID: existing.ID, Role: existing.Role, CraftsmanID: craftsman.ID, Response: r}, nil
	}

	r := &usersmod.RegularUserResponse{
		UserResponse: usersmod.UserResponse{
			Username:       existing.Username,
			Firstname:      existing.Firstname,
			Lastname:       existing.Lastname,
			Email:          existing.Email,
			ProfilePicture: existing.ProfilePicture,
			Gender:         existing.Gender,
		},
	}

	userCart, cartErr := uc.cartReader.GetUserCart(ctx, existing.ID)
	if cartErr != nil {
		return LoginResult{}, cartErr
	}

	r.Cart = userCart

	return LoginResult{ID: existing.ID, Role: existing.Role, Response: r}, nil

}
