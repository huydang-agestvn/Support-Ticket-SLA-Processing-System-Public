package router

import (
	"github.com/gin-gonic/gin"
	handler "support-ticket.com/internal/handler"
)

func InitRouter(r *gin.Engine, eventHandler *handler.TicketEventHandler, ticketHandler *handler.TicketHandler) *gin.Engine {
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api/v1")
	{
		eventGroup := api.Group("/ticket-events")
		{
			eventGroup.POST("/import", eventHandler.ImportEvents)
		}

		ticketGroup := api.Group("/tickets")
		{
			ticketGroup.POST("", ticketHandler.HandleCreateTicket)
			ticketGroup.GET("", ticketHandler.HandleListTickets)
			ticketGroup.GET("/:id", ticketHandler.HandleGetTicket)
			ticketGroup.PATCH("/:id/status", ticketHandler.HandleUpdateStatus)
		}
	}

	return r
}
