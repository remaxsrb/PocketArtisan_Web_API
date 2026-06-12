# Redis Versioned Write-Through Patch

## Patch Artifact

- Patch file: `redis_cache_versioned_write_through.patch`
- Location: project root

## Why This Patch Exists

This patch removes delete-based Redis cache invalidation and replaces it with versioned cache namespaces.

Goal:
- Avoid stale reads after DB writes
- Avoid expensive key pattern deletion and repopulate cycles
- Keep read path safe: cache first, DB second

## Core Design

### 1) Namespace Versioning

A new helper was added in `internal/modules/utils/cache.go`:

- `GetCacheVersion(ctx, cache, namespace)`
- `BumpCacheVersion(ctx, cache, namespaces...)`

Each cache namespace has a version key in Redis:

- `cache:version:users`
- `cache:version:craftsmen`
- `cache:version:products`
- `cache:version:crafts`
- `cache:version:product_categories`

### 2) Getter Strategy (Cache First)

Getters now build versioned keys, for example:

- `user:username:v:<ver>:<username>`
- `craftsmen:all:v:<ver>:skip:<skip>:limit:<limit>`
- `products:category:search:v:<ver>:<search>:skip:<skip>:limit:<limit>`

Flow:
1. Read current version
2. Try Redis with versioned key
3. On miss/unmarshal failure, query DB
4. Write fresh response to Redis under the same versioned key

### 3) Write Strategy (No Deletes)

After successful DB create/update/delete, write use-cases now call `BumpCacheVersion(...)` for affected namespaces.

This makes old cache entries logically obsolete without deleting them directly.

## Safety Check Updates

Registration flow now includes a cache-first username uniqueness check before DB fallback.

## Main Areas Included

- Users: register/login/change password/set profile picture/delete account/set role
- Craftsmen: list/search/create/rate
- Products: create/delete/list by craftsman/list by category
- Crafts and Product Categories: get all + write paths

## How To Apply This Patch

From project root:

```bash
git apply --index redis_cache_versioned_write_through.patch
```

If you do not want to stage immediately:

```bash
git apply redis_cache_versioned_write_through.patch
```

## Quick Validation

```bash
go test ./... -run '^$'
```

And then verify manually:

1. Call a getter endpoint twice and confirm second call is served from cache behavior-wise.
2. Perform a DB-changing operation (create/update/delete).
3. Call getter again and confirm new data is returned immediately.

## Notes

- Old versioned keys are intentionally not deleted. They become unused once the namespace version is bumped.
- This favors consistency and simpler write behavior; optional background TTL/cleanup can be added later if needed.
