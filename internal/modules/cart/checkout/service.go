package checkout

import (
	"PocketArtisan/internal/entities"
	cartmod "PocketArtisan/internal/modules/cart"
	"PocketArtisan/internal/modules/order/create"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo        cartmod.Repository
	cache       *redis.Client
	cartReader  cartmod.CartReader
	orderCreate *create.Service
}

func NewService(db *gorm.DB, cache *redis.Client, orderCreate *create.Service) *Service {
	repo := cartmod.NewGormRepository(db)
	return &Service{
		repo:        repo,
		cache:       cache,
		cartReader:  repo,
		orderCreate: orderCreate,
	}
}

func (uc *Service) Execute(ctx context.Context, req CheckoutRequest) ([]OrderResult, error) {
	customerID := ctx.Value("user_id").(uint64)
	saga := NewCheckoutSaga(uc.orderCreate.CompensateOrder)

	userCart, err := uc.cartReader.GetUserCart(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("fetch cart: %w", err)
	}
	if len(userCart.Items) == 0 {
		return nil, errors.New("cart is empty")
	}

	groups := groupByCraftsman(userCart.Items)

	var results []OrderResult
	for craftsmanID, items := range groups {
		orderReq := create.NewOrderRequest{
			CraftsmanID:     craftsmanID,
			Items:           toOrderItems(items),
			PaymentType:     req.PaymentType,
			ShippingAddress: req.ShippingAddress,
		}

		result, err := uc.orderCreate.Execute(ctx, orderReq)
		if err != nil {
			saga.Compensate(ctx)
			return nil, fmt.Errorf("create order for craftsman %d: %w", craftsmanID, err)
		}

		saga.Record(result.OrderID, result.PaymentReservationID)

		results = append(results, OrderResult{
			OrderID:     result.OrderID,
			CraftsmanID: craftsmanID,
			Total:       result.TotalPrice,
			PDFURL:      result.PDFURL,
		})
	}

	if err := uc.clearCart(ctx, userCart); err != nil {
		log.Printf("checkout: cart clear failed for user %d: %v", customerID, err)
	}

	return results, nil
}

func groupByCraftsman(items []entities.CartItem) map[uint64][]entities.CartItem {
	groups := make(map[uint64][]entities.CartItem)
	for _, item := range items {
		id := item.Product.CraftsmanID
		groups[id] = append(groups[id], item)
	}
	return groups
}

func toOrderItems(items []entities.CartItem) []create.NewOrderItemRequest {
	out := make([]create.NewOrderItemRequest, len(items))
	for i, item := range items {
		out[i] = create.NewOrderItemRequest{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}
	return out
}

func (uc *Service) clearCart(ctx context.Context, userCart *entities.Cart) error {
	if err := uc.repo.ClearCartItems(ctx, userCart.ID); err != nil {
		return err
	}
	return uc.repo.UpdateCartTotal(ctx, userCart.ID, 0)
}
