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
