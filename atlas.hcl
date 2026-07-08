// Atlas configuration for PocketArtisan.
//
// The desired schema is derived from the GORM models via ./cmd/atlasloader.
// Versioned migrations live in ./migrations and are the single source of truth
// applied to every environment (local, Neon branches, staging, prod).
//
// Common commands (see Makefile for wrappers):
//   atlas migrate diff <name> --env local   # author a new migration from model changes
//   atlas migrate lint       --env local    # lint pending migrations (CI drift guard)
//   atlas migrate apply      --env local    # apply pending migrations to $DATABASE_URL

data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "./cmd/atlasloader",
  ]
}

// `local` is used for authoring/linting migrations. Atlas spins up a throwaway
// Postgres via Docker as the "dev" database to compute a clean diff, so Docker
// must be available when running `migrate diff`/`migrate lint`.
env "local" {
  src = data.external_schema.gorm.url
  dev = "docker://postgres/15/dev?search_path=public"

  migration {
    dir = "file://migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

// `deploy` is used to apply migrations to a real database (Neon branch, staging,
// prod). The target URL comes from $DATABASE_URL. No dev database is needed for
// apply, but `migrate lint` in CI still uses a dev DB (see the CI workflow).
//
// We bind the connection to the `public` schema (search_path=public) so Atlas
// stores its `atlas_schema_revisions` bookkeeping table there. Without a bound
// schema Atlas tries to use a separate `atlas_schema_revisions` schema, which
// fails on Neon with: relation "atlas_schema_revisions.atlas_schema_revisions"
// does not exist (42P01).
env "deploy" {
  url = urlqueryset(getenv("DATABASE_URL"), "search_path", "public")

  migration {
    dir = "file://migrations"
  }
}
