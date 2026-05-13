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
// @Summary Create a new ticket
// @Description Create a new support ticket with title, description, requestor, assignee and priority.
// @Tags Tickets
// @Accept json
// @Produce json
// @Param request body dto.CreateTicketReq true "Create ticket request"
// @Success 201 {object} map[string]interface{} "Ticket created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or invalid priority"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /tickets [post]
func (h *TicketHandler) HandleCreateTicket(c *gin.Context) {
	var req dto.CreateTicketReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}

	// Validate priority
	if !req.Priority.IsValid() {
		c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
			Success: false,
			Error:   "invalid priority value",
		})
		return
	}

	ticket, err := h.ticketService.Create(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, errmsgs.ErrValidation) {
			c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse[interface{}]{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse[*domain.Ticket]{
		Success: true,
		Data:    ticket,
	})
}

// HandleListTickets godoc
// @Summary List tickets
// @Description Get a list of support tickets with optional filters by status, priority and assignee ID.
// @Tags Tickets
// @Produce json
// @Param status query string false "Filter by ticket status" Enums(new, assigned, in_progress, resolved, closed, cancelled)
// @Param priority query string false "Filter by ticket priority" Enums(low, medium, high, urgent)
// @Param assignee_id query string false "Filter by assignee ID"
// @Success 200 {object} map[string]interface{} "List tickets successfully"
// @Failure 500 {object} map[string]interface{} "Internal server error"
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
		c.JSON(http.StatusInternalServerError, dto.APIResponse[interface{}]{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if tickets == nil {
		tickets = []domain.Ticket{}
	}
	c.JSON(http.StatusOK, dto.APIResponse[[]domain.Ticket]{
		Success: true,
		Data:    tickets,
	})
}

// HandleGetTicket godoc
// @Summary Get ticket detail
// @Description Get support ticket detail by ticket ID.
// @Tags Tickets
// @Produce json
// @Param id path int true "Ticket ID"
// @Success 200 {object} map[string]interface{} "Get ticket detail successfully"
// @Failure 400 {object} map[string]interface{} "Invalid ticket ID format"
// @Failure 404 {object} map[string]interface{} "Ticket not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /tickets/{id} [get]
func (h *TicketHandler) HandleGetTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
			Success: false,
			Error:   "invalid ticket ID format",
		})
		return
	}

	ticket, err := h.ticketService.FindById(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, errmsgs.ErrTicketNotFound) {
			c.JSON(http.StatusNotFound, dto.APIResponse[interface{}]{
				Success: false,
				Error:   "ticket not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse[interface{}]{
			Success: false,
			Error:   "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse[*domain.Ticket]{
		Success: true,
		Data:    ticket,
	})
}

// HandleUpdateStatus godoc
// @Summary Update ticket status
// @Description Update support ticket status by ticket ID. The status transition must follow the required ticket status flow.
// @Tags Tickets
// @Accept json
// @Produce json
// @Param id path int true "Ticket ID"
// @Param request body dto.UpdateStatusReq true "Update ticket status request"
// @Success 200 {object} map[string]interface{} "Ticket status updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body, invalid ticket ID or invalid status transition"
// @Failure 404 {object} map[string]interface{} "Ticket not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /tickets/{id}/status [patch]
func (h *TicketHandler) HandleUpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
			Success: false,
			Error:   "invalid ticket ID format",
		})
		return
	}

	var req dto.UpdateStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}

	err = h.ticketService.UpdateTicketStatus(c.Request.Context(), uint(id), req)
	if err != nil {
		if errors.Is(err, errmsgs.ErrTicketNotFound) {
			c.JSON(http.StatusNotFound, dto.APIResponse[interface{}]{
				Success: false,
				Error:   "ticket not found",
			})
			return
		}
		if errors.Is(err, errmsgs.ErrInvalidStatusTransition) || errors.Is(err, errmsgs.ErrValidation) {
			c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.APIResponse[interface{}]{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse[interface{}]{
		Success: true,
		Message: "ticket status updated successfully",
	})
}
