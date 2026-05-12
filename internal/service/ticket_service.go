package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"support-ticket.com/internal/domain"
	"support-ticket.com/internal/dto"
	"support-ticket.com/internal/errmsgs"
	"support-ticket.com/internal/repository"
)

type TicketService interface {
	Create(ctx context.Context, req dto.CreateTicketReq) (*domain.Ticket, error)
	FindById(ctx context.Context, id uint) (*domain.Ticket, error)
	FindAll(ctx context.Context, filters map[string]interface{}) ([]domain.Ticket, error)
	UpdateTicketStatus(ctx context.Context, id uint, req dto.UpdateStatusReq) error
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

func (s *ticketServiceImpl) UpdateTicketStatus(ctx context.Context, id uint, req dto.UpdateStatusReq) error {
	ticket, err := s.repo.FindById(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get ticket: %w", err)
	}

	if ticket.Status == "new" && req.Status == "assigned" {
		if strings.TrimSpace(req.AssigneeID) == "" {
			return fmt.Errorf("assigneeId is required and cannot be empty when assigning a ticket")
		}
	}

	if err := ticket.UpdateStatusValidate(req.Status, time.Now()); err != nil {
		return fmt.Errorf("invalid ticket data: %w", err)
	}

	if ticket.Status == "new" && req.Status == "assigned" {
		if strings.TrimSpace(req.AssigneeID) == "" {
			return fmt.Errorf("assigneeId is required and cannot be empty when assigning a ticket")
		}
	}

	ticket.AssigneeID = req.AssigneeID

	// Cập nhật status mới cho ticket
	ticket.Status = req.Status

	// Build event
	event := &domain.TicketEvent{
		TicketID:   ticket.ID,
		FromStatus: ticket.Status,
		ToStatus:   req.Status,
		CreatedAt:  time.Now(),
	}
	if req.Note != "" {
		event.Note = &req.Note
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
