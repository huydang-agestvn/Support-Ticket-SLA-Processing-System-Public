package app

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"support-ticket.com/internal/config"
	"support-ticket.com/internal/handler"
	"support-ticket.com/internal/migration"
	"support-ticket.com/internal/repository"
	"support-ticket.com/internal/router"
	"support-ticket.com/internal/service"
)

type App struct {
	cfg    *config.Config
	db     *gorm.DB
	router *gin.Engine
}

func NewApp() *App {
	return &App{}
}

func (a *App) Run() error {
	// 1. Load Configuration
	a.cfg = config.LoadConfig()

	// 2. Initialize Database
	if err := a.initDB(); err != nil {
		return err
	}

	sqlDB, err := a.db.DB()
	if err == nil {
		defer func() {
			if closeErr := sqlDB.Close(); closeErr != nil {
				log.Printf("warning: failed to close database connection: %v", closeErr)
			}
		}()
	}

	// 3. Run Migrations
	if err := migration.RunMigrations(a.db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// 4. Setup Dependency Injection
	a.setupDependencies()

	// 5. Start HTTP Server
	return a.startServer()
}

func (a *App) initDB() error {
	var err error
	a.db, err = a.cfg.GetDatabase()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	log.Printf("database connected: %s:%d/%s", a.cfg.DBHost, a.cfg.DBPort, a.cfg.DBName)
	return nil
}

func (a *App) setupDependencies() {
	ticketRepo := repository.NewTicketRepository(a.db)
	eventRepo := repository.NewTicketEventRepository(a.db)
	reportRepo := repository.NewReportRepository(a.db)

	ticketService := service.NewTicketService(ticketRepo, eventRepo)
	eventService := service.NewTicketEventService(eventRepo, ticketRepo)
	reportService := service.NewReportService(reportRepo)

	ticketHandler := handler.NewTicketHandler(ticketService)
	eventHandler := handler.NewTicketEventHandler(eventService)
	reportHandler := handler.NewReportHandler(reportService)

	r := gin.New()
	a.router = router.InitRouter(r, eventHandler, ticketHandler, reportHandler)
}

func (a *App) startServer() error {
	serverPort := a.cfg.ServerPort
	if serverPort == 0 {
		serverPort = 8080
	}
	addr := fmt.Sprintf(":%d", serverPort)

	log.Printf("worker pool size: %d", a.cfg.WorkerPoolSize)
	log.Printf("starting HTTP server on %s", addr)

	// Khởi chạy server (blocking operation)
	return a.router.Run(addr)
}
