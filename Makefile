# Database migration helpers (Atlas + GORM).
#
# The desired schema is derived from the GORM models via ./cmd/atlasloader and
# realized as versioned SQL migrations in ./migrations. Reference-data seeds are
# generated from config/seed.go via ./cmd/seedgen.
#
# `migrate diff` and `migrate lint` need Docker (Atlas spins a throwaway Postgres
# as its dev database). `migrate apply` only needs DATABASE_URL.

ATLAS ?= atlas
MIGRATIONS_DIR := file://migrations
TS := $(shell date +%Y%m%d%H%M%S)

# Loads DATABASE_URL without typing it on the command line. Pass env=<name> and
# the connection string is sourced from the gitignored file .env.atlas.<name>
# (e.g. `make migrate-status env=prod` reads .env.atlas.prod). That file should
# contain a single line: DATABASE_URL=postgres://...:...@...neon.tech/db?sslmode=require
# The value is read literally (not shell-sourced), so special characters such as
# the `&` in Neon's `&channel_binding=require` are safe and need no quoting.
# If env= is omitted, an already-exported $$DATABASE_URL is used instead.
define LOAD_DB
if [ -n "$(env)" ]; then \
  f=".env.atlas.$(env)"; \
  test -f "$$f" || { echo "missing $$f" >&2; exit 1; }; \
  DATABASE_URL="$$(sed -n 's/^[[:space:]]*DATABASE_URL=//p' "$$f" | tail -n1)"; \
  DATABASE_URL="$${DATABASE_URL%\"}"; DATABASE_URL="$${DATABASE_URL#\"}"; \
  export DATABASE_URL; \
fi; \
test -n "$$DATABASE_URL" || { echo "pass env=<name> (reads .env.atlas.<name>) or export DATABASE_URL" >&2; exit 1; };
endef

.PHONY: migrate-diff migrate-lint migrate-apply migrate-status migrate-hash seed-migration migrate-baseline

## migrate-baseline: one-time adoption on an EXISTING database that already has
## the schema (created by the old AutoMigrate). Marks the baseline migration as
## already-applied, then runs later migrations (e.g. the seed). Run once per
## pre-existing environment; brand-new/Neon-cloned DBs do NOT need this.
##   make migrate-baseline env=dev     (reads .env.atlas.dev)
##   make migrate-baseline env=prod    (reads .env.atlas.prod)
migrate-baseline:
	@$(LOAD_DB) $(ATLAS) migrate apply --env deploy --baseline 20260708080000


## migrate-diff name=<change>: author a new migration from model changes
migrate-diff:
	@test -n "$(name)" || (echo "usage: make migrate-diff name=<change>" && exit 1)
	$(ATLAS) migrate diff $(name) --env local

## migrate-lint: lint pending migrations for destructive/irreversible changes
migrate-lint:
	$(ATLAS) migrate lint --env local --latest 1

## migrate-apply: apply pending migrations. Pass env=<name> to read the URL from
## .env.atlas.<name>, or export DATABASE_URL yourself.
migrate-apply:
	@$(LOAD_DB) $(ATLAS) migrate apply --env deploy

## migrate-status: show applied/pending migrations. Pass env=<name> to read the
## URL from .env.atlas.<name>, or export DATABASE_URL yourself.
migrate-status:
	@$(LOAD_DB) $(ATLAS) migrate status --env deploy

## migrate-hash: recompute atlas.sum after manual edits
migrate-hash:
	$(ATLAS) migrate hash --dir $(MIGRATIONS_DIR)

## seed-migration: regenerate reference-data seed as a NEW versioned migration
seed-migration:
	go run ./cmd/seedgen > migrations/$(TS)_seed_reference_data.sql
	$(ATLAS) migrate hash --dir $(MIGRATIONS_DIR)
	@echo "created migrations/$(TS)_seed_reference_data.sql"
