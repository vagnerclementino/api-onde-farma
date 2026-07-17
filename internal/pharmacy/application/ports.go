package application

import (
	"context"

	"github.com/vagner/api-onde-farma/internal/pharmacy/domain"
)

type Reader interface {
	ListPharmacies(ctx context.Context, filters domain.PharmacyFilters) (domain.PharmacyPage, error)
	ListStates(ctx context.Context) ([]string, error)
	ListCities(ctx context.Context, state string) ([]string, error)
	ListNeighborhoods(ctx context.Context, state string, city string) ([]string, error)
	FindByCNPJs(ctx context.Context, cnpjs []string) ([]domain.Pharmacy, error)
}
