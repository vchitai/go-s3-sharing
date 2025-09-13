# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=go-s3-sharing
BINARY_SERVER=bin/server
BINARY_CLI=bin/cli

# Docker parameters
DOCKER_IMAGE=vchitai/go-s3-sharing
DOCKER_TAG=latest

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

.PHONY: all build clean test deps lint security docker docker-push help

all: clean deps test build

# Build the application
build: build-server build-cli

build-server:
	@echo "Building server..."
	@mkdir -p bin
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_SERVER) ./cmd/server

build-cli:
	@echo "Building CLI..."
	@mkdir -p bin
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_CLI) ./cmd/cli

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf bin/
	@rm -f coverage.out

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./...

test-coverage:
	@echo "Generating coverage report..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

deps-update:
	@echo "Updating dependencies..."
	$(GOMOD) tidy
	$(GOMOD) verify

# Linting
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# Security scanning
security:
	@echo "Running security scan..."
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	gosec ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	$(GOCMD) vet ./...

# Docker operations
docker:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-push:
	@echo "Pushing Docker image..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

# Development
dev:
	@echo "Starting development server..."
	$(GOCMD) run ./cmd/server

dev-cli:
	@echo "Starting CLI..."
	$(GOCMD) run ./cmd/cli

# Kubernetes
k8s-deploy:
	@echo "Deploying to Kubernetes..."
	kubectl apply -f deployments/kubernetes/

k8s-delete:
	@echo "Deleting from Kubernetes..."
	kubectl delete -f deployments/kubernetes/

# Docker Compose
compose-up:
	@echo "Starting with Docker Compose..."
	docker-compose up -d

compose-down:
	@echo "Stopping Docker Compose..."
	docker-compose down

compose-logs:
	@echo "Showing Docker Compose logs..."
	docker-compose logs -f

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  build-server   - Build the server binary"
	@echo "  build-cli      - Build the CLI binary"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-coverage  - Generate coverage report"
	@echo "  deps           - Download dependencies"
	@echo "  deps-update    - Update dependencies"
	@echo "  lint           - Run linter"
	@echo "  security       - Run security scan"
	@echo "  fmt            - Format code"
	@echo "  vet            - Vet code"
	@echo "  docker         - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-push    - Push Docker image"
	@echo "  dev            - Start development server"
	@echo "  dev-cli        - Start CLI"
	@echo "  k8s-deploy     - Deploy to Kubernetes"
	@echo "  k8s-delete     - Delete from Kubernetes"
	@echo "  compose-up     - Start with Docker Compose"
	@echo "  compose-down   - Stop Docker Compose"
	@echo "  compose-logs   - Show Docker Compose logs"
	@echo "  help           - Show this help message"
