package users

import (
	"PocketArtisan/internal/entities"
	"context"
	"time"

	"gorm.io/gorm"
)

// Repository covers all user and craftsman data access.
// Modules that depend on only part of this interface should define their own
// narrower interface and accept it by value (Interface Segregation).
type Repository interface {
	// ── User ────────────────────────────────────────────────────────────────

	FindUserByID(ctx context.Context, id uint64) (*entities.User, error)
	FindUserByUsername(ctx context.Context, username string) (*entities.User, error)
	FindUserByEmail(ctx context.Context, email string) (*entities.User, error)
	ExistsUserByEmail(ctx context.Context, email string) (bool, error)
	ExistsUserByUsername(ctx context.Context, username string) (bool, error)
	SaveUser(ctx context.Context, user *entities.User) error
	DeleteUser(ctx context.Context, user *entities.User) error
	CreateUserWithCart(ctx context.Context, user *entities.User) error
	CountNonAdminUsers(ctx context.Context, from, to *time.Time) (int64, error)
	ListAllUsers(ctx context.Context, skip, limit int) ([]*entities.User, error)

	// ── Craftsman ───────────────────────────────────────────────────────────

	FindCraftsmanByUserID(ctx context.Context, userID uint64) (*entities.Craftsman, error)
	FindCraftsmanByID(ctx context.Context, craftsmanID uint64) (*entities.Craftsman, error)
	FindCraftsmanProfileByUsername(ctx context.Context, username string) (*CraftsmanResponse, error)
	FindCraftByName(ctx context.Context, name string) (*entities.Craft, error)
	SaveCraftsman(ctx context.Context, craftsman *entities.Craftsman) error
	CreateCraftsman(ctx context.Context, craftsman *entities.Craftsman) error
	CountCraftsmen(ctx context.Context, from, to *time.Time) (int64, error)
	ListCraftsmen(ctx context.Context, skip, limit int) ([]*CraftsmanResponse, error)
	CountCraftsmenByCraft(ctx context.Context, normalizedCraft string) (int64, error)
	ListCraftsmenByCraft(ctx context.Context, normalizedCraft string, skip, limit int) ([]*CraftsmanResponse, error)
	CountCraftsmenTotal(ctx context.Context) (int64, error)
	ListCraftsmenByRating(ctx context.Context, direction string, skip, limit int) ([]*CraftsmanResponse, error)

	// ── Rating ──────────────────────────────────────────────────────────────

	FindRatingRecord(ctx context.Context, customerID, craftsmanID uint64) (*entities.CraftsmanRatingRecord, error)
	RateCraftsman(ctx context.Context, craftsmanUserID uint64, customerID uint64, rating int) (*entities.Craftsman, error)
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

// ── User ────────────────────────────────────────────────────────────────────

func (r *GormRepository) FindUserByID(ctx context.Context, id uint64) (*entities.User, error) {
	var u entities.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *GormRepository) FindUserByUsername(ctx context.Context, username string) (*entities.User, error) {
	var u entities.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *GormRepository) FindUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	var u entities.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *GormRepository) ExistsUserByEmail(ctx context.Context, email string) (bool, error) {
	var u entities.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err == nil {
		return true, nil
	}
	if isNotFound(err) {
		return false, nil
	}
	return false, err
}

func (r *GormRepository) ExistsUserByUsername(ctx context.Context, username string) (bool, error) {
	var u entities.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if err == nil {
		return true, nil
	}
	if isNotFound(err) {
		return false, nil
	}
	return false, err
}

func (r *GormRepository) SaveUser(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *GormRepository) DeleteUser(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Delete(user).Error
}

func (r *GormRepository) CreateUserWithCart(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return tx.Create(&entities.Cart{UserID: user.ID}).Error
	})
}

func (r *GormRepository) CountNonAdminUsers(ctx context.Context, from, to *time.Time) (int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&entities.User{}).Where("role != ?", "admin")
	if from != nil {
		q = q.Where("created_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("created_at < ?", *to)
	}
	return total, q.Count(&total).Error
}

func (r *GormRepository) ListAllUsers(ctx context.Context, skip, limit int) ([]*entities.User, error) {
	list := make([]*entities.User, 0, limit)
	err := r.db.WithContext(ctx).
		Offset(skip).
		Limit(limit).
		Order("created_at desc, id asc").
		Find(&list).Error
	return list, err
}

// ── Craftsman ────────────────────────────────────────────────────────────────

