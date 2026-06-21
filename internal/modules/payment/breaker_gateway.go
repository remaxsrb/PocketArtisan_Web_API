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
