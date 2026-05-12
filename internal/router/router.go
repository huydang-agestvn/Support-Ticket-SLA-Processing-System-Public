package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"support-ticket.com/internal/handler"

	_ "support-ticket.com/docs"
)

func InitRouter(r *gin.Engine, eventHandler *handler.TicketEventHandler, 
	ticketHandler *handler.TicketHandler, reportHandler *handler.ReportHandler) *gin.Engine {
	r.Use(gin.Logger(), gin.Recovery())

	// Swagger API documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
		reportGroup := api.Group("/reports")
		{
			reportGroup.GET("/daily", reportHandler.GetDaily)
		}
	}

	return r
}
