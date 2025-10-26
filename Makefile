# Project variables
BINARY_NAME=k8s-tui
MAIN_PATH=cmd/k8s-tui/main.go
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Go variables
GO=go
GOTEST=$(GO) test
GOVET=$(GO) vet
GOFMT=gofmt
GOLINT=golangci-lint

# OS detection for cross-compilation
UNAME_S := $(shell uname -s)
GOARCH := $(shell go env GOARCH)

.PHONY: all build clean test lint fmt run install help

## Default target
all: clean lint test build

## Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary built at $(BUILD_DIR)/$(BINARY_NAME)"

## Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Multi-platform build complete"

## Run the application
run:
	@$(GO) run $(LDFLAGS) $(MAIN_PATH)

## Run with race detector
run-race:
	@$(GO) run -race $(LDFLAGS) $(MAIN_PATH)

## Install the binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@$(GO) install $(LDFLAGS) $(MAIN_PATH)
	@echo "Installed to $$(go env GOPATH)/bin/$(BINARY_NAME)"

## Run tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v -cover -race ./...

## Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -v -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

## Run benchmarks
bench:
	@echo "Running benchmarks..."
	@$(GOTEST) -bench=. -benchmem ./...

## Format code
fmt:
	@echo "Formatting code..."
	@$(GOFMT) -w -s .
	@$(GO) mod tidy

## Lint code
lint:
	@echo "Linting code..."
	@if ! command -v $(GOLINT) &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@$(GOLINT) run ./...

## Run go vet
vet:
	@echo "Running go vet..."
	@$(GOVET) ./...

## Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy

## Update dependencies
update-deps:
	@echo "Updating dependencies..."
	@$(GO) get -u ./...
	@$(GO) mod tidy

## Initialize project (first time setup)
init: deps
	@echo "Initializing project..."
	@mkdir -p cmd/k8s-tui
	@mkdir -p internal/app
	@mkdir -p internal/k8s
	@mkdir -p internal/ui/components
	@mkdir -p internal/ui/styles
	@mkdir -p internal/ui/keys
	@mkdir -p internal/models
	@mkdir -p internal/config
	@mkdir -p pkg/utils
	@echo "Project structure created"

## Run development environment with auto-reload
dev:
	@if ! command -v air &> /dev/null; then \
		echo "Installing air for auto-reload..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	@air

## Display help
help:
	@echo "Available targets:"
	@echo "  make build        - Build the binary"
	@echo "  make build-all    - Build for multiple platforms"
	@echo "  make run          - Run the application"
	@echo "  make run-race     - Run with race detector"
	@echo "  make install      - Install the binary to GOPATH/bin"
	@echo "  make test         - Run tests"
	@echo "  make test-coverage- Run tests with coverage report"
	@echo "  make bench        - Run benchmarks"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Lint code"
	@echo "  make vet          - Run go vet"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make deps         - Download dependencies"
	@echo "  make update-deps  - Update dependencies"
	@echo "  make init         - Initialize project structure"
	@echo "  make dev          - Run development environment"
	@echo "  make help         - Display this help"