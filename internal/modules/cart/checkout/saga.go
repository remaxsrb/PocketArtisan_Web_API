package checkout

import (
	"context"
	"log"
)

type compensatorFunc func(ctx context.Context, orderID uint64, reservationID string)

type sagaEntry struct {
	orderID       uint64
	reservationID string
}

type CheckoutSaga struct {
	entries     []sagaEntry
	compensator compensatorFunc
}

func NewCheckoutSaga(compensator compensatorFunc) *CheckoutSaga {
	return &CheckoutSaga{compensator: compensator}
}

func (s *CheckoutSaga) Record(orderID uint64, reservationID string) {
	s.entries = append(s.entries, sagaEntry{orderID: orderID, reservationID: reservationID})
}

func (s *CheckoutSaga) Compensate(ctx context.Context) {
	for i := len(s.entries) - 1; i >= 0; i-- {
		entry := s.entries[i]
		log.Printf("checkout saga: compensating order %d", entry.orderID)
		s.compensator(ctx, entry.orderID, entry.reservationID)
	}
}
