.PHONY: help build run test docker-build docker-up docker-down clean

help:
	@echo "Available commands:"
	@echo "  make build          - Build the project"
	@echo "  make run            - Run the service locally"
	@echo "  make test           - Run tests"
	@echo "  make docker-build   - Build docker image"
	@echo "  make docker-up      - Start docker containers"
	@echo "  make docker-down    - Stop docker containers"
	@echo "  make clean          - Clean build artifacts"

build:
	@echo "Building project..."
	go build -o bin/api ./cmd/api
	@echo "Build complete!"

run:
	@echo "Running service..."
	go run ./cmd/api

test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

docker-build:
	@echo "Building docker image..."
	docker build -t pr-reviewers:latest .
	@echo "Image built!"

docker-up:
	@echo "Starting docker containers..."
	docker-compose up -d
	@echo "Containers started!"

docker-down:
	@echo "Stopping docker containers..."
	docker-compose down
	@echo "Containers stopped!"

docker-logs:
	docker-compose logs -f api

clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean
	@echo "Clean complete!"

fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Format complete!"

lint:
	@echo "Running linter..."
	golangci-lint run ./...

vet:
	@echo "Running go vet..."
	go vet ./...
