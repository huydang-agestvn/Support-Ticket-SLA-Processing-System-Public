package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"support-ticket.com/internal/config"
	"support-ticket.com/internal/model"
	"support-ticket.com/internal/repository"
	"support-ticket.com/internal/service"
)

func main() {
	// Parse flag --date=2026-05-05
	dateStr := flag.String("date", time.Now().Format("2006-01-02"), "Date to generate report for (YYYY-MM-DD)")
	flag.Parse()

	date, err := time.ParseInLocation("2006-01-02", *dateStr, time.Local)
	if err != nil {
		log.Fatalf("invalid date format, expected YYYY-MM-DD: %v", err)
	}

	// Load config từ .env
	cfg := config.LoadConfig()

	// Connect DB
	db, err := cfg.GetDatabase()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// AutoMigrate report table
	if err := db.AutoMigrate(&domain.TicketReport{}); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	// Wire dependencies
	reportRepo := repository.NewReportRepository(db)
	reportSvc := service.NewReportService(reportRepo)

	// Generate report
	report, err := reportSvc.GenerateReport(date)
	if err != nil {
		log.Fatalf("failed to generate report: %v", err)
	}

	fmt.Printf("Report generated for %s\n", date.Format("2006-01-02"))
	fmt.Printf("  new:       %d\n", report.NewCount)
	fmt.Printf("  resolved:  %d\n", report.ResolvedCount)
	fmt.Printf("  cancelled: %d\n", report.CancelledCount)
	fmt.Printf("  overdue:   %d\n", report.OverdueCount)
	fmt.Printf("  sla breach: %d\n", report.SlaBreacheCount)
	fmt.Printf("  avg time:  %.2f hours\n", report.AvgResolutionTime)
	fmt.Printf("  high:      %d\n", report.HighPriorityCount)
	fmt.Printf("  medium:    %d\n", report.MediumPriorityCount)
	fmt.Printf("  low:       %d\n", report.LowPriorityCount)
}
