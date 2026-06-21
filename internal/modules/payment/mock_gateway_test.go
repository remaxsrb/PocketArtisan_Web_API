package payment_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"PocketArtisan/internal/modules/payment"
)

var (
	bgCtx          = context.Background()
	testReserveReq = payment.ReserveRequest{
		OrderID:     42,
		Amount:      3200.00,
		Currency:    "RSD",
		Description: "Order #42",
	}
)

// ── Reserve ───────────────────────────────────────────────────────────────────

func TestMockGateway_Reserve_returnsNonEmptyIDAndCorrectAmount(t *testing.T) {
	mock := payment.NewMockGateway()

	res, err := mock.Reserve(bgCtx, testReserveReq)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID == "" {
		t.Error("reservation ID must not be empty")
	}
	if res.Amount != testReserveReq.Amount {
		t.Errorf("amount: want %.2f, got %.2f", testReserveReq.Amount, res.Amount)
	}
}

func TestMockGateway_Reserve_forceError_returnsErrAndClearsField(t *testing.T) {
	mock := payment.NewMockGateway()
	forced := errors.New("gateway timeout")
	mock.ForceNextError = forced

	_, err := mock.Reserve(bgCtx, testReserveReq)
	if !errors.Is(err, forced) {
		t.Fatalf("want forced error, got %v", err)
	}

	// ForceNextError is cleared after one use — next call must succeed.
	_, err = mock.Reserve(bgCtx, testReserveReq)
	if err != nil {
		t.Errorf("second call should succeed after force-error consumed, got %v", err)
	}
}

func TestMockGateway_Reserve_multipleCallsProduceUniqueIDs(t *testing.T) {
	mock := payment.NewMockGateway()
	seen := make(map[string]bool)

	for i := range 5 {
		req := payment.ReserveRequest{OrderID: uint(i + 1), Amount: float64(i * 100), Currency: "RSD"}
		res, err := mock.Reserve(bgCtx, req)
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if seen[res.ID] {
			t.Errorf("duplicate reservation ID %q on call %d", res.ID, i)
		}
		seen[res.ID] = true
	}
}

// ── Capture ───────────────────────────────────────────────────────────────────

func TestMockGateway_Capture_success_removesReservation(t *testing.T) {
	mock := payment.NewMockGateway()
	res, _ := mock.Reserve(bgCtx, testReserveReq)

	if err := mock.Capture(bgCtx, res.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Reservation deleted — second Capture must fail.
	if err := mock.Capture(bgCtx, res.ID); err == nil {
		t.Error("expected error on second Capture of same ID")
	}
}

func TestMockGateway_Capture_unknownID_returnsError(t *testing.T) {
	mock := payment.NewMockGateway()

	if err := mock.Capture(bgCtx, "nonexistent"); err == nil {
		t.Error("expected error for unknown reservation ID")
	}
}

// ── Refund ────────────────────────────────────────────────────────────────────

func TestMockGateway_Refund_success_removesReservation(t *testing.T) {
	mock := payment.NewMockGateway()
	res, _ := mock.Reserve(bgCtx, testReserveReq)

	if err := mock.Refund(bgCtx, res.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.Refund(bgCtx, res.ID); err == nil {
		t.Error("expected error on second Refund of same ID")
	}
}

func TestMockGateway_Refund_unknownID_returnsError(t *testing.T) {
	mock := payment.NewMockGateway()

	if err := mock.Refund(bgCtx, "nonexistent"); err == nil {
		t.Error("expected error for unknown reservation ID")
	}
}

// ── Concurrency ───────────────────────────────────────────────────────────────

func TestMockGateway_Reserve_concurrent_producesUniqueIDs(t *testing.T) {
	mock := payment.NewMockGateway()

	const n = 50
	type result struct {
		id  string
		err error
	}
	results := make(chan result, n)

	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := payment.ReserveRequest{OrderID: uint(i + 1), Amount: float64(i * 100), Currency: "RSD"}
			res, err := mock.Reserve(bgCtx, req)
			results <- result{id: res.ID, err: err}
		}(i)
	}

	wg.Wait()
	close(results)

	seen := make(map[string]bool)
	for r := range results {
		if r.err != nil {
			t.Errorf("goroutine error: %v", r.err)
			continue
		}
		if seen[r.id] {
			t.Errorf("duplicate reservation ID: %q", r.id)
		}
		seen[r.id] = true
	}
	if len(seen) != n {
		t.Errorf("want %d unique IDs, got %d", n, len(seen))
	}
}
