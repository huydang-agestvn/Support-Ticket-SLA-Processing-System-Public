# TODO: Setup Docker in Week 6
# For now, use local PostgreSQL development setup

# To run PostgreSQL locally on macOS (using Homebrew):
#   brew install postgresql
#   brew services start postgresql
#   createdb ticket_sla
#
# Or using Docker for just PostgreSQL:
#   docker run --name ticket-db -e POSTGRES_DB=ticket_sla -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres:16-alpine
#
# Then run: go run ./cmd/api/main.go

# =========================
# Stage 1: Build Go binary
# =========================
FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ticket-sla-api ./cmd/api


# =========================
# Stage 2: Run application
# =========================
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/ticket-sla-api .

EXPOSE 8080

CMD ["./ticket-sla-api"]