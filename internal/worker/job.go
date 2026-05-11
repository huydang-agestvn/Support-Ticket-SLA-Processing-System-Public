package worker

import (
	"sync"

	"support-ticket.com/internal/config"
)

type JobResult[R any] struct {
	Value R
	Err   error
}

var worker = config.GetPoolSize("WORKER_POOL_SIZE")

func Run[T any, R any](items []T, job func(T) R) []R {
	jobs := make(chan T, len(items))
	results := make(chan R, len(items))
	var wg sync.WaitGroup

	for i := 0; i < worker; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range jobs {
				results <- job(item)
			}
		}()
	}

	for _, item := range items {
		jobs <- item
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var allResults []R
	for r := range results {
		allResults = append(allResults, r)
	}
	return allResults
}
