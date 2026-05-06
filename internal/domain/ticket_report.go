package domain

import (
	"fmt"
	"time"
)

type TicketReport struct {
	ID                  int       `json:"id"`
	ReportDate          time.Time `json:"report_date"`
	NewCount            int       `json:"new_count"`
	ResolvedCount       int       `json:"resolved_count"`
	CancelledCount      int       `json:"cancelled_count"`
	OverdueCount        int       `json:"overdue_count"`
	AvgResolutionTime   int       `json:"avg_resolution_time"`
	HighPriorityCount   int       `json:"high_priority_count"`
	MediumPriorityCount int       `json:"medium_priority_count"`
	LowPriorityCount    int       `json:"low_priority_count"`
	CreatedAt           time.Time `json:"created_at"`
}

func (r *TicketReport) Validate() error {
	if r.ReportDate.IsZero() {
		return fmt.Errorf("%w: report date is required", ErrValidation)
	}
	if r.NewCount < 0 || r.ResolvedCount < 0 || r.CancelledCount < 0 {
		return fmt.Errorf("%w: status counts cannot be negative", ErrValidation)
	}
	if r.OverdueCount < 0 {
		return fmt.Errorf("%w: overdue count cannot be negative", ErrValidation)
	}
	if r.AvgResolutionTime < 0 {
		return fmt.Errorf("%w: average resolution time cannot be negative", ErrValidation)
	}
	if r.HighPriorityCount < 0 || r.MediumPriorityCount < 0 || r.LowPriorityCount < 0 {
		return fmt.Errorf("%w: priority counts cannot be negative", ErrValidation)
	}
	return nil
}
