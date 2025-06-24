# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Binary info
BINARY_NAME=i18ngen
BINARY_PATH=.
BUILD_DIR=./build

# Build info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: deps
deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) verify

.PHONY: tidy
tidy: ## Clean up dependencies
	$(GOMOD) tidy

.PHONY: fmt
fmt: ## Format code
	$(GOFMT) -s -w .

.PHONY: lint
lint: ## Run linter
	$(GOLINT) run --timeout=5m

.PHONY: test
test: ## Run tests
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

.PHONY: test-short
test-short: ## Run tests without race detector
	$(GOTEST) -v ./...

.PHONY: coverage
coverage: test ## Generate coverage report
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: bench
bench: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

.PHONY: build
build: ## Build binary
	$(GOBUILD) -v $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)

.PHONY: build-all
build-all: ## Build for all platforms
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

.PHONY: install
install: ## Install binary to GOPATH/bin
	$(GOCMD) install $(LDFLAGS) .

.PHONY: clean
clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

.PHONY: run
run: ## Run the application with example config
	$(GOCMD) run . --help

.PHONY: dev-setup
dev-setup: deps ## Setup development environment
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin; \
	}

.PHONY: check
check: fmt lint test ## Run all checks (format, lint, test)

.PHONY: ci
ci: deps check ## Run CI pipeline locally

.PHONY: pre-commit
pre-commit: fmt lint test-short ## Quick checks before commit

.PHONY: release-dry-run
release-dry-run: ## Test release build
	@echo "Testing release build..."
	$(MAKE) clean
	$(MAKE) build-all
	@echo "Release build test completed successfully"

.PHONY: docker-build
docker-build: ## Build Docker image (if Dockerfile exists)
	@if [ -f Dockerfile ]; then \
		docker build -t $(BINARY_NAME):$(VERSION) .; \
	else \
		echo "Dockerfile not found"; \
	fi

.PHONY: security
security: ## Run security scan
	@command -v gosec >/dev/null 2>&1 || { \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	}
	gosec ./...

# Development helpers
.PHONY: watch
watch: ## Watch for changes and run tests
	@command -v fswatch >/dev/null 2>&1 || { \
		echo "fswatch not found. Install with: brew install fswatch (macOS) or apt-get install inotify-tools (Linux)"; \
		exit 1; \
	}
	@echo "Watching for changes..."
	@fswatch -o . -e ".*" -i "\\.go$$" | xargs -n1 -I{} sh -c 'clear && echo "Running tests..." && make test-short'

.PHONY: update-deps
update-deps: ## Update all dependencies
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Default target
.DEFAULT_GOAL := help 