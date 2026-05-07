package worker

import (
	"sync"

	"support-ticket.com/internal/domain"
	"support-ticket.com/internal/service"
)

type jobResult struct {
	result service.EventResult
}
// Run worker pool
func Run(events []domain.TicketEvent, numWorkers int, svc *service.TicketService) domain.BatchImportResult {
	jobs := make(chan domain.TicketEvent, len(events))
	results := make(chan jobResult, len(events))

	// spawn N workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for event := range jobs {
				e := event
				r, _ := svc.ProcessEvent(&e)
				results <- jobResult{result: r}
			}
		}()
	}

	//Close
	for _, e := range events {
		jobs <- e
	}
	close(jobs) 

	go func() {
		wg.Wait()
		close(results)
	}()

	// Output
	var summary domain.BatchImportResult
	for r := range results {
		switch r.result {
		case service.ResultAccepted:
			summary.AcceptedCount++
		case service.ResultRejected:
			summary.RejectedCount++
		case service.ResultDuplicate:
			summary.DuplicateCount++
		}
	}

	return summary
}