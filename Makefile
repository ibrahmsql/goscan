# Makefile for goscan project

.PHONY: all build clean install test help

# Default target
all: build

# Build all binaries
build: goscan apiscan

# Build goscan (directory scanner)
goscan:
	@echo "Building goscan..."
	go build -o goscan ./cmd/goscan
	@echo "✅ goscan built successfully"

# Build apiscan (API endpoint scanner)
apiscan:
	@echo "Building apiscan..."
	go build -o apiscan ./cmd/apiscan
	@echo "✅ apiscan built successfully"

# Build with optimizations for release
release:
	@echo "Building optimized release binaries..."
	go build -ldflags="-s -w" -o goscan ./cmd/goscan
	go build -ldflags="-s -w" -o apiscan ./cmd/apiscan
	@echo "✅ Release binaries built successfully"

# Install binaries to GOPATH/bin
install: build
	@echo "Installing binaries..."
	cp goscan $(GOPATH)/bin/
	cp apiscan $(GOPATH)/bin/
	@echo "✅ Binaries installed to $(GOPATH)/bin/"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f goscan apiscan
	@echo "✅ Clean completed"

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Show help
help:
	@echo "Available targets:"
	@echo "  all      - Build all binaries (default)"
	@echo "  build    - Build all binaries"
	@echo "  goscan   - Build goscan binary"
	@echo "  apiscan  - Build apiscan binary"
	@echo "  release  - Build optimized release binaries"
	@echo "  install  - Install binaries to GOPATH/bin"
	@echo "  clean    - Clean build artifacts"
	@echo "  test     - Run tests"
	@echo "  fmt      - Format code"
	@echo "  lint     - Run linter"
	@echo "  tidy     - Tidy dependencies"
	@echo "  help     - Show this help"

# Example usage targets
examples:
	@echo "Example usage:"
	@echo ""
	@echo "Directory scanning:"
	@echo "  ./goscan wordlists/common-web-paths.txt https://example.com"
	@echo ""
	@echo "API endpoint scanning:"
	@echo "  ./apiscan wordlists/api-endpoints.txt https://api.example.com"
	@echo "  ./apiscan wordlists/api-endpoints.txt https://api.example.com --output results.json"
	@echo "  ./apiscan wordlists/api-endpoints.txt https://api.example.com --threads 20 --timeout 15"