package dto

import "support-ticket.com/internal/domain"

type CreateTicketReq struct {
	RequestorID string          `json:"requestor_id" binding:"required"`
	Title       string          `json:"title" binding:"required,min=5,max=255"`
	Description string          `json:"description" binding:"max=5000"`
	Priority    domain.Priority `json:"priority" binding:"required"`
}

type UpdateStatusReq struct {
	AssigneeID string              `json:"assignee_id"`
	Status     domain.TicketStatus `json:"status"`
	Note       string              `json:"note"`
}

type TicketFilter struct {
	Status     string `form:"status" binding:"omitempty,oneof=new assigned in_progress resolved closed canceled"`
	Priority   string `form:"priority" binding:"omitempty,oneof=low medium high"`
	AssigneeID string `form:"assignee_id" binding:"omitempty,min=1"`
}
