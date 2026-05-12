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

// HandleCreateTicket godoc
// @Summary Create ticket
// @Description Create a new support ticket
// @Tags tickets
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Create ticket request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /tickets [post]
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
		if errors.Is(err, errmsgs.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": ticket})
}

// HandleListTickets godoc
// @Summary List tickets
// @Description Get list of support tickets
// @Tags tickets
// @Produce json
// @Param status query string false "Ticket status"
// @Param priority query string false "Ticket priority"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /tickets [get]
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

// HandleGetTicket godoc
// @Summary Get ticket detail
// @Description Get ticket detail by ticket ID
// @Tags tickets
// @Produce json
// @Param id path int true "Ticket ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /tickets/{id} [get]
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

// HandleUpdateStatus godoc
// @Summary Update ticket status
// @Description Update ticket status by ticket ID
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path int true "Ticket ID"
// @Param request body map[string]interface{} true "Update status request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /tickets/{id}/status [patch]
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

	err = h.ticketService.UpdateTicketStatus(c.Request.Context(), uint(id), req)
	if err != nil {
		if errors.Is(err, errmsgs.ErrTicketNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
			return
		}
		if errors.Is(err, errmsgs.ErrInvalidStatusTransition) || errors.Is(err, errmsgs.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ticket status updated successfully",
		"status": req.Status})
}
