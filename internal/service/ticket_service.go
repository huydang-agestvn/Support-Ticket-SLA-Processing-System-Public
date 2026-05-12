package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"support-ticket.com/internal/domain"
	"support-ticket.com/internal/dto"
	"support-ticket.com/internal/errmsgs"
	"support-ticket.com/internal/repository"
)

var (
	ErrTicketNotFound               = errors.New("ticket not found")
	ErrEventTransitionAlreadyExists = errors.New("event transition already exists")
)

type TicketService interface {
	Create(ctx context.Context, req dto.CreateTicketReq) (*domain.Ticket, error)
	FindById(ctx context.Context, id uint) (*domain.Ticket, error)
	FindAll(ctx context.Context, filters map[string]interface{}) ([]domain.Ticket, error)
	UpdateTicketStatus(ctx context.Context, id uint, newStatus domain.TicketStatus, actorID string, assigneeID string, note string) error
}

type ticketServiceImpl struct {
	repo      repository.TicketRepository
	eventRepo repository.TicketEventRepository
}

func NewTicketService(repo repository.TicketRepository, eventRepo repository.TicketEventRepository) TicketService {
	return &ticketServiceImpl{
		repo:      repo,
		eventRepo: eventRepo,
	}
}

func (s *ticketServiceImpl) Create(ctx context.Context, req dto.CreateTicketReq) (*domain.Ticket, error) {
	now := time.Now()

	ticket := &domain.Ticket{
		RequestorID: req.RequestorID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Status:      domain.StatusNew,
		CreatedAt:   now,
	}

	// SLA: High = 4h, Medium = 24h, Low = 48h
	var slaDuration time.Duration
	switch req.Priority {
	case domain.PriorityHigh:
		slaDuration = 4 * time.Hour
	case domain.PriorityMedium:
		slaDuration = 24 * time.Hour
	case domain.PriorityLow:
		slaDuration = 48 * time.Hour
	default:
		slaDuration = 48 * time.Hour
	}

	slaDueAt := now.Add(slaDuration)
	ticket.SLADueAt = &slaDueAt

	// Domain Validation
	if err := ticket.Validate(); err != nil {
		return nil, fmt.Errorf("invalid ticket data: %w", err)
	}

	// Persistence: DB layer
	if err := s.repo.Create(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to create ticket in db: %w", err)
	}

	return ticket, nil
}

func (s *ticketServiceImpl) FindById(ctx context.Context, id uint) (*domain.Ticket, error) {
	ticket, err := s.repo.FindById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket from db: %w", err)
	}

	if ticket == nil {
		return nil, errmsgs.ErrTicketNotFound
	}

	return ticket, nil
}

func (s *ticketServiceImpl) FindAll(ctx context.Context, filters map[string]interface{}) ([]domain.Ticket, error) {
	tickets, err := s.repo.FindAll(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list tickets: %w", err)
	}

	return tickets, nil
}

func (s *ticketServiceImpl) UpdateTicketStatus(ctx context.Context, id uint, newStatus domain.TicketStatus, actorID string, assigneeID string, note string) error {
	// 1. Validate input
	if newStatus == "" || actorID == "" {
		return fmt.Errorf("%w: status and actorID are required", domain.ErrValidation)
	}

	// 2. Lấy ticket
	ticket, err := s.FindById(ctx, id)
	if err != nil {
		return err
	}

	oldStatus := ticket.Status

	// 3. Không thay đổi gì nếu status giống nhau
	if oldStatus == newStatus {
		return nil
	}

	// 4. Validate assignee_id chỉ cho assigned
	if assigneeID != "" && newStatus != domain.StatusAssigned {
		return fmt.Errorf("%w: assignee_id is only allowed when transitioning to assigned", domain.ErrValidation)
	}

	if newStatus == domain.StatusAssigned {
		if assigneeID == "" {
			return fmt.Errorf("%w: assignee_id is required when transitioning to assigned", domain.ErrValidation)
		}
		ticket.AssigneeID = assigneeID
	}

	// 5. Validate và update status trên struct
	now := time.Now()
	if err := ticket.UpdateStatus(newStatus, now); err != nil {
		return fmt.Errorf("domain validation failed: %w", err)
	}

	// 6. Check duplicate transition event
	existingEvent, err := s.eventRepo.FindTransitionEvent(ctx, ticket.ID, oldStatus, newStatus)
	if err != nil {
		return fmt.Errorf("failed to check existing event: %w", err)
	}
	if existingEvent != nil {
		return fmt.Errorf("%w: transition from %s to %s already exists", ErrEventTransitionAlreadyExists, oldStatus, newStatus)
	}

	// 7. Build event
	event := &domain.TicketEvent{
		TicketID:   ticket.ID,
		FromStatus: oldStatus,
		ToStatus:   newStatus,
		ActorID:    actorID,
		CreatedAt:  now,
	}
	if note != "" {
		event.Note = &note
	}
	if err := event.Validate(); err != nil {
		return fmt.Errorf("failed to validate event: %w", err)
	}

	// 8. Update ticket + insert event trong transaction
	if err := s.repo.UpdateStatusWithEvent(ctx, ticket, event); err != nil {
		return fmt.Errorf("failed to update ticket status with event: %w", err)
	}

	return nil
}
