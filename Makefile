SHELL := /bin/sh

-include .env
export

.PHONY: fmt test tidy build-lambda migrate seed db-reset local-up local-down sqlc-generate localstack-deploy run stop

fmt:
	gofmt -w $$(find . -type f -name '*.go' -not -path './vendor/*')

test:
	go test ./...

tidy:
	go mod tidy

build-lambda:
	mkdir -p build
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o build/bootstrap ./cmd/lambda

migrate:
	chmod +x scripts/migrate.sh
	./scripts/migrate.sh

seed:
	chmod +x scripts/seed_csv.sh
	./scripts/seed_csv.sh

db-reset:
	psql "$(DATABASE_URL)" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	$(MAKE) migrate
	$(MAKE) seed

sqlc-generate:
	docker run --rm -v "$(PWD):/src" -w /src sqlc/sqlc:1.30.0 generate

local-up:
	docker compose up -d

local-down:
	docker compose down -v

localstack-deploy: build-lambda
	chmod +x scripts/deploy_localstack.sh
	./scripts/deploy_localstack.sh

run: local-up migrate seed localstack-deploy
	@echo "Project started locally with Postgres + LocalStack + Lambda deployment."

stop:
	docker compose down -v
