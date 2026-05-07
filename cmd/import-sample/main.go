package main

import (
	"fmt"
	"math/rand"
	"time"

	"support-ticket.com/internal/worker"
)

func generateMockEvents() []worker.Event {
	rand.Seed(time.Now().UnixNano())

	statuses := []string{
		"accepted",
		"rejected",
		"duplicate",
	}

	events := make([]worker.Event, 1000)

	for i := 0; i < 1000; i++ {
		events[i] = worker.Event{
			ID:     i + 1,
			Status: statuses[rand.Intn(len(statuses))],
		}
	}

	return events
}

func main() {
	start := time.Now()
	defer func() {
		fmt.Println()
		fmt.Printf("Total time: %v\n", time.Since(start))
	}()
	events := generateMockEvents()

	result := worker.StartTicketPool(events, 5)

	fmt.Printf(
		`{"accepted_count": %d, "rejected_count": %d, "duplicate_count": %d}`,
		result.Accepted,
		result.Rejected,
		result.Duplicate,
	)

}

