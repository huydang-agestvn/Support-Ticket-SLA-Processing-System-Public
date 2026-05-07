package domain

import (
	"fmt"
	"strings"
	"time"
)

type TicketEvent struct {
	ID         uint         `json:"event_id" gorm:"primarykey"`
	TicketID   uint         `json:"ticket_id"`
	Note       string       `json:"note"`
	FromStatus TicketStatus `json:"from_status"`
	ToStatus   TicketStatus `json:"to_status"`
	ActorID    string       `json:"actor_id"`
	CreatedAt  time.Time    `json:"created_at"`
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
	return nil
}
