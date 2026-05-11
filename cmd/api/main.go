package main

import (
	"log"
	"net/http"

	"support-ticket.com/internal/app"
)

// @title Support Ticket SLA Processing System API
// @version 1.0
// @description API documentation for Support Ticket SLA Processing System.
// @host localhost:8080
// @BasePath /api/v1
func main() {
	application := app.NewApp()

	if err := application.Run(); err != nil {
		log.Fatalf("application failed to start: %v", err)
	}
}
