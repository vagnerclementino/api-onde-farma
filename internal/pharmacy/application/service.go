package application

import (
	"context"
	"regexp"
	"strings"

	"github.com/vagner/api-onde-farma/internal/pharmacy/domain"
)

const (
	defaultPage  = 1
	defaultLimit = 50
	maxLimit     = 50
	maxCNPJBatch = 500
)

var nonDigits = regexp.MustCompile(`\D`)

type Service struct {
	repo Reader
}

func NewService(repo Reader) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListPharmacies(ctx context.Context, filters domain.PharmacyFilters) (domain.PharmacyPage, error) {
	if filters.Page < 1 {
		filters.Page = defaultPage
	}
	if filters.Limit < 1 {
		filters.Limit = defaultLimit
	}
	if filters.Limit > maxLimit {
		filters.Limit = maxLimit
	}

	filters.State = strings.ToUpper(strings.TrimSpace(filters.State))
	filters.City = strings.ToUpper(strings.TrimSpace(filters.City))
	filters.Neighborhood = strings.ToUpper(strings.TrimSpace(filters.Neighborhood))

	return s.repo.ListPharmacies(ctx, filters)
}

func (s *Service) ListStates(ctx context.Context) ([]string, error) {
	return s.repo.ListStates(ctx)
}

func (s *Service) ListCities(ctx context.Context, state string) ([]string, error) {
	state = strings.ToUpper(strings.TrimSpace(state))
	return s.repo.ListCities(ctx, state)
}

func (s *Service) ListNeighborhoods(ctx context.Context, state string, city string) ([]string, error) {
	state = strings.ToUpper(strings.TrimSpace(state))
	city = strings.ToUpper(strings.TrimSpace(city))
	return s.repo.ListNeighborhoods(ctx, state, city)
}

func (s *Service) FindByCNPJs(ctx context.Context, cnpjs []string) ([]domain.Pharmacy, error) {
	if len(cnpjs) == 0 {
		return []domain.Pharmacy{}, nil
	}
	if len(cnpjs) > maxCNPJBatch {
		cnpjs = cnpjs[:maxCNPJBatch]
	}

	normalized := make([]string, 0, len(cnpjs))
	seen := make(map[string]struct{}, len(cnpjs))
	for _, cnpj := range cnpjs {
		clean := nonDigits.ReplaceAllString(strings.TrimSpace(cnpj), "")
		if len(clean) != 14 {
			continue
		}
		if _, exists := seen[clean]; exists {
			continue
		}
		seen[clean] = struct{}{}
		normalized = append(normalized, clean)
	}
	if len(normalized) == 0 {
		return []domain.Pharmacy{}, nil
	}

	return s.repo.FindByCNPJs(ctx, normalized)
}
