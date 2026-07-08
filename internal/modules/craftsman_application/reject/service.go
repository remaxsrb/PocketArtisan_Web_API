package reject

import (
	camod "PocketArtisan/internal/modules/craftsman_application"
	"PocketArtisan/internal/modules/mail"
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo   camod.Repository
	cache  *redis.Client
	mailer mail.Service
	logo   []byte
}

func NewService(db *gorm.DB, cache *redis.Client, mailer mail.Service) *Service {
	return &Service{repo: camod.NewGormRepository(db), cache: cache, mailer: mailer, logo: camod.LoadLogo()}
}

func (uc *Service) Execute(ctx context.Context, req Request) error {
	ca, err := uc.repo.FindByID(ctx, uint64(req.ApplicationID))

	if err != nil {
		return errors.New("application not found")
	}

	nextStatus, err := camod.NextApplicationStatus(ca.Status, camod.ApplicationActionReject)
	if err != nil {
		return err
	}
	ca.Status = nextStatus
	resolvedAt := time.Now()
	ca.ResolvedAt = &resolvedAt

	if err := uc.repo.Save(ctx, ca); err != nil {
		return err
	}

	camod.SendDecisionEmail(ctx, uc.mailer, uc.logo, ca.Email, false, req.Message)
	return nil
}
