.PHONY: help build run clean test docker-build docker-run docker-stop install deps fmt vet lint

APP_NAME=zindex
CMD_PATH=./cmd/run
BUILD_DIR=./bin
DOCKER_IMAGE=zindex:latest
DOCKER_CONTAINER=zindex-container
CONFIG_PATH=configs/config.yaml

help:
	@echo "Available targets:"
	@echo "  make build         - Build the zindex binary"
	@echo "  make run           - Run the indexer locally"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make test          - Run tests"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run Docker container"
	@echo "  make docker-stop   - Stop and remove Docker container"
	@echo "  make install       - Install dependencies"
	@echo "  make deps          - Download Go dependencies"
	@echo "  make fmt           - Format code"
	@echo "  make vet           - Run go vet"
	@echo "  make lint          - Run linter (requires golangci-lint)"

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

.DEFAULT_GOAL := help
