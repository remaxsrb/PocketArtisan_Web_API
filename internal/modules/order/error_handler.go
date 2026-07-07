package order

import (
	"PocketArtisan/internal/http/response"
	"PocketArtisan/internal/modules/payment"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ErrorHandler provides centralized error handling for order-related operations.
// It converts domain errors into appropriate HTTP responses with structured error details.
type ErrorHandler struct{}

// NewErrorHandler creates a new ErrorHandler.
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// HandleOrderCreationError handles errors from order creation, with special handling for payment failures.
func (eh *ErrorHandler) HandleOrderCreationError(c *gin.Context, err error) {
	if isPaymentError(err) {
		eh.respondWithPaymentError(c, err)
		return
	}
	eh.respondWithGenericError(c, err)
}

// HandleOrderOperationError handles errors from order operations (ship, decline) that involve payment.
func (eh *ErrorHandler) HandleOrderOperationError(c *gin.Context, err error) {
	if isPaymentError(err) {
		eh.respondWithPaymentError(c, err)
		return
	}
	eh.respondWithGenericError(c, err)
}

// respondWithPaymentError converts a payment error to a structured HTTP response.
func (eh *ErrorHandler) respondWithPaymentError(c *gin.Context, err error) {
	// Check if it's a typed PaymentError
	var paymentErr *payment.PaymentError
	if errors.As(err, &paymentErr) {
		statusCode, userMessage := eh.mapPaymentErrorToHTTP(paymentErr)
		errResp := response.NewPaymentError(
			userMessage,
			string(paymentErr.Code()),
			string(paymentErr.Reason()),
			paymentErr.IsRetryable(),
		)
		response.NewBuilder(statusCode).WithError(errResp).Send(c)
		return
	}

	// Handle wrapped payment errors (string-based from service layer)
	statusCode, userMessage, code, reason, retryable := eh.categorizeWrappedPaymentError(err.Error())
	errResp := response.NewPaymentError(userMessage, code, reason, retryable)
	response.NewBuilder(statusCode).WithError(errResp).Send(c)
}

// respondWithGenericError returns a generic error response for non-payment errors.
func (eh *ErrorHandler) respondWithGenericError(c *gin.Context, err error) {
	response.Error(c, http.StatusBadRequest, err.Error())
}

// mapPaymentErrorToHTTP converts a typed PaymentError to HTTP status and user message.
func (eh *ErrorHandler) mapPaymentErrorToHTTP(err *payment.PaymentError) (statusCode int, userMessage string) {
	switch err.Code() {
	case payment.ErrCodeCircuitOpen:
		return http.StatusServiceUnavailable,
			"Payment gateway is temporarily unavailable. Please try again in a few moments."
	case payment.ErrCodeCircuitHalfOpen:
		return http.StatusServiceUnavailable,
			"Payment gateway is recovering. Please try again in a moment."
	case payment.ErrCodeReservationFailed:
		if err.Reason() == payment.ReasonInsufficientFunds {
			return http.StatusPaymentRequired,
				"Insufficient funds on your card. Please use a different payment method."
		}
		return http.StatusPaymentRequired,
			"Payment was declined. Please verify your card details and try again."
	case payment.ErrCodeCaptureFailed:
		return http.StatusServiceUnavailable,
			"Payment capture failed. Please try again or contact support."
	case payment.ErrCodeRefundFailed:
		return http.StatusServiceUnavailable,
			"Refund processing failed. Please try again or contact support."
	case payment.ErrCodeGatewayError:
		if err.IsRetryable() {
			return http.StatusServiceUnavailable,
				"Temporary payment processing error. Please try again."
		}
		return http.StatusBadGateway,
			"Payment gateway error. Please try again later."
	default:
		return http.StatusInternalServerError,
			"Payment processing encountered an unexpected error."
	}
}

// categorizeWrappedPaymentError examines a wrapped payment error message to determine
// the appropriate HTTP response, returning (statusCode, userMessage, code, reason, retryable).
func (eh *ErrorHandler) categorizeWrappedPaymentError(errMsg string) (int, string, string, string, bool) {
	lowerErr := strings.ToLower(errMsg)

	// Check for circuit breaker states first (highest priority)
	if strings.Contains(lowerErr, "circuit open") {
		return http.StatusServiceUnavailable,
			"Payment gateway is temporarily unavailable. Please try again in a few moments.",
			string(payment.ErrCodeCircuitOpen),
			string(payment.ReasonCircuitOpen),
			true
	}

	if strings.Contains(lowerErr, "circuit half-open") {
		return http.StatusServiceUnavailable,
			"Payment gateway is recovering. Please try again in a moment.",
			string(payment.ErrCodeCircuitHalfOpen),
			string(payment.ReasonGatewayUnavailable),
			true
	}

	// Reservation errors
	if strings.Contains(lowerErr, "payment reservation failed") {
		return http.StatusPaymentRequired,
			"Payment was declined. Please verify your card details and try again.",
			string(payment.ErrCodeReservationFailed),
			string(payment.ReasonCardDeclined),
			false
	}

	// Capture errors
	if strings.Contains(lowerErr, "capture payment") {
		return http.StatusServiceUnavailable,
			"Payment capture failed. Please try again or contact support.",
			string(payment.ErrCodeCaptureFailed),
			string(payment.ReasonGatewayUnavailable),
			true
	}

	// Refund errors
	if strings.Contains(lowerErr, "refund payment") {
		return http.StatusServiceUnavailable,
			"Refund processing failed. Please contact support.",
			string(payment.ErrCodeRefundFailed),
			string(payment.ReasonGatewayUnavailable),
			true
	}

	// Generic payment error
	return http.StatusBadGateway,
		"Payment processing encountered an error. Please try again.",
		string(payment.ErrCodeGatewayError),
		string(payment.ReasonUnknown),
		true
}

// isPaymentError checks if an error is related to payment processing.
func isPaymentError(err error) bool {
	// Check for typed PaymentError
	var paymentErr *payment.PaymentError
	if errors.As(err, &paymentErr) {
		return true
	}

	// Check for wrapped payment error messages
	errMsg := strings.ToLower(err.Error())
	paymentKeywords := []string{
		"payment reservation failed",
		"capture payment",
		"refund payment",
		"payment gateway circuit",
	}
	for _, kw := range paymentKeywords {
		if strings.Contains(errMsg, kw) {
			return true
		}
	}
	return false
}
