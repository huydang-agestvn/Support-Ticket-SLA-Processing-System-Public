package domain

import (
	"fmt"
	"time"

	"support-ticket.com/internal/errmsgs"
)

type TicketEvent struct {
	ID         uint         `json:"event_id,omitempty" gorm:"primaryKey"`
	TicketID   uint         `json:"ticket_id" gorm:"column:ticket_id;not null"`
	Note       *string      `json:"note" gorm:"column:note;type:text"`
	FromStatus TicketStatus `json:"from_status" gorm:"column:from_status;type:varchar(20);not null"`
	ToStatus   TicketStatus `json:"to_status" gorm:"column:to_status;type:varchar(20);not null"`
	AssigneeID string       `json:"assignee_id" gorm:"column:assignee_id;type:varchar(255);not null"`
	CreatedAt  time.Time    `json:"created_at" gorm:"column:created_at;not null;autoCreateTime:milli"`

	// Relations
	Ticket *Ticket `json:"-" gorm:"foreignKey:TicketID;constraint:OnDelete:CASCADE"`
}
type BatchImportResult struct {
	AcceptedCount   int              `json:"accepted_count"`
	RejectedCount   int              `json:"rejected_count"`
	DuplicateCount  int              `json:"duplicate_count"`
	RejectedDetails []RejectedDetail `json:"rejected_details"`
}

type RejectedDetail struct {
	ErrorName string        `json:"error_name"`
	Events    []TicketEvent `json:"events"`
}

func (e *TicketEvent) Validate() error {
	if e.TicketID == 0 {
		return fmt.Errorf("%w: Ticket ID is required", errmsgs.ErrInvalidInput)
	}
	if e.AssigneeID == "" {
		return fmt.Errorf("%w: Assignee ID is required", errmsgs.ErrInvalidInput)
	}

	if !e.FromStatus.IsValid() {
		return fmt.Errorf("%w: Unknown From Status '%s'", errmsgs.ErrInvalidInput, e.FromStatus)
	}
	if !e.ToStatus.IsValid() {
		return fmt.Errorf("%w: Unknown To Status '%s'", errmsgs.ErrInvalidInput, e.ToStatus)
	}
	if e.FromStatus == e.ToStatus {
		return fmt.Errorf("%w: From Status and To Status cannot be the same ('%s')", errmsgs.ErrInvalidStatusTransition, e.FromStatus)
	}
	if !e.FromStatus.CanTransitionTo(e.ToStatus) {
		return fmt.Errorf("%w: Illegal event transition intent from '%s' to '%s'", errmsgs.ErrInvalidStatusTransition, e.FromStatus, e.ToStatus)
	}
	if e.CreatedAt.IsZero() {
		return fmt.Errorf("%w: Event created_at is required", errmsgs.ErrInvalidInput)
	}
	return nil
}
