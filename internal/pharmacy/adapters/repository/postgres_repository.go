package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/vagner/api-onde-farma/internal/pharmacy/domain"
	"github.com/vagner/api-onde-farma/internal/platform/persistence/postgres/sqlc"
)

type PostgresRepository struct {
	q *sqlc.Queries
}

func NewPostgresRepository(q *sqlc.Queries) *PostgresRepository {
	return &PostgresRepository{q: q}
}

func (r *PostgresRepository) ListPharmacies(ctx context.Context, filters domain.PharmacyFilters) (domain.PharmacyPage, error) {
	offset := (filters.Page - 1) * filters.Limit
	count, err := r.q.CountPharmacies(ctx, sqlc.CountPharmaciesParams{
		State:        nullable(filters.State),
		City:         nullable(filters.City),
		Neighborhood: nullable(filters.Neighborhood),
	})
	if err != nil {
		return domain.PharmacyPage{}, err
	}

	rows, err := r.q.ListPharmacies(ctx, sqlc.ListPharmaciesParams{
		State:        nullable(filters.State),
		City:         nullable(filters.City),
		Neighborhood: nullable(filters.Neighborhood),
		Limit:        int32(filters.Limit),
		Offset:       int32(offset),
	})
	if err != nil {
		return domain.PharmacyPage{}, err
	}

	data := make([]domain.Pharmacy, 0, len(rows))
	for _, row := range rows {
		id, parseErr := uuid.Parse(row.ID)
		if parseErr != nil {
			continue
		}
		data = append(data, domain.Pharmacy{
			ID:           id,
			CNPJ:         row.Cnpj,
			Name:         row.Name,
			Address:      row.Address,
			Neighborhood: row.Neighborhood,
			City:         row.City,
			State:        row.State,
		})
	}

	totalPages := int(count) / filters.Limit
	if int(count)%filters.Limit != 0 {
		totalPages++
	}

	return domain.PharmacyPage{
		Data: data,
		Pagination: domain.Pagination{
			Page:        filters.Page,
			Limit:       filters.Limit,
			Total:       int(count),
			TotalPages:  totalPages,
			HasNextPage: filters.Page < totalPages,
			HasPrevPage: filters.Page > 1,
		},
	}, nil
}

func (r *PostgresRepository) ListStates(ctx context.Context) ([]string, error) {
	rows, err := r.q.ListStates(ctx)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PostgresRepository) ListCities(ctx context.Context, state string) ([]string, error) {
	rows, err := r.q.ListCities(ctx, nullable(state))
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PostgresRepository) ListNeighborhoods(ctx context.Context, state string, city string) ([]string, error) {
	rows, err := r.q.ListNeighborhoods(ctx, sqlc.ListNeighborhoodsParams{State: nullable(state), City: nullable(city)})
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PostgresRepository) FindByCNPJs(ctx context.Context, cnpjs []string) ([]domain.Pharmacy, error) {
	rows, err := r.q.FindPharmaciesByCnpjs(ctx, cnpjs)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Pharmacy, 0, len(rows))
	for _, row := range rows {
		id, parseErr := uuid.Parse(row.ID)
		if parseErr != nil {
			continue
		}
		out = append(out, domain.Pharmacy{
			ID:           id,
			CNPJ:         row.Cnpj,
			Name:         row.Name,
			Address:      row.Address,
			Neighborhood: row.Neighborhood,
			City:         row.City,
			State:        row.State,
		})
	}
	return out, nil
}

func nullable(input string) *string {
	if input == "" {
		return nil
	}
	return &input
}
