package domain

import "time"

type TicketEvent struct {
	EventID    string       `json:"event_id"`
	TicketID   string       `json:"ticket_id"`
	FromStatus TicketStatus `json:"from_status"`
	ToStatus   TicketStatus `json:"to_status"`
	CreatedBy  string       `json:"created_by"`
	CreatedAt  time.Time    `json:"created_at"`
}
