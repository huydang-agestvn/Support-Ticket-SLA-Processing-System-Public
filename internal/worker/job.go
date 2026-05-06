package worker

import (
	"sync"
)

type ImportResult struct {
	Accepted  int
	Rejected  int
	Duplicate int
}

type Event struct {
	ID     int
	Status string
}

func ticketWorker(jobs <-chan Event, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for event := range jobs {
		switch event.Status {
		case "accepted":
			results <- "accepted"
		case "rejected":
			results <- "rejected"
		default:
			results <- "duplicate"
		}
	}
}
func StartTicketPool(events []Event, workerCount int) ImportResult {
	jobs := make(chan Event, len(events))
	results := make(chan string, len(events))
	var wg sync.WaitGroup

	for w := 1; w <= workerCount; w++ {
		wg.Add(1)
		go ticketWorker(jobs, results, &wg)
	}

	for _, event := range events {
		jobs <- event
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	res := ImportResult{}
	for r := range results {
		switch r {
		case "accepted":
			res.Accepted++
		case "rejected":
			res.Rejected++
		case "duplicate":
			res.Duplicate++
		}
	}

	return res
}
