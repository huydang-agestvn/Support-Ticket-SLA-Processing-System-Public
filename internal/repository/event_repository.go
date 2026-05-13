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
	FetchLatestEventPerTicket(ctx context.Context, ticketIDs []int) ([]domain.TicketEvent, error)
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

	rows, err := r.db.WithContext(ctx).Raw("SELECT DISTINCT ON (ticket_id) ticket_id, to_status, assignee_id, created_at FROM ticket_events WHERE ticket_id IN (?) ORDER BY ticket_id, created_at DESC", ticketIDs).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.TicketEvent
	for rows.Next() {
		var ev domain.TicketEvent
		if err := rows.Scan(&ev.TicketID, &ev.ToStatus, &ev.AssigneeID, &ev.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, ev)
	}
	return results, rows.Err()
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
