package service

import (
	"fmt"
	"time"

	"support-ticket.com/internal/domain"
	"support-ticket.com/internal/repository"
)

type ReportService interface {
	GenerateReport(date time.Time) (*domain.TicketReport, error)
	GetReport(date time.Time) (*domain.TicketReport, error)
}

type reportService struct {
	repo repository.ReportRepository
}

func NewReportService(repo repository.ReportRepository) ReportService {
	return &reportService{repo: repo}
}

// GenerateReport chạy aggregate từ tickets, ghi vào daily_ticket_reports
func (s *reportService) GenerateReport(date time.Time) (*domain.TicketReport, error) {
	report, err := s.repo.AggregateByDate(date)
	if err != nil {
		return nil, fmt.Errorf("aggregate report: %w", err)
	}

	if err := s.repo.Upsert(report); err != nil {
		return nil, fmt.Errorf("save report: %w", err)
	}

	return report, nil
}

// GetReport đọc report đã được generate từ daily_ticket_reports
func (s *reportService) GetReport(date time.Time) (*domain.TicketReport, error) {
	report, err := s.repo.GetByDate(date)
	if err != nil {
		return nil, err
	}
	return report, nil
}
