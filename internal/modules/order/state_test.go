package order

import (
	"PocketArtisan/internal/entities"
	"testing"
)

func TestNextOrderStatus_AllowsValidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		current entities.OrderStatus
		action  OrderAction
		want    entities.OrderStatus
	}{
		{name: "pending to accepted", current: entities.OrderPending, action: OrderActionAccept, want: entities.OrderAccepted},
		{name: "pending to declined", current: entities.OrderPending, action: OrderActionDecline, want: entities.OrderDeclined},
		{name: "reserved to accepted", current: entities.OrderPaymentReserved, action: OrderActionAccept, want: entities.OrderAccepted},
		{name: "reserved to declined", current: entities.OrderPaymentReserved, action: OrderActionDecline, want: entities.OrderDeclined},
		{name: "accepted to shipped", current: entities.OrderAccepted, action: OrderActionShip, want: entities.OrderShipped},
		{name: "shipped to completed", current: entities.OrderShipped, action: OrderActionComplete, want: entities.OrderCompleted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NextOrderStatus(tc.current, tc.action)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("want %s, got %s", tc.want, got)
			}
		})
	}
}

func TestNextOrderStatus_RejectsInvalidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		current entities.OrderStatus
		action  OrderAction
	}{
		{name: "pending cannot ship", current: entities.OrderPending, action: OrderActionShip},
		{name: "accepted cannot decline", current: entities.OrderAccepted, action: OrderActionDecline},
		{name: "declined cannot accept", current: entities.OrderDeclined, action: OrderActionAccept},
		{name: "completed cannot ship", current: entities.OrderCompleted, action: OrderActionShip},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NextOrderStatus(tc.current, tc.action)
			if err == nil {
				t.Fatal("expected transition error")
			}
		})
	}
}

func TestInitialOrderStatus(t *testing.T) {
	status, err := InitialOrderStatus(entities.PaymentCreditCard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != entities.OrderPaymentReserved {
		t.Fatalf("want %s, got %s", entities.OrderPaymentReserved, status)
	}

	status, err = InitialOrderStatus(entities.CashOnDelivery)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != entities.OrderPending {
		t.Fatalf("want %s, got %s", entities.OrderPending, status)
	}
}
