package request

import (
	"time"

	domain "support-ticket.com/internal/model"
)

type CreateTicketReq struct {
	RequestorID string          `json:"-" swaggerignore:"true"`
	Title       string          `json:"title" binding:"required,min=5,max=255"`
	Description string          `json:"description" binding:"max=5000"`
	Priority    domain.Priority `json:"priority" binding:"required"`
	SlaDueAt    *time.Time      `json:"sla_due_at,omitempty" binding:"required"`
}

type UpdateStatusReq struct {
	Status     domain.TicketStatus `json:"status" binding:"required"`
	Note       string              `json:"note,omitempty"`
	AssigneeID string              `json:"-" swaggerignore:"true"`
}

type TicketFilter struct {
	Status     string `form:"status" binding:"omitempty,oneof=new assigned in_progress resolved closed canceled"`
	Priority   string `form:"priority" binding:"omitempty,oneof=low medium high"`
	AssigneeID string `form:"assignee_id" binding:"omitempty,min=1"`
}
