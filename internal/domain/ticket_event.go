package domain

import "time"

type TicketEvent struct {
    EventID    string
    TicketID   string
    FromStatus TicketStatus
    ToStatus   TicketStatus
	CreatedBy  string  
    CreatedAt  time.Time
}