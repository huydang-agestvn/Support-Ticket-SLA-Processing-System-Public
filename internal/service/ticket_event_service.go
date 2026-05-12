package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"support-ticket.com/internal/domain"
	"support-ticket.com/internal/repository"
	"support-ticket.com/internal/worker"
)

type TicketEventService interface {
	Import(ctx context.Context, data []byte) (domain.BatchImportResult, error)
}

type ticketEventService struct {
	eventRepo repository.TicketEventRepository
}

func NewTicketEventService(eventRepo repository.TicketEventRepository) TicketEventService {
	return &ticketEventService{
		eventRepo: eventRepo,
	}
}

type workerOutput struct {
	Event  domain.TicketEvent
	Status EventResult
}

type EventResult string

const (
	ResultAccepted  EventResult = "accepted"
	ResultRejected  EventResult = "rejected"
	ResultDuplicate EventResult = "duplicate"
)

func (s *ticketEventService) processEvent(
	event *domain.TicketEvent,
	seen map[string]bool,
	mu *sync.Mutex,
) (EventResult, error) {
	key := fmt.Sprintf("%d|%s|%s", event.TicketID, event.FromStatus, event.ToStatus)

	mu.Lock()
	isDup := seen[key]
	if !isDup {
		seen[key] = true
	}
	mu.Unlock()

	if isDup {
		return ResultDuplicate, nil
	}

	if err := event.Validate(); err != nil {
		return ResultRejected, err
	}

	return ResultAccepted, nil
}

func (s *ticketEventService) parseAndValidateEvents(data []byte) ([]domain.TicketEvent, error) {
	var events []domain.TicketEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	var validEvents []domain.TicketEvent
	for _, e := range events {
		if err := e.Validate(); err != nil {
			continue
		}
		validEvents = append(validEvents, e)
	}
	return validEvents, nil
}

func (s *ticketEventService) Import(ctx context.Context, data []byte) (domain.BatchImportResult, error) {
	validEvents, err := s.parseAndValidateEvents(data)
	if err != nil {
		return domain.BatchImportResult{}, err
	}

	seenEvents := make(map[string]bool)
	var mu sync.Mutex

	results := worker.Run(validEvents, func(event domain.TicketEvent) workerOutput {
		status, _ := s.processEvent(&event, seenEvents, &mu)
		return workerOutput{
			Event:  event,
			Status: status,
		}
	})

	var eventsToInsert []domain.TicketEvent
	finalResult := domain.BatchImportResult{}
	for _, res := range results {
		switch res.Status {
		case ResultAccepted:
			eventsToInsert = append(eventsToInsert, res.Event)
			finalResult.AcceptedCount++
		case ResultDuplicate:
			finalResult.DuplicateCount++
		case ResultRejected:
			finalResult.RejectedCount++
		}
	}

	if len(eventsToInsert) > 0 {
		if err := s.eventRepo.CreateBatch(eventsToInsert); err != nil {
			return finalResult, err
		}
	}

	return finalResult, nil
}
