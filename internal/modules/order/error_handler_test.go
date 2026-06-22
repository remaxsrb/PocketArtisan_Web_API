package order

import (
	"PocketArtisan/internal/modules/payment"
	"errors"
	"net/http"
	"testing"
)

func TestErrorHandler_IsPaymentError_TypedError(t *testing.T) {
	pe := payment.NewPaymentError(
		payment.ErrCodeReservationFailed,
		payment.ReasonCardDeclined,
		false,
		"Card declined",
	)
	if !isPaymentError(pe) {
		t.Error("isPaymentError should return true for PaymentError")
	}
}

func TestErrorHandler_IsPaymentError_WrappedError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{"payment reservation failed", "payment reservation failed: test", true},
		{"capture payment", "capture payment: circuit open", true},
		{"refund payment", "refund payment: gateway error", true},
		{"circuit state", "payment gateway circuit open", true},
		{"generic error", "order not found", false},
		{"database error", "connection refused", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			if result := isPaymentError(err); result != tt.expected {
				t.Errorf("isPaymentError(%q): want %v, got %v", tt.errMsg, tt.expected, result)
			}
		})
	}
}

func TestErrorHandler_MapPaymentErrorToHTTP_CircuitOpen(t *testing.T) {
	eh := NewErrorHandler()
	pe := payment.NewPaymentError(
		payment.ErrCodeCircuitOpen,
		payment.ReasonCircuitOpen,
		true,
		"Gateway unavailable",
	)

	statusCode, userMsg := eh.mapPaymentErrorToHTTP(pe)

	if statusCode != http.StatusServiceUnavailable {
		t.Errorf("statusCode: want %d, got %d", http.StatusServiceUnavailable, statusCode)
	}
	if userMsg == "" {
		t.Error("userMsg should not be empty")
	}
}

func TestErrorHandler_MapPaymentErrorToHTTP_ReservationFailed_InsufficientFunds(t *testing.T) {
	eh := NewErrorHandler()
	pe := payment.NewPaymentError(
		payment.ErrCodeReservationFailed,
		payment.ReasonInsufficientFunds,
		false,
		"Insufficient funds",
	)

	statusCode, userMsg := eh.mapPaymentErrorToHTTP(pe)

	if statusCode != http.StatusPaymentRequired {
		t.Errorf("statusCode: want %d, got %d", http.StatusPaymentRequired, statusCode)
	}
	if userMsg == "" {
		t.Error("userMsg should not be empty")
	}
	// User message should mention insufficient funds
	if !containsStr(userMsg, "Insufficient") {
		t.Errorf("userMsg should mention insufficient funds, got: %s", userMsg)
	}
}

func TestErrorHandler_MapPaymentErrorToHTTP_ReservationFailed_CardDeclined(t *testing.T) {
	eh := NewErrorHandler()
	pe := payment.NewPaymentError(
		payment.ErrCodeReservationFailed,
		payment.ReasonCardDeclined,
		false,
		"Card declined",
	)

	statusCode, userMsg := eh.mapPaymentErrorToHTTP(pe)

	if statusCode != http.StatusPaymentRequired {
		t.Errorf("statusCode: want %d, got %d", http.StatusPaymentRequired, statusCode)
	}
	if !containsStr(userMsg, "declined") {
		t.Errorf("userMsg should mention declined, got: %s", userMsg)
	}
}

func TestErrorHandler_MapPaymentErrorToHTTP_CaptureFailed(t *testing.T) {
	eh := NewErrorHandler()
	pe := payment.NewPaymentError(
		payment.ErrCodeCaptureFailed,
		payment.ReasonGatewayUnavailable,
		true,
		"Capture failed",
	)

	statusCode, userMsg := eh.mapPaymentErrorToHTTP(pe)

	if statusCode != http.StatusServiceUnavailable {
		t.Errorf("statusCode: want %d, got %d", http.StatusServiceUnavailable, statusCode)
	}
	if userMsg == "" {
		t.Error("userMsg should not be empty")
	}
}

func TestErrorHandler_MapPaymentErrorToHTTP_RefundFailed(t *testing.T) {
	eh := NewErrorHandler()
	pe := payment.NewPaymentError(
		payment.ErrCodeRefundFailed,
		payment.ReasonGatewayUnavailable,
		true,
		"Refund failed",
	)

	statusCode, userMsg := eh.mapPaymentErrorToHTTP(pe)

	if statusCode != http.StatusServiceUnavailable {
		t.Errorf("statusCode: want %d, got %d", http.StatusServiceUnavailable, statusCode)
	}
	if userMsg == "" {
		t.Error("userMsg should not be empty")
	}
}

