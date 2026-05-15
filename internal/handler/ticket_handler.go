package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"support-ticket.com/internal/auth"
	"support-ticket.com/internal/dto/common"
	"support-ticket.com/internal/dto/request"
	"support-ticket.com/internal/errmsgs"
	domain "support-ticket.com/internal/model"
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

// respondWithError deduplicates error handling and hides raw internal errors
func respondWithError(c *gin.Context, err error) {
	if errors.Is(err, errmsgs.ErrTicketNotFound) {
		c.JSON(http.StatusNotFound, common.APIResponse[interface{}]{
			Success: false,
			Error:   errmsgs.ErrTicketNotFound.Error(),
		})
		return
	}
	if errors.Is(err, errmsgs.ErrValidation) || errors.Is(err, errmsgs.ErrInvalidInput) || errors.Is(err, errmsgs.ErrInvalidStatusTransition) {
		c.JSON(http.StatusBadRequest, common.APIResponse[interface{}]{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Never expose raw internal errors to the client
	// In a real system, you would log the raw `err` here for debugging
	c.JSON(http.StatusInternalServerError, common.APIResponse[interface{}]{
		Success: false,
		Error:   errmsgs.ErrInternal.Error(),
	})
}

func parseTicketID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, errmsgs.ErrInvalidInput
	}
	return uint(id), nil
}

// HandleCreateTicket godoc
// @Summary Create ticket
// @Description Create a new support ticket
// @Tags Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateTicketReq true "Create ticket request"
// @Success 201 {object} map[string]interface{} "Ticket created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or invalid priority"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /tickets [post]
func (h *TicketHandler) HandleCreateTicket(c *gin.Context) {
	var req request.CreateTicketReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.APIResponse[interface{}]{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}

	// Validate priority
	if !req.Priority.IsValid() {
		c.JSON(http.StatusBadRequest, common.APIResponse[interface{}]{
			Success: false,
			Error:   "invalid priority value",
		})
		return
	}

	currentUser := auth.UserFromContext(c.Request.Context())

	req.RequestorID = currentUser.UserID

	ticket, err := h.ticketService.Create(c.Request.Context(), req)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, common.APIResponse[*domain.Ticket]{
		Success: true,
		Data:    ticket,
	})
}

// HandleListTickets godoc
// @Summary List tickets
// @Description Get tickets with filters and pagination
// @Tags Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by ticket status"
// @Param priority query string false "Filter by priority"
// @Param requestor_id query int false "Filter by requestor ID"
// @Param assignee_id query int false "Filter by assignee ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} map[string]interface{} "Get tickets successfully"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /tickets [get]
func (h *TicketHandler) HandleListTickets(c *gin.Context) {
	var query struct {
		request.TicketFilter
		common.PaginationQuery
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, common.APIResponse[interface{}]{
			Success: false,
			Error:   "invalid query parameters: " + err.Error(),
		})
		return
	}

	tickets, err := h.ticketService.FindAll(c.Request.Context(), query.TicketFilter, query.PaginationQuery)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, common.APIResponse[*common.PaginatedResult[domain.Ticket]]{
		Success: true,
		Message: "Get tickets successfully",
		Data:    tickets,
	})
}

// HandleGetTicket godoc
// @Summary Get ticket detail
// @Description Get ticket detail by ID
// @Tags Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Success 200 {object} map[string]interface{} "Get ticket successfully"
// @Failure 400 {object} map[string]interface{} "Invalid ticket ID"
// @Failure 404 {object} map[string]interface{} "Ticket not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /tickets/{id} [get]
func (h *TicketHandler) HandleGetTicket(c *gin.Context) {
	id, err := parseTicketID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.APIResponse[interface{}]{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ticket, err := h.ticketService.FindById(c.Request.Context(), id)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, common.APIResponse[*domain.Ticket]{
		Success: true,
		Data:    ticket,
	})
}

// HandleUpdateStatus godoc
// @Summary Update ticket status
// @Description Update status of a ticket by ID
// @Tags Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body request.UpdateStatusReq true "Update status request"
// @Success 200 {object} map[string]interface{} "Ticket status updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or invalid status transition"
// @Failure 404 {object} map[string]interface{} "Ticket not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /tickets/{id}/status [patch]
func (h *TicketHandler) HandleUpdateStatus(c *gin.Context) {
	id, err := parseTicketID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.APIResponse[interface{}]{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req request.UpdateStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.APIResponse[interface{}]{
			Success: false,
			Error:   errmsgs.ErrInvalidInput.Error() + ": " + err.Error(),
		})
		return
	}

	currentUser := auth.UserFromContext(c.Request.Context())

	req.AssigneeID = currentUser.UserID

	err = h.ticketService.UpdateTicketStatus(c.Request.Context(), id, req)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, common.APIResponse[interface{}]{
		Success: true,
		Message: "ticket status updated successfully",
	})
}
