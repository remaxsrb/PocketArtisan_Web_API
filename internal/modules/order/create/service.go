package create

import (
	"PocketArtisan/internal/entities"
	orderPDF "PocketArtisan/internal/modules/files/generate_pdf/order"
	"PocketArtisan/internal/modules/files/storage"
	"PocketArtisan/internal/modules/payment"
	"PocketArtisan/internal/modules/utils/fonts"
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	db         *gorm.DB
	cache      *redis.Client
	pdfService *orderPDF.Service
	gateway    payment.Gateway
}

type ProductPrice struct {
	ID    uint64
	Price float64
}

func NewService(db *gorm.DB, cache *redis.Client, s storage.Storage, f *fonts.Service, gw payment.Gateway) *Service {
	return &Service{db: db, cache: cache, pdfService: orderPDF.NewService(s, f), gateway: gw}
}

func (uc *Service) Execute(ctx context.Context, req NewOrderRequest) (OrderCreationResult, error) {

	var order entities.Order
	order.CustomerID = ctx.Value("user_id").(uint64)
	order.CraftsmanID = req.CraftsmanID

	switch req.PaymentType {
	case "CC":
		order.PaymentType = entities.PaymentCreditCard
		order.Status = entities.OrderPaymentReserved
	case "COD":
		order.PaymentType = entities.CashOnDelivery
		order.Status = entities.OrderPending
	default:
		return OrderCreationResult{}, fmt.Errorf("invalid payment type: %s", req.PaymentType)
	}

	productIDs := make([]uint64, len(req.Items))
	quantities := make(map[uint64]int)
	// Map to store product prices in case craftsman changes them during the order creation process
	prices := make(map[uint64]float64)

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
		return OrderCreationResult{}, fmt.Errorf("failed to fetch product prices: %w", err)
	}

	if len(products) != len(productIDs) {
		return OrderCreationResult{}, fmt.Errorf("one or more products do not exist")
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
	var customer entities.User
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

		if err := tx.First(&customer, order.CustomerID).Error; err != nil {
			return fmt.Errorf("fetch customer: %w", err)
		}

		return nil
	})

	if err != nil {
		return OrderCreationResult{}, err
	}

	// Reserve funds for CC orders. On failure, compensate by deleting the committed order.
	var reservationID string
	if order.PaymentType == entities.PaymentCreditCard {
		res, err := uc.gateway.Reserve(ctx, payment.ReserveRequest{
			OrderID:     uint(order.ID),
			Amount:      order.TotalPrice,
			Currency:    "RSD",
			Description: fmt.Sprintf("Order #%d", order.ID),
		})
		if err != nil {
			uc.deleteOrder(ctx, order.ID)
			return OrderCreationResult{}, fmt.Errorf("payment reservation failed: %w", err)
		}
		reservationID = res.ID
		if err := uc.db.WithContext(ctx).Model(&order).Update("payment_reservation_id", reservationID).Error; err != nil {
			log.Printf("order %d: failed to persist reservation id: %v", order.ID, err)
		}
	}

	// Reload order items with product details for the PDF
	if err := uc.db.WithContext(ctx).Preload("Product").Find(&orderItems, "order_id = ?", order.ID).Error; err != nil {
		log.Printf("order %d: failed to preload items for PDF: %v", order.ID, err)
		return OrderCreationResult{OrderID: order.ID, TotalPrice: order.TotalPrice, PaymentReservationID: reservationID}, nil
	}

	pdfData := orderPDF.OrderData{
		OrderID:         order.ID,
		CustomerName:    customer.Username,
		CustomerEmail:   customer.Email,
		ShippingAddress: req.ShippingAddress,
		Items:           orderItems,
		OrderDate:       order.CreatedAt.Format("02/01/2006"),
		TotalPrice:      order.TotalPrice,
	}

	switch req.PaymentType {
	case "CC":
		pdfData.PaymentType = "Platna kartica"
	case "COD":
		pdfData.PaymentType = "Plaćanje pouzećem"
	}

	pdfURL, err := uc.pdfService.Generate(pdfData)
	if err != nil {
		log.Printf("order %d: pdf generation failed: %v", order.ID, err)
		return OrderCreationResult{OrderID: order.ID, TotalPrice: order.TotalPrice, PaymentReservationID: reservationID}, nil
	}

	if err := uc.db.WithContext(ctx).Model(&order).Update("url", pdfURL).Error; err != nil {
		log.Printf("order %d: failed to persist pdf url: %v", order.ID, err)
	}

	return OrderCreationResult{
		OrderID:              order.ID,
		TotalPrice:           order.TotalPrice,
		PDFURL:               pdfURL,
		PaymentReservationID: reservationID,
	}, nil
}

// deleteOrder removes a committed order and its items when a post-transaction step fails.
func (uc *Service) deleteOrder(ctx context.Context, orderID uint64) {
	if err := uc.db.WithContext(ctx).Where("order_id = ?", orderID).Delete(&entities.OrderItem{}).Error; err != nil {
		log.Printf("order %d: compensation delete items failed: %v", orderID, err)
	}
	if err := uc.db.WithContext(ctx).Delete(&entities.Order{}, orderID).Error; err != nil {
		log.Printf("order %d: compensation delete order failed: %v", orderID, err)
	}
}
