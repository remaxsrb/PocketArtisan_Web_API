package mail

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const resendAPIURL = "https://api.resend.com/emails"

type ResendService struct {
	apiKey     string
	fromAddr   string
	httpClient *http.Client
}

func NewResendService() Service {
	return &ResendService{
		apiKey:     os.Getenv("RESEND_API_KEY"),
		fromAddr:   os.Getenv("MAIL_FROM_ADDRESS"),
		httpClient: &http.Client{},
	}
}

type resendAttachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	ContentType string `json:"content_type,omitempty"`
	ContentID   string `json:"content_id,omitempty"`
}

type resendRequest struct {
	From        string             `json:"from"`
	To          []string           `json:"to"`
	Subject     string             `json:"subject"`
	HTML        string             `json:"html"`
	Attachments []resendAttachment `json:"attachments,omitempty"`
}

func (s *ResendService) Send(ctx context.Context, msg Message) error {
	if s.apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY is not set")
	}
	if s.fromAddr == "" {
		return fmt.Errorf("MAIL_FROM_ADDRESS is not set")
	}

	payload := resendRequest{
		From:    s.fromAddr,
		To:      []string{msg.To},
		Subject: msg.Subject,
		HTML:    msg.HTML,
	}

	for _, att := range msg.Attachments {
		payload.Attachments = append(payload.Attachments, resendAttachment{
			Filename:    att.Filename,
			Content:     base64.StdEncoding.EncodeToString(att.Content),
			ContentType: att.ContentType,
			ContentID:   att.ContentID,
		})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal resend payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build resend request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("resend request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
