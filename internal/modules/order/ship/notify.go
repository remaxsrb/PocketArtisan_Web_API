package ship

import (
	"context"
	"fmt"
	"html"
	"log"
	"os"

	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/mail"
)

const logoPath = "./assets/logo.png"

func loadLogo() []byte {
	logo, err := os.ReadFile(logoPath)
	if err != nil {
		log.Printf("order ship: logo not loaded from %s: %v (emails will be sent without it)", logoPath, err)
	}
	return logo
}

func sendShippedEmail(ctx context.Context, mailer mail.Service, logo []byte, to string, order *entities.Order) {
	if mailer == nil || to == "" {
		return
	}

	body := fmt.Sprintf(`<div style="font-family: Arial, sans-serif; color: #333; line-height: 1.5;">
  <img src="cid:logo" alt="Направи Ми" style="max-width: 220px; margin-bottom: 16px;" />
  <p>Ваша поруџбина #%d је послата.</p>
  <p>Биће испоручена на адресу: %s</p>
</div>`, order.ID, html.EscapeString(order.CustomerAddress))

	msg := mail.Message{To: to, Subject: "Ваша поруџбина је послата", HTML: body}
	if len(logo) > 0 {
		msg.Attachments = []mail.Attachment{{
			Filename:    "logo.png",
			ContentType: "image/png",
			Content:     logo,
			ContentID:   "logo",
		}}
	}

	if err := mailer.Send(ctx, msg); err != nil {
		log.Printf("order ship: shipped email to %s failed for order %d: %v", to, order.ID, err)
	}
}