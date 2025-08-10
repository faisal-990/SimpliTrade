.PHONY: build run clean dev

build:
	@echo "Building Process Starting....."
	@go build -o bin/app cmd/main.go
	@echo "Build completed."

run:
	@echo "Run mode starting to compile and run"
	@go run cmd/main.go

clean:
	@echo "Cleaning build artifacts ....."
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

