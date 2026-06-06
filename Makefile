.PHONY: build run docker-build docker-push help

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
	@echo "  run                - Run the API server"
	@echo "  docker-build       - Build Docker image"
	@echo "  clean              - Remove binary files"
	@echo "  create-migration   - Create a new database migration (usage: make create-migration add_email_column)"

# Build the Go project
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME) ./cmd
	@echo "Build complete: bin/$(BINARY_NAME)"

# Run the API
api: build
	@echo "Running API server..."
	./bin/$(BINARY_NAME)

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
