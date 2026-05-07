package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type TicketStatus string
type Priority string

const (
	StatusNew        TicketStatus = "new"
	StatusAssigned   TicketStatus = "assigned"
	StatusInProgress TicketStatus = "in_progress"
	StatusResolved   TicketStatus = "resolved"
	StatusClosed     TicketStatus = "closed"
	StatusCancelled  TicketStatus = "cancelled"
)

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Ticket struct {
	ID          uint         `json:"id" gorm:"primarykey"`
	AssigneeID  string       `json:"assignee_id"`
	RequestorID string       `json:"requestor_id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Priority    Priority     `json:"priority"`
	Status      TicketStatus `json:"status"`
	Deadline    time.Time    `json:"deadline"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	ResolvedAt  time.Time    `json:"resolved_at"`
	SLADueAt    time.Time    `json:"sla_due_at"`
	CancelledAt time.Time    `json:"cancelled_at"`
}

var (
	ErrValidation        = errors.New("ticket validation failed")
	ErrInvalidTransition = errors.New("invalid status transition")
)

func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	}
	return false
}

func (s TicketStatus) IsValid() bool {
	switch s {
	case StatusNew, StatusAssigned, StatusInProgress, StatusResolved, StatusClosed, StatusCancelled:
		return true
	}
	return false
}

var ticketTransitions = map[TicketStatus]map[TicketStatus]bool{
	StatusNew: {
		StatusAssigned:  true,
		StatusCancelled: true,
	},
	StatusAssigned: {
		StatusInProgress: true,
		StatusCancelled:  true,
	},
	StatusInProgress: {
		StatusResolved: true,
	},
	StatusResolved: {
		StatusClosed: true,
	},
}

func (s TicketStatus) CanTransitionTo(next TicketStatus) bool {
	allowed, ok := ticketTransitions[s]
	if !ok {
		return false
	}
	return allowed[next]
}

func (t *Ticket) Validate() error {
	if strings.TrimSpace(t.Title) == "" {
		return fmt.Errorf("%w: Title is required", ErrValidation)
	}
	if strings.TrimSpace(t.Description) == "" {
		return fmt.Errorf("%w: Description is required", ErrValidation)
	}
	if strings.TrimSpace(t.RequestorID) == "" {
		return fmt.Errorf("%w: Requestor ID is required", ErrValidation)
	}
	if !t.Priority.IsValid() {
		return fmt.Errorf("%w: Unknown priority '%s'", ErrValidation, t.Priority)
	}
	if !t.Status.IsValid() {
		return fmt.Errorf("%w: Unknown status '%s'", ErrValidation, t.Status)
	}
	if t.CreatedAt.IsZero() {
		return fmt.Errorf("%w: Created At is required", ErrValidation)
	}
	if t.Deadline.IsZero() {
		return fmt.Errorf("%w: Deadline is required for SLA tracking", ErrValidation)
	}
	if t.Deadline.Before(t.CreatedAt) {
		return fmt.Errorf("%w: Deadline cannot be before creation time", ErrValidation)
	}
	if t.Status == StatusResolved {
		if t.ResolvedAt.IsZero() {
			return fmt.Errorf("%w: Resolved At is required when status is resolved", ErrValidation)
		}
		if t.ResolvedAt.Before(t.CreatedAt) {
			return fmt.Errorf("%w: Resolved At cannot be before Created At", ErrValidation)
		}
	}
	return nil
}

func (t *Ticket) UpdateStatus(newStatus TicketStatus, timestamp time.Time) error {
	if t.Status == newStatus {
		return nil
	}
	if !newStatus.IsValid() {
		return fmt.Errorf("cannot transition to unknown status '%s': %w", newStatus, ErrInvalidTransition)
	}
	if !t.Status.CanTransitionTo(newStatus) {
		return fmt.Errorf("cannot transition from '%s' to '%s': %w", t.Status, newStatus, ErrInvalidTransition)
	}
	t.Status = newStatus
	t.UpdatedAt = timestamp
	switch newStatus {
	case StatusResolved:
		t.ResolvedAt = timestamp
	case StatusCancelled:
		t.CancelledAt = timestamp
	}
	return nil
}
