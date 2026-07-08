package reviewreminder

import (
	"PocketArtisan/internal/modules/mail"
	ordermod "PocketArtisan/internal/modules/order"
	usersmod "PocketArtisan/internal/modules/users"
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

const logoPath = "./assets/logo.png"

const emailBody = `<div style="font-family: Arial, sans-serif; color: #333; line-height: 1.5;">
  <img src="cid:logo" alt="Направи Ми" style="max-width: 220px; margin-bottom: 16px;" />
  <p>Прошло је седам дана од како је ваша поруџбина послата. Уколико сте задовољни производом, молимо Вас да оцените занатлију од ког сте га поручили.</p>
</div>`

type Service struct {
	orders ordermod.Repository
	users  usersmod.Repository
	mailer mail.Service
	delay  time.Duration
	logo   []byte
}

func NewService(orders ordermod.Repository, users usersmod.Repository, mailer mail.Service, delay time.Duration) *Service {
	logo, err := os.ReadFile(logoPath)
	if err != nil {
		log.Printf("review reminder: logo not loaded from %s: %v (emails will be sent without it)", logoPath, err)
	}
	return &Service{orders: orders, users: users, mailer: mailer, delay: delay, logo: logo}
}

func (s *Service) Execute(ctx context.Context) error {
	cutoff := time.Now().Add(-s.delay)

	pending, err := s.orders.ListPendingReviewReminders(ctx, cutoff)
	if err != nil {
		return fmt.Errorf("list pending review reminders: %w", err)
	}

	for _, order := range pending {
		customer, err := s.users.FindUserByID(ctx, order.CustomerID)
		if err != nil {
			log.Printf("review reminder: customer %d not found for order %d: %v", order.CustomerID, order.ID, err)
			continue
		}

		msg := mail.Message{
			To:      customer.Email,
			Subject: "Оцените занатлију",
			HTML:    emailBody,
		}
		if len(s.logo) > 0 {
			msg.Attachments = []mail.Attachment{{
				Filename:    "logo.png",
				ContentType: "image/png",
				Content:     s.logo,
				ContentID:   "logo",
			}}
		}

		sendErr := s.mailer.Send(ctx, msg)
		if sendErr != nil {
			log.Printf("review reminder: send failed for order %d: %v", order.ID, sendErr)
			continue // leave review_reminder_sent_at unset, retried next tick
		}

		if err := s.orders.MarkReviewReminderSent(ctx, order.ID, time.Now()); err != nil {
			log.Printf("review reminder: failed to mark order %d as sent: %v", order.ID, err)
		}
	}

	return nil
}
