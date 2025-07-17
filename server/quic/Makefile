# QUIC Server Makefile

.PHONY: all build test bench clean certs example client deps lint fmt vet

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Build targets
EXAMPLE_BINARY=example/quic-server
CLIENT_BINARY=client_example/quic-client

all: deps fmt vet test build

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	$(GOFMT) -s -w .

# Vet code
vet:
	$(GOVET) ./...

# Lint code (requires golangci-lint)
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with race detection
test-race:
	$(GOTEST) -race -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
bench:
	$(GOTEST) -bench=. -benchmem ./...

# Build example server
build: build-example build-client

build-example:
	$(GOBUILD) -o $(EXAMPLE_BINARY) ./example

build-client:
	$(GOBUILD) -o $(CLIENT_BINARY) ./client_example

# Generate certificates
certs:
	@if [ -f "example/server.crt" ]; then \
		echo "Certificates already exist. Remove them first if you want to regenerate."; \
	else \
		echo "Generating certificates..."; \
		chmod +x scripts/generate_certs.sh; \
		./scripts/generate_certs.sh; \
	fi

# Run example server (requires certificates)
example: build-example certs
	@echo "Starting QUIC server..."
	@echo "Make sure to run 'make client' in another terminal to test"
	cd example && ./quic-server

# Run client example
client: build-client
	@echo "Starting QUIC client..."
	cd client_example && ./quic-client

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(EXAMPLE_BINARY)
	rm -f $(CLIENT_BINARY)
	rm -f coverage.out
	rm -f coverage.html

# Clean certificates
clean-certs:
	rm -f example/server.crt
	rm -f example/server.key

# Full clean
clean-all: clean clean-certs

# Development setup
dev-setup: deps certs
	@echo "Development environment setup complete"
	@echo "Run 'make example' to start the server"
	@echo "Run 'make client' to start the client"

# Docker build
docker-build:
	docker build -t quic-server .

# Docker run
docker-run: docker-build
	docker run -p 8443:8443/udp quic-server

# Help
help:
	@echo "Available targets:"
	@echo "  all          - Run deps, fmt, vet, test, and build"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  fmt          - Format code"
	@echo "  vet          - Vet code"
	@echo "  lint         - Lint code (requires golangci-lint)"
	@echo "  test         - Run tests"
	@echo "  test-race    - Run tests with race detection"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  bench        - Run benchmarks"
	@echo "  build        - Build example server and client"
	@echo "  certs        - Generate test certificates"
	@echo "  example      - Run example server"
	@echo "  client       - Run client example"
	@echo "  clean        - Clean build artifacts"
	@echo "  clean-certs  - Clean certificates"
	@echo "  clean-all    - Clean everything"
	@echo "  dev-setup    - Setup development environment"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  help         - Show this help"