package repository

import (
	"support-ticket.com/internal/domain"

	"gorm.io/gorm"
)

type TicketEventRepository interface {
	GetAll(limit, offset int) ([]domain.TicketEvent, int64, error)
	GetByID(id uint) (*domain.TicketEvent, error)
	GetByTicketID(ticketID uint) ([]domain.TicketEvent, error)
	Create(event *domain.TicketEvent) error
	Update(event *domain.TicketEvent) error
	Delete(event *domain.TicketEvent) error
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

func (r *ticketEventRepository) GetAll(limit, offset int) ([]domain.TicketEvent, int64, error) {
	var events []domain.TicketEvent
	var total int64

	r.db.Model(&domain.TicketEvent{}).Count(&total)

	err := r.db.
		Preload("Ticket").
		Limit(limit).
		Offset(offset).
		Find(&events).Error

	return events, total, err
}

func (r *ticketEventRepository) GetByID(id uint) (*domain.TicketEvent, error) {
	var event domain.TicketEvent

	err := r.db.
		Preload("Ticket").
		First(&event, id).Error
	return &event, err
}

func (r *ticketEventRepository) GetByTicketID(ticketID uint) ([]domain.TicketEvent, error) {
	var events []domain.TicketEvent

	err := r.db.
		Where("ticket_id = ?", ticketID).
		Order("created_at ASC").
		Find(&events).Error

	return events, err
}

func (r *ticketEventRepository) Create(event *domain.TicketEvent) error {
	return r.db.Create(event).Error
}

func (r *ticketEventRepository) Update(event *domain.TicketEvent) error {
	return r.db.Save(event).Error
}

func (r *ticketEventRepository) Delete(event *domain.TicketEvent) error {
	return r.db.Delete(event).Error
}
