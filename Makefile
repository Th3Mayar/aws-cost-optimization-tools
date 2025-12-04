.PHONY: build clean install test help

# Variables
BINARY_NAME=cost-optimization
BUILD_DIR=./bin
CMD_DIR=./cmd/cost-optimization
INSTALL_PATH=/usr/local/bin

help: ## Show this help
	@echo "AWS Cost Optimization Tools - Makefile"
	@echo ""
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "📦 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "✅ Binary created: $(BUILD_DIR)/$(BINARY_NAME)"

build-linux: ## Build for Linux
	@echo "📦 Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	@echo "✅ Binary created: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

build-mac: ## Build for macOS
	@echo "📦 Building $(BINARY_NAME) for macOS..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	@echo "✅ Binaries created: $(BUILD_DIR)/$(BINARY_NAME)-darwin-*"

build-all: build-linux build-mac ## Build for all platforms

clean: ## Clean compiled files
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@echo "✅ Clean complete"

install: build ## Install the binary in the system
	@echo "📥 Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/
	@sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "✅ $(BINARY_NAME) installed successfully"
	@echo "   You can use it with: $(BINARY_NAME) --help"

uninstall: ## Uninstall the binary from the system
	@echo "🗑️  Uninstalling $(BINARY_NAME)..."
	@sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "✅ $(BINARY_NAME) uninstalled"

deps: ## Download dependencies
	@echo "📚 Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies updated"

run: ## Run in interactive mode
	@go run $(CMD_DIR) start

test-dry-run: ## Run tagging in dry-run mode
	@go run $(CMD_DIR) tagging all

test-show: ## Show resources without modifying
	@go run $(CMD_DIR) tagging show

test-help: ## Show help
	@go run $(CMD_DIR) --help

fmt: ## Format code
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

vet: ## Analyze code
	@echo "🔍 Analyzing code..."
	@go vet ./...
	@echo "✅ Analysis complete"

check: fmt vet ## Format and analyze code

version: ## Show version
	@go run $(CMD_DIR) --version

all: clean deps check build ## Run all tasks (clean, deps, check, build)
