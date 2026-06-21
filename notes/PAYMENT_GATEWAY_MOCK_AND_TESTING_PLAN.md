# Payment Gateway — Mock Design and Testing Plan

**Date**: June 15, 2026
**Status**: Proposed
**Scope**: Credit card validation, mock gateway implementation, Stripe migration path

---

## Goal

Introduce a `PaymentGateway` interface that the order flow depends on. Ship a mock implementation that covers all CC order states. Keep the interface narrow enough that dropping in Stripe later requires no changes to calling code.

---

## Interface Design

Place this in `internal/modules/payment/gateway.go`.

```go
package payment

import "context"

type ReserveRequest struct {
    OrderID     uint
    Amount      float64
    Currency    string // e.g. "RSD", "EUR"
    Description string
}

type Reservation struct {
    ID     string  // opaque token — Stripe calls this a PaymentIntent ID
    Amount float64
}

type Gateway interface {
    Reserve(ctx context.Context, req ReserveRequest) (Reservation, error)
    Capture(ctx context.Context, reservationID string) error
    Refund(ctx context.Context, reservationID string) error
}
```

**Why three methods only:**
- `Reserve` → called when customer confirms CC order (status: `PAYMENT_RESERVED`)
- `Capture` → called when craftsman marks order `SHIPPED`
- `Refund` → called when craftsman `DECLINED` the order

These map directly to Stripe's `PaymentIntent` lifecycle:
`create (capture_method=manual)` → `capture` → `cancel`.

---

## Mock Implementation

Place in `internal/modules/payment/mock_gateway.go`.

The mock stores reservations in a thread-safe in-memory map. It exposes a `ForceNextError` field so individual tests can inject failures without changing any global state.

```go
package payment

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type MockGateway struct {
    mu             sync.Mutex
    reservations   map[string]Reservation
    ForceNextError error // set in tests to simulate gateway failure
}

func NewMockGateway() *MockGateway {
    return &MockGateway{reservations: make(map[string]Reservation)}
}

func (m *MockGateway) Reserve(ctx context.Context, req ReserveRequest) (Reservation, error) {
    m.mu.Lock()
    defer m.mu.Unlock()

    if m.ForceNextError != nil {
        err := m.ForceNextError
        m.ForceNextError = nil
        return Reservation{}, err
    }

    id := fmt.Sprintf("mock_res_%d_%d", req.OrderID, time.Now().UnixMilli())
    r := Reservation{ID: id, Amount: req.Amount}
    m.reservations[id] = r
    return r, nil
}

func (m *MockGateway) Capture(ctx context.Context, reservationID string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if _, ok := m.reservations[reservationID]; !ok {
        return fmt.Errorf("reservation %s not found", reservationID)
    }
    delete(m.reservations, reservationID)
    return nil
}

func (m *MockGateway) Refund(ctx context.Context, reservationID string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if _, ok := m.reservations[reservationID]; !ok {
        return fmt.Errorf("reservation %s not found", reservationID)
    }
    delete(m.reservations, reservationID)
    return nil
}
```

---

## Circuit Breaker Integration

Wrap `Gateway` with the circuit breaker before injecting it into the container. The breaker sits between the calling service and the gateway — the calling code never sees it.

### Previous version — bugs

The original `call` method had a `switch` with only a `case "OPEN"`. This caused two silent bugs:

1. **No probe limiting in HALF_OPEN.** The `"HALF_OPEN"` state was never cased, so every concurrent request fell through and reached the inner gateway. HALF_OPEN is supposed to allow exactly one probe at a time.
2. **Failed probe did not immediately re-open.** A failure in HALF_OPEN incremented `failures` toward the threshold instead of re-opening the circuit immediately. One failed probe is sufficient to conclude the service is still down.

### Corrected implementation

```go
// internal/modules/payment/breaker_gateway.go
package payment

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type BreakerGateway struct {
    inner            Gateway
    failureThreshold int
    openTimeout      time.Duration

    mu            sync.Mutex
    failures      int
    state         string // "CLOSED", "OPEN", "HALF_OPEN"
    openedAt      time.Time
    probeInFlight bool
}

func NewBreakerGateway(inner Gateway, failureThreshold int, openTimeout time.Duration) *BreakerGateway {
    return &BreakerGateway{
        inner:            inner,
        failureThreshold: failureThreshold,
        openTimeout:      openTimeout,
        state:            "CLOSED",
    }
}

func (b *BreakerGateway) call(ctx context.Context, fn func() error) error {
    b.mu.Lock()

    switch b.state {
    case "OPEN":
        if time.Since(b.openedAt) < b.openTimeout {
            b.mu.Unlock()
            return fmt.Errorf("payment gateway circuit open")
        }
        // Timeout expired — transition to HALF_OPEN and let this request be the probe.
        b.state = "HALF_OPEN"
        b.probeInFlight = true

    case "HALF_OPEN":
        // Only one probe at a time; reject all others until the probe resolves.
        if b.probeInFlight {
            b.mu.Unlock()
            return fmt.Errorf("payment gateway circuit half-open, probe in flight")
        }
        b.probeInFlight = true

    // default (CLOSED): fall through, no gate needed.
    }

    b.mu.Unlock()
    err := fn()
    b.mu.Lock()
    defer b.mu.Unlock()

    b.probeInFlight = false

    if err != nil {
        b.failures++
        // In HALF_OPEN a single failure is enough to re-open immediately.
        if b.state == "HALF_OPEN" || b.failures >= b.failureThreshold {
            b.state = "OPEN"
            b.openedAt = time.Now()
            b.failures = 0
        }
        return err
    }

    b.failures = 0
    b.state = "CLOSED"
    return nil
}

func (b *BreakerGateway) Reserve(ctx context.Context, req ReserveRequest) (Reservation, error) {
    var result Reservation
    err := b.call(ctx, func() error {
        var e error
        result, e = b.inner.Reserve(ctx, req)
        return e
    })
    return result, err
}

func (b *BreakerGateway) Capture(ctx context.Context, reservationID string) error {
    return b.call(ctx, func() error { return b.inner.Capture(ctx, reservationID) })
}

func (b *BreakerGateway) Refund(ctx context.Context, reservationID string) error {
    return b.call(ctx, func() error { return b.inner.Refund(ctx, reservationID) })
}
```

