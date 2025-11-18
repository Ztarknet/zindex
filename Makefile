.PHONY: help build run clean test docker-build docker-run docker-stop install deps fmt vet lint docker-build-prod docker-push helm-install helm-upgrade helm-uninstall helm-template docker-compose-up docker-compose-down docker-compose-logs

APP_NAME=zindex
CMD_PATH=./cmd/run
BUILD_DIR=./bin
DOCKER_IMAGE=zindex:latest
DOCKER_IMAGE_PROD=brandonjroberts/zindex
DOCKER_CONTAINER=zindex-container
CONFIG_PATH=configs/config.yaml
APP_VERSION?=v0.1.0
COMMIT_SHA?=$(shell git rev-parse --short HEAD)
POSTGRES_PASSWORD?=changeme

help:
	@echo "Available targets:"
	@echo ""
	@echo "Development:"
	@echo "  make build              - Build the zindex binary"
	@echo "  make run                - Run the indexer locally"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make test               - Run tests"
	@echo "  make fmt                - Format code"
	@echo "  make vet                - Run go vet"
	@echo "  make lint               - Run linter (requires golangci-lint)"
	@echo ""
	@echo "Docker (Local):"
	@echo "  make docker-build       - Build Docker image"
	@echo "  make docker-run         - Run Docker container"
	@echo "  make docker-stop        - Stop and remove Docker container"
	@echo ""
	@echo "Docker Compose:"
	@echo "  make docker-compose-up  - Start services with docker-compose"
	@echo "  make docker-compose-down- Stop and remove docker-compose services"
	@echo "  make docker-compose-logs- View docker-compose logs"
	@echo ""
	@echo "Docker (Production):"
	@echo "  make docker-build-prod  - Build production Docker image"
	@echo "  make docker-push        - Push Docker image to registry"
	@echo ""
	@echo "Kubernetes/Helm:"
	@echo "  make helm-install       - Install Helm chart"
	@echo "  make helm-upgrade       - Upgrade Helm chart"
	@echo "  make helm-uninstall     - Uninstall Helm chart"
	@echo "  make helm-template      - Render Helm templates locally"

build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

run: build
	@echo "Running $(APP_NAME)..."
	@$(BUILD_DIR)/$(APP_NAME) --config $(CONFIG_PATH)

run-dev:
	@echo "Running $(APP_NAME) in development mode..."
	@go run $(CMD_PATH)/main.go --config $(CONFIG_PATH)

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

test:
	@echo "Running tests..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

docker-run: docker-build
	@echo "Running Docker container..."
	@docker run -d \
		--name $(DOCKER_CONTAINER) \
		-p 8080:8080 \
		-v $(PWD)/configs:/root/configs \
		$(DOCKER_IMAGE)
	@echo "Docker container started: $(DOCKER_CONTAINER)"
	@echo "API available at: http://localhost:8080"

docker-stop:
	@echo "Stopping Docker container..."
	@docker stop $(DOCKER_CONTAINER) 2>/dev/null || true
	@docker rm $(DOCKER_CONTAINER) 2>/dev/null || true
	@echo "Docker container stopped"

docker-logs:
	@docker logs -f $(DOCKER_CONTAINER)

# Docker Compose targets
docker-compose-up:
	@echo "Starting services with docker-compose..."
	@docker-compose up --build -d
	@echo "Services started. Use 'make docker-compose-logs' to view logs"

docker-compose-down:
	@echo "Stopping docker-compose services..."
	@docker-compose down
	@echo "Services stopped"

docker-compose-logs:
	@docker-compose logs -f

install: deps
	@echo "Installing $(APP_NAME)..."
	@go install $(CMD_PATH)
	@echo "Install complete"

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies downloaded"

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vet complete"

lint:
	@echo "Running linter..."
	@golangci-lint run ./...
	@echo "Lint complete"

db-migrate:
	@echo "Running database migrations..."
	@echo "Migrations will be applied automatically on startup"

db-reset:
	@echo "Resetting database..."
	@echo "WARNING: This will drop all tables!"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ]
	@psql -h localhost -U zindex -d zindex -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@echo "Database reset complete"

# Production Docker targets
docker-build-prod:
	@echo "Building production Docker image..."
	@docker build --platform linux/amd64 -f Dockerfile.prod \
		-t $(DOCKER_IMAGE_PROD):latest \
		-t $(DOCKER_IMAGE_PROD):$(APP_VERSION)-$(COMMIT_SHA) .
	@echo "Production Docker image built:"
	@echo "  $(DOCKER_IMAGE_PROD):latest"
	@echo "  $(DOCKER_IMAGE_PROD):$(APP_VERSION)-$(COMMIT_SHA)"

docker-push:
	@echo "Pushing Docker images..."
	@docker push $(DOCKER_IMAGE_PROD):latest
	@docker push $(DOCKER_IMAGE_PROD):$(APP_VERSION)-$(COMMIT_SHA)
	@echo "Docker images pushed"

# Helm targets
helm-install:
	@echo "Installing Helm chart..."
	@helm install zindex deploy/zindex-infra \
		--set postgres.password=$(POSTGRES_PASSWORD) \
		--set deployments.zindex.commit_sha=$(COMMIT_SHA) \
		--set deployments.zindex.tag=$(APP_VERSION)
	@echo "Helm chart installed"

helm-upgrade:
	@echo "Upgrading Helm chart..."
	@helm upgrade zindex deploy/zindex-infra \
		--set postgres.password=$(POSTGRES_PASSWORD) \
		--set deployments.zindex.commit_sha=$(COMMIT_SHA) \
		--set deployments.zindex.tag=$(APP_VERSION)
	@echo "Helm chart upgraded"

helm-uninstall:
	@echo "Uninstalling Helm chart..."
	@helm uninstall zindex
	@echo "Helm chart uninstalled"

helm-template:
	@echo "Rendering Helm templates..."
	@helm template zindex deploy/zindex-infra \
		--set postgres.password=$(POSTGRES_PASSWORD) \
		--set deployments.zindex.commit_sha=$(COMMIT_SHA) \
		--set deployments.zindex.tag=$(APP_VERSION)

.DEFAULT_GOAL := help
