# 5. Enforce Secure-by-Default API Controls for Public OSS Deployment

Date: 2026-03-05

## Status

Accepted

## Context

The project is open source with a public repository and internet-facing API. Inputs are attacker-controlled and endpoints are reachable through API Gateway. Security defaults must be embedded from the first implementation.

## Decision

Apply secure-by-default controls aligned with Go backend best practices:

- Explicit allowlist CORS (`ALLOWED_ORIGINS`), no wildcard by default
- Security response headers (`X-Content-Type-Options`, `X-Frame-Options`, `CSP`, `HSTS`, etc.)
- Request body size cap (`MAX_BODY_BYTES`) for POST payloads
- Strict input validation and normalization for query/body parameters
- Batch-size cap and normalization for CNPJ lookup
- Parameterized SQL access only
- Externalized secrets/configuration via environment variables (`DATABASE_URL`)
- Short request timeouts and context propagation

## Consequences

Positive:

- Reduces attack surface for common abuse vectors in public APIs.
- Improves consistency between local, staging, and production behavior.
- Keeps security controls visible in code and ADR documentation.

Trade-offs and risks:

- Misconfigured origin allowlist can block legitimate clients.
- Strict limits can require tuning for future use cases.
