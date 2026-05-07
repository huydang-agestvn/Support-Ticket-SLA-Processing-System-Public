package service

import (
	"sync"
	"fmt"
	"support-ticket.com/internal/domain"
)

type EventResult string

const (
	ResultAccepted  EventResult = "accepted"
	ResultRejected  EventResult = "rejected"
	ResultDuplicate EventResult = "duplicate"
)

type TicketService struct {
	seenMu     sync.Mutex
	seenEvents map[string]bool
}

func NewTicketService() *TicketService {
	return &TicketService{
		seenEvents: make(map[string]bool),
	}
}


func (s *TicketService) ProcessEvent(event *domain.TicketEvent) (EventResult, error) {
    // 1. Check duplicate
    key := fmt.Sprintf("%d|%s|%s", event.TicketID, event.FromStatus, event.ToStatus)

    s.seenMu.Lock()
    if s.seenEvents[key] {
        s.seenMu.Unlock()
        return ResultDuplicate, nil
    }
    s.seenEvents[key] = true
    s.seenMu.Unlock()

    // 2. Validate transition
    if err := event.Validate(); err != nil {
        return ResultRejected, err
    }

    return ResultAccepted, nil
}