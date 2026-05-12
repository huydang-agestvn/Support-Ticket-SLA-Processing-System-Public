package repository

import (
	"support-ticket.com/internal/domain"

	"gorm.io/gorm"
)

type TicketEventRepository interface {
	CreateBatch(events []domain.TicketEvent) error
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
