# 4. Adopt sqlc and Go 1.25 for Data Access and Build Baseline

Date: 2026-03-05

## Status

Accepted

## Context

The service must use modern Go, explicit SQL, and strong compile-time safety for database access. The repository is public OSS and should avoid runtime query surprises.

## Decision

- Set Go baseline to `1.25.0` in `go.mod`.
- Use `sqlc` with PostgreSQL/pgx v5 for typed query interfaces.
- Keep SQL in `db/queries` and schema/migrations in `db/migrations`.
- Generate data-access code under `internal/platform/persistence/postgres/sqlc`.
- Keep repository adapters dependent on generated query methods and domain mapping only.

## Consequences

Positive:

- Better type safety and early failure at compile time.
- SQL remains explicit, reviewable, and performance-tunable.
- Reduced risk of SQL injection through parameterized generated queries.

Trade-offs and risks:

- Requires `sqlc` generation step in developer workflow.
- Go 1.25 toolchain must be available locally/CI.
