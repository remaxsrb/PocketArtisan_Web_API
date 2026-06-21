package payment_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"PocketArtisan/internal/modules/payment"
)

// ── Test helpers ──────────────────────────────────────────────────────────────

// countingGateway wraps MockGateway and counts how many times Reserve reaches the inner.
// Used to verify that OPEN/HALF_OPEN states stop calls from reaching the real gateway.
type countingGateway struct {
	inner        *payment.MockGateway
	reserveCalls atomic.Int32
}

func newCounting() *countingGateway {
	return &countingGateway{inner: payment.NewMockGateway()}
}

func (g *countingGateway) Reserve(ctx context.Context, req payment.ReserveRequest) (payment.Reservation, error) {
	g.reserveCalls.Add(1)
	return g.inner.Reserve(ctx, req)
}
func (g *countingGateway) Capture(ctx context.Context, id string) error {
	return g.inner.Capture(ctx, id)
}
func (g *countingGateway) Refund(ctx context.Context, id string) error {
	return g.inner.Refund(ctx, id)
}

// stateGateway has two modes:
//   - "fail"  — returns an error immediately (used to trip the breaker)
//   - "block" — signals entered, then waits on release (used to hold a probe open
//     so a second goroutine can race it in HALF_OPEN)
//   - anything else — succeeds immediately
type stateGateway struct {
	mu      sync.Mutex
	mode    string
	entered chan struct{} // receives once when a "block" call enters
	release chan struct{} // close to let a "block" call return
}

