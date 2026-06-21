package payment

import "context"

type ReserveRequest struct {
	OrderID     uint    `json:"order_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
}

type Reservation struct {
	ID     string  `json:"id"`
	Amount float64 `json:"amount"`
}

type Gateway interface {
	Reserve(ctx context.Context, req ReserveRequest) (Reservation, error)
	Capture(ctx context.Context, reservationID string) error
	Refund(ctx context.Context, reservationID string) error
}
