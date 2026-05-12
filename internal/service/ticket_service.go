package service

import (
	"context"
	"fmt"
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
	UpdateTicketStatus(ctx context.Context, id uint, newStatus domain.TicketStatus) error
}

type ticketServiceImpl struct {
	repo repository.TicketRepository
}

func NewTicketService(repo repository.TicketRepository) TicketService {
	return &ticketServiceImpl{
		repo: repo,
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

func (s *ticketServiceImpl) UpdateTicketStatus(ctx context.Context, id uint, newStatus domain.TicketStatus) error {
	ticket, err := s.FindById(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	if err := ticket.UpdateStatus(newStatus, now); err != nil {
		return fmt.Errorf("domain validation failed: %w", err)
	}

	if err := ticket.Validate(); err != nil {
		return fmt.Errorf("ticket state corrupted after status update: %w", err)
	}

	if err := s.repo.UpdateTicket(ctx, ticket); err != nil {
		return fmt.Errorf("failed to update ticket in db: %w", err)
	}

	return nil
}
