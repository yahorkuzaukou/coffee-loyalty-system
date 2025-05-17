.PHONY: build build-docker clean clean-docker test migrate

# Build the application
build:
	go build -o bin/server

start:
	./bin/server

# Build Docker image
build-docker:
	docker compose build

start-docker:
	docker compose up

# Clean build artifacts
clean:
	rm -rf bin/

# Clean Docker images
clean-docker:
	docker rmi coffee-loyalty-system:latest || true

# Run tests
test:
	go test -v ./...

# Run database migrations
migrate:
	go run cmd/migrate/main.go 