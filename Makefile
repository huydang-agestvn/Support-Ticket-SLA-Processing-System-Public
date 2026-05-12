# Run API locally
run:
	go run ./cmd/api

# Run all tests
test:
	go test ./...

# Run tests with coverage
test-cover:
	go test ./... -cover

# Generate Swagger API docs
swagger:
	swag init -g cmd/api/main.go --parseInternal

# Start app and PostgreSQL with Docker Compose
docker-up:
# 	sudo docker compose up --build ( TODO: if code update will run this)
	sudo docker compose up
	sudo docker compose up

# Start app and PostgreSQL with Docker Compose and build images
docker-up-build:
	sudo docker compose up --build

# Stop Docker Compose services
docker-down:
	sudo docker compose down

# View API container logs
docker-logs:
	sudo docker logs ticket-sla-api