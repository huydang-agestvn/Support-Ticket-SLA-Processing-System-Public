package domain

import (
	"fmt"
	"strings"
	"time"

	"support-ticket.com/internal/errmsgs"
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
	ResolvedAt  *time.Time   `json:"resolved_at" gorm:"column:resolved_at"`
	SLADueAt    *time.Time   `json:"sla_due_at" gorm:"column:sla_due_at"`
	CancelledAt *time.Time   `json:"cancelled_at" gorm:"column:cancelled_at"`

	// TODO:Relations
	Events []TicketEvent `json:"events" gorm:"foreignKey:TicketID;constraint:OnDelete:CASCADE"`
}

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
		return fmt.Errorf("%w: Title is required", errmsgs.ErrInvalidInput)
	}
	if strings.TrimSpace(t.Description) == "" {
		return fmt.Errorf("%w: Description is required", errmsgs.ErrInvalidInput)
	}
	if strings.TrimSpace(t.RequestorID) == "" {
		return fmt.Errorf("%w: Requestor ID is required", errmsgs.ErrInvalidInput)
	}
	if !t.Priority.IsValid() {
		return fmt.Errorf("%w: Unknown priority '%s'", errmsgs.ErrInvalidInput, t.Priority)
	}
	if !t.Status.IsValid() {
		return fmt.Errorf("%w: Unknown status '%s'", errmsgs.ErrInvalidInput, t.Status)
	}
	if t.CreatedAt.IsZero() {
		return fmt.Errorf("%w: Created At is required", errmsgs.ErrInvalidInput)
	}
	if t.SLADueAt == nil || t.SLADueAt.IsZero() {
		return fmt.Errorf("%w: SLA Due At is required for SLA tracking", errmsgs.ErrInvalidInput)
	}
	if t.SLADueAt.Before(t.CreatedAt) {
		return fmt.Errorf("%w: SLA Due At cannot be before creation time", errmsgs.ErrInvalidInput)
	}
	if t.Status == StatusResolved {
		if err := validateTimestampAfterCreation(t.ResolvedAt, "Resolved At", t.CreatedAt); err != nil {
			return err
		}
	}
	if t.Status == StatusCancelled {
		if err := validateTimestampAfterCreation(t.CancelledAt, "Cancelled At", t.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}

func validateTimestampAfterCreation(ts *time.Time, fieldName string, createdAt time.Time) error {
	if ts == nil || ts.IsZero() {
		return fmt.Errorf("%w: %s is required", errmsgs.ErrInvalidInput, fieldName)
	}
	if ts.Before(createdAt) {
		return fmt.Errorf("%w: %s cannot be before Created At", errmsgs.ErrInvalidInput, fieldName)
	}
	return nil
}

func (t *Ticket) ValidateStatusTransition(newStatus TicketStatus, reqAssigneeId string, timestamp time.Time) error {
	reqAssigneeId = strings.TrimSpace(reqAssigneeId)

	if t.Status == StatusNew && newStatus == StatusAssigned {
		if reqAssigneeId == "" {
			return fmt.Errorf("%w: Assignee ID is required when assigning a ticket", errmsgs.ErrInvalidInput)
		}
		t.AssigneeID = reqAssigneeId
	} else if reqAssigneeId != "" && reqAssigneeId != t.AssigneeID {
		return fmt.Errorf("%w: Cannot change assignee to '%s' during status transition to '%s'. Current assignee is '%s'",
			errmsgs.ErrInvalidInput, reqAssigneeId, newStatus, t.AssigneeID)
	}

	if t.Status == newStatus {
		return fmt.Errorf("%w: Status is already set to '%s'", errmsgs.ErrInvalidStatusTransition, newStatus)
	}
	if !newStatus.IsValid() {
		return fmt.Errorf("%w: Cannot transition to unknown status '%s'", errmsgs.ErrInvalidStatusTransition, newStatus)
	}
	if !t.Status.CanTransitionTo(newStatus) {
		return fmt.Errorf("%w: Cannot transition from '%s' to '%s'", errmsgs.ErrInvalidStatusTransition, t.Status, newStatus)
	}

	switch newStatus {
	case StatusResolved:
		t.ResolvedAt = &timestamp
		if err := t.ValidateResolvedAt(t.CreatedAt); err != nil {
			return err
		}
	case StatusCancelled:
		t.CancelledAt = &timestamp
		if err := t.ValidateCancelledAt(t.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}

func (t *Ticket) ValidateResolvedAt(createdAt time.Time) error {
	return validateTimestampAfterCreation(t.ResolvedAt, "Resolved At", createdAt)
}

func (t *Ticket) ValidateCancelledAt(createdAt time.Time) error {
	return validateTimestampAfterCreation(t.CancelledAt, "Cancelled At", createdAt)
}
