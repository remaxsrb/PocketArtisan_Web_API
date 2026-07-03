package turnstile

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const siteVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

// ErrMissingToken is returned when no Turnstile token is supplied for verification.
var ErrMissingToken = errors.New("turnstile: missing response token")

// ErrVerificationFailed is returned when Cloudflare reports the token as invalid.
var ErrVerificationFailed = errors.New("turnstile: verification failed")

// VerifyResponse mirrors the JSON body returned by the Siteverify API.
type VerifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
	Action      string   `json:"action"`
	CData       string   `json:"cdata"`
}

// Verifier validates Cloudflare Turnstile tokens against the Siteverify API.
type Verifier struct {
	secret string
	client *http.Client
}

// NewVerifier creates a Verifier using the provided widget secret key.
func NewVerifier(secret string) *Verifier {
	return &Verifier{
		secret: secret,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Verify sends the token to Cloudflare for validation. remoteIP is optional and
// may be left empty. It returns ErrVerificationFailed when the token is rejected.
func (v *Verifier) Verify(ctx context.Context, token, remoteIP string) (*VerifyResponse, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrMissingToken
	}

	form := url.Values{}
	form.Set("secret", v.secret)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, siteVerifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return &result, ErrVerificationFailed
	}

	return &result, nil
}
