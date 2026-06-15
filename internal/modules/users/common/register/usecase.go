package register

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/utils"
	"PocketArtisan/internal/validators"
	"context"
	"errors"
	"fmt"
	"time"

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

func (uc *UseCase) Execute(ctx context.Context, req RegisterRequest) (*entities.User, error) {
	var existing entities.User

	if !validators.IsValidEmail(req.Email) {
		return nil, errors.New("invalid email")
	}

	if err := validators.ValidatePassword(req.Password); err != nil {
		return nil, errors.New(err.Error())
	}

	if err := uc.db.WithContext(ctx).Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return nil, errors.New("email already used")
	}

	if uc.cache != nil {
		cacheVersion := utils.GetCacheVersion(ctx, uc.cache, "users")
		usernameCacheKey := fmt.Sprintf("user:username:v:%d:%s", cacheVersion, req.Username)
		if _, err := uc.cache.Get(ctx, usernameCacheKey).Result(); err == nil {
			return nil, errors.New("username already used")
		}
	}

	if err := uc.db.WithContext(ctx).Where("username = ?", req.Username).First(&existing).Error; err == nil {
		return nil, errors.New("username already used")
	}

	dob, err := time.Parse("2006-01-02", req.DateOfBirth)

	if err != nil {
		return nil, errors.New("invalid date of birth string")
	}

	user := &entities.User{
		Username:    req.Username,
		Email:       req.Email,
		Firstname:   req.Firstname,
		Lastname:    req.Lastname,
		DateOfBirth: dob,
		Gender:      req.Gender,
		Role:        req.Role,
	}

	if req.Gender == "male" {
		user.ProfilePicture = "http://localhost:8080/assets/avatars/default_male.png"
	} else {
		user.ProfilePicture = "http://localhost:8080/assets/avatars/default_female.png"
	}

	if err := user.SetPassword(req.Password); err != nil {
		return nil, err
	}

	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		newCart := &entities.Cart{UserID: user.ID}
		return tx.Create(newCart).Error
	})
	if err != nil {
		return nil, err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "users")

	return user, nil
}
