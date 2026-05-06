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
	id int  
    assignee_id  string // Keycloak user ID
    requestor_id   string     // Keycloak user ID
    title string
    description string
    priority    Priority
    status      TicketStatus
    deadline    time.Time
	created_at   time.Time
	update_at   time.Time
	resolved_at  time.Time
    sla_due_at    time.Time
	cancelled_at time.Time
}