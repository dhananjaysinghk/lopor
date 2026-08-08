.PHONY: dev-infra stop-infra backend-dev frontend-dev migrate-up migrate-down test build help

# Help command
help:
	@echo "Lopor AI Workspace - Make Commands"
	@echo "  make dev-infra    Start PostgreSQL, Redis, Mailpit, Prometheus in Docker"
	@echo "  make stop-infra   Stop all Docker infrastructure containers"
	@echo "  make backend-dev  Run Go API backend server in live-reload mode"
	@echo "  make frontend-dev Run Next.js frontend app"
	@echo "  make migrate-up   Apply database SQL schema migrations"
	@echo "  make migrate-down Rollback last database SQL migration"
	@echo "  make test         Run backend unit & integration tests"
	@echo "  make build        Build backend Go static binary and Next.js bundle"

# Run Infrastructure
dev-infra:
	docker-compose up -d

stop-infra:
	docker-compose down

# Run Apps
backend-dev:
	cd apps/backend && go run cmd/server/main.go

frontend-dev:
	cd apps/frontend && npm run dev

# Testing & Building
test:
	cd apps/backend && go test -v ./...

build:
	cd apps/backend && go build -o bin/server cmd/server/main.go
	cd apps/frontend && npm run build
