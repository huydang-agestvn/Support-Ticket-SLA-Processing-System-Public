package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"support-ticket.com/internal/domain"
	"support-ticket.com/internal/dto"
	"support-ticket.com/internal/errmsgs"
	"support-ticket.com/internal/service"
)

type TicketHandler struct {
	ticketService service.TicketService
}

func NewTicketHandler(s service.TicketService) *TicketHandler {
	return &TicketHandler{
		ticketService: s,
	}
}

// API: POST /tickets
func (h *TicketHandler) HandleCreateTicket(c *gin.Context) {
	var req dto.CreateTicketReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	// Validate priority
	if !req.Priority.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid priority value"})
		return
	}

	ticket, err := h.ticketService.Create(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": ticket})
}

// API: GET /tickets
func (h *TicketHandler) HandleListTickets(c *gin.Context) {
	filters := map[string]interface{}{}

	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if priority := c.Query("priority"); priority != "" {
		filters["priority"] = priority
	}
	if assigneeID := c.Query("assignee_id"); assigneeID != "" {
		filters["assignee_id"] = assigneeID
	}

	tickets, err := h.ticketService.FindAll(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if tickets == nil {
		tickets = []domain.Ticket{}
	}
	c.JSON(http.StatusOK, gin.H{"data": tickets})
}

// API: GET /tickets/:id
func (h *TicketHandler) HandleGetTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket ID format"})
		return
	}

	ticket, err := h.ticketService.FindById(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, errmsgs.ErrTicketNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ticket})
}

// API: PATCH /tickets/:id/status
func (h *TicketHandler) HandleUpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket ID format"})
		return
	}

	var req dto.UpdateStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	newStatus := domain.TicketStatus(req.Status)

	err = h.ticketService.UpdateTicketStatus(c.Request.Context(), uint(id), newStatus, req.ActorID, req.AssigneeID, req.Note)
	if err != nil {
		if errors.Is(err, errmsgs.ErrTicketNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
			return
		}
		if errors.Is(err, service.ErrEventTransitionAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domain.ErrInvalidTransition) || errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update ticket status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ticket status updated successfully",
		"status": newStatus})
}
