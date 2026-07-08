package mail

import (
    "fmt"
    "strings"
)

const (
    ProviderResend  = "resend"
    ProviderSMTPDev = "smtp4dev"
)

func NewService(provider string) (Service, error) {
    switch strings.ToLower(strings.TrimSpace(provider)) {
    case ProviderResend:
        return NewResendService(), nil
    case "", ProviderSMTPDev:
        return NewSMTPService(), nil
    default:
        return nil, fmt.Errorf("unsupported mail provider %q", provider)
    }
}