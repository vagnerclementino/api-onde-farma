SHELL := /bin/sh

-include .env
export

.DEFAULT_GOAL := help

.PHONY: help fmt test tidy build-lambda migrate seed db-reset local-up local-down sqlc-generate localstack-deploy run stop

# Colors for help output.
GREEN  := $(shell tput -Txterm setaf 2)
WHITE  := $(shell tput -Txterm setaf 7)
YELLOW := $(shell tput -Txterm setaf 3)
RESET  := $(shell tput -Txterm sgr0)

HELP_DOCS = \
	%help; \
	while(<>) { push @{$$help{$$2 // 'other'}}, [$$1, $$3] if /^([a-zA-Z0-9_.-]+)\s*:.*\#\#(?:@([a-zA-Z\-]+))?\s(.*)$$/ }; \
	print "usage: make [target]\n\n"; \
	for (sort keys %help) { \
	print "${WHITE}$$_:${RESET}\n"; \
	for (@{$$help{$$_}}) { \
	$$sep = " " x (32 - length $$_->[0]); \
	print "  ${YELLOW}$$_->[0]${RESET}$$sep${GREEN}$$_->[1]${RESET}\n"; \
	}; \
	print "\n"; }

help: ##@other Show this help.
	@perl -e '$(HELP_DOCS)' $(MAKEFILE_LIST)

fmt: ##@quality Formata o codigo Go
	gofmt -w $$(find . -type f -name '*.go' -not -path './vendor/*')

test: ##@quality Roda os testes
	go test ./...

tidy: ##@helper Organiza dependencias do Go module
	go mod tidy

build-lambda: ##@application Gera o bootstrap da Lambda em ./build
	mkdir -p build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o build/bootstrap ./cmd/lambda

migrate: ##@database Aplica as migracoes no banco
	chmod +x scripts/migrate.sh
	./scripts/migrate.sh

seed: ##@database Popula o banco com dados iniciais
	chmod +x scripts/seed_csv.sh
	./scripts/seed_csv.sh

db-reset: ##@database Reseta schema e reaplica migrate + seed
	sh ./scripts/psql.sh "$(DATABASE_URL)" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	$(MAKE) migrate
	$(MAKE) seed

sqlc-generate: ##@helper Gera codigo a partir do sqlc.yaml
	docker run --rm -v "$(PWD):/src" -w /src sqlc/sqlc:1.30.0 generate

local-up: ##@local Sobe servicos locais com Docker Compose
	docker compose up -d

local-down: ##@local Derruba servicos locais e volumes
	docker compose down -v

localstack-deploy: build-lambda ##@local Faz deploy da Lambda no LocalStack
	chmod +x scripts/deploy_localstack.sh
	./scripts/deploy_localstack.sh

run: local-up migrate seed localstack-deploy ##@application Inicializa ambiente local completo
	@echo "Project started locally with Postgres + LocalStack + Lambda deployment."

stop: ##@local Para ambiente local e remove volumes
	docker compose down -v
