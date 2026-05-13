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
		if errors.Is(err, errmsgs.ErrValidation) || errors.Is(err, errmsgs.ErrInvalidInput) {
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

	var paging dto.PaginationQuery
	if err := c.ShouldBindQuery(&paging); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
			Success: false,
			Error:   "invalid pagination parameters: " + err.Error(),
		})
		return
	}

	tickets, err := h.ticketService.FindAll(c.Request.Context(), filters, paging)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse[interface{}]{
			Success: false,
			Error:   errmsgs.ErrInternal.Error(),
		})
		return
	}
	if tickets == nil {
        tickets = &dto.PaginatedResult[domain.Ticket]{
            Items: []domain.Ticket{}, 
        }
    } else if tickets.Items == nil {
        tickets.Items = []domain.Ticket{} 
    }
	message := "Get tickets successfully"
    if len(tickets.Items) == 0 {
        message = "No tickets found matching the criteria"
    }
	c.JSON(http.StatusOK, dto.APIResponse[*dto.PaginatedResult[domain.Ticket]]{
		Success: true,
		Message: message,
		Data:    tickets,
	})
}

// API: GET /tickets/:id
func (h *TicketHandler) HandleGetTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
			Success: false,
			Error:   errmsgs.ErrInvalidInput.Error(),
		})
		return
	}

	ticket, err := h.ticketService.FindById(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, errmsgs.ErrTicketNotFound) {
			c.JSON(http.StatusNotFound, dto.APIResponse[interface{}]{
				Success: false,
				Error:   errmsgs.ErrTicketNotFound.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse[interface{}]{
			Success: false,
			Error:   errmsgs.ErrInternal.Error() + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse[*domain.Ticket]{
		Success: true,
		Data:    ticket,
	})
}

// API: PATCH /tickets/:id/status
func (h *TicketHandler) HandleUpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
			Success: false,
			Error:   errmsgs.ErrInvalidInput.Error(),
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
		if errors.Is(err, errmsgs.ErrTicketNotFound) {
			c.JSON(http.StatusNotFound, dto.APIResponse[interface{}]{
				Success: false,
				Error:   errmsgs.ErrTicketNotFound.Error(),
			})
			return
		}
		if errors.Is(err, errmsgs.ErrInvalidStatusTransition) || errors.Is(err, errmsgs.ErrValidation) || errors.Is(err, errmsgs.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
				Success: false,
				Error:   err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.APIResponse[interface{}]{
			Success: false,
			Error:   errmsgs.ErrInternal.Error() + ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse[interface{}]{
		Success: true,
		Message: "ticket status updated successfully",
	})
}
