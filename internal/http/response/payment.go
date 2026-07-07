package response

import "time"

// NewPaymentError creates an APIError compatible with the standard response envelope.
func NewPaymentError(userMessage string, code, reason string, retryable bool) *APIError {
	return &APIError{
		Message:   userMessage,
		Code:      code,
		Reason:    reason,
		Retryable: &retryable,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}
