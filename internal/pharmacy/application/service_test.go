package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/vagner/api-onde-farma/internal/pharmacy/domain"
)

type fakeRepo struct {
	filtersReceived domain.PharmacyFilters
	cnpjsReceived   []string
}

func (f *fakeRepo) ListPharmacies(_ context.Context, filters domain.PharmacyFilters) (domain.PharmacyPage, error) {
	f.filtersReceived = filters
	return domain.PharmacyPage{}, nil
}

func (f *fakeRepo) ListStates(_ context.Context) ([]string, error) {
	return []string{"MG"}, nil
}

func (f *fakeRepo) ListCities(_ context.Context, _ string) ([]string, error) {
	return []string{"BELO HORIZONTE"}, nil
}

func (f *fakeRepo) ListNeighborhoods(_ context.Context, _, _ string) ([]string, error) {
	return []string{"CENTRO"}, nil
}

func (f *fakeRepo) FindByCNPJs(_ context.Context, cnpjs []string) ([]domain.Pharmacy, error) {
	f.cnpjsReceived = cnpjs
	return []domain.Pharmacy{{ID: uuid.New(), CNPJ: cnpjs[0], Name: "A", Address: "B", Neighborhood: "C", City: "D", State: "MG"}}, nil
}

func TestListPharmaciesNormalizesFiltersAndPagination(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	_, err := svc.ListPharmacies(context.Background(), domain.PharmacyFilters{
		State:        " mg ",
		City:         " belo horizonte ",
		Neighborhood: " centro ",
		Page:         0,
		Limit:        500,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.filtersReceived.Page != 1 {
		t.Fatalf("expected page 1, got %d", repo.filtersReceived.Page)
	}
	if repo.filtersReceived.Limit != 50 {
		t.Fatalf("expected limit 50, got %d", repo.filtersReceived.Limit)
	}
	if repo.filtersReceived.State != "MG" || repo.filtersReceived.City != "BELO HORIZONTE" || repo.filtersReceived.Neighborhood != "CENTRO" {
		t.Fatalf("unexpected normalized filters: %+v", repo.filtersReceived)
	}
}

func TestFindByCNPJsNormalizesAndDeduplicates(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	_, err := svc.FindByCNPJs(context.Background(), []string{"21.651.625/0001-93", "21651625000193", "invalid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.cnpjsReceived) != 1 {
		t.Fatalf("expected one unique cnpj, got %d", len(repo.cnpjsReceived))
	}
	if repo.cnpjsReceived[0] != "21651625000193" {
		t.Fatalf("unexpected cnpj %s", repo.cnpjsReceived[0])
	}
}
