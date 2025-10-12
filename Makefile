# Makefile for UDP Logger Relay

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: build clean test deps help

# Default target
all: build

# Build the application
build:
	go build $(LDFLAGS) -o udp-logger-relay .

# Build for Windows
build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o udp-logger-relay-windows.exe .

# Build for Linux
build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o udp-logger-relay-linux .

# Build for macOS
build-macos:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o udp-logger-relay-macos .

# Build for all platforms
build-all: build-windows build-linux build-macos

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	rm -f udp-logger-relay udp-logger-relay-* coverage.out coverage.html

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Create example config
config:
	cp config-example.yaml .udp-logger-relay.yaml

# Run the application with verbose logging
run:
	go run . --verbose

# Display help
help:
	@echo "Available targets:"
	@echo "  build         Build the application for current platform"
	@echo "  build-all     Build for Windows, Linux, and macOS"
	@echo "  test          Run tests"
	@echo "  test-coverage Run tests with coverage report"
	@echo "  deps          Install and tidy dependencies"
	@echo "  clean         Remove build artifacts"
	@echo "  fmt           Format source code"
	@echo "  lint          Run linter (requires golangci-lint)"
	@echo "  config        Create example config file"
	@echo "  run           Run application with verbose logging"
	@echo "  help          Show this help message"