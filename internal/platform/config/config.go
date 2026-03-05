package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL    string
	AllowedOrigins []string
	MaxBodyBytes   int
}

func Load() (Config, error) {
	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	origins := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS"))
	allowedOrigins := []string{"https://ondefarma.com.br"}
	if origins != "" {
		allowedOrigins = splitCSV(origins)
	}

	maxBodyBytes := 65536
	if raw := strings.TrimSpace(os.Getenv("MAX_BODY_BYTES")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return Config{}, fmt.Errorf("MAX_BODY_BYTES must be a positive integer")
		}
		maxBodyBytes = parsed
	}

	return Config{
		DatabaseURL:    dbURL,
		AllowedOrigins: allowedOrigins,
		MaxBodyBytes:   maxBodyBytes,
	}, nil
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			out = append(out, item)
		}
	}
	if len(out) == 0 {
		return []string{"https://ondefarma.com.br"}
	}
	return out
}
