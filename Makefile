.PHONY: build run docker-build docker-push help

# Variables
BINARY_NAME=uninaquiz-backend
IMAGE_NAME=uninaquiz-backend
DOCKER_REGISTRY?=
TAG?=latest

# Setup docker alias
dc = docker

# Default target
help:
	@echo "Available targets:"
	@echo "  build              - Build the Go project"
	@echo "  run                - Run the API server"
	@echo "  docker-build       - Build Docker image"
	@echo "  clean              - Remove binary files"

# Build the Go project
build:
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/$(BINARY_NAME) ./cmd
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
