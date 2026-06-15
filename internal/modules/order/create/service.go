package create

import (
	"PocketArtisan/internal/entities"
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	db    *gorm.DB
	cache *redis.Client
}

type ProductPrice struct {
	ID    uint
	Price float64
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{db: db, cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req NewOrderRequest) error {

	var order entities.Order
	order.CustomerID = ctx.Value("user_id").(uint)
	order.CraftsmanID = req.CraftsmanID

	switch req.PaymentType {
	case "CREDIT_CARD":
		order.PaymentType = entities.PaymentCreditCard
		order.Status = entities.OrderPaymentReserved
	case "CASH_ON_DELIVERY":
		order.PaymentType = entities.CashOnDelivery
		order.Status = entities.OrderPending
	default:
		return fmt.Errorf("invalid payment type: %s", req.PaymentType)
	}

	productIDs := make([]uint, len(req.Items))
	quantities := make(map[uint]int)
	// Map to store product prices in case craftsman changes them during the order creation process
	prices := make(map[uint]float64)

	for i, item := range req.Items {
		productIDs[i] = item.ProductID
		quantities[item.ProductID] = item.Quantity
	}

	var products []ProductPrice

	err := uc.db.Model(&entities.Product{}).
		Select("id, price").
		Where("id IN ? AND craftsman_id = ?", productIDs, req.CraftsmanID).
		Find(&products).Error

	if err != nil {
		return fmt.Errorf("failed to fetch product prices: %w", err)
	}

	if len(products) != len(productIDs) {
		return fmt.Errorf("one or more products do not exist")
	}

	total := 0.0

	for _, p := range products {
		prices[p.ID] = p.Price
		total += p.Price * float64(quantities[p.ID])
	}

	order.TotalPrice = total

	orderItems := make([]entities.OrderItem, len(req.Items))

	for i, item := range req.Items {
		orderItems[i] = entities.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: prices[item.ProductID],
		}
	}

	// transaction to ensure atomicity
	err = uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&order).Error; err != nil {
			return fmt.Errorf("create order: %w", err)
		}

		for i := range orderItems {
			orderItems[i].OrderID = order.ID
		}

		if err := tx.Create(&orderItems).Error; err != nil {
			return fmt.Errorf("create order items: %w", err)
		}

		var customer entities.User
		if err := tx.First(&customer, order.CustomerID).Error; err != nil {
			return fmt.Errorf("fetch customer: %w", err)
		}

		customer.Cart.Total = 0
		customer.Cart.Items = []entities.CartItem{}

		if err := tx.Save(&customer).Error; err != nil {
			return fmt.Errorf("clear cart: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// if err := uc.cache.Del(ctx, fmt.Sprintf("cart:%d", order.CustomerID)).Err(); err != nil {
	// 	// log only
	// }

	return nil

}
