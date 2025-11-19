.PHONY: run test bench lint fmt clean build help

# Variables
BINARY_NAME=rate-limiter
BUILD_DIR=./bin
CMD_DIR=./cmd/server/main.go

# Default target
all: fmt lint test build

# Run the application
run:
	go run $(CMD_DIR)

# Run tests
test:
	go test -v ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Lint code
lint:
	golangci-lint run --timeout=5m

# Format code
fmt:
	go fmt ./...
	go mod tidy

# Build binary
build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	go clean

# Show help
help:
	@echo "Available targets:"
	@echo "  run    - Run the application"
	@echo "  test   - Run tests"
	@echo "  bench  - Run benchmarks"
	@echo "  lint   - Run linter"
	@echo "  fmt    - Format code"
	@echo "  build  - Build binary"
	@echo "  clean  - Clean build artifacts"
	@echo "  all    - Format, lint, test, and build"
