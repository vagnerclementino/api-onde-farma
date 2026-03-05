package httpx

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

func JSON(status int, origin string, payload any, allowedOrigins []string) (events.APIGatewayV2HTTPResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, err
	}

	headers := map[string]string{
		"Content-Type":                 "application/json; charset=utf-8",
		"X-Content-Type-Options":       "nosniff",
		"X-Frame-Options":              "DENY",
		"Referrer-Policy":              "no-referrer",
		"Strict-Transport-Security":    "max-age=63072000; includeSubDomains; preload",
		"Content-Security-Policy":      "default-src 'none'; frame-ancestors 'none'",
		"Cross-Origin-Resource-Policy": "same-site",
	}

	if corsOrigin, ok := resolveCORSOrigin(origin, allowedOrigins); ok {
		headers["Access-Control-Allow-Origin"] = corsOrigin
		headers["Access-Control-Allow-Headers"] = "Content-Type, Authorization"
		headers["Access-Control-Allow-Methods"] = "GET,POST,OPTIONS"
		headers["Vary"] = "Origin"
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: status,
		Headers:    headers,
		Body:       string(body),
	}, nil
}

func ResolvePreflight(origin string, allowedOrigins []string) events.APIGatewayV2HTTPResponse {
	headers := map[string]string{}
	if corsOrigin, ok := resolveCORSOrigin(origin, allowedOrigins); ok {
		headers["Access-Control-Allow-Origin"] = corsOrigin
		headers["Access-Control-Allow-Headers"] = "Content-Type, Authorization"
		headers["Access-Control-Allow-Methods"] = "GET,POST,OPTIONS"
		headers["Access-Control-Max-Age"] = "86400"
		headers["Vary"] = "Origin"
	}
	return events.APIGatewayV2HTTPResponse{StatusCode: 204, Headers: headers}
}

func resolveCORSOrigin(origin string, allowedOrigins []string) (string, bool) {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return "", false
	}
	for _, candidate := range allowedOrigins {
		if origin == candidate {
			return origin, true
		}
	}
	return "", false
}
