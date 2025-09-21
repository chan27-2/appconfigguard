.PHONY: build test clean help setup-auth demo preview apply

# Binary name and paths
BINARY_NAME=appconfigguard
BINARY_PATH=bin/$(BINARY_NAME)
EXAMPLE_CONFIG=config.example.json

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Default Azure App Configuration values (customize these)
APP_CONFIG_ENDPOINT ?= https://your-store.azconfig.io
APP_CONFIG_NAME ?= your-store-name
RESOURCE_GROUP ?= your-resource-group

# =============================================================================
# DEVELOPMENT TARGETS
# =============================================================================

# Build the project
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_PATH) -v ./cmd
	@echo "✅ Binary built: $(BINARY_PATH)"

# Run tests
test:
	@echo "🧪 Running tests..."
	$(GOTEST) -v ./...

# Clean build files
clean:
	@echo "🧹 Cleaning build files..."
	$(GOCLEAN)
	rm -f $(BINARY_PATH) bin/*-*
	@echo "✅ Clean complete"

# Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ Dependencies ready"

# =============================================================================
# AUTHENTICATION SETUP
# =============================================================================

# Setup authentication using Azure CLI (recommended for development)
setup-auth-cli:
	@echo "🔐 Setting up Azure CLI authentication..."
	@echo "Run the following commands:"
	@echo "  az login"
	@echo "  az account set --subscription <your-subscription-id>"
	@echo ""
	@echo "✅ Azure CLI authentication is now ready!"
	@echo "   The demo will automatically use this authentication."

# Setup authentication using access keys (recommended for production)
setup-auth-keys:
	@echo "🔑 Setting up Access Key authentication..."
	@echo "1. Get your access keys:"
	@echo "   az appconfig credential list --name $(APP_CONFIG_NAME) --resource-group $(RESOURCE_GROUP)"
	@echo ""
	@echo "2. Set the connection string:"
	@echo "   export APP_CONFIG_CONNECTION_STRING='Endpoint=$(APP_CONFIG_ENDPOINT);Id=<your-id>;Secret=<your-secret>'"
	@echo ""
	@echo "3. Test with: make demo-keys"

# Check authentication status
auth-check:
	@echo "🔍 Checking authentication methods..."
	@if [ -n "$$APP_CONFIG_CONNECTION_STRING" ]; then \
		echo "✅ Access Key authentication: Configured"; \
		echo "   APP_CONFIG_CONNECTION_STRING is set"; \
	else \
		echo "⚠️  Access Key authentication: Not configured"; \
		if az account show --query "name" 2>/dev/null >/dev/null; then \
			echo "✅ Azure CLI authentication: Available"; \
			az account show --query "name" --output tsv | xargs echo "   Logged in as:"; \
		else \
			echo "❌ Azure CLI authentication: Not available (run 'az login')"; \
		fi; \
	fi
	@echo ""
	@echo "💡 The app will automatically try these methods in order:"

# =============================================================================
# DEMO TARGETS
# =============================================================================

# Quick demo with automatic authentication
demo: build
	@echo ""
	@echo "🚀 Running demo with $(EXAMPLE_CONFIG)..."
	@echo "📍 Using endpoint: $(APP_CONFIG_ENDPOINT)"
	@echo "📄 Using config: $(EXAMPLE_CONFIG)"
	@echo ""
	@./$(BINARY_PATH) --file=$(EXAMPLE_CONFIG) --endpoint=$(APP_CONFIG_ENDPOINT) || \
	(echo ""; echo "❌ Demo failed. Make sure you're authenticated:"; \
	 echo "   • Azure CLI: Run 'az login'"; \
	 echo "   • Access Key: Set APP_CONFIG_CONNECTION_STRING"; \
	 echo "   • See 'make help' for more options"; exit 1)

# Demo with access key auth
demo-keys: build
	@echo "🚀 Running demo with access key authentication..."
	@if [ -z "$$APP_CONFIG_CONNECTION_STRING" ]; then \
		echo "❌ APP_CONFIG_CONNECTION_STRING not set. Run 'make setup-auth-keys' first."; \
		exit 1; \
	fi
	./$(BINARY_PATH) --file=$(EXAMPLE_CONFIG) --endpoint=$(APP_CONFIG_ENDPOINT)

# Preview changes only
preview: build
	@echo "👀 Previewing changes..."
	./$(BINARY_PATH) --file=$(EXAMPLE_CONFIG) --endpoint=$(APP_CONFIG_ENDPOINT)

# Apply changes (with confirmation)
apply: build
	@echo "⚠️  This will apply changes to Azure App Configuration!"
	@echo "   Config: $(EXAMPLE_CONFIG)"
	@echo "   Endpoint: $(APP_CONFIG_ENDPOINT)"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ] || (echo "❌ Cancelled"; exit 1)
	./$(BINARY_PATH) --file=$(EXAMPLE_CONFIG) --endpoint=$(APP_CONFIG_ENDPOINT) --apply

# =============================================================================
# BUILD TARGETS
# =============================================================================

# Cross-platform builds
build-linux:
	@echo "🐧 Building for Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-linux-amd64 -v ./cmd

build-windows:
	@echo "🪟 Building for Windows..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-windows-amd64.exe -v ./cmd

build-darwin:
	@echo "🍎 Building for macOS..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-darwin-amd64 -v ./cmd

# Build for all platforms
build-all: build-linux build-windows build-darwin
	@echo "✅ All binaries built in ./bin/"

# =============================================================================
# DEVELOPMENT TOOLS
# =============================================================================

# Install development dependencies
dev-deps: deps
	@echo "🔧 Installing development tools..."
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint

# Lint the code
lint:
	@echo "🔍 Linting code..."
	golangci-lint run

# Format code
fmt:
	@echo "💅 Formatting code..."
	$(GOCMD) fmt ./...

# =============================================================================
# RELEASE TARGETS
# =============================================================================

# Install GoReleaser
install-goreleaser:
	@echo "📦 Installing GoReleaser..."
	$(GOCMD) install github.com/goreleaser/goreleaser@latest

# Create a snapshot release (for testing)
release-snapshot: install-goreleaser
	@echo "📸 Creating release snapshot..."
	goreleaser release --snapshot --clean

# Create a full release (requires GITHUB_TOKEN)
release: install-goreleaser
	@echo "🚀 Creating full release..."
	@if [ -z "$$GITHUB_TOKEN" ]; then \
		echo "❌ GITHUB_TOKEN not set. Please set it first:"; \
		echo "   export GITHUB_TOKEN=your_github_token_here"; \
		exit 1; \
	fi
	goreleaser release --clean

# Check if release would work
release-check: install-goreleaser
	@echo "🔍 Checking release configuration..."
	goreleaser check

# Create a new release using the release script
release-tag:
	@echo "🏷️  Creating a new release tag..."
	@if [ -z "$$VERSION" ]; then \
		echo "❌ VERSION not set. Usage: make release-tag VERSION=v1.2.3"; \
		exit 1; \
	fi
	./scripts/release.sh $(VERSION)

# =============================================================================
# HELP
# =============================================================================

help:
	@echo "🤖 $(BINARY_NAME) - Azure App Configuration Management Tool"
	@echo ""
	@echo "QUICK START:"
	@echo "  1. make build                    # Build the tool"
	@echo "  2. az login                      # Login to Azure (optional)"
	@echo "  3. make demo                     # Run a demo"
	@echo ""
	@echo "ESSENTIAL COMMANDS:"
	@echo "  build          - Build the binary"
	@echo "  demo           - Run demo with example config"
	@echo "  preview        - Preview changes without applying"
	@echo "  apply          - Apply changes (with confirmation)"
	@echo "  auth-check     - Check authentication status"
	@echo ""
	@echo "AUTHENTICATION:"
	@echo "  setup-auth-cli   - Setup Azure CLI authentication"
	@echo "  setup-auth-keys  - Setup access key authentication"
	@echo ""
	@echo "BUILD & DEVELOPMENT:"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build files"
	@echo "  build-all      - Build for all platforms"
	@echo "  lint           - Lint the code"
	@echo "  fmt            - Format code"
	@echo "  deps           - Download dependencies"
	@echo ""
	@echo "RELEASE & DISTRIBUTION:"
	@echo "  release-check    - Check release configuration"
	@echo "  release-snapshot - Create test release locally"
	@echo "  release-tag      - Create and push a new release tag"
	@echo "  release          - Create full GitHub release (manual)"
	@echo ""
	@echo "CONFIGURATION:"
	@echo "  Set these variables to customize:"
	@echo "    APP_CONFIG_ENDPOINT=$(APP_CONFIG_ENDPOINT)"
	@echo "    APP_CONFIG_NAME=$(APP_CONFIG_NAME)"
	@echo "    RESOURCE_GROUP=$(RESOURCE_GROUP)"
	@echo ""
	@echo "EXAMPLE USAGE:"
	@echo "  # Preview changes"
	@echo "  $(BINARY_NAME) --file=config.json --endpoint=https://mystorage.azconfig.io"
	@echo ""
	@echo "  # Apply changes"
	@echo "  $(BINARY_NAME) --file=config.json --endpoint=https://mystorage.azconfig.io --apply"
