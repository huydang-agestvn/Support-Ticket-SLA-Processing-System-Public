package repository

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"support-ticket.com/internal/domain"
)

type TicketRepository interface {
	Create(ctx context.Context, ticket *domain.Ticket) error
	FindById(ctx context.Context, id uint) (*domain.Ticket, error)
	FindAll(ctx context.Context, filters map[string]interface{}) ([]domain.Ticket, error)
	UpdateTicket(ctx context.Context, ticket *domain.Ticket) error
}

type ticketRepositoryImpl struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &ticketRepositoryImpl{db: db}
}

func (r *ticketRepositoryImpl) Create(ctx context.Context, ticket *domain.Ticket) error {
	return r.db.WithContext(ctx).Create(ticket).Error
}

func (r *ticketRepositoryImpl) FindById(ctx context.Context, id uint) (*domain.Ticket, error) {
	var ticket domain.Ticket

	err := r.db.WithContext(ctx).Preload("Events").First(&ticket, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil 
		}
		return nil, err
	}

	return &ticket, nil
}

func (r *ticketRepositoryImpl) FindAll(ctx context.Context, filters map[string]interface{}) ([]domain.Ticket, error) {
	var tickets []domain.Ticket
	query := r.db.WithContext(ctx)

	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if priority, ok := filters["priority"]; ok && priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if assigneeID, ok := filters["assignee_id"]; ok && assigneeID != "" {
		query = query.Where("assignee_id = ?", assigneeID)
	}

	err := query.Preload("Events").Order("created_at DESC").Find(&tickets).Error
	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func (r *ticketRepositoryImpl) UpdateTicket(ctx context.Context, ticket *domain.Ticket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}