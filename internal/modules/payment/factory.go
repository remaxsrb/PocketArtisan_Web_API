package payment

import (
	"fmt"
	"strings"
)

const (
	ProviderMock = "mock"
)

func NewGateway(provider string) (Gateway, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", ProviderMock:
		return NewMockGateway(), nil
	default:
		return nil, fmt.Errorf("unsupported payment provider %q", provider)
	}
}
