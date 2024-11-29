# Go parameters
BINARY_NAME=github-copilot-invite
MAIN_PACKAGE=.
GO=go
GOBUILD=$(GO) build
GOTEST=$(GO) test
GOCLEAN=$(GO) clean
GOGET=$(GO) get
GOMOD=$(GO) mod
GOLINT=golangci-lint

# Build flags
LDFLAGS=-ldflags "-s -w"

# Output directory for binaries
BIN_DIR=bin

# Get the current git commit hash
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Aliases
b: build  ## Alias for build
c: clean  ## Alias for clean

.PHONY: all build clean test coverage deps lint help b c cert

all: clean deps build test ## Build the project and run tests

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

clean: ## Clean build files
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	$(GOCLEAN)

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -cover -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

deps: ## Download and verify dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) verify
	$(GOMOD) tidy

lint: ## Run linter
	@echo "Running linter..."
	$(GOLINT) run

run: build ## Run the application
	@echo "Running $(BINARY_NAME)..."
	./$(BIN_DIR)/$(BINARY_NAME)

dev: ## Run the application in development mode with hot reload
	@echo "Running in development mode..."
	air

config: ## Create config file from template if it doesn't exist
	@if [ ! -f config.yaml ]; then \
		echo "Creating config.yaml from template..."; \
		cp config.yaml.template config.yaml; \
	else \
		echo "config.yaml already exists"; \
	fi

cert: ## Generate self-signed SSL certificates
	@echo "Generating self-signed SSL certificates..."
	@cd scripts && ./generate_cert.sh

help: ## Display this help message
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_-]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Default target
.DEFAULT_GOAL := help
