-- name: CountPharmacies :one
SELECT COUNT(*)
FROM pharmacy_search_mv
WHERE
  ($1::text IS NULL OR state_normalized = $1)
  AND ($2::text IS NULL OR city_normalized = $2)
  AND ($3::text IS NULL OR neighborhood_normalized = $3);

-- name: ListPharmacies :many
SELECT id, cnpj, name, address, neighborhood, city, state
FROM pharmacy_search_mv
WHERE
  ($1::text IS NULL OR state_normalized = $1)
  AND ($2::text IS NULL OR city_normalized = $2)
  AND ($3::text IS NULL OR neighborhood_normalized = $3)
ORDER BY name ASC
LIMIT $4 OFFSET $5;

-- name: ListStates :many
SELECT DISTINCT state
FROM pharmacy_search_mv
ORDER BY state ASC;

-- name: ListCities :many
SELECT DISTINCT city
FROM pharmacy_search_mv
WHERE ($1::text IS NULL OR state_normalized = $1)
ORDER BY city ASC;

-- name: ListNeighborhoods :many
SELECT DISTINCT neighborhood
FROM pharmacy_search_mv
WHERE
  ($1::text IS NULL OR state_normalized = $1)
  AND ($2::text IS NULL OR city_normalized = $2)
ORDER BY neighborhood ASC;

-- name: FindPharmaciesByCnpjs :many
SELECT id, cnpj, name, address, neighborhood, city, state
FROM pharmacy_search_mv
WHERE cnpj = ANY($1::text[])
ORDER BY name ASC;
