package product

import (
	"PocketArtisan/internal/entities"
	"context"
	"errors"

	"gorm.io/gorm"
)

func GetCraftsmanIDByUsername(ctx context.Context, db *gorm.DB, username string) (uint64, error) {
	var craftsman entities.Craftsman
	if err := db.WithContext(ctx).
		Joins("JOIN users ON users.id = craftsmen.user_id").
		Where("users.username = ?", username).
		First(&craftsman).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("craftsman not found")
		}
		return 0, err
	}
	return craftsman.ID, nil
}
