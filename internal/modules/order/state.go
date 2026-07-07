package order

import (
	"PocketArtisan/internal/entities"
	"fmt"
)

type OrderAction string

const (
	OrderActionAccept   OrderAction = "accept"
	OrderActionDecline  OrderAction = "decline"
	OrderActionShip     OrderAction = "ship"
	OrderActionComplete OrderAction = "complete"
)

type orderState interface {
	status() entities.OrderStatus
	transition(action OrderAction) (entities.OrderStatus, error)
}

type pendingOrderState struct{}
type paymentReservedOrderState struct{}
type acceptedOrderState struct{}
type shippedOrderState struct{}
type declinedOrderState struct{}
type completedOrderState struct{}

func (pendingOrderState) status() entities.OrderStatus         { return entities.OrderPending }
func (paymentReservedOrderState) status() entities.OrderStatus { return entities.OrderPaymentReserved }
func (acceptedOrderState) status() entities.OrderStatus        { return entities.OrderAccepted }
func (shippedOrderState) status() entities.OrderStatus         { return entities.OrderShipped }
func (declinedOrderState) status() entities.OrderStatus        { return entities.OrderDeclined }
func (completedOrderState) status() entities.OrderStatus       { return entities.OrderCompleted }

func (pendingOrderState) transition(action OrderAction) (entities.OrderStatus, error) {
	switch action {
	case OrderActionAccept:
		return entities.OrderAccepted, nil
	case OrderActionDecline:
		return entities.OrderDeclined, nil
	default:
		return "", invalidOrderTransition(entities.OrderPending, action)
	}
}

func (paymentReservedOrderState) transition(action OrderAction) (entities.OrderStatus, error) {
	switch action {
	case OrderActionAccept:
		return entities.OrderAccepted, nil
	case OrderActionDecline:
		return entities.OrderDeclined, nil
	default:
		return "", invalidOrderTransition(entities.OrderPaymentReserved, action)
	}
}

func (acceptedOrderState) transition(action OrderAction) (entities.OrderStatus, error) {
	if action == OrderActionShip {
		return entities.OrderShipped, nil
	}
	return "", invalidOrderTransition(entities.OrderAccepted, action)
}

func (shippedOrderState) transition(action OrderAction) (entities.OrderStatus, error) {
	if action == OrderActionComplete {
		return entities.OrderCompleted, nil
	}
	return "", invalidOrderTransition(entities.OrderShipped, action)
}

func (declinedOrderState) transition(action OrderAction) (entities.OrderStatus, error) {
	return "", invalidOrderTransition(entities.OrderDeclined, action)
}

func (completedOrderState) transition(action OrderAction) (entities.OrderStatus, error) {
	return "", invalidOrderTransition(entities.OrderCompleted, action)
}

func NextOrderStatus(current entities.OrderStatus, action OrderAction) (entities.OrderStatus, error) {
	state, err := resolveOrderState(current)
	if err != nil {
		return "", err
	}
	return state.transition(action)
}

func InitialOrderStatus(paymentType entities.PaymentType) (entities.OrderStatus, error) {
	switch paymentType {
	case entities.PaymentCreditCard:
		return entities.OrderPaymentReserved, nil
	case entities.CashOnDelivery:
		return entities.OrderPending, nil
	default:
		return "", fmt.Errorf("invalid payment type: %s", paymentType)
	}
}

func resolveOrderState(status entities.OrderStatus) (orderState, error) {
	switch status {
	case entities.OrderPending:
		return pendingOrderState{}, nil
	case entities.OrderPaymentReserved:
		return paymentReservedOrderState{}, nil
	case entities.OrderAccepted:
		return acceptedOrderState{}, nil
	case entities.OrderShipped:
		return shippedOrderState{}, nil
	case entities.OrderDeclined:
		return declinedOrderState{}, nil
	case entities.OrderCompleted:
		return completedOrderState{}, nil
	default:
		return nil, fmt.Errorf("unknown order status: %s", status)
	}
}

func invalidOrderTransition(current entities.OrderStatus, action OrderAction) error {
	return fmt.Errorf("invalid order transition: status=%s action=%s", current, action)
}
