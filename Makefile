# Binary name
BINARY_NAME=github-copilot-invite

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build directory
BUILD_DIR=bin

# Security files
ENCRYPTION_KEY=.encryption_key
CONFIG_FILE=config.yaml
CONFIG_TEMPLATE=config.yaml.template

# Aliases
b: build  ## Alias for build
c: clean  ## Alias for clean

.PHONY: all build clean test coverage deps lint help b c cert key init-config

all: clean deps build test ## Build the project and run tests

build: ## Build the application
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v

clean: ## Clean build files
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

test: ## Run tests
	@echo "Running tests..."
	@$(GOTEST) -v ./...

coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@$(GOTEST) -cover ./...

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GOMOD) download

lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint is not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

# Security targets
key: ## Generate new encryption key
	@echo "Generating encryption key..."
	@if [ -f "$(ENCRYPTION_KEY)" ]; then \
		echo "Warning: Encryption key already exists. Remove it first to generate a new one."; \
		exit 1; \
	else \
		$(GOCMD) run scripts/generate_encryption_key.go; \
	fi

key-force: ## Force generate new encryption key (overwrites existing)
	@echo "Force generating new encryption key..."
	@$(GOCMD) run scripts/generate_encryption_key.go

cert: ## Generate self-signed SSL certificates
	@echo "Generating self-signed SSL certificates..."
	@cd scripts && ./generate_cert.sh

cert-force: ## Force generate new SSL certificates (overwrites existing)
	@echo "Force generating new SSL certificates..."
	@rm -rf certs/*
	@cd scripts && ./generate_cert.sh

init-config: ## Initialize configuration file from template
	@echo "Initializing configuration..."
	@if [ ! -f "$(CONFIG_FILE)" ]; then \
		cp $(CONFIG_TEMPLATE) $(CONFIG_FILE); \
		echo "Created $(CONFIG_FILE) from template"; \
	else \
		echo "$(CONFIG_FILE) already exists"; \
	fi

secure-check: ## Check security configuration
	@echo "Checking security configuration..."
	@echo "Checking encryption key..."
	@if [ ! -f "$(ENCRYPTION_KEY)" ]; then \
		echo "Warning: Encryption key not found. Run 'make key' to generate one."; \
	else \
		if [ "$$(stat -c %a $(ENCRYPTION_KEY))" != "600" ]; then \
			echo "Warning: Encryption key has incorrect permissions. Setting to 600..."; \
			chmod 600 $(ENCRYPTION_KEY); \
		else \
			echo "Encryption key permissions OK (600)"; \
		fi \
	fi
	@echo "Checking SSL certificates..."
	@if [ ! -d "certs" ]; then \
		echo "Warning: SSL certificates not found. Run 'make cert' to generate them."; \
	else \
		echo "SSL certificates directory exists"; \
	fi
	@echo "Checking configuration..."
	@if [ ! -f "$(CONFIG_FILE)" ]; then \
		echo "Warning: Configuration file not found. Run 'make init-config' to create one."; \
	else \
		echo "Configuration file exists"; \
	fi

# Configuration targets
config-view: ## View all configuration values (including non-sensitive)
	@echo "Viewing configuration..."
	@$(GOCMD) run scripts/view_config.go -all

config-view-sensitive: ## View only sensitive configuration values
	@echo "Viewing sensitive configuration values..."
	@$(GOCMD) run scripts/view_config.go -sensitive

config-view-decrypted: ## View decrypted configuration values (use with caution)
	@echo "CAUTION: Viewing decrypted configuration values..."
	@echo "This will display sensitive information. Make sure you are in a secure environment."
	@read -p "Are you sure you want to continue? [y/N] " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		$(GOCMD) run scripts/view_config.go -all -decrypt; \
	else \
		echo "Operation cancelled."; \
	fi

run: secure-check ## Run the application with security checks
	@echo "Starting application..."
	@$(BUILD_DIR)/$(BINARY_NAME)

help: ## Display this help message
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_-]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Default target
.DEFAULT_GOAL := help
