# Run API locally
run:
	go run ./cmd/api

# Run all tests
test:
	go test ./...

# Run tests with coverage
test-cover:
	go test ./... -cover

# Start app and PostgreSQL with Docker Compose
docker-up:
	sudo docker compose up --build

# Stop Docker Compose services
docker-down:
	sudo docker compose down

# View API container logs
docker-logs:
	sudo docker logs ticket-sla-api