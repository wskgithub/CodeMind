# ============================================================
# CodeMind Backend Makefile
# ============================================================

APP_NAME    := codemind
VERSION     := $(shell cat ../VERSION)
BUILD_TIME  := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS     := -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)

.PHONY: help build run test lint clean swagger migrate

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the application binary
	go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME) cmd/server/main.go

run: ## Run the application in development mode
	go run cmd/server/main.go

test: ## Run all tests with coverage
	go test ./... -v -cover -coverprofile=coverage.out

test-coverage: test ## Generate and open test coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linter
	golangci-lint run ./...

swagger: ## Generate Swagger documentation
	swag init -g cmd/server/main.go -o docs/swagger

migrate: ## Run database migrations
	go run cmd/server/main.go --migrate

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html

docker-build: ## Build Docker image
	docker build -t $(APP_NAME):$(VERSION) .

.DEFAULT_GOAL := help
