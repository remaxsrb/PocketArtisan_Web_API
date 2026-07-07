package approve

import (
	camod "PocketArtisan/internal/modules/craftsman_application"
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo  camod.Repository
	cache *redis.Client
}

func NewService(db *gorm.DB, cache *redis.Client) *Service {
	return &Service{repo: camod.NewGormRepository(db), cache: cache}
}

func (uc *Service) Execute(ctx context.Context, req Request) error {
	ca, err := uc.repo.FindByID(ctx, uint64(req.ApplicationID))

	if err != nil {
		return errors.New("application not found")
	}

	nextStatus, err := camod.NextApplicationStatus(ca.Status, camod.ApplicationActionApprove)
	if err != nil {
		return err
	}
	ca.Status = nextStatus
	resolvedAt := time.Now()
	ca.ResolvedAt = &resolvedAt

	return uc.repo.Save(ctx, ca)
}
