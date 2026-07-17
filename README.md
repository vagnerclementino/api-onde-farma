# api-onde-farma

Backend REST API for Onde Farma using AWS Lambda behind API Gateway, with Postgres (Supabase-compatible), hexagonal architecture, ADR documentation, and secure defaults.

## Stack

- Go 1.25.0
- AWS Lambda + API Gateway (HTTP API)
- PostgreSQL (Supabase)
- sqlc + pgx/v5
- Local development: Docker, Docker Compose, LocalStack, Make

## Architecture

```
cmd/lambda                 # composition root
internal/pharmacy/domain   # pure domain
internal/pharmacy/application
internal/pharmacy/adapters/http
internal/pharmacy/adapters/repository
internal/platform/*        # config/http/persistence infra
db/migrations              # schema + materialized view
db/queries                 # sqlc queries
docs/adr                   # architecture decisions
```

## Endpoints

- `GET /v1/pharmacies?state=MG&city=BELO%20HORIZONTE&neighborhood=CENTRO&page=1&limit=50`
- `GET /v1/pharmacies/states`
- `GET /v1/pharmacies/cities?state=MG`
- `GET /v1/pharmacies/neighborhoods?state=MG&city=BELO%20HORIZONTE`
- `POST /v1/pharmacies/by-cnpj`

Body example:

```json
{
  "cnpjs": ["21651625000193", "11442517000157"]
}
```

## Local setup

1. Prerequisites:

```bash
docker --version
docker compose version
go version
psql --version
aws --version
```

`psql` local is optional. If not installed, the project uses `psql` inside the `postgres` container.

2. Configure environment variables:

```bash
cp .env.example .env
```

The `make` targets load `.env` automatically. Make sure `DATABASE_URL` exists in this file.

3. Start all local dependencies:

```bash
make local-up
```

4. Run database migrations and seed:

```bash
make migrate
make seed
```

5. Build and deploy Lambda to LocalStack:

```bash
make localstack-deploy
```

6. Or run everything in one command:

```bash
make run
```

7. Stop local environment:

```bash
make stop
```

## Troubleshooting local run

If you get this error:

```text
./scripts/migrate.sh: line 4: DATABASE_URL: DATABASE_URL must be set
```

Check:

1. You are running `make` from the project root.
2. The `.env` file exists.
3. The `.env` has `DATABASE_URL`, for example:

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/ondefarma?sslmode=disable
```

If you get this error:

```text
./scripts/migrate.sh: line X: psql: command not found
```

Run:

```bash
make local-up
```

Then run `make run` again. The migration/seed commands now fallback to `docker compose exec postgres psql` when local `psql` is not installed.

If you get this error:

```text
src/data/pharmacies.csv: No such file or directory
```

Recreate containers to apply updated `docker-compose` mounts:

```bash
make stop
make local-up
make run
```

## Developer commands

- `make fmt`
- `make tidy`
- `make test`
- `make migrate`
- `make seed`
- `make db-reset`
- `make sqlc-generate`
- `make run`
- `make stop`

## OpenAPI

The API contract is documented in [`openapi.yaml`](openapi.yaml).

## ADR

ADRs are managed with `adr-tools` in [`docs/adr`](docs/adr).
