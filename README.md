# PocketArtisan_Web_API

Backend API for **NapraviMi** — a two-sided marketplace connecting customers with craftsmen. Built in Go with Gin, organized as a modular monolith using vertical slices.

Live at `api.napravimi.com`. Frontend client: [NapraviMi_web_client](https://github.com/remaxsrb/NapraviMi_web_client).

---

## Architecture

The application is a **modular monolith**. Each business capability lives in its own vertical slice under `internal/modules/`, and every slice follows the same internal shape:

- `usecase` — business logic
- `controller` — HTTP handlers
- `dto` — request/response types (shared response shapes go in `common_dto.go` per module)
- `repository.go` — a `Repository` interface + `GormRepository` implementation, injected into the module's services

**Modules:** `auth`, `cart`, `crafts`, `craftsman_application`, `files`, `health`, `mail`, `order`, `payment`, `product`, `product_categories`, `users`, `utils`.

Three cross-cutting patterns apply across the codebase:

1. **Repository pattern** — every module talks to Postgres through a `Repository` interface backed by a `GormRepository` struct, never raw GORM calls scattered through business logic.
2. **Response envelope** — `internal/http/response` (`response.Data` / `response.Error` / `response.Empty`) standardizes API responses instead of ad-hoc `c.JSON`.
3. **Factory + decorator DI for external gateways** — external integrations (payment, mail) are built via a `New<X>Gateway(provider)` factory and wrapped by decorators (e.g. `payment.NewBreakerGateway(gateway, threshold, timeout)` for circuit breaking) wired once in `main.go` and injected as interfaces via the app container.

User roles: `admin`, `user`, `craftsman` (a `moderator` role is a planned future addition).

### Notable modules

- **`payment`** — gateway abstraction with a mock gateway for the current no-real-payment-processor prototype stage, wrapped in a circuit breaker decorator (`breaker_gateway.go`) to fail fast against a struggling payment provider. See `notes/Circuit Breaker pattern.md` for the state machine this implements.
- **`mail`** — factory + decorator gateway supporting Resend (production) and smtp4dev (local dev) providers.
- **`cart`** — supports splitting a single multi-craftsman cart into separate per-craftsman orders at checkout (`cart/checkout`).
- **Redis caching** — versioned write-through pattern (`cache:version:*` keys, bumped on writes) instead of delete-based invalidation, avoiding cache-stampede/race issues on invalidation.
- **Search** — Cyrillic/accented-text-aware matching via `utils.NormalizeForSearch`, backed by GIN-indexable `search_keywords` array columns on `Craft` and `ProductCategory`.

## Database & Migrations

Schema is managed with [Atlas](https://atlasgo.io) rather than GORM `AutoMigrate`:

- The desired schema is derived directly from GORM models via `cmd/atlasloader` (`gormschema.Load`).
- Versioned SQL migrations live in `migrations/`, with `atlas.sum` as the integrity hash — this is the source of truth, not the Go structs at runtime.
- `atlas.hcl` defines two environments: `local` (a throwaway Dockerized Postgres used for `migrate diff` / `migrate lint`) and `deploy` (targets `$DATABASE_URL`, pinned to `search_path=public` for Neon compatibility).
- Reference-data seeds (crafts, product categories) are generated via `cmd/seedgen` from `config.BuildSeedSQL()`. Only the admin user is seeded at application runtime — everything else is data-driven through migrations.

Common commands (see `Makefile` for the full set and required env files):

```bash
make migrate-diff env=dev      # generate a new migration from GORM model changes
make migrate-lint env=dev      # lint pending migrations
make migrate-apply env=dev     # apply migrations to a target database
make migrate-status env=dev
make migrate-baseline env=dev  # one-time adoption for a pre-existing DB
```

`env=<name>` reads connection details from a gitignored `.env.atlas.<name>` file; alternatively export `DATABASE_URL` directly.

### CI/CD for schema changes

Three GitHub Actions workflows under `.github/workflows/`:

- **`db-migrations.yml`** — runs on PRs touching migrations/entities/seed files. Spins up an ephemeral Neon branch, validates migration integrity, lints with Atlas (if `ATLAS_TOKEN` is set), applies migrations to the branch, runs a drift check (GORM models vs. committed migrations), then tears the branch down.
- **`db-deploy-dev.yml`** — applies migrations to the Neon `development` branch on push to `development`.
- **`db-deploy.yml`** — applies migrations to production on push to `master`.

## Tech Stack

| Concern | Choice |
|---|---|
| HTTP framework | Gin |
| ORM | GORM |
| Database | PostgreSQL (hosted on [Neon](https://neon.com)) |
| Migrations | Atlas |
| Cache | Redis (hosted on [Upstash](https://upstash.com)) |
| Document store | MongoDB (planned — see below) |
| Email | Resend (prod) / smtp4dev (local) |
| File storage | Cloudflare R2 (`cdn.napravimi.com`) |
| Deployment | Render (free tier), Docker |
| Scheduling | `robfig/cron/v3` |

## Local Setup

Requires Docker.

1. Create a `.env` file in the project root:

    ```
    ADMIN_EMAIL=sample@protonmail.com
    ADMIN_PASSWORD=Secure_pass1
    ADMIN_USERNAME=admin

    BASE_URL=http://localhost:8080

    CORS_ALLOWED_ORIGINS=http://localhost:4200

    POSTGRES_HOST=postgres
    POSTGRES_USER=test_user
    POSTGRES_PASSWORD=test_password
    POSTGRES_DB=test_db_name
    POSTGRES_PORT=5432

    REDIS_HOST=redis_cache
    REDIS_PORT=6379

    APP_PORT=8080
    APP_ENV=development

    JWT_SECRET=test_jwt_secret
    ```

2. Build and run:

    ```bash
    docker compose up --build
    ```

3. After any `docker compose down -v`, run `./app_init.sh` from the project root to clear the Go test cache and repopulate the database with seed data.

> Cloudflare's bot detection (Turnstile) is disabled in local builds so the init script can bulk-insert users.

## Notes

Design notes and implementation plans written during development (often with AI assistance) live in `/notes`, kept as reference and study material. Highlights: circuit breaker pattern, multi-craftsman checkout flow, Redis versioned cache invalidation, search optimization, payment error handling, and the (planned) end-to-end encrypted messaging and scheduled email digest features.
