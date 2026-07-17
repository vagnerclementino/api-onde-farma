package http

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/vagner/api-onde-farma/internal/pharmacy/application"
	"github.com/vagner/api-onde-farma/internal/pharmacy/domain"
	"github.com/vagner/api-onde-farma/internal/platform/httpx"
)

type Handler struct {
	service        *application.Service
	allowedOrigins []string
	maxBodyBytes   int
}

func NewHandler(service *application.Service, allowedOrigins []string, maxBodyBytes int) *Handler {
	return &Handler{service: service, allowedOrigins: allowedOrigins, maxBodyBytes: maxBodyBytes}
}

func (h *Handler) Handle(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	origin := req.Headers["origin"]
	if req.RequestContext.HTTP.Method == "OPTIONS" {
		return httpx.ResolvePreflight(origin, h.allowedOrigins), nil
	}

	requestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	path := normalizePath(req.RawPath)
	method := req.RequestContext.HTTP.Method

	switch {
	case method == "GET" && path == "/v1/pharmacies":
		return h.listPharmacies(requestCtx, req, origin)
	case method == "GET" && path == "/v1/pharmacies/states":
		return h.listStates(requestCtx, origin)
	case method == "GET" && path == "/v1/pharmacies/cities":
		return h.listCities(requestCtx, req, origin)
	case method == "GET" && path == "/v1/pharmacies/neighborhoods":
		return h.listNeighborhoods(requestCtx, req, origin)
	case method == "POST" && path == "/v1/pharmacies/by-cnpj":
		return h.byCNPJ(requestCtx, req, origin)
	default:
		return httpx.JSON(404, origin, map[string]string{"error": "not found"}, h.allowedOrigins)
	}
}

func (h *Handler) listPharmacies(ctx context.Context, req events.APIGatewayV2HTTPRequest, origin string) (events.APIGatewayV2HTTPResponse, error) {
	page := parsePositiveInt(req.QueryStringParameters["page"], 1)
	limit := parsePositiveInt(req.QueryStringParameters["limit"], 50)

	result, err := h.service.ListPharmacies(ctx, domain.PharmacyFilters{
		State:        req.QueryStringParameters["state"],
		City:         req.QueryStringParameters["city"],
		Neighborhood: req.QueryStringParameters["neighborhood"],
		Page:         page,
		Limit:        limit,
	})
	if err != nil {
		return httpx.JSON(500, origin, map[string]string{"error": "internal server error"}, h.allowedOrigins)
	}

	return httpx.JSON(200, origin, map[string]any{"data": toResponseList(result.Data), "pagination": result.Pagination}, h.allowedOrigins)
}

func (h *Handler) listStates(ctx context.Context, origin string) (events.APIGatewayV2HTTPResponse, error) {
	states, err := h.service.ListStates(ctx)
	if err != nil {
		return httpx.JSON(500, origin, map[string]string{"error": "internal server error"}, h.allowedOrigins)
	}
	return httpx.JSON(200, origin, states, h.allowedOrigins)
}

func (h *Handler) listCities(ctx context.Context, req events.APIGatewayV2HTTPRequest, origin string) (events.APIGatewayV2HTTPResponse, error) {
	state := strings.TrimSpace(req.QueryStringParameters["state"])
	if state == "" {
		return httpx.JSON(400, origin, map[string]string{"error": "state parameter is required"}, h.allowedOrigins)
	}
	cities, err := h.service.ListCities(ctx, state)
	if err != nil {
		return httpx.JSON(500, origin, map[string]string{"error": "internal server error"}, h.allowedOrigins)
	}
	return httpx.JSON(200, origin, cities, h.allowedOrigins)
}

func (h *Handler) listNeighborhoods(ctx context.Context, req events.APIGatewayV2HTTPRequest, origin string) (events.APIGatewayV2HTTPResponse, error) {
	state := strings.TrimSpace(req.QueryStringParameters["state"])
	if state == "" {
		return httpx.JSON(400, origin, map[string]string{"error": "state parameter is required"}, h.allowedOrigins)
	}
	city := strings.TrimSpace(req.QueryStringParameters["city"])
	if city == "" {
		return httpx.JSON(400, origin, map[string]string{"error": "city parameter is required"}, h.allowedOrigins)
	}
	neighborhoods, err := h.service.ListNeighborhoods(ctx, state, city)
	if err != nil {
		return httpx.JSON(500, origin, map[string]string{"error": "internal server error"}, h.allowedOrigins)
	}
	return httpx.JSON(200, origin, neighborhoods, h.allowedOrigins)
}

func (h *Handler) byCNPJ(ctx context.Context, req events.APIGatewayV2HTTPRequest, origin string) (events.APIGatewayV2HTTPResponse, error) {
	body := []byte(req.Body)
	if req.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			return httpx.JSON(400, origin, map[string]string{"error": "invalid base64 body"}, h.allowedOrigins)
		}
		body = decoded
	}

	if len(body) > h.maxBodyBytes {
		return httpx.JSON(413, origin, map[string]string{"error": "payload too large"}, h.allowedOrigins)
	}

	var payload struct {
		CNPJs []string `json:"cnpjs"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return httpx.JSON(400, origin, map[string]string{"error": "invalid JSON body"}, h.allowedOrigins)
	}
	if payload.CNPJs == nil {
		return httpx.JSON(400, origin, map[string]string{"error": "cnpjs array is required"}, h.allowedOrigins)
	}
	data, err := h.service.FindByCNPJs(ctx, payload.CNPJs)
	if err != nil {
		return httpx.JSON(500, origin, map[string]string{"error": "internal server error"}, h.allowedOrigins)
	}
	return httpx.JSON(200, origin, map[string]any{"data": toResponseList(data)}, h.allowedOrigins)
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if path != "/" {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}

func parsePositiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

type pharmacyResponse struct {
	ID           string `json:"id"`
	CNPJ         string `json:"cnpj"`
	Name         string `json:"name"`
	Address      string `json:"address"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
}

func toResponseList(values []domain.Pharmacy) []pharmacyResponse {
	out := make([]pharmacyResponse, 0, len(values))
	for _, item := range values {
		out = append(out, pharmacyResponse{
			ID:           item.ID.String(),
			CNPJ:         item.CNPJ,
			Name:         item.Name,
			Address:      item.Address,
			Neighborhood: item.Neighborhood,
			City:         item.City,
			State:        item.State,
		})
	}
	return out
}
