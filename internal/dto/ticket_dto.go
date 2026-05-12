package dto

import "support-ticket.com/internal/domain"

type CreateTicketReq struct {
	RequestorID string          `json:"requestor_id" binding:"required"`
	Title       string          `json:"title" binding:"required,min=5,max=255"`
	Description string          `json:"description" binding:"max=5000"`
	Priority    domain.Priority `json:"priority" binding:"required"`
}

type UpdateStatusReq struct {
	ActorID    string              `json:"actor_id" binding:"required"`
	Status     domain.TicketStatus `json:"status"`
	Note       string              `json:"note"`
	AssigneeID string              `json:"assignee_id"` // optional
}
