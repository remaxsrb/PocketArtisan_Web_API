package get_registered_users_by_month

import (
	usersmod "PocketArtisan/internal/modules/users"
	"PocketArtisan/internal/modules/utils"
	"PocketArtisan/internal/modules/utils/timeutil"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Service struct {
	repo        usersmod.Repository
	cache       *redis.Client
	timeService timeutil.Service
}

func NewService(db *gorm.DB, cache *redis.Client, timeService timeutil.Service) *Service {
	return &Service{repo: usersmod.NewGormRepository(db), cache: cache, timeService: timeService}
}

func (uc *Service) Execute(ctx context.Context, req Request) (Response, error) {
	const cacheTTL = 5 * time.Minute

	from, to, err := uc.timeService.ParseDateRange(req.From, req.To)
	if err != nil {
		return Response{}, err
	}

	cacheVersion := utils.GetCacheVersion(ctx, uc.cache, "users")
	cacheKey := fmt.Sprintf("users:registered:by-month:v:%d:%s:%s", cacheVersion, req.From, req.To)

	if cached, err := uc.cache.Get(ctx, cacheKey).Result(); err == nil {
		var cachedResp Response
		if jsonErr := json.Unmarshal([]byte(cached), &cachedResp); jsonErr == nil {
			return cachedResp, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		fmt.Printf("Redis error: %v\n", err)
	}

	rows, err := uc.repo.CountNonAdminUsersByMonth(ctx, from, to)
	if err != nil {
		return Response{}, err
	}

	buckets := make([]Bucket, 0, len(rows))
	for _, row := range rows {
		buckets = append(buckets, Bucket{Month: row.Month.Format("2006-01"), Total: row.Total})
	}
	resp := Response{Buckets: buckets}

	if data, err := json.Marshal(resp); err == nil {
		_ = uc.cache.Set(ctx, cacheKey, data, cacheTTL).Err()
	}

	return resp, nil
}
