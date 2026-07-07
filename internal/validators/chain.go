package validators

import "errors"

// ValidationContext carries values that can be validated by chained handlers.
type ValidationContext struct {
	Email    string
	Password string
}

// Handler defines a chainable validation step.
type Handler interface {
	SetNext(next Handler) Handler
	Handle(ctx *ValidationContext) error
}

type baseHandler struct {
	next Handler
}

func (h *baseHandler) SetNext(next Handler) Handler {
	h.next = next
	return next
}

func (h *baseHandler) nextOrNil(ctx *ValidationContext) error {
	if h.next == nil {
		return nil
	}
	return h.next.Handle(ctx)
}

type EmailFormatHandler struct {
	baseHandler
}

func NewEmailFormatHandler() *EmailFormatHandler {
	return &EmailFormatHandler{}
}

func (h *EmailFormatHandler) Handle(ctx *ValidationContext) error {
	if !IsValidEmail(ctx.Email) {
		return errors.New("invalid email")
	}
	return h.nextOrNil(ctx)
}

type PasswordPolicyHandler struct {
	baseHandler
}

func NewPasswordPolicyHandler() *PasswordPolicyHandler {
	return &PasswordPolicyHandler{}
}

func (h *PasswordPolicyHandler) Handle(ctx *ValidationContext) error {
	if err := ValidatePassword(ctx.Password); err != nil {
		return err
	}
	return h.nextOrNil(ctx)
}
