package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"support-ticket.com/internal/dto"
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

// ImportEvents godoc
// @Summary Import ticket events
// @Description Import ticket events in batch using worker pool
// @Tags ticket-events
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Import ticket events request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /ticket-events/import [post]
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
		c.JSON(http.StatusInternalServerError, dto.APIResponse[interface{}]{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse[interface{}]{
		Success: true,
		Data:    result,
		Message: "import completed",
	})
}
