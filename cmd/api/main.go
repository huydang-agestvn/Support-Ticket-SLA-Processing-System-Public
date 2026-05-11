package main

import (
	"fmt"
	"log"

	"support-ticket.com/internal/config"
	"support-ticket.com/internal/migrations"
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

	// TODO: Setup HTTP server
	// TODO: Setup routes
	// TODO: Setup worker pool
	// TODO: Setup middleware (auth, logging, etc)

	log.Println("API server is ready. TODO: Start listening on port", cfg.ServerPort)
	select {} // Keep running
}
