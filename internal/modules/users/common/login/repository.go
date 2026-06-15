package login

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/users"
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	FindByUsername(ctx context.Context, username string) (*entities.User, error)
	UpdateLastLogin(ctx context.Context, user *entities.User) error
	GetCraftsmanProfileByUsername(ctx context.Context, username string) (*users.CraftsmanResponse, error)
	GetCraftsmanByUserID(ctx context.Context, userID uint64) (*entities.Craftsman, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	var user entities.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *gormRepository) UpdateLastLogin(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *gormRepository) GetCraftsmanProfileByUsername(ctx context.Context, username string) (*users.CraftsmanResponse, error) {
	var profile users.CraftsmanResponse
	if err := r.db.WithContext(ctx).
		Table("users").
		Select(`
        users.username,
        users.firstname,
        users.lastname,
        users.email,
        users.profile_picture,
        users.gender,
        crafts.name AS craft,
        craftsmen.rating,
        craftsmen.number_of_ratings
    `).
		Joins("INNER JOIN craftsmen ON craftsmen.user_id = users.id").
		Joins("INNER JOIN crafts ON crafts.id = craftsmen.craft_id").
		Where("users.username = ?", username).
		Scan(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *gormRepository) GetCraftsmanByUserID(ctx context.Context, userID uint64) (*entities.Craftsman, error) {
	var craftsman entities.Craftsman
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&craftsman).Error; err != nil {
		return nil, err
	}
	return &craftsman, nil
}