func (r *GormRepository) FindCraftsmanByUserID(ctx context.Context, userID uint64) (*entities.Craftsman, error) {
	var c entities.Craftsman
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *GormRepository) FindCraftsmanByID(ctx context.Context, craftsmanID uint64) (*entities.Craftsman, error) {
	var c entities.Craftsman
	if err := r.db.WithContext(ctx).Where("id = ?", craftsmanID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *GormRepository) FindCraftsmanProfileByUsername(ctx context.Context, username string) (*CraftsmanResponse, error) {
	var profile CraftsmanResponse
	err := r.db.WithContext(ctx).
		Table("users").
		Select(`
			users.username,
			users.firstname,
			users.lastname,
			users.email,
			users.profile_picture,
			users.gender,
			craftsmen.id as craftsman_id,
			crafts.name AS craft,
			craftsmen.rating,
			craftsmen.number_of_ratings
		`).
		Joins("INNER JOIN craftsmen ON craftsmen.user_id = users.id").
		Joins("INNER JOIN crafts ON crafts.id = craftsmen.craft_id").
		Where("users.username = ?", username).
		Scan(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *GormRepository) FindCraftByName(ctx context.Context, name string) (*entities.Craft, error) {
	var c entities.Craft
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *GormRepository) SaveCraftsman(ctx context.Context, craftsman *entities.Craftsman) error {
	return r.db.WithContext(ctx).Save(craftsman).Error
}

func (r *GormRepository) CreateCraftsman(ctx context.Context, craftsman *entities.Craftsman) error {
	return r.db.WithContext(ctx).Create(craftsman).Error
}

func (r *GormRepository) CountCraftsmen(ctx context.Context, from, to *time.Time) (int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&entities.Craftsman{})
	if from != nil {
		q = q.Where("approved_at >= ?", *from)
	}
	if to != nil {
		q = q.Where("approved_at < ?", *to)
	}
	return total, q.Count(&total).Error
}

func (r *GormRepository) ListCraftsmen(ctx context.Context, skip, limit int) ([]*CraftsmanResponse, error) {
	list := make([]*CraftsmanResponse, 0, limit)
	err := r.db.WithContext(ctx).
		Table("users").
		Select(`
			users.firstname,
			users.lastname,
			users.username,
			users.email,
			users.profile_picture,
			craftsmen.id as craftsman_id,
			crafts.name as craft,
			craftsmen.rating,
			craftsmen.number_of_ratings
		`).
		Joins("INNER JOIN craftsmen ON craftsmen.user_id = users.id").
		Joins("INNER JOIN crafts ON crafts.id = craftsmen.craft_id").
		Where("users.role = ?", "craftsman").
		Offset(skip).
		Limit(limit).
		Order("users.created_at desc, users.id asc").
		Scan(&list).Error
	return list, err
}

func (r *GormRepository) CountCraftsmenByCraft(ctx context.Context, normalizedCraft string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Table("users").
		Joins("INNER JOIN craftsmen ON craftsmen.user_id = users.id").
		Joins("INNER JOIN crafts ON crafts.id = craftsmen.craft_id").
		Where("users.role = ?", "craftsman").
		Where("? = ANY(crafts.search_keywords)", normalizedCraft).
		Count(&total).Error
	return total, err
}

func (r *GormRepository) ListCraftsmenByCraft(ctx context.Context, normalizedCraft string, skip, limit int) ([]*CraftsmanResponse, error) {
	list := make([]*CraftsmanResponse, 0, limit)
	err := r.db.WithContext(ctx).
		Table("users").
		Select(`
			users.firstname,
			users.lastname,
			users.username,
			users.email,
			users.profile_picture,
			craftsmen.id as craftsman_id,
			crafts.name as craft,
			craftsmen.rating,
			craftsmen.number_of_ratings
		`).
		Joins("INNER JOIN craftsmen ON craftsmen.user_id = users.id").
		Joins("INNER JOIN crafts ON crafts.id = craftsmen.craft_id").
		Where("users.role = ?", "craftsman").
		Where("? = ANY(crafts.search_keywords)", normalizedCraft).
		Offset(skip).
		Limit(limit).
		Order("users.created_at desc, users.id asc").
		Scan(&list).Error
	return list, err
}

func (r *GormRepository) CountCraftsmenTotal(ctx context.Context) (int64, error) {
	var total int64
	return total, r.db.WithContext(ctx).Model(&entities.Craftsman{}).Count(&total).Error
}

func (r *GormRepository) ListCraftsmenByRating(ctx context.Context, direction string, skip, limit int) ([]*CraftsmanResponse, error) {
	list := make([]*CraftsmanResponse, 0, limit)
	err := r.db.WithContext(ctx).
		Table("users").
		Select(`
			users.firstname,
			users.lastname,
			users.username,
			users.email,
			users.profile_picture,
			craftsmen.id as craftsman_id,
			crafts.name as craft,
			craftsmen.rating,
			craftsmen.number_of_ratings
		`).
		Joins("INNER JOIN craftsmen ON craftsmen.user_id = users.id").
		Joins("INNER JOIN crafts ON crafts.id = craftsmen.craft_id").
		Where("users.role = ?", "craftsman").
		Offset(skip).
		Limit(limit).
		Order("craftsmen.rating " + direction).
		Scan(&list).Error
	return list, err
}

// ── Rating ───────────────────────────────────────────────────────────────────

func (r *GormRepository) FindRatingRecord(ctx context.Context, customerID, craftsmanID uint64) (*entities.CraftsmanRatingRecord, error) {
	var rec entities.CraftsmanRatingRecord
	if err := r.db.WithContext(ctx).Where("customer_id = ? AND craftsman_id = ?", customerID, craftsmanID).First(&rec).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *GormRepository) RateCraftsman(ctx context.Context, craftsmanUserID uint64, customerID uint64, rating int) (*entities.Craftsman, error) {
	var craftsman entities.Craftsman
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", craftsmanUserID).First(&craftsman).Error; err != nil {
			return err
		}

		var existing entities.CraftsmanRatingRecord
		err := tx.Where("customer_id = ? AND craftsman_id = ?", customerID, craftsmanUserID).First(&existing).Error
		if err == nil {
			return errAlreadyRated
		}
		if !isNotFound(err) {
			return err
		}

		record := entities.CraftsmanRatingRecord{CustomerID: customerID, CraftsmanID: craftsmanUserID}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		craftsman.Rating = ((craftsman.Rating * float64(craftsman.NumberOfRatings)) + float64(rating)) / float64(craftsman.NumberOfRatings+1)
		craftsman.NumberOfRatings++
		return tx.Save(&craftsman).Error
	})
	if err != nil {
		return nil, err
	}
	return &craftsman, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

var errAlreadyRated = errAlreadyRatedType{}

type errAlreadyRatedType struct{}

func (errAlreadyRatedType) Error() string { return "you have already rated this craftsman" }

func IsAlreadyRatedError(err error) bool {
	_, ok := err.(errAlreadyRatedType)
	return ok
}

func isNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}
