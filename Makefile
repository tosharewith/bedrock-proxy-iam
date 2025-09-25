# Makefile for Bedrock IAM Proxy

# Variables
APP_NAME := bedrock-proxy
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S_UTC')
GO_VERSION := 1.21

# Docker variables
REGISTRY := ghcr.io
IMAGE_NAME := $(REGISTRY)/bedrock-proxy/$(APP_NAME)
DOCKERFILE := build/Dockerfile

# Go build variables
LDFLAGS := -w -s -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)
BUILD_FLAGS := -ldflags "$(LDFLAGS)" -trimpath

.PHONY: help
help: ## Show this help message
	@echo "Bedrock IAM Proxy - Make Commands"
	@echo "================================="
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ { printf "  %-20s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: clean
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -f $(APP_NAME)
	@rm -rf dist/
	@rm -rf coverage.out
	@go clean -cache -testcache -modcache

.PHONY: deps
deps: ## Download and verify dependencies
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod verify
	@go mod tidy

.PHONY: fmt
fmt: ## Format Go code
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@goimports -w -local github.com/bedrock-proxy/bedrock-iam-proxy .

.PHONY: lint
lint: ## Run linters
	@echo "🔍 Running linters..."
	@golangci-lint run ./...

.PHONY: test
test: ## Run tests
	@echo "🧪 Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "🧪 Running integration tests..."
	@go test -v -race -tags=integration ./...

.PHONY: security
security: ## Run security scans
	@echo "🔒 Running security scans..."
	@gosec ./...
	@nancy sleuth

.PHONY: build
build: deps ## Build the application
	@echo "🔨 Building $(APP_NAME) $(VERSION)..."
	@CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(APP_NAME) ./cmd/$(APP_NAME)

.PHONY: build-linux
build-linux: deps ## Build for Linux
	@echo "🔨 Building $(APP_NAME) for Linux..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(APP_NAME)-linux-amd64 ./cmd/$(APP_NAME)

.PHONY: build-all
build-all: deps ## Build for multiple platforms
	@echo "🔨 Building $(APP_NAME) for multiple platforms..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o dist/$(APP_NAME)-linux-amd64 ./cmd/$(APP_NAME)
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o dist/$(APP_NAME)-linux-arm64 ./cmd/$(APP_NAME)
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o dist/$(APP_NAME)-darwin-amd64 ./cmd/$(APP_NAME)
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o dist/$(APP_NAME)-darwin-arm64 ./cmd/$(APP_NAME)

.PHONY: run
run: build ## Build and run the application
	@echo "🚀 Running $(APP_NAME)..."
	@./$(APP_NAME)

.PHONY: dev
dev: ## Run in development mode
	@echo "🚀 Running $(APP_NAME) in development mode..."
	@GIN_MODE=debug LOG_LEVEL=debug go run ./cmd/$(APP_NAME)

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build \
		--build-arg BUILD_VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t $(IMAGE_NAME):$(VERSION) \
		-t $(IMAGE_NAME):latest \
		-f $(DOCKERFILE) .

.PHONY: docker-run
docker-run: docker-build ## Build and run Docker container
	@echo "🐳 Running Docker container..."
	@docker run -p 8080:8080 --rm $(IMAGE_NAME):latest

.PHONY: docker-scan
docker-scan: docker-build ## Scan Docker image for vulnerabilities
	@echo "🔍 Scanning Docker image..."
	@trivy image --severity HIGH,CRITICAL $(IMAGE_NAME):latest

.PHONY: docker-push
docker-push: docker-build ## Push Docker image to registry
	@echo "🚀 Pushing Docker image..."
	@docker push $(IMAGE_NAME):$(VERSION)
	@docker push $(IMAGE_NAME):latest

.PHONY: k8s-deploy
k8s-deploy: ## Deploy to Kubernetes
	@echo "☸️  Deploying to Kubernetes..."
	@kubectl apply -f deployments/kubernetes/

.PHONY: k8s-delete
k8s-delete: ## Delete from Kubernetes
	@echo "☸️  Deleting from Kubernetes..."
	@kubectl delete -f deployments/kubernetes/

.PHONY: k8s-logs
k8s-logs: ## Show Kubernetes logs
	@kubectl logs -f deployment/bedrock-proxy -n bedrock-system

.PHONY: terraform-plan
terraform-plan: ## Run Terraform plan
	@echo "🏗️  Running Terraform plan..."
	@cd deployments/terraform && terraform plan

.PHONY: terraform-apply
terraform-apply: ## Apply Terraform configuration
	@echo "🏗️  Applying Terraform configuration..."
	@cd deployments/terraform && terraform apply

.PHONY: terraform-destroy
terraform-destroy: ## Destroy Terraform resources
	@echo "💥 Destroying Terraform resources..."
	@cd deployments/terraform && terraform destroy

.PHONY: bench
bench: ## Run benchmarks
	@echo "📊 Running benchmarks..."
	@go test -bench=. -benchmem ./...

.PHONY: profile-cpu
profile-cpu: build ## Run CPU profiling
	@echo "🔍 Running CPU profiling..."
	@./$(APP_NAME) -cpuprofile=cpu.prof &
	@sleep 10
	@pkill $(APP_NAME)
	@go tool pprof cpu.prof

.PHONY: profile-mem
profile-mem: build ## Run memory profiling
	@echo "🔍 Running memory profiling..."
	@./$(APP_NAME) -memprofile=mem.prof &
	@sleep 10
	@pkill $(APP_NAME)
	@go tool pprof mem.prof

.PHONY: health-check
health-check: ## Check application health
	@echo "🏥 Checking application health..."
	@curl -f http://localhost:8080/health || echo "❌ Health check failed"
	@curl -f http://localhost:8080/ready || echo "❌ Readiness check failed"

.PHONY: load-test
load-test: ## Run load test
	@echo "⚡ Running load test..."
	@hey -n 1000 -c 10 http://localhost:8080/health

.PHONY: release
release: clean test security build-all docker-build docker-scan ## Prepare release
	@echo "🚀 Release $(VERSION) prepared successfully!"
	@echo "Artifacts:"
	@ls -la dist/ 2>/dev/null || true
	@echo "Docker images:"
	@docker images $(IMAGE_NAME)

.PHONY: ci
ci: deps fmt lint test security build ## Run CI pipeline
	@echo "✅ CI pipeline completed successfully!"

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "🛠️  Installing development tools..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/sast-scan/cmd/gosec@latest
	@go install github.com/sonatypecommunity/nancy@latest

.DEFAULT_GOAL := help