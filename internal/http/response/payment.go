package response

import "time"

// PaymentErrorDetails contains structured information about a payment failure.
type PaymentErrorDetails struct {
	Code      string `json:"code"`      // ErrorCode (e.g., "PAYMENT_CIRCUIT_OPEN")
	Reason    string `json:"reason"`    // ErrorReason (e.g., "circuit_open")
	Retryable bool   `json:"retryable"` // Whether the operation can be safely retried
	Timestamp string `json:"timestamp"` // RFC3339 timestamp of the error
}

// PaymentErrorResponse is the structure sent to clients when payment processing fails.
type PaymentErrorResponse struct {
	Error   string              `json:"error"` // User-friendly message
	Details PaymentErrorDetails `json:"details"`
}

// NewPaymentErrorResponse creates a structured payment error response.
func NewPaymentErrorResponse(userMessage string, code, reason string, retryable bool) PaymentErrorResponse {
	return PaymentErrorResponse{
		Error: userMessage,
		Details: PaymentErrorDetails{
			Code:      code,
			Reason:    reason,
			Retryable: retryable,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
}