func TestErrorHandler_MapPaymentErrorToHTTP_GatewayError_Retryable(t *testing.T) {
	eh := NewErrorHandler()
	pe := payment.NewPaymentError(
		payment.ErrCodeGatewayError,
		payment.ReasonTimeout,
		true, // retryable
		"Timeout",
	)

	statusCode, userMsg := eh.mapPaymentErrorToHTTP(pe)

	if statusCode != http.StatusServiceUnavailable {
		t.Errorf("statusCode: want %d, got %d", http.StatusServiceUnavailable, statusCode)
	}
	if userMsg == "" {
		t.Error("userMsg should not be empty")
	}
}

func TestErrorHandler_MapPaymentErrorToHTTP_GatewayError_NotRetryable(t *testing.T) {
	eh := NewErrorHandler()
	pe := payment.NewPaymentError(
		payment.ErrCodeGatewayError,
		payment.ReasonUnknown,
		false, // not retryable
		"Unknown error",
	)

	statusCode, userMsg := eh.mapPaymentErrorToHTTP(pe)

	if statusCode != http.StatusBadGateway {
		t.Errorf("statusCode: want %d, got %d", http.StatusBadGateway, statusCode)
	}
	if userMsg == "" {
		t.Error("userMsg should not be empty")
	}
}

func TestErrorHandler_CategorizeWrappedPaymentError_ReservationFailed(t *testing.T) {
	eh := NewErrorHandler()

	statusCode, userMsg, code, _, retryable := eh.categorizeWrappedPaymentError(
		"payment reservation failed: card declined",
	)

	if statusCode != http.StatusPaymentRequired {
		t.Errorf("statusCode: want %d, got %d", http.StatusPaymentRequired, statusCode)
	}
	if code != string(payment.ErrCodeReservationFailed) {
		t.Errorf("code: want %s, got %s", payment.ErrCodeReservationFailed, code)
	}
	if retryable {
		t.Error("retryable should be false for card declined")
	}
	if userMsg == "" {
		t.Error("userMsg should not be empty")
	}
}

func TestErrorHandler_CategorizeWrappedPaymentError_CircuitOpen(t *testing.T) {
	eh := NewErrorHandler()

	statusCode, userMsg, code, _, retryable := eh.categorizeWrappedPaymentError(
		"capture payment: payment gateway circuit open",
	)

	if statusCode != http.StatusServiceUnavailable {
		t.Errorf("statusCode: want %d, got %d", http.StatusServiceUnavailable, statusCode)
	}
	if code != string(payment.ErrCodeCircuitOpen) {
		t.Errorf("code: want %s, got %s", payment.ErrCodeCircuitOpen, code)
	}
	if !retryable {
		t.Error("retryable should be true for circuit open")
	}
	if userMsg == "" {
		t.Error("userMsg should not be empty")
	}
}

func TestErrorHandler_CategorizeWrappedPaymentError_CaptureFailed(t *testing.T) {
	eh := NewErrorHandler()

	statusCode, userMsg, code, _, retryable := eh.categorizeWrappedPaymentError(
		"capture payment: gateway timeout",
	)

	if statusCode != http.StatusServiceUnavailable {
		t.Errorf("statusCode: want %d, got %d", http.StatusServiceUnavailable, statusCode)
	}
	if code != string(payment.ErrCodeCaptureFailed) {
		t.Errorf("code: want %s, got %s", payment.ErrCodeCaptureFailed, code)
	}
	if !retryable {
		t.Error("retryable should be true for capture failure")
	}
	if userMsg == "" {
		t.Error("userMsg should not be empty")
	}
}

func TestErrorHandler_CategorizeWrappedPaymentError_RefundFailed(t *testing.T) {
	eh := NewErrorHandler()

	statusCode, userMsg, code, _, retryable := eh.categorizeWrappedPaymentError(
		"refund payment: gateway error",
	)

	if statusCode != http.StatusServiceUnavailable {
		t.Errorf("statusCode: want %d, got %d", http.StatusServiceUnavailable, statusCode)
	}
	if code != string(payment.ErrCodeRefundFailed) {
		t.Errorf("code: want %s, got %s", payment.ErrCodeRefundFailed, code)
	}
	if !retryable {
		t.Error("retryable should be true for refund failure")
	}
	if userMsg == "" {
		t.Error("userMsg should not be empty")
	}
}

// Helper functions

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if matchStr(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func matchStr(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if toLowerChar(a[i]) != toLowerChar(b[i]) {
			return false
		}
	}
	return true
}

func toLowerChar(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}
