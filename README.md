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

1. Copy envs:

```bash
cp .env.example .env
```

2. Start dependencies:

```bash
make local-up
```

3. Apply DB migrations and seed from CSV:

```bash
make migrate
make seed
```

4. Build lambda and deploy to LocalStack:

```bash
make build-lambda
make localstack-deploy
```

5. Run tests:

```bash
make test
```

## Developer commands

- `make fmt`
- `make tidy`
- `make test`
- `make migrate`
- `make seed`
- `make db-reset`
- `make sqlc-generate`

## ADR

ADRs are managed with `adr-tools` in [`docs/adr`](docs/adr).
