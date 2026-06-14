.PHONY: build run run-engine seed clean watch test test-race cover vet lint tidy check tools

# ---- Build & run ----
build:
	@echo "Building..."
	@go build -o bin/server ./cmd/server
	@go build -o bin/engine ./cmd/engine
	@echo "Build completed."

run:
	@go run ./cmd/server

# Run Tower 2 (the strategy engine daemon).
run-engine:
	@go run ./cmd/engine

# Seed the stock universe (run before the engine on a fresh DB).
seed:
	@go run ./cmd/seed

clean:
	@rm -rf ./bin
	@echo "Build artifacts cleaned"

# Dev mode with Air (installs Air only if not already installed)
watch:
	@if ! command -v air >/dev/null 2>&1; then \
		echo "Installing Air..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@echo "Starting Air for hot reload..."
	@air

# ---- Quality gates (the per-milestone "done bar") ----
test:
	@go test ./...

test-race:
	@go test ./... -race

cover:
	@go test ./... -coverprofile=coverage.out
	@go tool cover -func=coverage.out | tail -1

vet:
	@go vet ./...

lint:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not installed — run 'make tools'"; exit 1; \
	fi
	@golangci-lint run ./...

tidy:
	@go mod tidy

# check is the gate that must pass before any milestone is committed.
check: tidy vet test-race
	@echo "✅ check passed (tidy + vet + test -race)"

# ---- Dev tooling ----
tools:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install go.uber.org/mock/mockgen@latest
