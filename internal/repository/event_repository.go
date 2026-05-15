package repository

import (
	"context"

	"gorm.io/gorm"
	domain "support-ticket.com/internal/model"
)

type TicketEventRepository interface {
	CreateBatch(events []domain.TicketEvent) error
	Create(ctx context.Context, event *domain.TicketEvent) error
	GetExistingEventKeys(ctx context.Context, keys []string) (map[string]bool, error)
	FetchLatestEventPerTicket(ctx context.Context, ticketIDs []int) ([]domain.TicketEvent, error)
	FetchLatestResolvedEventPerTicket(ctx context.Context, ticketIDs []int) ([]domain.TicketEvent, error)
}

type ticketEventRepository struct {
	db *gorm.DB
}

func NewTicketEventRepository(db *gorm.DB) TicketEventRepository {
	return &ticketEventRepository{db}
}

// FetchLatestEventPerTicket implements [TicketEventRepository].
func (r *ticketEventRepository) FetchLatestEventPerTicket(ctx context.Context, ticketIDs []int) ([]domain.TicketEvent, error) {
	if len(ticketIDs) == 0 {
		return nil, nil
	}

	var results []domain.TicketEvent
	err := r.db.WithContext(ctx).
		Model(&domain.TicketEvent{}).
		Select("DISTINCT ON (ticket_id) ticket_id, to_status, assignee_id, created_at").
		Where("ticket_id IN ?", ticketIDs).
		Order("ticket_id, created_at DESC").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *ticketEventRepository) FetchLatestResolvedEventPerTicket(ctx context.Context, ticketIDs []int) ([]domain.TicketEvent, error) {
	if len(ticketIDs) == 0 {
		return nil, nil
	}

	var results []domain.TicketEvent
	err := r.db.WithContext(ctx).
		Model(&domain.TicketEvent{}).
		Select("DISTINCT ON (ticket_id) ticket_id, created_at").
		Where("ticket_id IN ? AND to_status = ?", ticketIDs, domain.StatusResolved).
		Order("ticket_id, created_at DESC").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	return results, nil
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
