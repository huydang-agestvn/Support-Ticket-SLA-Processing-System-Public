package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"support-ticket.com/internal/dto"
	"support-ticket.com/internal/errmsgs"
	"support-ticket.com/internal/service"
)

type TicketEventHandler struct {
	service service.TicketEventService
}

func NewTicketEventHandler(service service.TicketEventService) *TicketEventHandler {
	return &TicketEventHandler{
		service: service,
	}
}

func (h *TicketEventHandler) ImportEvents(c *gin.Context) {
	ctx := c.Request.Context()
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
			Success: false,
			Error:   "invalid input",
		})
		return
	}
	defer c.Request.Body.Close()
	
	result, err := h.service.Import(ctx, data)
	if err != nil {
		if errors.Is(err, errmsgs.ErrEmptyBatch) || errors.Is(err, errmsgs.ErrBatchTooLarge) || errors.Is(err, errmsgs.ErrEmptyBody) {
			c.JSON(http.StatusBadRequest, dto.APIResponse[interface{}]{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		
		// Hide internal errors from client
		c.JSON(http.StatusInternalServerError, dto.APIResponse[interface{}]{
			Success: false,
			Error:   errmsgs.ErrInternal.Error(),
		})
		return
	}

	response := dto.NewTicketImportResponse(result)
	c.JSON(http.StatusOK, dto.APIResponse[interface{}]{
		Success: true,
		Data:    response,
		Message: "import completed",
	})
}
