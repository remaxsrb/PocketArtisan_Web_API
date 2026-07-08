# Database migrations (Atlas + GORM)

This project uses **[Atlas](https://atlasgo.io)** for versioned, up/down database
migrations, with **[Neon](https://neon.tech)** branches providing throwaway
copies of production for CI verification and drift detection.

The application **no longer runs `AutoMigrate`** at startup. Schema changes are
applied exclusively through the versioned SQL files in this directory.

---

## How it works

```
GORM models (internal/entities)            ← the desired schema
        │
        │  cmd/atlasloader  (gormschema.Load)
        ▼
   desired-state DDL  ──►  atlas migrate diff  ──►  migrations/*.sql (+ atlas.sum)
                                   ▲                          │
                          docker dev Postgres                 │  atlas migrate apply
                          (ephemeral, local/CI)               ▼
                                                     target database
```

- **Desired schema** = the GORM models. `cmd/atlasloader` dumps them to DDL so
  Atlas can compare "what the models say" against "what the migrations produce".
- **Migrations** in this folder are the source of truth for every environment.
  `atlas.sum` is an integrity checksum — CI fails if a file is edited without
  rehashing.
- **Reference-data seeds** (crafts, product categories, and their links) are
  generated from the authoritative Go data in `config/seed.go` via
  `cmd/seedgen` → `config.BuildSeedSQL()`. This keeps `search_keywords`
  normalization (`utils.NormalizeForSearch`) consistent with the app.
- The only thing seeded at **runtime** is the admin user (needs bcrypt + env),
  in `config.runSeeds()`.

### Configuration

- `atlas.hcl`
  - `env "local"` — used by `migrate diff` / `migrate lint`. Loads the schema
    from `cmd/atlasloader` and uses a **Docker** Postgres as a throwaway dev DB.
  - `env "deploy"` — used by `migrate apply` / `migrate status`. Targets
    `$DATABASE_URL`.

---

## Migration files

| File | Purpose |
| --- | --- |
| `20260708080000_baseline.sql` | Full initial schema (all tables, indexes, FKs) generated from the GORM models. |
| `20260708080100_seed_reference_data.sql` | Idempotent reference data: upserts crafts/categories and **rebuilds** the `craft_product_categories` join table (DELETE + INSERT…SELECT). |
| `atlas.sum` | Integrity checksum for all migrations. Never edit by hand. |

> The seed migration's rebuild of `craft_product_categories` permanently fixes
> the "empty categories array" bug that the old `ON CONFLICT DO NOTHING` seeding
> left behind (stale/removed links were never reconciled).

---

## Common tasks (Makefile)

`migrate diff` and `migrate lint` need **Docker** (Atlas starts a throwaway
Postgres as its dev database). `migrate apply` / `status` only need
`DATABASE_URL`.

```bash
# Author a new migration after changing GORM models
make migrate-diff name=add_something

# Lint pending migrations for destructive/irreversible changes
make migrate-lint

# Regenerate the reference-data seed as a NEW versioned migration
make seed-migration

# Recompute atlas.sum after a manual edit
make migrate-hash

# Apply / inspect against a database
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=disable make migrate-apply
DATABASE_URL=...                                              make migrate-status
```

### Adding a schema change

1. Edit the GORM models under `internal/entities/`.
2. `make migrate-diff name=<short_description>` (needs Docker).
3. Review the generated SQL, then commit it **together with `atlas.sum`**.

### Changing reference data

Because seeds are versioned, they no longer refresh on every boot — you author a
new migration instead:

1. Edit the seed data in `config/seed.go`.
2. `make seed-migration` (writes a new timestamped file + rehashes).
3. Review and commit.

---

## Deployment

### docker-compose

A one-shot **`migrate`** init container (`arigaio/atlas`) applies pending
migrations before the app starts:

- `postgres` has a healthcheck.
- `migrate` runs `migrate apply` once `postgres` is healthy.
- `app` waits for `migrate` to complete successfully.

### CI/CD (GitHub Actions)

- **`.github/workflows/db-migrations.yml`** (on PR): creates an ephemeral **Neon
  branch** (a copy of prod), lints migrations, applies them to the branch, runs
  a **drift check** (`atlas migrate diff` must produce nothing — i.e. migrations
  match the models), then deletes the branch. This is the "test against a copy
  of prod" + "drift check" that Neon gives us cheaply.
- **`.github/workflows/db-deploy.yml`** (on push to `master`): applies migrations
  to the **production** Neon branch.
- **`.github/workflows/db-deploy-dev.yml`** (on push to `development`): applies
  migrations to the **development** Neon branch.

### Branch mapping (Neon ⇄ repo)

| Repo branch | Neon branch | Workflow |
| --- | --- | --- |
| `master` | `production` | `db-deploy.yml` → applies to prod |
| `development` | `development` | `db-deploy-dev.yml` → applies to dev |
| pull requests | ephemeral clone of `production` (`vars.NEON_PARENT_BRANCH`) | `db-migrations.yml` → lint + apply + drift, then deleted |

PRs are validated against a copy of **production** (`NEON_PARENT_BRANCH=production`)
so a migration is verified against the exact schema/data it will ultimately be
applied to on `master`.

### End-to-end lifecycle

A change flows through the branches like this:

```
feature branch ──PR──► development ──PR──► master
      │                     │                  │
      │  db-migrations.yml  │  db-migrations   │  db-migrations
      │  (PR checks)        │  (PR checks)     │  (PR checks)
      ▼                     ▼                  ▼
  runs while the PR is open (lint + apply to an ephemeral
  clone of production + drift check), then the clone is deleted
                            │                  │
              merge = push to development       merge = push to master
                            ▼                  ▼
                   db-deploy-dev.yml     db-deploy.yml
                   apply → Neon dev      apply → Neon prod
                   (if migrations/**)    (if migrations/**)
```

**When you open a PR (into `development` or `master`):**
`db-migrations.yml` runs *during* the PR (on open and on every push to the PR
branch) — **if** the PR touches a watched path (`migrations/**`,
`internal/entities/**`, `cmd/atlasloader/**`, `config/seed.go`,
`config/seedsql.go`, `atlas.hcl`, or the workflow itself). It lints, applies to
an ephemeral production clone, and runs the drift check. It does **not** re-run
at merge (merge is a `push`, not a `pull_request`).

**When you merge/push to `development`:**
`db-deploy-dev.yml` runs — **if** the push changed `migrations/**` — and applies
pending migrations to the Neon **development** branch. No lint/drift here (that
lives in the PR workflow).

**When you merge/push to `master`:**
`db-deploy.yml` runs — **if** the push changed `migrations/**` — and applies
pending migrations to the Neon **production** branch.

> ⚠️ **Direct pushes bypass the checks.** A `git push` straight to `development`
> or `master` (not via a PR) triggers only the corresponding deploy workflow —
> **no lint, no drift check**. To guarantee validation, protect both branches
> and require pull requests (and the `db-migrations.yml` check) before merging.

Required GitHub configuration:

| Type | Name | Used by |
| --- | --- | --- |
| Secret | `NEON_API_KEY` | PR workflow (create/delete Neon branch) |
| Secret | `PROD_DATABASE_URL` | Deploy workflow (`atlas migrate apply` → production) |
| Secret | `DEV_DATABASE_URL` | Dev deploy workflow (`atlas migrate apply` → development) |
| Variable | `NEON_PROJECT_ID` | PR workflow |
| Variable | `NEON_PARENT_BRANCH` | PR workflow (set to `production`) |

> **Where these come from — not `.env`.** The workflows do **not** read your
> local `.env` (it is gitignored and never uploaded to GitHub). `${{ secrets.* }}`
> is resolved from **Settings → Secrets and variables → Actions → Secrets**, and
> `${{ vars.* }}` from the **Variables** tab in that same page. The deploy job
> injects the secret into the process environment:
>
> ```yaml
> env:
>   DATABASE_URL: ${{ secrets.PROD_DATABASE_URL }}
> run: atlas migrate apply --env deploy
> ```
>
> `atlas.hcl`'s `env "deploy"` then reads it via `getenv("DATABASE_URL")`. Your
> local `.env` is only used at runtime by the Go app / docker-compose on your
> machine — the two are entirely separate, so production values must be added to
> GitHub's settings for the workflows to work.

### Running migrations locally (without pasting connection strings)

To avoid typing Neon URLs on the command line (they'd leak into shell history and
`ps`), the Makefile can source them from **gitignored** per-environment files.
Copy the templates and fill in the direct (non-pooled) Neon URLs:

```bash
cp .env.atlas.dev.example  .env.atlas.dev    # edit → DATABASE_URL=postgres://…dev…
cp .env.atlas.prod.example .env.atlas.prod   # edit → DATABASE_URL=postgres://…prod…
```

Then pass `env=<name>` to any DB target — the URL is read from `.env.atlas.<name>`
and never appears on the command line:

```bash
make migrate-status   env=dev
make migrate-apply    env=prod
make migrate-baseline env=prod
```

`.env.atlas.*` is ignored by `.gitignore` (only the `*.example` templates are
committed). Passing no `env=` still works if you've exported `DATABASE_URL`.

---

## Adopting an existing database (run once)

If a database already has the schema (created by the old `AutoMigrate`), a plain
`apply` of the baseline would fail because the tables already exist. Mark the
baseline as already-applied **once per pre-existing environment**:

```bash
make migrate-baseline env=dev     # reads .env.atlas.dev
make migrate-baseline env=prod    # reads .env.atlas.prod
```

This registers `20260708080000_baseline` as applied and then runs the later
migrations (e.g. the seed). Brand-new databases and Neon branches cloned **after**
this step inherit the revision history and do **not** need baselining.

---

## Notes / gotchas

- `atlas.sum` must always match the migration files — run `make migrate-hash`
  after any manual edit, or CI will reject the change.
- `cmd/atlasloader`'s entity list must stay in sync with the GORM models. When
  you add a new entity, add it there too.
- Running `migrate diff`/`lint` locally requires Docker access. If `docker`
  needs `sudo` on your machine, add yourself to the group:
  `sudo usermod -aG docker $USER` (re-login required).
