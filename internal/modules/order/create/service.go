package create

import (
	"PocketArtisan/internal/entities"
	orderPDF "PocketArtisan/internal/modules/files/generate_pdf/order"
	"PocketArtisan/internal/modules/files/storage"
	ordermod "PocketArtisan/internal/modules/order"
	"PocketArtisan/internal/modules/payment"
	"PocketArtisan/internal/modules/utils/fonts"
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo       ordermod.Repository
	cache      *redis.Client
	pdfService *orderPDF.Service
	gateway    payment.Gateway
}

func NewService(db *gorm.DB, cache *redis.Client, s storage.Storage, f *fonts.Service, gw payment.Gateway) *Service {
	return &Service{repo: ordermod.NewGormRepository(db), cache: cache, pdfService: orderPDF.NewService(s, f), gateway: gw}
}

func (uc *Service) Execute(ctx context.Context, req NewOrderRequest) (OrderCreationResult, error) {
	customerID := ctx.Value("user_id").(uint64)

	productIDs := make([]uint64, len(req.Items))

	for i, item := range req.Items {
		productIDs[i] = item.ProductID
	}

	products, err := uc.repo.FindProductPricesByCraftsman(ctx, productIDs, req.CraftsmanID)
	if err != nil {
		return OrderCreationResult{}, fmt.Errorf("failed to fetch product prices: %w", err)
	}

	builder := NewOrderBuilder().
		WithCustomer(customerID).
		WithCraftsman(req.CraftsmanID).
		WithPaymentType(req.PaymentType).
		WithItems(req.Items).
		WithPrices(products)

	order, orderItems, err := builder.Build()
	if err != nil {
		return OrderCreationResult{}, err
	}

	customer, err := uc.repo.CreateOrderWithItemsAndCustomer(ctx, &order, orderItems)
	if err != nil {
		return OrderCreationResult{}, fmt.Errorf("create order transaction: %w", err)
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
			uc.compensateOrder(ctx, order.ID, "")
			return OrderCreationResult{}, fmt.Errorf("payment reservation failed: %w", err)
		}
		reservationID = res.ID
		if err := uc.repo.UpdatePaymentReservationID(ctx, order.ID, reservationID); err != nil {
			log.Printf("order %d: failed to persist reservation id: %v", order.ID, err)
		}
	}

	// Reload order items with product details for the PDF
	orderItems, err = uc.repo.FindOrderItemsWithProduct(ctx, order.ID)
	if err != nil {
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

	if err := uc.repo.UpdateURL(ctx, order.ID, pdfURL); err != nil {
		log.Printf("order %d: failed to persist pdf url: %v", order.ID, err)
	}

	return OrderCreationResult{
		OrderID:              order.ID,
		TotalPrice:           order.TotalPrice,
		PDFURL:               pdfURL,
		PaymentReservationID: reservationID,
	}, nil
}

// CompensateOrder executes compensation for an already-created order.
// It attempts reservation refund first (when reservationID is present), then
// removes the order and its items.
func (uc *Service) CompensateOrder(ctx context.Context, orderID uint64, reservationID string) {
	if err := uc.compensateOrder(ctx, orderID, reservationID); err != nil {
		log.Printf("order %d: compensation failed: %v", orderID, err)
	}
}

func (uc *Service) compensateOrder(ctx context.Context, orderID uint64, reservationID string) error {
	if reservationID != "" {
		if err := uc.gateway.Refund(ctx, reservationID); err != nil {
			return fmt.Errorf("refund reservation %s: %w", reservationID, err)
		}
	}

	if err := uc.repo.DeleteOrderWithItems(ctx, orderID); err != nil {
		return fmt.Errorf("delete order with items: %w", err)
	}

	return nil
}
