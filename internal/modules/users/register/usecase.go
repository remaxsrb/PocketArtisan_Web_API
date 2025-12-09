package register

import (
	"PocketArtisan/internal/modules/users/common"
	"PocketArtisan/internal/modules/users/validator"
	"context"
	"errors"

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

func (uc *UseCase) Execute(ctx context.Context, req RegisterRequest) (*common.User, error) {
	var existing common.User

	if !validator.IsValidEmail(req.Email) {
		return nil, errors.New("invalid email")
	}

	if err := validator.ValidatePassword(req.Password); err != nil {
		return nil, errors.New(err.Error())
	}

	if err := uc.db.WithContext(ctx).Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return nil, errors.New("email already used")
	}

	if err := uc.db.WithContext(ctx).Where("username = ?", req.Username).First(&existing).Error; err == nil {
		return nil, errors.New("username already used")
	}

	user := &common.User{
		Username:    req.Username,
		Email:       req.Email,
		Firstname:   req.Firstname,
		Lastname:    req.Lastname,
		DateOfBirth: req.DateOfBirth,
		Role:        "users",
		Gender:      req.Gender,
	}

	if req.Gender == "male" {
		user.ProfilePicture = "http://localhost:8080/files/avatars/default_male.png"
	} else {
		user.ProfilePicture = "http://localhost:8080/files/avatars/default_female.png"
	}

	if err := user.SetPassword(req.Password); err != nil {
		return nil, err
	}
	if err := uc.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}
