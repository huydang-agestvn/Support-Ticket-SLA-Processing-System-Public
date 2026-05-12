package repository

import (
	"context"

	"gorm.io/gorm"
	"support-ticket.com/internal/domain"
)

type TicketEventRepository interface {
	CreateBatch(events []domain.TicketEvent) error
	Create(ctx context.Context, event *domain.TicketEvent) error
	// FindTransitionEvent checks if an event with this exact transition already exists
	FindTransitionEvent(ctx context.Context, ticketID uint, fromStatus, toStatus domain.TicketStatus) (*domain.TicketEvent, error)
}

type ticketEventRepository struct {
	db *gorm.DB
}

func NewTicketEventRepository(db *gorm.DB) TicketEventRepository {
	return &ticketEventRepository{db}
}

func (r *ticketEventRepository) CreateBatch(events []domain.TicketEvent) error {
	return r.db.CreateInBatches(events, 100).Error
}

func (r *ticketEventRepository) Create(ctx context.Context, event *domain.TicketEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// FindTransitionEvent checks if an event with this exact transition already exists
func (r *ticketEventRepository) FindTransitionEvent(ctx context.Context, ticketID uint, fromStatus, toStatus domain.TicketStatus) (*domain.TicketEvent, error) {
	var event domain.TicketEvent
	err := r.db.WithContext(ctx).
		Where("ticket_id = ? AND from_status = ? AND to_status = ?", ticketID, fromStatus, toStatus).
		First(&event).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &event, nil
}
