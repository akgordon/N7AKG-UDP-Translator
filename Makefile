# Makefile for UDP Logger Relay

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Output directory for all generated files
OUTPUT_DIR = output

.PHONY: build clean test deps help prepare build-windows build-linux build-macos build-all test-coverage fmt lint config run run-binary run-example-varac package

# Default target
all: build

# Create output directory if it doesn't exist
prepare:
	@mkdir -p $(OUTPUT_DIR)

# Build the application
build: prepare
	go build $(LDFLAGS) -o $(OUTPUT_DIR)/N7AKG-UDP-Translator .

# Build for Windows
build-windows: prepare
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/N7AKG-UDP-Translator-windows.exe .

# Build for Linux
build-linux: prepare
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/N7AKG-UDP-Translator-linux .

# Build for macOS
build-macos: prepare
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/N7AKG-UDP-Translator-macos .

# Build for all platforms
build-all: build-windows build-linux build-macos

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage: prepare
	go test -v -coverprofile=$(OUTPUT_DIR)/coverage.out ./...
	go tool cover -html=$(OUTPUT_DIR)/coverage.out -o $(OUTPUT_DIR)/coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	rm -rf $(OUTPUT_DIR)

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Create example config
config:
	cp config-example.yaml .N7AKG-UDP-Translator.yaml

# Run the application with verbose logging
run:
	go run . --verbose

# Run the built binary (must build first)
run-binary: build
	$(OUTPUT_DIR)/N7AKG-UDP-Translator --verbose

# Build and run examples
run-example-varac:
	go run examples/varac_demo.go

# Package release binaries
package: build-all
	cd $(OUTPUT_DIR) && tar -czf N7AKG-UDP-Translator-$(VERSION)-windows.tar.gz N7AKG-UDP-Translator-windows.exe
	cd $(OUTPUT_DIR) && tar -czf N7AKG-UDP-Translator-$(VERSION)-linux.tar.gz N7AKG-UDP-Translator-linux
	cd $(OUTPUT_DIR) && tar -czf N7AKG-UDP-Translator-$(VERSION)-macos.tar.gz N7AKG-UDP-Translator-macos

# Display help
help:
	@echo "Available targets:"
	@echo "  build            Build the application for current platform (output/)"
	@echo "  build-all        Build for Windows, Linux, and macOS (output/)"
	@echo "  build-windows    Build for Windows (output/)"
	@echo "  build-linux      Build for Linux (output/)"
	@echo "  build-macos      Build for macOS (output/)"
	@echo "  test             Run tests"
	@echo "  test-coverage    Run tests with coverage report (output/)"
	@echo "  deps             Install and tidy dependencies"
	@echo "  clean            Remove all build artifacts (output/)"
	@echo "  fmt              Format source code"
	@echo "  lint             Run linter (requires golangci-lint)"
	@echo "  config           Create example config file"
	@echo "  run              Run application with verbose logging"
	@echo "  run-binary       Run the built binary (must build first)"
	@echo "  run-example-varac Run VarAC format demo"
	@echo "  package          Create release packages (output/)"
	@echo "  help             Show this help message"
	@echo ""
	@echo "All build artifacts are placed in the 'output/' directory"