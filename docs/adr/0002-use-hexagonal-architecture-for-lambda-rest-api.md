# 2. Use Hexagonal Architecture for Lambda REST API

Date: 2026-03-05

## Status

Accepted

## Context

The legacy `ondefarma` project exposes local API endpoints in Next.js route handlers. The new backend must run behind API Gateway and Lambda, while remaining maintainable for future adapters and providers.

The project requires alignment with the architecture constraints defined in `arch.md`: domain-focused structure, dependency injection in `cmd`, pure domain entities, explicit transport mapping, and `context.Context` propagation.

## Decision

Adopt a hexagonal architecture with explicit boundaries:

- `internal/pharmacy/domain`: pure business entities and value objects (no JSON or DB tags)
- `internal/pharmacy/application`: use-case service and repository interfaces (input/output ports)
- `internal/pharmacy/adapters/http`: API Gateway/Lambda transport mapping and request validation
- `internal/pharmacy/adapters/repository`: Postgres adapter implementing repository ports
- `internal/platform/*`: cross-cutting concerns (`config`, `httpx`, persistence bootstrap)
- `cmd/lambda/main.go`: composition root and dependency wiring

## Consequences

Positive:

- Business logic remains independent from AWS/event format and DB details.
- Enables easier tests with fake repository implementations.
- Future adapters (e.g., batch importer, gRPC, CLI) can be added without rewriting domain logic.

Trade-offs and risks:

- Higher number of packages and indirection for a small initial API.
- Requires discipline to keep transport/DB details out of domain and application layers.
