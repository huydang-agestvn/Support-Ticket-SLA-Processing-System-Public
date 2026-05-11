package migrations

import (
	"fmt"

	"gorm.io/gorm"
	"support-ticket.com/internal/domain"
)

// RunMigrations 
func RunMigrations(db *gorm.DB) error {
	fmt.Println("Running database migrations...")

	if err := db.AutoMigrate(
		&domain.Ticket{},
		&domain.TicketEvent{},
		&domain.TicketReport{},
	); err != nil {
		return fmt.Errorf("failed to run auto migrations: %w", err)
	}

	fmt.Println("✓ Database migrations completed successfully")
	return nil
}
