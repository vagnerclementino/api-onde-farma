package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jackc/pgx/v5/pgxpool"
	httpadapter "github.com/vagner/api-onde-farma/internal/pharmacy/adapters/http"
	"github.com/vagner/api-onde-farma/internal/pharmacy/adapters/repository"
	"github.com/vagner/api-onde-farma/internal/pharmacy/application"
	"github.com/vagner/api-onde-farma/internal/platform/config"
	"github.com/vagner/api-onde-farma/internal/platform/persistence/postgres"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	defer pool.Close()

	if err = postgres.Ping(ctx, pool); err != nil {
		log.Fatalf("database ping error: %v", err)
	}

	queries := postgres.NewQueries(pool)
	repo := repository.NewPostgresRepository(queries)
	service := application.NewService(repo)
	handler := httpadapter.NewHandler(service, cfg.AllowedOrigins, cfg.MaxBodyBytes)

	lambda.Start(func(lambdaCtx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return handler.Handle(lambdaCtx, req)
	})
}
