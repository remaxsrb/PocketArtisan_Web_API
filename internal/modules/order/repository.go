package order

import (
	"PocketArtisan/internal/entities"
	"context"
	"time"

	"gorm.io/gorm"
)

type ProductPriceRow struct {
	ID    uint64
	Price float64
}

type CategoryCountRow struct {
	Category string
	Count    int64
}

type MonthlyCountRow struct {
	Month string
	Count int64
}

type MonthlyCategoryCountRow struct {
	Month    string
	Category string
	Count    int64
}

type Repository interface {
	FindByID(ctx context.Context, orderID uint64) (*entities.Order, error)
	Save(ctx context.Context, order *entities.Order) error

	FindProductPricesByCraftsman(ctx context.Context, productIDs []uint64, craftsmanID uint64) ([]ProductPriceRow, error)
	CreateOrderWithItemsAndCustomer(ctx context.Context, order *entities.Order, items []entities.OrderItem) (*entities.User, error)
	UpdatePaymentReservationID(ctx context.Context, orderID uint64, reservationID string) error
	UpdateURL(ctx context.Context, orderID uint64, url string) error
	FindOrderItemsWithProduct(ctx context.Context, orderID uint64) ([]entities.OrderItem, error)
	DeleteOrderWithItems(ctx context.Context, orderID uint64) error

	CountByCraftsman(ctx context.Context, craftsmanID uint64) (int64, error)
	ListByCraftsman(ctx context.Context, craftsmanID uint64, skip, limit int) ([]*entities.Order, error)
	ShippedCategoryCountsByCraftsman(ctx context.Context, craftsmanID uint64) ([]CategoryCountRow, error)

	CountByCustomer(ctx context.Context, customerID uint64) (int64, error)
	ListByCustomer(ctx context.Context, customerID uint64, skip, limit int) ([]*entities.Order, error)

	MonthlyShippedCounts(ctx context.Context, from, to *time.Time) ([]MonthlyCountRow, error)
	MonthlyShippedByCategory(ctx context.Context, craftsmanID uint64, from, to *time.Time) ([]MonthlyCategoryCountRow, error)

	ListPendingReviewReminders(ctx context.Context, shippedBefore time.Time) ([]*entities.Order, error)
	MarkReviewReminderSent(ctx context.Context, orderID uint64, sentAt time.Time) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindByID(ctx context.Context, orderID uint64) (*entities.Order, error) {
	var existing entities.Order
	if err := r.db.WithContext(ctx).Where("id = ?", orderID).First(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func (r *GormRepository) Save(ctx context.Context, order *entities.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *GormRepository) FindProductPricesByCraftsman(ctx context.Context, productIDs []uint64, craftsmanID uint64) ([]ProductPriceRow, error) {
	products := make([]ProductPriceRow, 0, len(productIDs))
	err := r.db.WithContext(ctx).Model(&entities.Product{}).
		Select("id, price").
		Where("id IN ? AND craftsman_id = ?", productIDs, craftsmanID).
		Find(&products).Error
	return products, err
}

func (r *GormRepository) CreateOrderWithItemsAndCustomer(ctx context.Context, order *entities.Order, items []entities.OrderItem) (*entities.User, error) {
	var customer entities.User
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		for i := range items {
			items[i].OrderID = order.ID
		}

		if err := tx.Create(&items).Error; err != nil {
			return err
		}

		if err := tx.First(&customer, order.CustomerID).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &customer, nil
}

func (r *GormRepository) UpdatePaymentReservationID(ctx context.Context, orderID uint64, reservationID string) error {
	return r.db.WithContext(ctx).Model(&entities.Order{}).Where("id = ?", orderID).Update("payment_reservation_id", reservationID).Error
}

func (r *GormRepository) UpdateURL(ctx context.Context, orderID uint64, url string) error {
	return r.db.WithContext(ctx).Model(&entities.Order{}).Where("id = ?", orderID).Update("url", url).Error
}

func (r *GormRepository) FindOrderItemsWithProduct(ctx context.Context, orderID uint64) ([]entities.OrderItem, error) {
	items := make([]entities.OrderItem, 0)
	err := r.db.WithContext(ctx).Preload("Product").Find(&items, "order_id = ?", orderID).Error
	return items, err
}

func (r *GormRepository) DeleteOrderWithItems(ctx context.Context, orderID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", orderID).Delete(&entities.OrderItem{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&entities.Order{}, orderID).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *GormRepository) CountByCraftsman(ctx context.Context, craftsmanID uint64) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&entities.Order{}).Where("craftsman_id = ?", craftsmanID).Count(&total).Error
	return total, err
}

func (r *GormRepository) ListByCraftsman(ctx context.Context, craftsmanID uint64, skip, limit int) ([]*entities.Order, error) {
	raw := make([]*entities.Order, 0, limit)
	err := r.db.WithContext(ctx).
		Where("craftsman_id = ?", craftsmanID).
		Offset(skip).
		Limit(limit).
		Order("created_at desc").
		Find(&raw).Error
	return raw, err
}

func (r *GormRepository) ShippedCategoryCountsByCraftsman(ctx context.Context, craftsmanID uint64) ([]CategoryCountRow, error) {
	categoryCounts := make([]CategoryCountRow, 0)
	err := r.db.WithContext(ctx).
		Table("orders").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Joins("JOIN product_categories ON product_categories.id = products.category_id").
		Select("product_categories.name as category, COUNT(DISTINCT orders.id) as count").
		Where("orders.craftsman_id = ? AND orders.status = ?", craftsmanID, entities.OrderShipped).
		Group("product_categories.name").
		Order("product_categories.name").
		Scan(&categoryCounts).Error
	return categoryCounts, err
}

func (r *GormRepository) CountByCustomer(ctx context.Context, customerID uint64) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&entities.Order{}).Where("customer_id = ?", customerID).Count(&total).Error
	return total, err
}

func (r *GormRepository) ListByCustomer(ctx context.Context, customerID uint64, skip, limit int) ([]*entities.Order, error) {
	raw := make([]*entities.Order, 0, limit)
	err := r.db.WithContext(ctx).
		Where("customer_id = ?", customerID).
		Offset(skip).
		Limit(limit).
		Order("created_at desc").
		Find(&raw).Error
	return raw, err
}

func (r *GormRepository) MonthlyShippedCounts(ctx context.Context, from, to *time.Time) ([]MonthlyCountRow, error) {
	query := r.db.WithContext(ctx).
		Model(&entities.Order{}).
		Select("to_char(date_trunc('month', completed_at), 'YYYY-MM') as month, COUNT(*) as count").
		Where("status = ?", entities.OrderShipped)

	if from != nil {
		query = query.Where("completed_at >= ?", *from)
	}
	if to != nil {
		query = query.Where("completed_at < ?", *to)
	}

	results := make([]MonthlyCountRow, 0)
	if err := query.Group("month").Order("month").Scan(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func (r *GormRepository) MonthlyShippedByCategory(ctx context.Context, craftsmanID uint64, from, to *time.Time) ([]MonthlyCategoryCountRow, error) {
	query := r.db.WithContext(ctx).
		Table("orders").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Joins("JOIN product_categories ON product_categories.id = products.category_id").
		Select(`
			to_char(date_trunc('month', orders.completed_at), 'YYYY-MM') as month,
			product_categories.name as category,
			COUNT(DISTINCT orders.id) as count
		`).
		Where("orders.craftsman_id = ? AND orders.status = ?", craftsmanID, entities.OrderShipped)

	if from != nil {
		query = query.Where("orders.completed_at >= ?", *from)
	}
	if to != nil {
		query = query.Where("orders.completed_at < ?", *to)
	}

	results := make([]MonthlyCategoryCountRow, 0)
	if err := query.Group("month, product_categories.name").Order("month").Scan(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func (r *GormRepository) ListPendingReviewReminders(ctx context.Context, shippedBefore time.Time) ([]*entities.Order, error) {
	var orders []*entities.Order
	err := r.db.WithContext(ctx).
		Where("status = ?", entities.OrderShipped).
		Where("shipped_at IS NOT NULL AND shipped_at <= ?", shippedBefore).
		Where("review_reminder_sent_at IS NULL").
		Where("NOT EXISTS (?)", r.db.
			Model(&entities.CraftsmanRatingRecord{}).
			Select("1").
			Where("craftsman_rating_records.customer_id = orders.customer_id AND craftsman_rating_records.craftsman_id = orders.craftsman_id"),
		).
		Find(&orders).Error
	return orders, err
}

func (r *GormRepository) MarkReviewReminderSent(ctx context.Context, orderID uint64, sentAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&entities.Order{}).
		Where("id = ?", orderID).
		Update("review_reminder_sent_at", sentAt).Error
}
