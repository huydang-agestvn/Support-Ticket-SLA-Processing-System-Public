package response

import (
	"time"

	"support-ticket.com/internal/model"
)

type TicketEventImportResponse struct {
	TicketID   uint                `json:"ticket_id"`
	FromStatus domain.TicketStatus `json:"from_status"`
	ToStatus   domain.TicketStatus `json:"to_status"`
	AssigneeID string              `json:"assignee_id"`
	Note       *string             `json:"note,omitempty"`
	CreatedAt  time.Time           `json:"created_at"`
}

type TicketEventRejectedDetail struct {
	ErrorName string                      `json:"error_name"`
	Events    []TicketEventImportResponse `json:"events"`
}

type TicketImportResponse struct {
	AcceptedCount   int                         `json:"accepted_count"`
	RejectedCount   int                         `json:"rejected_count"`
	DuplicateCount  int                         `json:"duplicate_count"`
	RejectedDetails []TicketEventRejectedDetail `json:"rejected_details"`
}

func NewTicketImportResponse(result domain.BatchImportResult) TicketImportResponse {
	details := make([]TicketEventRejectedDetail, 0, len(result.RejectedDetails))
	for _, rejected := range result.RejectedDetails {
		events := make([]TicketEventImportResponse, 0, len(rejected.Events))
		for _, event := range rejected.Events {
			events = append(events, TicketEventImportResponse{
				TicketID:   event.TicketID,
				FromStatus: event.FromStatus,
				ToStatus:   event.ToStatus,
				AssigneeID: event.AssigneeID,
				Note:       event.Note,
				CreatedAt:  event.CreatedAt,
			})
		}
		details = append(details, TicketEventRejectedDetail{
			ErrorName: rejected.ErrorName,
			Events:    events,
		})
	}

	return TicketImportResponse{
		AcceptedCount:   result.AcceptedCount,
		RejectedCount:   result.RejectedCount,
		DuplicateCount:  result.DuplicateCount,
		RejectedDetails: details,
	}
}
