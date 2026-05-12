package dto

import "support-ticket.com/internal/domain"

type CreateTicketReq struct {
	RequestorID string          `json:"requestor_id" binding:"required"`
	Title       string          `json:"title" binding:"required"`
	Description string          `json:"description" binding:"required"`
	Priority    domain.Priority `json:"priority" binding:"required"`
}

type UpdateStatusReq struct {
	Status string `json:"status" binding:"required"`
}
