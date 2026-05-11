package domain

import (
	"fmt"
	"strings"
	"time"
)

type TicketEvent struct {
	ID         uint         `json:"event_id" gorm:"primaryKey"`
	TicketID   uint         `json:"ticket_id" gorm:"column:ticket_id;not null"`
	Note       *string      `json:"note" gorm:"column:note;type:text"`
	FromStatus TicketStatus `json:"from_status" gorm:"column:from_status;type:varchar(20);not null"`
	ToStatus   TicketStatus `json:"to_status" gorm:"column:to_status;type:varchar(20);not null"`
	ActorID    string       `json:"actor_id" gorm:"column:actor_id;type:varchar(255);not null"`
	CreatedAt  time.Time    `json:"created_at" gorm:"column:created_at;not null;autoCreateTime:milli"`

	// Relations
	Ticket *Ticket `json:"-" gorm:"foreignKey:TicketID;constraint:OnDelete:CASCADE"`
}
type BatchImportResult struct {
	AcceptedCount  int `json:"accepted_count"`
	RejectedCount  int `json:"rejected_count"`
	DuplicateCount int `json:"duplicate_count"`
}

func (e *TicketEvent) Validate() error {
	if strings.TrimSpace(e.ActorID) == "" {
		return fmt.Errorf("%w: Actor ID is required", ErrValidation)
	}
	if !e.FromStatus.IsValid() {
		return fmt.Errorf("%w: Unknown From Status '%s'", ErrValidation, e.FromStatus)
	}
	if !e.ToStatus.IsValid() {
		return fmt.Errorf("%w: Unknown To Status '%s'", ErrValidation, e.ToStatus)
	}
	if e.FromStatus != e.ToStatus && !e.FromStatus.CanTransitionTo(e.ToStatus) {
		return fmt.Errorf("%w: Illegal event transition intent from '%s' to '%s'", ErrInvalidTransition, e.FromStatus, e.ToStatus)
	}
	if e.CreatedAt.IsZero() {
		return fmt.Errorf("%w: Event created_at is required", ErrValidation)
	}
	return nil
}
