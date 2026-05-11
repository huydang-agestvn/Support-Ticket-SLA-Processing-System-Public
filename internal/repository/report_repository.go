package repository

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"support-ticket.com/internal/domain"
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

// AggregateByDate scan bảng tickets ĐÚNG 1 LẦN để tính toán mọi chỉ số
func (r *reportRepository) AggregateByDate(date time.Time) (*domain.TicketReport, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)
	now := time.Now()

	report := &domain.TicketReport{
		ReportDate: start,
		CreatedAt:  now,
	}

	// SQL gom 7 queries thành 1.
	// WHERE ở cuối giúp giới hạn số dòng DB phải quét, tránh Full Table Scan.
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE created_at >= @start AND created_at < @end) AS new_count,
			COUNT(*) FILTER (WHERE resolved_at >= @start AND resolved_at < @end) AS resolved_count,
			COUNT(*) FILTER (WHERE cancelled_at >= @start AND cancelled_at < @end) AS cancelled_count,
			COUNT(*) FILTER (WHERE sla_due_at < @now AND status NOT IN ('resolved', 'closed', 'cancelled')) AS overdue_count,
			COALESCE(AVG(EXTRACT(EPOCH FROM (resolved_at - created_at)) / 3600) FILTER (WHERE resolved_at >= @start AND resolved_at < @end), 0) AS avg_resolution_time,
			COUNT(*) FILTER (WHERE created_at >= @start AND created_at < @end AND priority = 'high') AS high_priority_count,
			COUNT(*) FILTER (WHERE created_at >= @start AND created_at < @end AND priority = 'medium') AS medium_priority_count,
			COUNT(*) FILTER (WHERE created_at >= @start AND created_at < @end AND priority = 'low') AS low_priority_count
		FROM tickets 
		WHERE (created_at >= @start AND created_at < @end)
		   OR (resolved_at >= @start AND resolved_at < @end)
		   OR (cancelled_at >= @start AND cancelled_at < @end)
		   OR (sla_due_at < @now AND status NOT IN ('resolved', 'closed', 'cancelled'))
	`

	// Dùng sql.Named để truyền biến an toàn và rõ ràng vào raw query
	err := r.db.Raw(query,
		sql.Named("start", start),
		sql.Named("end", end),
		sql.Named("now", now),
	).Scan(report).Error

	if err != nil {
		return nil, fmt.Errorf("aggregate daily report failed: %w", err)
	}

	return report, nil
}

// Upsert sử dụng cơ chế ON CONFLICT của PostgreSQL
func (r *reportRepository) Upsert(report *domain.TicketReport) error {
	// LƯU Ý: Bạn CẦN đảm bảo bảng daily_ticket_reports có UNIQUE INDEX cho cột report_date
	// Ví dụ trong file migrate: db.Exec("CREATE UNIQUE INDEX idx_report_date ON daily_ticket_reports(report_date);")

	err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "report_date"}}, // Cột bắt trùng lặp
		UpdateAll: true,                                   // Nếu trùng ngày, cập nhật toàn bộ các chỉ số mới
	}).Create(report).Error

	if err != nil {
		return fmt.Errorf("upsert daily_ticket_reports: %w", err)
	}
	return nil
}

// GetByDate giữ nguyên logic
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