package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vagner/api-onde-farma/internal/platform/persistence/postgres/sqlc"
)

func NewQueries(pool *pgxpool.Pool) *sqlc.Queries {
	return sqlc.New(pool)
}

func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}
