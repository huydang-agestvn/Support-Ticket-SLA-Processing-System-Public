package router

import (
	"github.com/gin-gonic/gin"
	handler "support-ticket.com/internal/handler"
)

func InitRouter(r *gin.Engine, eventHandler *handler.TicketEventHandler) {
	api := r.Group("/api/v1")
	{
		api.POST("/ticket-events/import", eventHandler.ImportEvents)
	}
}
