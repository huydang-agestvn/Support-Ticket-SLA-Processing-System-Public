package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"support-ticket.com/internal/config"
	"support-ticket.com/internal/domain"
	"support-ticket.com/internal/errmsgs"
	"support-ticket.com/internal/repository"
	"support-ticket.com/internal/worker"
)

type TicketEventService interface {
	Import(ctx context.Context, data []byte) (domain.BatchImportResult, error)
}

type ticketEventService struct {
	eventRepo  repository.TicketEventRepository
	ticketRepo repository.TicketRepository
}

func NewTicketEventService(eventRepo repository.TicketEventRepository, ticketRepo repository.TicketRepository) TicketEventService {
	return &ticketEventService{
		eventRepo:  eventRepo,
		ticketRepo: ticketRepo,
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

var maxBatchSize = config.GetBatchSize("WORKER_BATCH_SIZE")

func (s *ticketEventService) processEvent(
	event *domain.TicketEvent,
	existingTickets map[uint]bool, // Cache từ DB
	existingDBEvents map[string]bool, // Cache từ DB
	localSeen map[string]bool, // Trạng thái của current batch
	mu *sync.Mutex,
) (EventResult, error) {

	// 1. Validate Business Logic (DB level)
	// Kiểm tra xem Ticket có thực sự tồn tại dưới DB không
	if !existingTickets[event.TicketID] {
		return ResultRejected, fmt.Errorf("ticket_id %d does not exist in DB", event.TicketID)
	}

	key := fmt.Sprintf("%d|%s|%s", event.TicketID, event.FromStatus, event.ToStatus)

	// 2. Kiểm tra Duplicate với DB (Những lần import trước đó)
	if existingDBEvents[key] {
		return ResultDuplicate, nil
	}

	// 3. Kiểm tra Duplicate trong chính batch đang xử lý (Concurrency safe)
	mu.Lock()
	isDupLocal := localSeen[key]
	if !isDupLocal {
		localSeen[key] = true
	}
	mu.Unlock()

	if isDupLocal {
		return ResultDuplicate, nil
	}

	// Logic e.Validate() đã chạy ở vòng parseEvents rồi, không cần chạy lại ở đây nữa

	return ResultAccepted, nil
}

type parsedEvent struct {
	Event domain.TicketEvent
	Err   error // nil = valid
}

func (s *ticketEventService) parseEvents(data []byte) ([]parsedEvent, error) {
	if len(data) == 0 {
		return nil, errmsgs.ErrEmptyBody
	}

	var events []domain.TicketEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if len(events) == 0 {
		return nil, errmsgs.ErrEmptyBatch
	}

	if len(events) > maxBatchSize {
		return nil, fmt.Errorf("%w: got %d, max %d", errmsgs.ErrBatchTooLarge, len(events), maxBatchSize)
	}

	parsed := make([]parsedEvent, len(events))
	for i, e := range events {
		parsed[i] = parsedEvent{
			Event: e,
			Err:   e.Validate(), // nil nếu valid
		}
	}
	return parsed, nil
}

func (s *ticketEventService) Import(ctx context.Context, data []byte) (domain.BatchImportResult, error) {
	parsedEvents, err := s.parseEvents(data)
	if err != nil {
		return domain.BatchImportResult{}, err
	}
	finalResult := domain.BatchImportResult{}
	var ticketIDs []uint
	var eventKeys []string
	validEvents := make([]domain.TicketEvent, 0, len(parsedEvents))

	for _, pe := range parsedEvents {
		if pe.Err != nil {
			finalResult.RejectedCount++
			continue
		}
		validEvents = append(validEvents, pe.Event)
		ticketIDs = append(ticketIDs, pe.Event.TicketID)

		key := fmt.Sprintf("%d|%s|%s", pe.Event.TicketID, pe.Event.FromStatus, pe.Event.ToStatus)
		eventKeys = append(eventKeys, key)
	}

	existingTickets, err := s.ticketRepo.GetExistingTicketIDs(ctx, ticketIDs)
	if err != nil {
		return domain.BatchImportResult{}, fmt.Errorf("failed to fetch tickets: %w", err)
	}

	existingDBEvents, err := s.eventRepo.GetExistingEventKeys(ctx, eventKeys)
	if err != nil {
		return domain.BatchImportResult{}, fmt.Errorf("failed to fetch existing events: %w", err)
	}

	localSeenEvents := make(map[string]bool)
	var mu sync.Mutex

	results := worker.Run(validEvents, func(event domain.TicketEvent) workerOutput {
		status, _ := s.processEvent(&event, existingTickets, existingDBEvents, localSeenEvents, &mu)
		return workerOutput{
			Event:  event,
			Status: status,
		}
	})

	var eventsToInsert []domain.TicketEvent

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