**What changed and why:**

| # | Change | Reason |
|---|--------|--------|
| 1 | Added `probeInFlight bool` field | Tracks whether a probe is already in progress so concurrent requests in HALF_OPEN are rejected |
| 2 | Added explicit `case "HALF_OPEN"` | Makes the intended behaviour visible instead of relying on switch fall-through |
| 3 | OPEN → HALF_OPEN transition sets `probeInFlight = true` immediately | The request that triggers the transition is itself the probe; no second gate needed |
| 4 | `probeInFlight` cleared before result evaluation | Ensures the flag is always reset even if the result handling panics |
| 5 | `b.state == "HALF_OPEN"` check before threshold in failure path | A single failed probe re-opens the circuit immediately without waiting for more failures |
| 6 | `b.failures = 0` on re-open | Resets the counter so the next CLOSED phase starts clean |

---

## Wiring into the Container

Add `Gateway payment.Gateway` to `AppContainer`. In `cmd/main.go` decide which implementation to use based on environment:

```go
// cmd/main.go

var gateway payment.Gateway = payment.NewMockGateway()
// When ready for Stripe:
// gateway = payment.NewStripeGateway(os.Getenv("STRIPE_SECRET_KEY"))

wrappedGateway := payment.NewBreakerGateway(gateway, 5, 30*time.Second)
```

Inject `wrappedGateway` into the order `create` and `set_status` services — pass it through `RegisterRoutes` the same way `Storage` is passed today.

---

## Where to Store the Reservation ID

Add `PaymentReservationID string` to the `Order` entity. Populate it after a successful `Reserve` call in the `order/create` service. The `set_status` service reads it when calling `Capture` or `Refund`.

```go
// entities/order.go
type Order struct {
    // ... existing fields
    PaymentReservationID string `json:"payment_reservation_id" gorm:"type:text"`
}
```

---

## Test Scenarios

These are service-level tests — call the service directly, no HTTP layer needed. Inject `MockGateway` for scenarios 1–6, `BreakerGateway` wrapping a failing mock for 7–8.

| # | Scenario | Setup | Expected outcome |
|---|----------|-------|-----------------|
| 1 | Happy path — COD | `PaymentType = COD` | `Reserve` never called; status `PENDING_CRAFTSMAN_REVIEW` |
| 2 | Happy path — CC | `PaymentType = CC` | `Reserve` called; `PaymentReservationID` persisted; status `PAYMENT_RESERVED` |
| 3 | Craftsman ships CC order | Status moves to `SHIPPED` | `Capture` called with stored reservation ID |
| 4 | Craftsman declines CC order | Status moves to `DECLINED` | `Refund` called with stored reservation ID |
| 5 | Gateway `Reserve` fails | `mock.ForceNextError = errors.New("timeout")` | Order creation returns error; no order row saved |
| 6 | Gateway `Capture` fails | `mock.ForceNextError = errors.New("timeout")` | Shipment returns error; order status unchanged |
| 7 | Circuit breaker trips | 5 consecutive Reserve failures | 6th call fails immediately with "circuit open"; inner gateway not called |
| 8 | Circuit breaker recovers | Wait `openTimeout`, next call succeeds | State resets to CLOSED; subsequent calls flow through |

---

## Stripe Migration Path

When integrating Stripe, create `internal/modules/payment/stripe_gateway.go` implementing the same `Gateway` interface:

```go
type StripeGateway struct{ secretKey string }

func NewStripeGateway(secretKey string) *StripeGateway { ... }

func (s *StripeGateway) Reserve(ctx context.Context, req ReserveRequest) (Reservation, error) {
    // stripe.PaymentIntents.New with capture_method=manual
    // return Reservation{ID: pi.ID, Amount: req.Amount}
}

func (s *StripeGateway) Capture(ctx context.Context, reservationID string) error {
    // stripe.PaymentIntents.Capture(reservationID, nil)
}

func (s *StripeGateway) Refund(ctx context.Context, reservationID string) error {
    // stripe.PaymentIntents.Cancel(reservationID, nil)
}
```

No changes to the order services, container signature, or test structure — only `cmd/main.go` flips which concrete type is constructed.

---

## File Map

```
internal/modules/payment/
    gateway.go            ← Gateway interface + ReserveRequest/Reservation types
    mock_gateway.go       ← MockGateway (in-memory, ForceNextError for tests)
    breaker_gateway.go    ← BreakerGateway wrapping any Gateway
    stripe_gateway.go     ← (future) Stripe implementation
```