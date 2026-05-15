package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"support-ticket.com/internal/auth"
	"support-ticket.com/internal/handler"
	"support-ticket.com/internal/middleware"

	_ "support-ticket.com/docs"
)

func InitRouter(
	r *gin.Engine,
	authHandler *handler.AuthHandler,
	eventHandler *handler.TicketEventHandler,
	ticketHandler *handler.TicketHandler,
	authMiddleware *middleware.AuthMiddleware,
	reportHandler *handler.ReportHandler,
) *gin.Engine {
	r.Use(gin.Logger(), gin.Recovery())

	// Swagger API documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", authHandler.Login)
		}

		// Agent: import event
		eventGroup := api.Group("/ticket-events")
		{
			eventGroup.POST(
				"/import",
				authMiddleware.RequireAuth(),
				authMiddleware.RequireRole(auth.RoleAgent),
				eventHandler.ImportEvents,
			)
		}

		ticketGroup := api.Group("/tickets")
		{
			// Requestor
			ticketGroup.POST(
				"",
				authMiddleware.RequireAuth(),
				authMiddleware.RequireRole(auth.RoleRequestor),
				ticketHandler.HandleCreateTicket,
			)

			// Requestor / Agent / Manager
			ticketGroup.GET(
				"",
				authMiddleware.RequireAuth(),
				authMiddleware.RequireRole(
					auth.RoleRequestor,
					auth.RoleAgent,
					auth.RoleManager,
				),
				ticketHandler.HandleListTickets,
			)

			// Requestor / Agent / Manager
			ticketGroup.GET(
				"/:id",
				authMiddleware.RequireAuth(),
				authMiddleware.RequireRole(
					auth.RoleRequestor,
					auth.RoleAgent,
					auth.RoleManager,
				),
				ticketHandler.HandleGetTicket,
			)

			// Agent: update status
			ticketGroup.PATCH(
				"/:id/status",
				authMiddleware.RequireAuth(),
				authMiddleware.RequireRole(auth.RoleAgent),
				ticketHandler.HandleUpdateStatus,
			)
		}

		reportGroup := api.Group("/reports")
		{
			reportGroup.GET(
				"/daily",
				authMiddleware.RequireAuth(),
				authMiddleware.RequireRole(auth.RoleManager),
				reportHandler.GetDaily,
			)
		}
	}
	return r
}
