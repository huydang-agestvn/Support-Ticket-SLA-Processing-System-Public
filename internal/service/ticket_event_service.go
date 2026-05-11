package service

import (
	"encoding/json"
	"fmt"
	"sync"

	"support-ticket.com/internal/domain"
	"support-ticket.com/internal/errors"
	"support-ticket.com/internal/repository"
	"support-ticket.com/internal/worker"
)

type TicketEventService interface {
	GetAll(limit, offset int) ([]domain.TicketEvent, int64, error)
	GetByID(id uint) (*domain.TicketEvent, error)
	GetByTicketID(ticketID uint) ([]domain.TicketEvent, error)
	Create(event *domain.TicketEvent) error
	Update(event *domain.TicketEvent) error
	Delete(id uint) error
	Import(data []byte) (domain.BatchImportResult, error)
}

type ticketEventService struct {
	eventRepo  repository.TicketEventRepository
}

func NewTicketEventService(eventRepo repository.TicketEventRepository) TicketEventService {
    return &ticketEventService{
        eventRepo:  eventRepo,
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

func (s *ticketEventService) Import(data []byte) (domain.BatchImportResult, error) {
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

func (s *ticketEventService) GetAll(limit, offset int) ([]domain.TicketEvent, int64, error) {
	return s.eventRepo.GetAll(limit, offset)
}

func (s *ticketEventService) GetByID(id uint) (*domain.TicketEvent, error) {
	event, err := s.eventRepo.GetByID(id)
	if err != nil {
		return nil, errors.ErrNotFound
	}
	return event, nil
}

func (s *ticketEventService) GetByTicketID(ticketID uint) ([]domain.TicketEvent, error) {
	// Kiểm tra ticket tồn tại trước
	// _, err := s.ticketRepo.GetByID(ticketID)
	// if err != nil {
	// 	return nil, errors.ErrNotFound
	// }

	return []domain.TicketEvent{}, nil
}

func (s *ticketEventService) Create(event *domain.TicketEvent) error {
	// Kiểm tra ticket tồn tại
	// ticket, err := s.ticketRepo.GetByID(event.TicketID)
	// if err != nil {
	// 	return errors.ErrNotFound
	// }

	// // from_status phải khớp với status hiện tại của ticket
	// if event.FromStatus != ticket.Status {
	// 	return errors.ErrInvalidStatusTransition
	// }
	return nil
	// return s.eventRepo.Create(event)
}

func (s *ticketEventService) Update(event *domain.TicketEvent) error {
	// Kiểm tra event tồn tại
	_, err := s.eventRepo.GetByID(event.ID)
	if err != nil {
		return errors.ErrNotFound
	}

	return s.eventRepo.Update(event)
}

func (s *ticketEventService) Delete(id uint) error {
	event, err := s.eventRepo.GetByID(id)
	if err != nil {
		return errors.ErrNotFound
	}

	return s.eventRepo.Delete(event)
}
