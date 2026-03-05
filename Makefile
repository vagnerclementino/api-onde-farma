SHELL := /bin/sh

-include .env
export

.DEFAULT_GOAL := help

.PHONY: help fmt test tidy build-lambda migrate seed db-reset local-up local-down sqlc-generate localstack-deploy run stop

help: ## Mostra todos os comandos disponiveis
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z0-9_.-]+:.*## / {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

fmt: ## Formata o codigo Go
	gofmt -w $$(find . -type f -name '*.go' -not -path './vendor/*')

test: ## Roda os testes
	go test ./...

tidy: ## Organiza dependencias do Go module
	go mod tidy

build-lambda: ## Gera o bootstrap da Lambda em ./build
	mkdir -p build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o build/bootstrap ./cmd/lambda

migrate: ## Aplica as migracoes no banco
	chmod +x scripts/migrate.sh
	./scripts/migrate.sh

seed: ## Popula o banco com dados iniciais
	chmod +x scripts/seed_csv.sh
	./scripts/seed_csv.sh

db-reset: ## Reseta schema e reaplica migrate + seed
	psql "$(DATABASE_URL)" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	$(MAKE) migrate
	$(MAKE) seed

sqlc-generate: ## Gera codigo a partir do sqlc.yaml
	docker run --rm -v "$(PWD):/src" -w /src sqlc/sqlc:1.30.0 generate

local-up: ## Sobe servicos locais com Docker Compose
	docker compose up -d

local-down: ## Derruba servicos locais e volumes
	docker compose down -v

localstack-deploy: build-lambda ## Faz deploy da Lambda no LocalStack
	chmod +x scripts/deploy_localstack.sh
	./scripts/deploy_localstack.sh

run: local-up migrate seed localstack-deploy ## Inicializa ambiente local completo
	@echo "Project started locally with Postgres + LocalStack + Lambda deployment."

stop: ## Para ambiente local e remove volumes
	docker compose down -v
