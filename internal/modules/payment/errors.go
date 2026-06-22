package payment

import (
	"context"
	"errors"
	"fmt"
)

// ErrorCode represents a structured payment error identifier.
type ErrorCode string

const (
	// Circuit breaker errors
	ErrCodeCircuitOpen     ErrorCode = "PAYMENT_CIRCUIT_OPEN"
	ErrCodeCircuitHalfOpen ErrorCode = "PAYMENT_CIRCUIT_HALF_OPEN"

	// Reservation/capture errors
	ErrCodeReservationFailed ErrorCode = "PAYMENT_RESERVATION_FAILED"
	ErrCodeCaptureFailed     ErrorCode = "PAYMENT_CAPTURE_FAILED"
	ErrCodeRefundFailed      ErrorCode = "PAYMENT_REFUND_FAILED"

	// Generic gateway errors
	ErrCodeGatewayError ErrorCode = "PAYMENT_GATEWAY_ERROR"
)

// ErrorReason describes the underlying cause of a payment failure for client diagnosis.
type ErrorReason string

const (
	ReasonCircuitOpen        ErrorReason = "circuit_open"
	ReasonInsufficientFunds  ErrorReason = "insufficient_funds"
	ReasonInvalidCard        ErrorReason = "invalid_card"
	ReasonCardDeclined       ErrorReason = "card_declined"
	ReasonGatewayUnavailable ErrorReason = "gateway_unavailable"
	ReasonNetworkError       ErrorReason = "network_error"
	ReasonTimeout            ErrorReason = "timeout"
	ReasonUnknown            ErrorReason = "unknown"
)

// PaymentError wraps payment gateway errors with structured metadata.
type PaymentError struct {
	code      ErrorCode
	reason    ErrorReason
	retryable bool
	message   string
	err       error
}

// NewPaymentError creates a new structured payment error.
func NewPaymentError(code ErrorCode, reason ErrorReason, retryable bool, message string) *PaymentError {
	return &PaymentError{
		code:      code,
		reason:    reason,
		retryable: retryable,
		message:   message,
	}
}

// WrapPaymentError wraps an existing error with payment context.
func WrapPaymentError(err error, code ErrorCode, reason ErrorReason, retryable bool, message string) *PaymentError {
	return &PaymentError{
		code:      code,
		reason:    reason,
		retryable: retryable,
		message:   message,
		err:       err,
	}
}

func (e *PaymentError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.message, e.err)
	}
	return e.message
}

func (e *PaymentError) Unwrap() error {
	return e.err
}

// Code returns the structured error code.
func (e *PaymentError) Code() ErrorCode {
	return e.code
}

// Reason returns the error reason for client diagnostics.
func (e *PaymentError) Reason() ErrorReason {
	return e.reason
}

// IsRetryable returns true if the operation can be safely retried.
func (e *PaymentError) IsRetryable() bool {
	return e.retryable
}

// UserMessage returns a user-friendly error message.
func (e *PaymentError) UserMessage() string {
	return e.message
}

// AsPaymentError checks if an error is a PaymentError and returns it.
// Returns nil if err is not a PaymentError.
func AsPaymentError(err error) *PaymentError {
	var pe *PaymentError
	if errors.As(err, &pe) {
		return pe
	}
	return nil
}

// CategorizeGatewayError examines a gateway error string and determines the appropriate
// error code, reason, and retryability. This is a heuristic for the mock and real gateways.
func CategorizeGatewayError(err error) (code ErrorCode, reason ErrorReason, retryable bool) {
	if err == nil {
		return ErrCodeGatewayError, ReasonUnknown, false
	}

	errStr := err.Error()

	// Circuit breaker states
	if errStr == "payment gateway circuit open" {
		return ErrCodeCircuitOpen, ReasonCircuitOpen, true
	}
	if errStr == "payment gateway circuit half-open, probe in flight" {
		return ErrCodeCircuitHalfOpen, ReasonGatewayUnavailable, true
	}

	// Timeout errors are retryable
	if errors.Is(err, context.DeadlineExceeded) || errStr == "context deadline exceeded" {
		return ErrCodeGatewayError, ReasonTimeout, true
	}

	// Network-like errors are retryable
	if errors.Is(err, context.Canceled) || errStr == "context canceled" {
		return ErrCodeGatewayError, ReasonNetworkError, true
	}

	// Generic timeout keywords
	if containsAny(errStr, "timeout", "deadline", "connection reset") {
		return ErrCodeGatewayError, ReasonTimeout, true
	}

	// Network keywords
	if containsAny(errStr, "connection", "network", "unreachable", "unavailable") {
		return ErrCodeGatewayError, ReasonGatewayUnavailable, true
	}

	// Card-specific errors (not retryable with same card)
	if containsAny(errStr, "insufficient", "funds") {
		return ErrCodeReservationFailed, ReasonInsufficientFunds, false
	}
	if containsAny(errStr, "invalid", "card", "declined", "rejected") {
		return ErrCodeReservationFailed, ReasonCardDeclined, false
	}

	// Unknown errors are not retryable without more context
	return ErrCodeGatewayError, ReasonUnknown, false
}

// containsAny checks if s contains any of the substrings (case-insensitive).
func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			// Simple case-insensitive check
			for i := 0; i <= len(s)-len(substr); i++ {
				if matchIgnoreCase(s[i:i+len(substr)], substr) {
					return true
				}
			}
		}
	}
	return false
}

// matchIgnoreCase checks if a and b match case-insensitively.
func matchIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if toLower(a[i]) != toLower(b[i]) {
			return false
		}
	}
	return true
}

// toLower converts a byte to lowercase if it's an ASCII letter.
func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}
