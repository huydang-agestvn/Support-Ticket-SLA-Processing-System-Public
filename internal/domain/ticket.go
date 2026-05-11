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
	ID          uint         `json:"id" gorm:"primaryKey"`
	AssigneeID  string       `json:"assignee_id" gorm:"column:assignee_id;type:varchar(255)"`
	RequestorID string       `json:"requestor_id" gorm:"column:requestor_id;type:varchar(255);not null"`
	Title       string       `json:"title" gorm:"column:title;type:varchar(255);not null"`
	Description string       `json:"description" gorm:"column:description;type:text"`
	Priority    Priority     `json:"priority" gorm:"column:priority;type:varchar(20);not null"`
	Status      TicketStatus `json:"status" gorm:"column:status;type:varchar(20);not null"`
	CreatedAt   time.Time    `json:"created_at" gorm:"column:created_at;not null;autoCreateTime:milli"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"column:updated_at"`
	ResolvedAt  *time.Time   `json:"resolved_at" gorm:"column:resolved_at"`
	SLADueAt    *time.Time   `json:"sla_due_at" gorm:"column:sla_due_at"`
	CancelledAt *time.Time   `json:"cancelled_at" gorm:"column:cancelled_at"`

	// Relations
	Events []TicketEvent `json:"events" gorm:"foreignKey:TicketID;constraint:OnDelete:CASCADE"`
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
	if t.SLADueAt == nil || t.SLADueAt.IsZero() {
		return fmt.Errorf("%w: SLA Due At is required for SLA tracking", ErrValidation)
	}
	if t.SLADueAt.Before(t.CreatedAt) {
		return fmt.Errorf("%w: SLA Due At cannot be before creation time", ErrValidation)
	}
	if t.Status == StatusResolved {
		if t.ResolvedAt == nil || t.ResolvedAt.IsZero() {
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
		t.ResolvedAt = &timestamp
	case StatusCancelled:
		t.CancelledAt = &timestamp
	}
	return nil
}
