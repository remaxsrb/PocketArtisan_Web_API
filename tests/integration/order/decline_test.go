//go:build integration

package order_test

import (
	"context"
	"testing"

	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/order/decline"
	"PocketArtisan/internal/modules/payment"
	"gorm.io/gorm"
)

func declineSvc(tx *gorm.DB, gw payment.Gateway) *decline.Service {
	return decline.NewService(tx, nil, gw)
}

// ── COD ───────────────────────────────────────────────────────────────────────

func TestDecline_COD_doesNotCallRefund_statusDeclined(t *testing.T) {
	withTx(t, func(tx *gorm.DB) {
		fix := seedFixtures(t, tx)
		mock := freshMock()
		order := seedOrder(t, tx, fix.Customer.ID, fix.Craftsman.ID, entities.CashOnDelivery, "")

		status, err := declineSvc(tx, mock).Execute(context.Background(), decline.DeclineOrderRequest{
			OrderID:     order.ID,
			CraftsmanID: fix.Craftsman.ID,
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if status != entities.OrderDeclined {
			t.Errorf("status: want %s, got %s", entities.OrderDeclined, status)
		}
	})
}

// ── Credit card ───────────────────────────────────────────────────────────────

func TestDecline_CC_refundsReservation_statusDeclined(t *testing.T) {
	withTx(t, func(tx *gorm.DB) {
		fix := seedFixtures(t, tx)
		mock := freshMock()

		res, _ := mock.Reserve(context.Background(), payment.ReserveRequest{
			OrderID: 1, Amount: 3200, Currency: "RSD",
		})
		order := seedOrder(t, tx, fix.Customer.ID, fix.Craftsman.ID, entities.PaymentCreditCard, res.ID)

		status, err := declineSvc(tx, mock).Execute(context.Background(), decline.DeclineOrderRequest{
			OrderID:     order.ID,
			CraftsmanID: fix.Craftsman.ID,
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if status != entities.OrderDeclined {
			t.Errorf("status: want %s, got %s", entities.OrderDeclined, status)
		}

		// Verify reservation was consumed (Refund removes it from mock).
		if refundErr := mock.Refund(context.Background(), res.ID); refundErr == nil {
			t.Error("reservation should have been refunded (consumed) by decline service")
		}
	})
}

func TestDecline_CC_refundFailure_returnsError_statusUnchanged(t *testing.T) {
	withTx(t, func(tx *gorm.DB) {
		fix := seedFixtures(t, tx)
		mock := freshMock()

		// No matching reservation in mock — Refund will fail with "unknown ID".
		order := seedOrder(t, tx, fix.Customer.ID, fix.Craftsman.ID, entities.PaymentCreditCard, "nonexistent-reservation")

		_, err := declineSvc(tx, mock).Execute(context.Background(), decline.DeclineOrderRequest{
			OrderID:     order.ID,
			CraftsmanID: fix.Craftsman.ID,
		})
		if err == nil {
			t.Fatal("expected error when Refund fails")
		}

		var updated entities.Order
		tx.First(&updated, order.ID)
		if updated.Status == entities.OrderDeclined {
			t.Error("order status must not be DECLINED after a failed refund")
		}
	})
}

// ── Authorization ─────────────────────────────────────────────────────────────

func TestDecline_wrongCraftsman_returnsForbidden(t *testing.T) {
	withTx(t, func(tx *gorm.DB) {
		fix := seedFixtures(t, tx)
		mock := freshMock()
		order := seedOrder(t, tx, fix.Customer.ID, fix.Craftsman.ID, entities.CashOnDelivery, "")

		_, err := declineSvc(tx, mock).Execute(context.Background(), decline.DeclineOrderRequest{
			OrderID:     order.ID,
			CraftsmanID: 99999, // wrong craftsman
		})
		if err == nil {
			t.Error("expected forbidden error for wrong craftsman")
		}
	})
}