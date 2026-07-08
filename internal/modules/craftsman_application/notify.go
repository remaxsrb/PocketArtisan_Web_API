package craftsman_application

import (
	"PocketArtisan/internal/modules/mail"
	"context"
	"fmt"
	"html"
	"log"
	"os"
	"strings"
)

const logoPath = "./assets/logo.png"

func LoadLogo() []byte {
	logo, err := os.ReadFile(logoPath)
	if err != nil {
		log.Printf("craftsman application: logo not loaded from %s: %v (emails will be sent without it)", logoPath, err)
	}
	return logo
}

func SendDecisionEmail(ctx context.Context, mailer mail.Service, logo []byte, to string, approved bool, message string) {
	subject := "Ваша пријава за занатлију је одбијена"
	statusLine := "Нажалост, ваша пријава за занатлију је одбијена."
	if approved {
		subject = "Ваша пријава за занатлију је одобрена"
		statusLine = "Честитамо, ваша пријава за занатлију је одобрена."
	}

	body := fmt.Sprintf(`<div style="font-family: Arial, sans-serif; color: #333; line-height: 1.5;">
  <img src="cid:logo" alt="Направи Ми" style="max-width: 220px; margin-bottom: 16px;" />
  <p>%s</p>%s
</div>`, statusLine, messageParagraph(message))

	msg := mail.Message{To: to, Subject: subject, HTML: body}
	if len(logo) > 0 {
		msg.Attachments = []mail.Attachment{{
			Filename:    "logo.png",
			ContentType: "image/png",
			Content:     logo,
			ContentID:   "logo",
		}}
	}

	if err := mailer.Send(ctx, msg); err != nil {
		log.Printf("craftsman application: decision email to %s failed: %v", to, err)
	}
}

func messageParagraph(message string) string {
	if message == "" {
		return ""
	}
	escaped := strings.ReplaceAll(html.EscapeString(message), "\n", "<br>")
	return fmt.Sprintf("\n  <p>%s</p>", escaped)
}