package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"support-ticket.com/internal/domain"
	"support-ticket.com/internal/service"
	"support-ticket.com/internal/worker"
)

const numWorkers = 5

func main() {
	events, err := loadEvents("./cmd/import-sample/ticket_events_sample.json")
	if err != nil {
		log.Fatalf("failed to load events: %v", err)
	}

	svc := service.NewTicketService()

	result := worker.Run(events, numWorkers, svc)

	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
}

func loadEvents(path string) ([]domain.TicketEvent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var events []domain.TicketEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, err
	}
	return events, nil
}