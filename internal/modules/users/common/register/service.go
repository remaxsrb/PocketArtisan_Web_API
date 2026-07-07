package register

import (
	"PocketArtisan/config"
	"PocketArtisan/internal/entities"
	usersmod "PocketArtisan/internal/modules/users"
	"PocketArtisan/internal/modules/utils"
	"PocketArtisan/internal/modules/utils/turnstile"
	"PocketArtisan/internal/validators"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo      usersmod.Repository
	cache     *redis.Client
	turnstile *turnstile.Verifier
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{
		repo:      usersmod.NewGormRepository(db),
		cache:     cache,
		turnstile: turnstile.NewVerifier(config.GetCrypto().TurnstileSecret),
	}
}

func (uc *Service) Execute(ctx context.Context, req RegisterRequest, remoteIP string) (*entities.User, error) {

	if os.Getenv("APP_ENV") == "production" {
		if _, err := uc.turnstile.Verify(ctx, req.TurnstileToken, remoteIP); err != nil {
			return nil, errors.New("captcha verification failed")
		}
	}

	if !validators.IsValidEmail(req.Email) {
		return nil, errors.New("invalid email")
	}

	if err := validators.ValidatePassword(req.Password); err != nil {
		return nil, errors.New(err.Error())
	}

	existsEmail, err := uc.repo.ExistsUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existsEmail {
		return nil, errors.New("email already used")
	}

	if uc.cache != nil {
		cacheVersion := utils.GetCacheVersion(ctx, uc.cache, "users")
		usernameCacheKey := fmt.Sprintf("user:username:v:%d:%s", cacheVersion, req.Username)
		if _, err := uc.cache.Get(ctx, usernameCacheKey).Result(); err == nil {
			return nil, errors.New("username already used")
		}
	}

	existsUsername, err := uc.repo.ExistsUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if existsUsername {
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

	switch req.Gender {
	case "male":
		user.ProfilePicture = defaultAvatarURL("default_male.png")
	case "female":
		user.ProfilePicture = defaultAvatarURL("default_female.png")
	}

	if err := user.SetPassword(req.Password); err != nil {
		return nil, err
	}

	if err := uc.repo.CreateUserWithCart(ctx, user); err != nil {
		return nil, err
	}

	utils.BumpCacheVersion(ctx, uc.cache, "users")

	return user, nil
}

func defaultAvatarURL(fileName string) string {
	base := "http://localhost:8080/api/assets/avatars"
	if publicURL := os.Getenv("R2_PUBLIC_URL"); publicURL != "" {
		base = strings.TrimRight(publicURL, "/") + "/avatars"
	}
	return base + "/" + fileName
}
