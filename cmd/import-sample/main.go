package importsample

import (
	"fmt"
	"math/rand"
	"time"
)

func generateMockEvents() []Event {
	rand.Seed(time.Now().UnixNano())

	statuses := []string{
		"accepted",
		"rejected",
		"duplicate",
	}

	events := make([]Event, 100)

	for i := 0; i < 100; i++ {
		events[i] = Event{
			ID:     i + 1,
			Status: statuses[rand.Intn(len(statuses))],
		}
	}

	return events
}

func main() {
	events := generateMockEvents()

	result := StartTicketPool(events, 5)

	fmt.Printf(
		`{"accepted_count": %d, "rejected_count": %d, "duplicate_count": %d}`,
		result.Accepted,
		result.Rejected,
		result.Duplicate,
	)
}