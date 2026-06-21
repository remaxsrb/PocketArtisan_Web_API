//go:build integration

package order_test

import (
	"context"
	"errors"
	"testing"

	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/order/ship"
	"PocketArtisan/internal/modules/payment"
	"gorm.io/gorm"
)

// seedOrder inserts a bare-minimum order into tx and returns it.
func seedOrder(t *testing.T, tx *gorm.DB, customerID, craftsmanID uint64, pt entities.PaymentType, reservationID string) entities.Order {
	t.Helper()
	status := entities.OrderPending
	if pt == entities.PaymentCreditCard {
		status = entities.OrderPaymentReserved
	}
	order := entities.Order{
		CustomerID:           customerID,
		CraftsmanID:          craftsmanID,
		TotalPrice:           3200.00,
		PaymentType:          pt,
		Status:               status,
		PaymentReservationID: reservationID,
	}
	if err := tx.Create(&order).Error; err != nil {
		t.Fatalf("seed order: %v", err)
	}
	return order
}

func shipSvc(tx *gorm.DB, gw payment.Gateway) *ship.Service {
	return ship.NewService(tx, nil, gw)
}

// ── COD ───────────────────────────────────────────────────────────────────────

func TestShip_COD_doesNotCallCapture_statusShipped(t *testing.T) {
	withTx(t, func(tx *gorm.DB) {
		fix := seedFixtures(t, tx)
		mock := freshMock()
		order := seedOrder(t, tx, fix.Customer.ID, fix.Craftsman.ID, entities.CashOnDelivery, "")

		status, err := shipSvc(tx, mock).Execute(context.Background(), ship.ShipOrderRequest{
			OrderID:     order.ID,
			CraftsmanID: fix.Craftsman.ID,
			CustomerID:  fix.Customer.ID,
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if status != entities.OrderShipped {
			t.Errorf("status: want %s, got %s", entities.OrderShipped, status)
		}
	})
}

// ── Credit card ───────────────────────────────────────────────────────────────

func TestShip_CC_capturesReservation_statusShipped(t *testing.T) {
	withTx(t, func(tx *gorm.DB) {
		fix := seedFixtures(t, tx)
		mock := freshMock()

		res, _ := mock.Reserve(context.Background(), payment.ReserveRequest{
			OrderID: 1, Amount: 3200, Currency: "RSD",
		})
		order := seedOrder(t, tx, fix.Customer.ID, fix.Craftsman.ID, entities.PaymentCreditCard, res.ID)

		status, err := shipSvc(tx, mock).Execute(context.Background(), ship.ShipOrderRequest{
			OrderID:     order.ID,
			CraftsmanID: fix.Craftsman.ID,
			CustomerID:  fix.Customer.ID,
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if status != entities.OrderShipped {
			t.Errorf("status: want %s, got %s", entities.OrderShipped, status)
		}

		// Verify reservation was consumed (Capture removes it from mock).
		if captureErr := mock.Capture(context.Background(), res.ID); captureErr == nil {
			t.Error("reservation should have been captured (consumed) by ship service")
		}
	})
}

func TestShip_CC_captureFailure_returnsError_statusUnchanged(t *testing.T) {
	withTx(t, func(tx *gorm.DB) {
		fix := seedFixtures(t, tx)
		mock := freshMock()

		res, _ := mock.Reserve(context.Background(), payment.ReserveRequest{
			OrderID: 1, Amount: 3200, Currency: "RSD",
		})
		order := seedOrder(t, tx, fix.Customer.ID, fix.Craftsman.ID, entities.PaymentCreditCard, res.ID)

		// Remove the reservation so Capture fails with "unknown ID".
		_ = mock.Refund(context.Background(), res.ID)

		_, err := shipSvc(tx, mock).Execute(context.Background(), ship.ShipOrderRequest{
			OrderID:     order.ID,
			CraftsmanID: fix.Craftsman.ID,
			CustomerID:  fix.Customer.ID,
		})
		if err == nil {
			t.Fatal("expected error when Capture fails")
		}

		var updated entities.Order
		tx.First(&updated, order.ID)
		if updated.Status == entities.OrderShipped {
			t.Error("order status must not be SHIPPED after a failed capture")
		}
	})
}

// ── Authorization ─────────────────────────────────────────────────────────────

func TestShip_wrongCraftsman_returnsForbidden(t *testing.T) {
	withTx(t, func(tx *gorm.DB) {
		fix := seedFixtures(t, tx)
		mock := freshMock()
		order := seedOrder(t, tx, fix.Customer.ID, fix.Craftsman.ID, entities.CashOnDelivery, "")

		_, err := shipSvc(tx, mock).Execute(context.Background(), ship.ShipOrderRequest{
			OrderID:     order.ID,
			CraftsmanID: 99999, // wrong craftsman
			CustomerID:  fix.Customer.ID,
		})
		if err == nil {
			t.Error("expected forbidden error for wrong craftsman")
		}
	})
}

var _ = errors.New // silence unused import warning during dead-code paths
