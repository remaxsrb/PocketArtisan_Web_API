package create

import (
	"PocketArtisan/internal/entities"
	ordermod "PocketArtisan/internal/modules/order"
	"fmt"
)

type OrderBuilder struct {
	customerID  uint64
	craftsmanID uint64
	paymentType entities.PaymentType
	items       []NewOrderItemRequest
	prices      map[uint64]float64
}

func NewOrderBuilder() *OrderBuilder {
	return &OrderBuilder{prices: make(map[uint64]float64)}
}

func (b *OrderBuilder) WithCustomer(customerID uint64) *OrderBuilder {
	b.customerID = customerID
	return b
}

func (b *OrderBuilder) WithCraftsman(craftsmanID uint64) *OrderBuilder {
	b.craftsmanID = craftsmanID
	return b
}

func (b *OrderBuilder) WithPaymentType(paymentType entities.PaymentType) *OrderBuilder {
	b.paymentType = paymentType
	return b
}

func (b *OrderBuilder) WithItems(items []NewOrderItemRequest) *OrderBuilder {
	b.items = items
	return b
}

func (b *OrderBuilder) WithPrices(products []ordermod.ProductPriceRow) *OrderBuilder {
	for _, p := range products {
		b.prices[p.ID] = p.Price
	}
	return b
}

func (b *OrderBuilder) Build() (entities.Order, []entities.OrderItem, error) {
	order := entities.Order{
		CustomerID:  b.customerID,
		CraftsmanID: b.craftsmanID,
		PaymentType: b.paymentType,
	}

	switch order.PaymentType {
	case entities.PaymentCreditCard, entities.CashOnDelivery:
	default:
		return entities.Order{}, nil, fmt.Errorf("invalid payment type: %s", b.paymentType)
	}

	initialStatus, err := ordermod.InitialOrderStatus(order.PaymentType)
	if err != nil {
		return entities.Order{}, nil, err
	}
	order.Status = initialStatus

	orderItems := make([]entities.OrderItem, len(b.items))
	total := 0.0
	for i, item := range b.items {
		price, ok := b.prices[item.ProductID]
		if !ok {
			return entities.Order{}, nil, fmt.Errorf("one or more products do not exist")
		}

		total += price * float64(item.Quantity)
		orderItems[i] = entities.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: price,
		}
	}

	order.TotalPrice = total
	return order, orderItems, nil
}
