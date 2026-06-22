//go:build integration

package order_test

import (
	"context"
	"errors"
	"testing"

	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/files/storage"
	"PocketArtisan/internal/modules/order/create"
	"PocketArtisan/internal/modules/payment"

	"gorm.io/gorm"
)

// createSvc builds an order/create.Service wired to tx so all DB writes are inside
// the rolled-back transaction. PDFs are written to t.TempDir() and cleaned up automatically.
func createSvc(t *testing.T, tx *gorm.DB, gw payment.Gateway) *create.Service {
	t.Helper()
	if testFonts == nil {
		t.Skip("font assets unavailable — set TEST_ASSETS_DIR to the project assets directory")
	}
	s := storage.NewLocalStorage(t.TempDir(), "http://localhost/api/files")
	return create.NewService(tx, nil, s, testFonts, gw)
}

// ── COD ───────────────────────────────────────────────────────────────────────

func TestCreate_COD_createsOrderPending(t *testing.T) {
	withTx(t, func(tx *gorm.DB) {
		fix := seedFixtures(t, tx)
		svc := createSvc(t, tx, freshMock())

		result, err := svc.Execute(ctxFor(fix.Customer.ID), create.NewOrderRequest{
			CraftsmanID:     fix.Craftsman.ID,
			Items:           []create.NewOrderItemRequest{{ProductID: fix.Product.ID, Quantity: 1}},
			PaymentType:     entities.CashOnDelivery,
			ShippingAddress: "Ulica Svetog Save 12, Beograd",
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}

		var order entities.Order
		tx.First(&order, result.OrderID)

		if order.Status != entities.OrderPending {
			t.Errorf("status: want %s, got %s", entities.OrderPending, order.Status)
		}
		if result.PaymentReservationID != "" {
			t.Error("COD order must not have a reservation ID")
		}
	})
}

func TestCreate_COD_totalPriceIsQuantityTimesUnitPrice(t *testing.T) {
	withTx(t, func(tx *gorm.DB) {
		fix := seedFixtures(t, tx)
		svc := createSvc(t, tx, freshMock())

		result, err := svc.Execute(ctxFor(fix.Customer.ID), create.NewOrderRequest{
			CraftsmanID:     fix.Craftsman.ID,
			Items:           []create.NewOrderItemRequest{{ProductID: fix.Product.ID, Quantity: 3}},
			PaymentType:     entities.CashOnDelivery,
			ShippingAddress: "Test Street 1",
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}

		want := fix.Product.Price * 3
		if result.TotalPrice != want {
			t.Errorf("total: want %.2f, got %.2f", want, result.TotalPrice)
		}
	})
}

// ── Credit card ───────────────────────────────────────────────────────────────

func TestCreate_CC_reservesFundsAndPersistsReservationID(t *testing.T) {
	withTx(t, func(tx *gorm.DB) {
		fix := seedFixtures(t, tx)
		mock := freshMock()
		svc := createSvc(t, tx, mock)

		result, err := svc.Execute(ctxFor(fix.Customer.ID), create.NewOrderRequest{
			CraftsmanID:     fix.Craftsman.ID,
			Items:           []create.NewOrderItemRequest{{ProductID: fix.Product.ID, Quantity: 1}},
			PaymentType:     entities.PaymentCreditCard,
			ShippingAddress: "Test Street 1",
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if result.PaymentReservationID == "" {
			t.Error("CC order must have a non-empty PaymentReservationID")
		}

		var order entities.Order
		tx.First(&order, result.OrderID)

		if order.Status != entities.OrderPaymentReserved {
			t.Errorf("status: want %s, got %s", entities.OrderPaymentReserved, order.Status)
		}
		if order.PaymentReservationID != result.PaymentReservationID {
			t.Error("PaymentReservationID must be persisted on the order row")
		}
	})
}

func TestCreate_CC_reserveFailure_compensatesWithOrderDeletion(t *testing.T) {
	withTx(t, func(tx *gorm.DB) {
		fix := seedFixtures(t, tx)
		mock := freshMock()
		mock.ForceNextError = errors.New("gateway down")
		svc := createSvc(t, tx, mock)

		_, err := svc.Execute(ctxFor(fix.Customer.ID), create.NewOrderRequest{
			CraftsmanID:     fix.Craftsman.ID,
			Items:           []create.NewOrderItemRequest{{ProductID: fix.Product.ID, Quantity: 1}},
			PaymentType:     entities.PaymentCreditCard,
			ShippingAddress: "Test Street 1",
		})
		if err == nil {
			t.Fatal("expected error when Reserve fails")
		}

		var count int64
		tx.Model(&entities.Order{}).Where("customer_id = ?", fix.Customer.ID).Count(&count)
		if count != 0 {
			t.Errorf("compensation must delete the committed order row, found %d", count)
		}
	})
}

// ── Validation ────────────────────────────────────────────────────────────────

func TestCreate_unknownProduct_returnsError(t *testing.T) {
	withTx(t, func(tx *gorm.DB) {
		fix := seedFixtures(t, tx)
		svc := createSvc(t, tx, freshMock())

		_, err := svc.Execute(ctxFor(fix.Customer.ID), create.NewOrderRequest{
			CraftsmanID:     fix.Craftsman.ID,
			Items:           []create.NewOrderItemRequest{{ProductID: 99999, Quantity: 1}},
			PaymentType:     entities.CashOnDelivery,
			ShippingAddress: "Test Street 1",
		})
		if err == nil {
			t.Error("expected error for non-existent product ID")
		}

		var count int64
		tx.Model(&entities.Order{}).Where("customer_id = ?", fix.Customer.ID).Count(&count)
		if count != 0 {
			t.Errorf("no order row should be created on validation failure, found %d", count)
		}
	})
}

// keep context.Background visible to avoid lint warning on unused import
var _ = context.Background
