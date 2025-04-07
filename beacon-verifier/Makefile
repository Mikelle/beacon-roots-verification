# Project details
PROJECT_NAME := beacon-verifier
MAIN_PACKAGE := ./cmd/beacon-verifier
GO_FILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*")

# Build options
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: all build clean test lint fmt vet install help

all: clean build test ## Build and run tests

build: ## Build the binary
	@echo "Building $(PROJECT_NAME) for $(GOOS)/$(GOARCH)..."
	@go build -o bin/$(PROJECT_NAME) $(MAIN_PACKAGE)
	@echo "Build complete: bin/$(PROJECT_NAME)"

build-all: ## Build for all major platforms
	@echo "Building for multiple platforms..."
	@GOOS=linux GOARCH=amd64 go build -o bin/$(PROJECT_NAME)-linux-amd64 $(MAIN_PACKAGE)
	@GOOS=darwin GOARCH=amd64 go build -o bin/$(PROJECT_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	@GOOS=windows GOARCH=amd64 go build -o bin/$(PROJECT_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	@echo "Multi-platform build complete!"

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean
	@echo "Clean complete!"

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

fmt: ## Run gofmt
	@echo "Formatting code..."
	@gofmt -s -w $(GO_FILES)

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

install: ## Install the binary
	@echo "Installing $(PROJECT_NAME)..."
	@go install $(MAIN_PACKAGE)
	@echo "Install complete!"

run: build ## Build and run the application
	@echo "Running $(PROJECT_NAME) with slot $(SLOT)..."
	@./bin/$(PROJECT_NAME) $(if $(SLOT),-slot $(SLOT))
	
# List all make targets with descriptions
help:
	@echo "$(PROJECT_NAME) make targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Default target
.DEFAULT_GOAL := help