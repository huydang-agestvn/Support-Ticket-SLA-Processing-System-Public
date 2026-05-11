package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"support-ticket.com/internal/config"
	"support-ticket.com/internal/handler"
	"support-ticket.com/internal/migrations"
	"support-ticket.com/internal/repository"
	"support-ticket.com/internal/router"
	"support-ticket.com/internal/service"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	fmt.Println("Initializing database connection...")
	db, err := cfg.GetDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Log configuration (without password)
	fmt.Printf("✓ Database connected: %s:%d/%s\n", cfg.DBHost, cfg.DBPort, cfg.DBName)
	fmt.Printf("✓ Server Port: %d\n", cfg.ServerPort)
	fmt.Printf("✓ Worker Pool Size: %d\n", cfg.WorkerPoolSize)

	// Log database info
	sqlDB, err := db.DB()
	if err == nil {
		fmt.Printf("✓ Database connection pool opened\n")
		defer sqlDB.Close()
	}

	// Run migrations
	if err := migrations.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	eventRepo := repository.NewTicketEventRepository(db)

	// Initialize services
	eventService := service.NewTicketEventService(eventRepo)

	// Initialize handlers
	eventHandler := handler.NewTicketEventHandler(eventService)

	// Setup routes
	r := gin.Default()

	router.InitRouter(r, eventHandler)

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf(" Server starting on port %s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
