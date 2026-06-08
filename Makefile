.PHONY: build run docker-build docker-push help api-dev setup-air test test-unit test-integration test-coverage generate-mocks lint

# Variables
BINARY_NAME=uninaquiz-backend
IMAGE_NAME=uninaquiz-backend
DOCKER_REGISTRY?=
TAG?=latest
DB_URL=postgres://root:root@localhost:5432/db_uninaquiz?sslmode=disable


# Setup aliases
dc = docker
gs = goose


ifeq (create-migration,$(firstword $(MAKECMDGOALS)))
  # Pega do segundo argumento em diante
  MIGRATION_NAME := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # Transforma o nome capturado em um alvo vazio para o make não dar erro
  $(eval $(MIGRATION_NAME):;@:)
endif

.PHONY: create-migration

# Default target
help:
	@echo "Available targets:"
	@echo "  build              - Build the Go project"
	@echo "  api                - Build and run the API server"
	@echo "  api-dev            - Run the API with hot reload (development mode)"
	@echo "  docker-build       - Build Docker image"
	@echo "  clean              - Remove binary files"
	@echo "  create-migration   - Create a new database migration (usage: make create-migration add_email_column)"
	@echo "  migrate-up         - Run up migrations"
	@echo "  migrate-down       - Run down migrations"
	@echo "  generate-mocks     - Generate gomock mocks for all interfaces (requires mockgen)"
	@echo "  lint               - Run linter on entire project (requires golangci-lint)"
	@echo "  test               - Run all tests (unit + integration)"
	@echo "  test-unit          - Run only unit tests with race detector and verbose output"
	@echo "  test-coverage      - Run only unit tests and output HTML coverage report"
	@echo "  test-integration   - Run only integration/E2E tests (requires Docker)"

# Build the Go project
build:
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/$(BINARY_NAME) ./cmd
	@echo "Build complete: bin/$(BINARY_NAME)"

# Run the API
api: build
	@echo "Running API server..."
	./bin/$(BINARY_NAME)

# Setup air dependency on machine
setup-air:
	@if ! command -v air > /dev/null 2>&1; then \
		echo "air não encontrado. Realizando instalação..."; \
		go install github.com/air-verse/air@latest; \
		echo "air instalado com sucesso!"; \
	else \
		echo "air já está instalado."; \
	fi

# Run API with hot reload (development mode)
api-dev: setup-air
	@echo "Running API server with hot reload..."
	air -c .air.toml

# Build Docker image
docker-build:
	@echo "Building Docker image $(IMAGE_NAME):$(TAG)..."
	docker build -t $(IMAGE_NAME):$(TAG) .
	@echo "Docker image built: $(IMAGE_NAME):$(TAG)"

# Push Docker image (requires DOCKER_REGISTRY to be set)
docker-push: docker-build
	@if [ -z "$(DOCKER_REGISTRY)" ]; then \
		echo "Error: DOCKER_REGISTRY not set"; \
		exit 1; \
	fi
	@echo "Pushing image to $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(TAG)..."
	docker tag $(IMAGE_NAME):$(TAG) $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(TAG)
	docker push $(DOCKER_REGISTRY)/$(IMAGE_NAME):$(TAG)
	@echo "Push complete"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f bin/$(BINARY_NAME)
	@echo "Clean complete"

# ─── Linting & Code Quality ───────────────────────────────────────────────────

# Check code formatting and potential bugs
lint:
	@echo "Running lint checks..."
	@echo "  Checking code format with gofmt..."
	@files=$$(gofmt -l cmd internal tests 2>/dev/null); \
	if [ -n "$$files" ]; then \
		echo "ERROR: Code is not properly formatted. Run 'go fmt ./...' to fix."; \
		echo "$$files"; \
		exit 1; \
	fi
	@echo "  Checking for bugs with go vet..."
	@go vet ./...
	@echo "Lint checks complete."

# ─── Testing ──────────────────────────────────────────────────────────────────

# Install mockgen if not available
setup-mockgen:
	@if ! command -v mockgen > /dev/null 2>&1; then \
		echo "mockgen not found. Installing..."; \
		go install go.uber.org/mock/mockgen@latest; \
		echo "mockgen installed successfully!"; \
	else \
		echo "mockgen already installed."; \
	fi

# Directories scanned for mockable interfaces.
# Add a new directory here whenever a new layer with interfaces is created.
MOCK_SOURCE_DIRS := \
	internal/domain/repositories \
	internal/application/services \
	internal/application/ports

# Generate all gomock mocks by scanning MOCK_SOURCE_DIRS automatically.
# Any .go file added to those directories will be picked up on the next run.
generate-mocks: setup-mockgen
	@echo "Generating mocks..."
	@mkdir -p internal/mocks
	@rm -f internal/mocks/mock_*.go
	@for src in $$(find $(MOCK_SOURCE_DIRS) -name "*.go" -not -name "*_test.go" 2>/dev/null); do \
		dest="internal/mocks/mock_$$(basename $$src)"; \
		echo "  mockgen $$src → $$dest"; \
		mockgen -source=$$src -destination=$$dest -package=mocks || exit 1; \
	done
	@echo "Mock generation complete."

# Run all unit tests (no integration, no race, fast)
test:
	@echo "Running unit tests..."
	go test ./internal/...
	@echo "Unit tests complete."
	@echo ""
	@echo "Running integration tests..."
	go test -tags=integration -v -timeout=300s ./tests/integration/...
	@echo "Integration tests complete."

# Run unit tests with race detector and verbose output
test-unit:
	@echo "Running unit tests (race detector + verbose)..."
	go test -race -v ./internal/...
	@echo "Unit tests complete."

# Run unit tests and produce an HTML coverage report
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"

# Run integration/E2E tests
test-integration:
	@echo "Running integration tests (requires Docker for testcontainers)..."
	go test -tags=integration -v -timeout=300s ./tests/integration/...
	@echo "Integration tests complete."


# Setup goose dependency on machine
.PHONY: setup-goose migrate

setup-goose:
	@if ! command -v goose > /dev/null 2>&1; then \
		echo "goose não encontrado. Realizando instalação..."; \
		go install github.com/pressly/goose/v3/cmd/goose@latest; \
	else \
		echo "goose já está instalado."; \
	fi

migrate-up: setup-goose
	goose -dir ./db/migrations postgres $(DB_URL) up

create-migration:
	@if [ ! -d "./db/migrations" ]; then \
		echo "Diretório de migrations não encontrado. Criando..."; \
		mkdir -p ./db/migrations; \
	fi
	@if [ -z "$(MIGRATION_NAME)" ]; then \
		echo "Erro: Informe o nome da migration."; \
		echo "Uso: make create-migration nome_da_tabela"; \
		exit 1; \
	fi
	goose -dir ./db/migrations create $(MIGRATION_NAME) sql

migrate-down: setup-goose
	goose -dir ./db/migrations postgres $(DB_URL) down
