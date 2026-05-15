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

// respondWithError deduplicates error handling and hides raw internal errors
func respondWithError(c *gin.Context, err error) {
	if errors.Is(err, errmsgs.ErrTicketNotFound) {
		c.JSON(http.StatusNotFound, dto.APIResponse[interface{}]{
			Success: false,
			Error:   errmsgs.ErrTicketNotFound.Error(),
		})
		return
	}
	if errors.Is(err, errmsgs.ErrValidation) || errors.Is(err, errmsgs.ErrInvalidInput) || errors.Is(err, errmsgs.ErrInvalidStatusTransition) {
		c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	
	// Never expose raw internal errors to the client
	// In a real system, you would log the raw `err` here for debugging
	c.JSON(http.StatusInternalServerError, dto.APIResponse[interface{}]{
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

// API: POST /tickets
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
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.APIResponse[*domain.Ticket]{
		Success: true,
		Data:    ticket,
	})
}

// API: GET /tickets
func (h *TicketHandler) HandleListTickets(c *gin.Context) {
	var query struct {
		dto.TicketFilter
		dto.PaginationQuery
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
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

	c.JSON(http.StatusOK, dto.APIResponse[*dto.PaginatedResult[domain.Ticket]]{
		Success: true,
		Message: "Get tickets successfully",
		Data:    tickets,
	})
}

// API: GET /tickets/:id
func (h *TicketHandler) HandleGetTicket(c *gin.Context) {
	id, err := parseTicketID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
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

	c.JSON(http.StatusOK, dto.APIResponse[*domain.Ticket]{
		Success: true,
		Data:    ticket,
	})
}

// API: PATCH /tickets/:id/status
func (h *TicketHandler) HandleUpdateStatus(c *gin.Context) {
	id, err := parseTicketID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req dto.UpdateStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
			Success: false,
			Error:   errmsgs.ErrInvalidInput.Error() + ": " + err.Error(),
		})
		return
	}

	err = h.ticketService.UpdateTicketStatus(c.Request.Context(), uint(id), req)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse[interface{}]{
		Success: true,
		Message: "ticket status updated successfully",
	})
}
