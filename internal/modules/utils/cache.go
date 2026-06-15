package utils

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

func cacheVersionKey(namespace string) string {
	return fmt.Sprintf("cache:version:%s", namespace)
}

// GetCacheVersion returns namespace cache version. On any Redis issue, it safely falls back to version 1.
func GetCacheVersion(ctx context.Context, cache *redis.Client, namespace string) int64 {
	if cache == nil {
		return 1
	}

	key := cacheVersionKey(namespace)
	v, err := cache.Get(ctx, key).Int64()
	if err == nil {
		if v < 1 {
			_ = cache.Set(ctx, key, 1, 0).Err()
			return 1
		}
		return v
	}

	if err == redis.Nil {
		_ = cache.Set(ctx, key, 1, 0).Err()
		return 1
	}

	return 1
}

// BumpCacheVersion increments version for provided namespaces.
// This avoids delete/repopulate flows and lets getters naturally move to fresh keys.
func BumpCacheVersion(ctx context.Context, cache *redis.Client, namespaces ...string) {
	if cache == nil || len(namespaces) == 0 {
		return
	}

	for _, ns := range namespaces {
		if ns == "" {
			continue
		}
		_ = cache.Incr(ctx, cacheVersionKey(ns)).Err()
	}
}