func newStateGateway(initialMode string) *stateGateway {
	return &stateGateway{
		mode:    initialMode,
		entered: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
}

func (g *stateGateway) setMode(m string) {
	g.mu.Lock()
	g.mode = m
	g.mu.Unlock()
}

func (g *stateGateway) Reserve(_ context.Context, req payment.ReserveRequest) (payment.Reservation, error) {
	g.mu.Lock()
	mode := g.mode
	g.mu.Unlock()

	switch mode {
	case "fail":
		return payment.Reservation{}, errors.New("stateGateway: forced failure")
	case "block":
		g.entered <- struct{}{}
		<-g.release
		return payment.Reservation{ID: "probe_res", Amount: req.Amount}, nil
	default:
		return payment.Reservation{ID: "ok", Amount: req.Amount}, nil
	}
}
func (g *stateGateway) Capture(_ context.Context, _ string) error { return nil }
func (g *stateGateway) Refund(_ context.Context, _ string) error  { return nil }

// tripBreaker forces n failures through the breaker using ForceNextError on inner.
func tripBreaker(b *payment.BreakerGateway, inner *countingGateway, n int) {
	for range n {
		inner.inner.ForceNextError = errors.New("forced trip")
		b.Reserve(bgCtx, testReserveReq)
	}
}

// newBreaker creates a breaker with threshold=3, timeout=30ms — short enough that
// recovery tests finish quickly without being flaky.
func newBreaker(inner payment.Gateway) *payment.BreakerGateway {
	return payment.NewBreakerGateway(inner, 3, 30*time.Millisecond)
}

// ── CLOSED state ──────────────────────────────────────────────────────────────

func TestBreakerGateway_CLOSED_requestsFlowThrough(t *testing.T) {
	cg := newCounting()
	b := newBreaker(cg)

	for range 5 {
		if _, err := b.Reserve(bgCtx, testReserveReq); err != nil {
			t.Fatalf("unexpected error in CLOSED state: %v", err)
		}
	}
	if got := cg.reserveCalls.Load(); got != 5 {
		t.Errorf("inner calls: want 5, got %d", got)
	}
}

func TestBreakerGateway_CLOSED_failuresBelowThreshold_staysClosed(t *testing.T) {
	cg := newCounting()
	b := newBreaker(cg) // threshold = 3

	// 2 failures — below threshold
	for range 2 {
		cg.inner.ForceNextError = errors.New("transient")
		b.Reserve(bgCtx, testReserveReq)
	}

	// Next request must still reach the inner gateway (still CLOSED)
	if _, err := b.Reserve(bgCtx, testReserveReq); err != nil {
		t.Errorf("circuit should be CLOSED after 2/3 failures, got %v", err)
	}
}

func TestBreakerGateway_CLOSED_tripsToOPEN_atThreshold(t *testing.T) {
	cg := newCounting()
	b := newBreaker(cg) // threshold = 3

	tripBreaker(b, cg, 3)

	callsBefore := cg.reserveCalls.Load()

	_, err := b.Reserve(bgCtx, testReserveReq)
	if err == nil || !strings.Contains(err.Error(), "circuit open") {
		t.Fatalf("want 'circuit open' error, got %v", err)
	}
	// Inner must NOT have been called — breaker short-circuited
	if cg.reserveCalls.Load() != callsBefore {
		t.Error("inner gateway was called despite OPEN state")
	}
}

// ── OPEN state ────────────────────────────────────────────────────────────────

func TestBreakerGateway_OPEN_fastFailWithoutCallingInner(t *testing.T) {
	cg := newCounting()
	b := newBreaker(cg)
	tripBreaker(b, cg, 3)

	snapshot := cg.reserveCalls.Load()

	for range 5 {
		_, err := b.Reserve(bgCtx, testReserveReq)
		if err == nil || !strings.Contains(err.Error(), "circuit open") {
			t.Fatalf("want 'circuit open', got %v", err)
		}
	}
	if cg.reserveCalls.Load() != snapshot {
		t.Errorf("inner was called %d extra time(s) while OPEN", cg.reserveCalls.Load()-snapshot)
	}
}

// ── OPEN → HALF_OPEN → CLOSED ─────────────────────────────────────────────────

func TestBreakerGateway_OPEN_transitionsToHALFOPEN_afterTimeout(t *testing.T) {
	cg := newCounting()
	b := newBreaker(cg) // timeout = 30ms
	tripBreaker(b, cg, 3)

	time.Sleep(40 * time.Millisecond) // wait for timeout

	// First call after timeout: probe allowed through
	if _, err := b.Reserve(bgCtx, testReserveReq); err != nil {
		t.Errorf("probe after timeout should succeed, got %v", err)
	}
}

func TestBreakerGateway_HALFOPEN_probeSuccess_resetsToCLOSED(t *testing.T) {
	cg := newCounting()
	b := newBreaker(cg)
	tripBreaker(b, cg, 3)

	time.Sleep(40 * time.Millisecond)

	// Successful probe
	if _, err := b.Reserve(bgCtx, testReserveReq); err != nil {
		t.Fatalf("probe should succeed, got %v", err)
	}

	// Subsequent requests must flow freely (CLOSED)
	for range 3 {
		if _, err := b.Reserve(bgCtx, testReserveReq); err != nil {
			t.Errorf("post-probe request should succeed, got %v", err)
		}
	}
}

func TestBreakerGateway_HALFOPEN_probeFailure_reopensImmediately(t *testing.T) {
	cg := newCounting()
	b := newBreaker(cg)
	tripBreaker(b, cg, 3)

	time.Sleep(40 * time.Millisecond)

	// Probe fails
	cg.inner.ForceNextError = errors.New("still down")
	if _, err := b.Reserve(bgCtx, testReserveReq); err == nil {
		t.Fatal("probe should have returned an error")
	}

	// Must be OPEN again — single probe failure is enough, no threshold needed
	_, err := b.Reserve(bgCtx, testReserveReq)
	if err == nil || !strings.Contains(err.Error(), "circuit open") {
		t.Errorf("circuit should re-open immediately on failed probe, got %v", err)
	}
}

// ── HALF_OPEN concurrency ─────────────────────────────────────────────────────

func TestBreakerGateway_HALFOPEN_concurrent_onlyOneProbeReachesInner(t *testing.T) {
	sg := newStateGateway("fail")
	b := payment.NewBreakerGateway(sg, 1, 20*time.Millisecond)

	// Trip to OPEN (threshold=1)
	b.Reserve(bgCtx, testReserveReq)

	// Wait for OPEN → HALF_OPEN eligibility
	time.Sleep(30 * time.Millisecond)

	// Switch inner to blocking mode so the probe hangs inside inner.Reserve
	sg.setMode("block")

	var wg sync.WaitGroup

	// Goroutine 1: the probe — will block inside sg.Reserve
	probeErr := make(chan error, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := b.Reserve(bgCtx, testReserveReq)
		probeErr <- err
	}()

	// Wait for the probe to be inside the inner gateway
	select {
	case <-sg.entered:
	case <-time.After(2 * time.Second):
		t.Fatal("probe never entered inner gateway")
	}

	// Goroutine 2: arrives while probe is in flight — must be rejected
	_, err := b.Reserve(bgCtx, testReserveReq)
	if err == nil || !strings.Contains(err.Error(), "probe in flight") {
		t.Errorf("concurrent request should get 'probe in flight', got %v", err)
	}

	// Release the probe
	close(sg.release)
	wg.Wait()

	if err := <-probeErr; err != nil {
		t.Errorf("probe should have succeeded, got %v", err)
	}
}

// ── Failure counter reset ─────────────────────────────────────────────────────

func TestBreakerGateway_CLOSED_successResetsFailureCounter(t *testing.T) {
	cg := newCounting()
	b := newBreaker(cg) // threshold = 3

	// 2 failures
	for range 2 {
		cg.inner.ForceNextError = errors.New("transient")
		b.Reserve(bgCtx, testReserveReq)
	}

	// 1 success — resets counter to 0
	if _, err := b.Reserve(bgCtx, testReserveReq); err != nil {
		t.Fatalf("success call failed: %v", err)
	}

	// 2 more failures — should still be CLOSED (counter at 2/3, not 4/3)
	for range 2 {
		cg.inner.ForceNextError = errors.New("transient")
		b.Reserve(bgCtx, testReserveReq)
	}

	// One more success must get through — counter was reset, only at 2/3 again
	if _, err := b.Reserve(bgCtx, testReserveReq); err != nil {
		t.Errorf("circuit should still be CLOSED (counter reset by earlier success), got %v", err)
	}
}
