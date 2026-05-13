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

type updateJob struct {
	TicketID   uint
	Status     domain.TicketStatus
	AssigneeID string
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

var maxBatchSize = config.GetBatchSize("MAX_BATCH_SIZE")

func (s *ticketEventService) processEvent(
	event *domain.TicketEvent,
	existingTickets map[uint]bool,
	existingDBEvents map[string]bool,
	localSeen map[string]bool,
	mu *sync.Mutex,
) (EventResult, error) {

	// Validate Business Logic (DB level)
	if !existingTickets[event.TicketID] {
		return ResultRejected, fmt.Errorf("ticket_id %d does not exist in DB", event.TicketID)
	}

	key := fmt.Sprintf("%d|%s|%s", event.TicketID, event.FromStatus, event.ToStatus)

	if existingDBEvents[key] {
		return ResultDuplicate, nil
	}

	mu.Lock()
	isDupLocal := localSeen[key]
	if !isDupLocal {
		localSeen[key] = true
	}
	mu.Unlock()

	if isDupLocal {
		return ResultDuplicate, nil
	}

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
			Err:   e.Validate(),
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
	rejectedEvents := make(map[string][]domain.TicketEvent)

	for _, pe := range parsedEvents {
		if pe.Err != nil {
			key := pe.Err.Error()
			rejectedEvents[key] = append(rejectedEvents[key], pe.Event)
			finalResult.RejectedCount++
			continue
		}
		validEvents = append(validEvents, pe.Event)
		ticketIDs = append(ticketIDs, pe.Event.TicketID)

		key := fmt.Sprintf("%d|%s|%s", pe.Event.TicketID, pe.Event.FromStatus, pe.Event.ToStatus)
		eventKeys = append(eventKeys, key)
	}

	// Convert rejectedEvents map to RejectedDetails
	for errorName, events := range rejectedEvents {
		finalResult.RejectedDetails = append(finalResult.RejectedDetails, domain.RejectedDetail{
			ErrorName: errorName,
			Events:    events,
		})
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

		// Sync ticket status after successful batch insert
		uniqueTicketIDs := make(map[uint]bool)
		for _, event := range eventsToInsert {
			uniqueTicketIDs[event.TicketID] = true
		}
		if len(uniqueTicketIDs) > 0 {
			ticketIDsInt := make([]int, 0, len(uniqueTicketIDs))
			for id := range uniqueTicketIDs {
				ticketIDsInt = append(ticketIDsInt, int(id))
			}
			latestEvents, err := s.eventRepo.FetchLatestEventPerTicket(ctx, ticketIDsInt)
			if err != nil {
				return finalResult, fmt.Errorf("failed to sync ticket status: %w", err)
			}
			jobs := make([]updateJob, 0, len(latestEvents))
			for _, ev := range latestEvents {
				jobs = append(jobs, updateJob{TicketID: uint(ev.TicketID), Status: ev.ToStatus, AssigneeID: ev.AssigneeID})
			}
			results := worker.Run(jobs, func(job updateJob) error {
				return s.ticketRepo.UpdateStatusAndAssignee(ctx, job.TicketID, job.Status, job.AssigneeID)
			})
			for _, err := range results {
				if err != nil {
					return finalResult, fmt.Errorf("failed to update ticket status: %w", err)
				}
			}
		}
	}

	return finalResult, nil
}
