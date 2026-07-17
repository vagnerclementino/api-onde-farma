package http

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/vagner/api-onde-farma/internal/pharmacy/application"
	"github.com/vagner/api-onde-farma/internal/pharmacy/domain"
)

type fakeRepo struct{}

func (f *fakeRepo) ListPharmacies(_ context.Context, _ domain.PharmacyFilters) (domain.PharmacyPage, error) {
	return domain.PharmacyPage{
		Data: []domain.Pharmacy{{
			ID:           uuid.MustParse("6d2f88fc-b58e-4ec1-9884-51ccce8b31f4"),
			CNPJ:         "21651625000193",
			Name:         "A BOTICA",
			Address:      "RUA X",
			Neighborhood: "CENTRO",
			City:         "BELO HORIZONTE",
			State:        "MG",
		}},
		Pagination: domain.Pagination{Page: 1, Limit: 50, Total: 1, TotalPages: 1},
	}, nil
}
func (f *fakeRepo) ListStates(_ context.Context) ([]string, error) { return []string{"MG"}, nil }
func (f *fakeRepo) ListCities(_ context.Context, _ string) ([]string, error) {
	return []string{"BELO HORIZONTE"}, nil
}
func (f *fakeRepo) ListNeighborhoods(_ context.Context, _, _ string) ([]string, error) {
	return []string{"CENTRO"}, nil
}
func (f *fakeRepo) FindByCNPJs(_ context.Context, _ []string) ([]domain.Pharmacy, error) {
	return []domain.Pharmacy{}, nil
}

func TestHandleListPharmacies(t *testing.T) {
	svc := application.NewService(&fakeRepo{})
	h := NewHandler(svc, []string{"http://localhost:3000"}, 1024)

	resp, err := h.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		RawPath: "/v1/pharmacies",
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: "GET"},
		},
		Headers: map[string]string{"origin": "http://localhost:3000"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(resp.Body), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if _, ok := body["data"]; !ok {
		t.Fatalf("missing data field")
	}
}

func TestHandleByCNPJRejectsLargeBody(t *testing.T) {
	svc := application.NewService(&fakeRepo{})
	h := NewHandler(svc, []string{"http://localhost:3000"}, 10)

	resp, err := h.Handle(context.Background(), events.APIGatewayV2HTTPRequest{
		RawPath: "/v1/pharmacies/by-cnpj",
		Body:    "{\"cnpjs\":[\"123\",\"456\",\"789\"]}",
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: "POST"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 413 {
		t.Fatalf("expected status 413, got %d", resp.StatusCode)
	}
}
