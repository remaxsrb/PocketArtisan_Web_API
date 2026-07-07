package response

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Envelope is the standard API response shape for all endpoints.
// Every response should include data/meta/error fields for consistency.
type Envelope struct {
	Data  interface{} `json:"data"`
	Meta  interface{} `json:"meta"`
	Error *APIError   `json:"error"`
}

type APIError struct {
	Message   string      `json:"message"`
	Code      string      `json:"code,omitempty"`
	Reason    string      `json:"reason,omitempty"`
	Retryable *bool       `json:"retryable,omitempty"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp string      `json:"timestamp,omitempty"`
}

type Builder struct {
	statusCode int
	envelope   Envelope
}

func NewBuilder(statusCode int) *Builder {
	return &Builder{
		statusCode: statusCode,
		envelope: Envelope{
			Data:  nil,
			Meta:  nil,
			Error: nil,
		},
	}
}

func (b *Builder) WithData(data interface{}) *Builder {
	b.envelope.Data = data
	return b
}

func (b *Builder) WithMeta(meta interface{}) *Builder {
	b.envelope.Meta = meta
	return b
}

func (b *Builder) WithError(err *APIError) *Builder {
	b.envelope.Error = err
	return b
}

func (b *Builder) Send(c *gin.Context) {
	c.JSON(b.statusCode, b.envelope)
}

func NewError(message string) *APIError {
	return &APIError{
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func Error(c *gin.Context, statusCode int, message string) {
	NewBuilder(statusCode).
		WithError(NewError(message)).
		Send(c)
}

func Data(c *gin.Context, statusCode int, data interface{}) {
	NewBuilder(statusCode).
		WithData(data).
		Send(c)
}

func Empty(c *gin.Context, statusCode int) {
	NewBuilder(statusCode).Send(c)
}
