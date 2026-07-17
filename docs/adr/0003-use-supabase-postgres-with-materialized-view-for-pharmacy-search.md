# 3. Use Supabase Postgres with Materialized View for Pharmacy Search

Date: 2026-03-05

## Status

Accepted

## Context

The migration target is Supabase Postgres. Legacy behavior requires fast filtering for pharmacy list endpoints by state, city, and neighborhood, plus lookup by CNPJ. The source CSV is denormalized and small now, but expected to evolve.

## Decision

Use a normalized relational schema and a materialized view for query serving:

- Tables: `states`, `cities`, `neighborhoods`, `pharmacies`
- Constraints:
  - uppercase state code validation
  - CNPJ digits-only check (`14` chars)
  - uniqueness on CNPJ and normalized location names
- Materialized view: `pharmacy_search_mv`
  - joins normalized tables
  - exposes flattened search fields
  - stores normalized columns for indexed filtering
- Dedicated indexes on CNPJ and `(state, city, neighborhood)` normalized columns
- Initial load from `src/data/pharmacies.csv` through `scripts/import_csv.sql`

## Consequences

Positive:

- Search endpoints read from a flattened structure optimized for API queries.
- Normalized base tables preserve data quality and future extensibility.
- Refreshing the materialized view gives predictable read performance.

Trade-offs and risks:

- Data refresh workflow is required after imports/updates.
- Materialized view increases storage and operational complexity.
