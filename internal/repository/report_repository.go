package repository

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"support-ticket.com/internal/model"
)

type ReportRepository interface {
	AggregateByDate(date time.Time) (*domain.TicketReport, error)
	Upsert(report *domain.TicketReport) error
	GetByDate(date time.Time) (*domain.TicketReport, error)
}

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) AggregateByDate(date time.Time) (*domain.TicketReport, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	report := &domain.TicketReport{
		ReportDate: start,
		CreatedAt:  time.Now(),
	}

	query := `
		SELECT 
			COUNT(*) FILTER (WHERE created_at >= @start AND created_at < @end) AS new_count,
			COUNT(*) FILTER (WHERE resolved_at >= @start AND resolved_at < @end) AS resolved_count,
			COUNT(*) FILTER (WHERE cancelled_at >= @start AND cancelled_at < @end) AS cancelled_count,
			COUNT(*) FILTER (WHERE sla_due_at < @end AND status NOT IN ('resolved', 'closed', 'cancelled')) AS overdue_count,
			COUNT(*) FILTER (WHERE resolved_at >= @start AND resolved_at < @end AND resolved_at > sla_due_at) AS sla_breache_count,
			COALESCE(AVG(EXTRACT(EPOCH FROM (resolved_at - created_at)) / 3600) FILTER (WHERE resolved_at >= @start AND resolved_at < @end), 0) AS avg_resolution_time,
			COUNT(*) FILTER (WHERE created_at >= @start AND created_at < @end AND priority = 'high') AS high_priority_count,
			COUNT(*) FILTER (WHERE created_at >= @start AND created_at < @end AND priority = 'medium') AS medium_priority_count,
			COUNT(*) FILTER (WHERE created_at >= @start AND created_at < @end AND priority = 'low') AS low_priority_count
		FROM tickets 
		WHERE (created_at >= @start AND created_at < @end)
		   OR (resolved_at >= @start AND resolved_at < @end)
		   OR (cancelled_at >= @start AND cancelled_at < @end)
	   OR (sla_due_at < @end AND status NOT IN ('resolved', 'closed', 'cancelled'))
	   OR (sla_due_at >= @start AND sla_due_at < @end AND status NOT IN ('resolved', 'closed', 'cancelled'))
	`
	err := r.db.Raw(query,
		sql.Named("start", start),
		sql.Named("end", end),
	).Scan(report).Error

	if err != nil {
		return nil, fmt.Errorf("aggregate daily report failed: %w", err)
	}

	return report, nil
}

func (r *reportRepository) Upsert(report *domain.TicketReport) error {
	err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "report_date"}}, // Cột bắt trùng lặp
		UpdateAll: true,                                   // Nếu trùng ngày, cập nhật toàn bộ các chỉ số mới
	}).Create(report).Error

	if err != nil {
		return fmt.Errorf("upsert daily_ticket_reports: %w", err)
	}
	return nil
}

func (r *reportRepository) GetByDate(date time.Time) (*domain.TicketReport, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	var report domain.TicketReport
	result := r.db.Where("report_date = ?", start).First(&report)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("report not found for date %s", start.Format("2006-01-02"))
	}
	if result.Error != nil {
		return nil, fmt.Errorf("get report by date: %w", result.Error)
	}
	return &report, nil
}
