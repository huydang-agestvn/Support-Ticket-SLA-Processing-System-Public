package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"support-ticket.com/internal/domain"
	"support-ticket.com/internal/dto"
)

type TicketRepository interface {
	Create(ctx context.Context, ticket *domain.Ticket) error
	FindById(ctx context.Context, id uint) (*domain.Ticket, error)
	FindAll(ctx context.Context, filter dto.TicketFilter, offset int, limit int) ([]domain.Ticket, int64, error)
	UpdateStatusWithEvent(ctx context.Context, ticket *domain.Ticket, event *domain.TicketEvent) error
	GetExistingTicketIDs(ctx context.Context, ticketIDs []uint) (map[uint]bool, error)
	GetTicketStatusAndCreatedAt(ctx context.Context, ticketIDs []uint) (map[uint]domain.TicketStatus, map[uint]time.Time, error)
	UpdateStatusAndAssignee(ctx context.Context, ticketID uint, status domain.TicketStatus, assigneeID string) error
	UpdateStatusAndResolvedAt(ctx context.Context, ticketID uint, status domain.TicketStatus, assigneeID string, resolvedAt time.Time) error
	UpdateStatusAndCancelledAt(ctx context.Context, ticketID uint, status domain.TicketStatus, assigneeID string, cancelledAt time.Time) error
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

func (r *ticketRepositoryImpl) FindAll(ctx context.Context, filter dto.TicketFilter, offset, limit int) ([]domain.Ticket, int64, error) {
	var tickets []domain.Ticket
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Ticket{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Priority != "" {
		query = query.Where("priority = ?", filter.Priority)
	}

	if filter.AssigneeID != "" {
		query = query.Where("assignee_id = ?", *&filter.AssigneeID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []domain.Ticket{}, 0, nil
	}
	err := query.Preload("Events").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&tickets).Error

	if err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

func (r *ticketRepositoryImpl) UpdateStatusWithEvent(ctx context.Context, ticket *domain.Ticket, event *domain.TicketEvent) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Update ticket
		if err := tx.Save(ticket).Error; err != nil {
			return fmt.Errorf("update ticket status: %w", err)
		}

		// 2. Insert event
		if err := tx.Create(event).Error; err != nil {
			return fmt.Errorf("insert ticket event: %w", err)
		}

		return nil
	})
}

func (r *ticketRepositoryImpl) GetExistingTicketIDs(ctx context.Context, ticketIDs []uint) (map[uint]bool, error) {
	var existingIDs []uint
	err := r.db.WithContext(ctx).Model(&domain.Ticket{}).Where("id IN ?", ticketIDs).Pluck("id", &existingIDs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint]bool)
	for _, id := range existingIDs {
		result[id] = true
	}

	return result, nil
}

func (r *ticketRepositoryImpl) GetTicketStatusAndCreatedAt(ctx context.Context, ticketIDs []uint) (map[uint]domain.TicketStatus, map[uint]time.Time, error) {
	if len(ticketIDs) == 0 {
		return make(map[uint]domain.TicketStatus), make(map[uint]time.Time), nil
	}

	type ticketMetadataRow struct {
		ID        uint               `gorm:"column:id"`
		Status    domain.TicketStatus `gorm:"column:status"`
		CreatedAt time.Time          `gorm:"column:created_at"`
	}

	var rows []ticketMetadataRow
	err := r.db.WithContext(ctx).Model(&domain.Ticket{}).
		Select("id, status, created_at").
		Where("id IN ?", ticketIDs).
		Find(&rows).Error
	if err != nil {
		return nil, nil, err
	}

	statuses := make(map[uint]domain.TicketStatus, len(rows))
	createdAt := make(map[uint]time.Time, len(rows))
	for _, row := range rows {
		statuses[row.ID] = row.Status
		createdAt[row.ID] = row.CreatedAt
	}

	return statuses, createdAt, nil
}

func (r *ticketRepositoryImpl) UpdateStatusAndAssignee(ctx context.Context, ticketID uint, status domain.TicketStatus, assigneeID string) error {
	return r.db.WithContext(ctx).Model(&domain.Ticket{}).Where("id = ?", ticketID).Updates(map[string]interface{}{"status": status, "assignee_id": assigneeID}).Error
}

func (r *ticketRepositoryImpl) updateStatusWithTimestamps(ctx context.Context, ticketID uint, status domain.TicketStatus, assigneeID string, resolvedAt, cancelledAt *time.Time) error {
	ticket, err := r.FindById(ctx, ticketID)
	if err != nil {
		return fmt.Errorf("failed to fetch ticket for status update: %w", err)
	}
	if ticket == nil {
		return fmt.Errorf("failed to fetch ticket for status update: %w", gorm.ErrRecordNotFound)
	}

	if resolvedAt != nil {
		validationTicket := &domain.Ticket{ResolvedAt: resolvedAt}
		if err := validationTicket.ValidateResolvedAt(ticket.CreatedAt); err != nil {
			return err
		}
	}
	if cancelledAt != nil {
		validationTicket := &domain.Ticket{CancelledAt: cancelledAt}
		if err := validationTicket.ValidateCancelledAt(ticket.CreatedAt); err != nil {
			return err
		}
	}

	updates := map[string]interface{}{"status": status, "assignee_id": assigneeID}
	if resolvedAt != nil {
		updates["resolved_at"] = *resolvedAt
	}
	if cancelledAt != nil {
		updates["cancelled_at"] = *cancelledAt
	}

	return r.db.WithContext(ctx).Model(&domain.Ticket{}).Where("id = ?", ticketID).Updates(updates).Error
}

func (r *ticketRepositoryImpl) UpdateStatusAndResolvedAt(ctx context.Context, ticketID uint, status domain.TicketStatus, assigneeID string, resolvedAt time.Time) error {
	return r.updateStatusWithTimestamps(ctx, ticketID, status, assigneeID, &resolvedAt, nil)
}

func (r *ticketRepositoryImpl) UpdateStatusAndCancelledAt(ctx context.Context, ticketID uint, status domain.TicketStatus, assigneeID string, cancelledAt time.Time) error {
	return r.updateStatusWithTimestamps(ctx, ticketID, status, assigneeID, nil, &cancelledAt)
}
