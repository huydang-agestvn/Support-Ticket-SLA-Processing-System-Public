package main

import (
	"log"

	"support-ticket.com/internal/app"
)

func main() {
	application := app.NewApp()

	if err := application.Run(); err != nil {
		log.Fatalf("application failed to start: %v", err)
	}
}
			