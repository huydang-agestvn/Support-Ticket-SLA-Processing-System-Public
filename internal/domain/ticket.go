package domain

import "time"

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
	ID          int          `json:"id"`
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
