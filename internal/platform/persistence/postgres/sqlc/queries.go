// Code generated manually to match sqlc expected output for this bootstrap.
package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Queries struct {
	db DBTX
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

type CountPharmaciesParams struct {
	State        *string
	City         *string
	Neighborhood *string
}

const countPharmacies = `-- name: CountPharmacies :one
SELECT COUNT(*)
FROM pharmacy_search_mv
WHERE
  ($1::text IS NULL OR state_normalized = $1)
  AND ($2::text IS NULL OR city_normalized = $2)
  AND ($3::text IS NULL OR neighborhood_normalized = $3)`

func (q *Queries) CountPharmacies(ctx context.Context, arg CountPharmaciesParams) (int64, error) {
	row := q.db.QueryRow(ctx, countPharmacies, arg.State, arg.City, arg.Neighborhood)
	var count int64
	err := row.Scan(&count)
	return count, err
}

type ListPharmaciesParams struct {
	State        *string
	City         *string
	Neighborhood *string
	Limit        int32
	Offset       int32
}

type ListPharmaciesRow struct {
	ID           string
	Cnpj         string
	Name         string
	Address      string
	Neighborhood string
	City         string
	State        string
}

const listPharmacies = `-- name: ListPharmacies :many
SELECT id, cnpj, name, address, neighborhood, city, state
FROM pharmacy_search_mv
WHERE
  ($1::text IS NULL OR state_normalized = $1)
  AND ($2::text IS NULL OR city_normalized = $2)
  AND ($3::text IS NULL OR neighborhood_normalized = $3)
ORDER BY name ASC
LIMIT $4 OFFSET $5`

func (q *Queries) ListPharmacies(ctx context.Context, arg ListPharmaciesParams) ([]ListPharmaciesRow, error) {
	rows, err := q.db.Query(ctx, listPharmacies, arg.State, arg.City, arg.Neighborhood, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListPharmaciesRow
	for rows.Next() {
		var i ListPharmaciesRow
		if err := rows.Scan(&i.ID, &i.Cnpj, &i.Name, &i.Address, &i.Neighborhood, &i.City, &i.State); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listStates = `-- name: ListStates :many
SELECT DISTINCT state
FROM pharmacy_search_mv
ORDER BY state ASC`

func (q *Queries) ListStates(ctx context.Context) ([]string, error) {
	rows, err := q.db.Query(ctx, listStates)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []string{}
	for rows.Next() {
		var item string
		if err := rows.Scan(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listCities = `-- name: ListCities :many
SELECT DISTINCT city
FROM pharmacy_search_mv
WHERE ($1::text IS NULL OR state_normalized = $1)
ORDER BY city ASC`

func (q *Queries) ListCities(ctx context.Context, state *string) ([]string, error) {
	rows, err := q.db.Query(ctx, listCities, state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []string{}
	for rows.Next() {
		var item string
		if err := rows.Scan(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type ListNeighborhoodsParams struct {
	State *string
	City  *string
}

const listNeighborhoods = `-- name: ListNeighborhoods :many
SELECT DISTINCT neighborhood
FROM pharmacy_search_mv
WHERE
  ($1::text IS NULL OR state_normalized = $1)
  AND ($2::text IS NULL OR city_normalized = $2)
ORDER BY neighborhood ASC`

func (q *Queries) ListNeighborhoods(ctx context.Context, arg ListNeighborhoodsParams) ([]string, error) {
	rows, err := q.db.Query(ctx, listNeighborhoods, arg.State, arg.City)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []string{}
	for rows.Next() {
		var item string
		if err := rows.Scan(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findPharmaciesByCnpjs = `-- name: FindPharmaciesByCnpjs :many
SELECT id, cnpj, name, address, neighborhood, city, state
FROM pharmacy_search_mv
WHERE cnpj = ANY($1::text[])
ORDER BY name ASC`

func (q *Queries) FindPharmaciesByCnpjs(ctx context.Context, cnpjs []string) ([]ListPharmaciesRow, error) {
	rows, err := q.db.Query(ctx, findPharmaciesByCnpjs, cnpjs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []ListPharmaciesRow{}
	for rows.Next() {
		var i ListPharmaciesRow
		if err := rows.Scan(&i.ID, &i.Cnpj, &i.Name, &i.Address, &i.Neighborhood, &i.City, &i.State); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
