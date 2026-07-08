package mail

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"mime/multipart"
	"net"
	"net/smtp"
	"net/textproto"
	"os"
)

type SMTPService struct {
	host     string
	port     string
	fromAddr string
}

func NewSMTPService() Service {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "2525"
	}

	return &SMTPService{
		host:     host,
		port:     port,
		fromAddr: os.Getenv("MAIL_FROM_ADDRESS"),
	}
}

func (s *SMTPService) Send(ctx context.Context, msg Message) error {
	addr := net.JoinHostPort(s.host, s.port)

	raw := s.buildMessage(msg)

	// smtp4dev accepts unauthenticated plain SMTP locally, so no smtp.Auth is used.
	if err := smtp.SendMail(addr, nil, s.fromAddr, []string{msg.To}, raw); err != nil {
		return fmt.Errorf("smtp send failed: %w", err)
	}
	return nil
}

func (s *SMTPService) buildMessage(msg Message) []byte {
	var out bytes.Buffer

	if len(msg.Attachments) == 0 {
		writeHeader(&out, "From", s.fromAddr)
		writeHeader(&out, "To", msg.To)
		writeHeader(&out, "Subject", mime.QEncoding.Encode("UTF-8", msg.Subject))
		writeHeader(&out, "MIME-Version", "1.0")
		writeHeader(&out, "Content-Type", `text/html; charset="UTF-8"`)
		writeHeader(&out, "Content-Transfer-Encoding", "base64")
		out.WriteString("\r\n")
		out.Write(encodeBase64Lines([]byte(msg.HTML)))
		return out.Bytes()
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	htmlHeader := textproto.MIMEHeader{}
	htmlHeader.Set("Content-Type", `text/html; charset="UTF-8"`)
	htmlHeader.Set("Content-Transfer-Encoding", "base64")
	if part, err := mw.CreatePart(htmlHeader); err == nil {
		part.Write(encodeBase64Lines([]byte(msg.HTML)))
	}

	for _, att := range msg.Attachments {
		ct := att.ContentType
		if ct == "" {
			ct = "application/octet-stream"
		}
		h := textproto.MIMEHeader{}
		h.Set("Content-Type", ct)
		h.Set("Content-Transfer-Encoding", "base64")
		if att.ContentID != "" {
			h.Set("Content-ID", "<"+att.ContentID+">")
			h.Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, att.Filename))
		} else {
			h.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, att.Filename))
		}
		if part, err := mw.CreatePart(h); err == nil {
			part.Write(encodeBase64Lines(att.Content))
		}
	}
	mw.Close()

	writeHeader(&out, "From", s.fromAddr)
	writeHeader(&out, "To", msg.To)
	writeHeader(&out, "Subject", mime.QEncoding.Encode("UTF-8", msg.Subject))
	writeHeader(&out, "MIME-Version", "1.0")
	writeHeader(&out, "Content-Type", `multipart/related; boundary="`+mw.Boundary()+`"`)
	out.WriteString("\r\n")
	out.Write(body.Bytes())
	return out.Bytes()
}

func writeHeader(w *bytes.Buffer, key, value string) {
	fmt.Fprintf(w, "%s: %s\r\n", key, value)
}

// encodeBase64Lines base64-encodes data and wraps it at 76 characters per line
// with CRLF endings, as required for MIME base64 transfer encoding.
func encodeBase64Lines(data []byte) []byte {
	encoded := base64.StdEncoding.EncodeToString(data)
	var out bytes.Buffer
	for len(encoded) > 76 {
		out.WriteString(encoded[:76])
		out.WriteString("\r\n")
		encoded = encoded[76:]
	}
	out.WriteString(encoded)
	out.WriteString("\r\n")
	return out.Bytes()
}
