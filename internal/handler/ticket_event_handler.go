package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"support-ticket.com/internal/service"
	"support-ticket.com/internal/errors"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": errors.ErrInvalidInput})
		return
	}
	defer c.Request.Body.Close()
	result, err := h.service.Import(ctx, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": errors.ErrInvalidInput,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "import completed",
		"data":    result,
	})
}
