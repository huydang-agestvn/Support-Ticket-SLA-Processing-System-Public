package repository

import (
	"context"

	"gorm.io/gorm"
	"support-ticket.com/internal/domain"
)

type TicketEventRepository interface {
	CreateBatch(events []domain.TicketEvent) error
	Create(ctx context.Context, event *domain.TicketEvent) error
	GetExistingEventKeys(ctx context.Context, keys []string) (map[string]bool, error)
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

func (r *ticketEventRepository) GetExistingEventKeys(ctx context.Context, keys []string) (map[string]bool, error) {
	var existingKeys []string
	err := r.db.WithContext(ctx).Raw("SELECT CONCAT(ticket_id, '|', from_status, '|', to_status) as key FROM ticket_events WHERE CONCAT(ticket_id, '|', from_status, '|', to_status) IN (?)", keys).Scan(&existingKeys).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool)
	for _, key := range existingKeys {
		result[key] = true
	}

	return result, nil
}
