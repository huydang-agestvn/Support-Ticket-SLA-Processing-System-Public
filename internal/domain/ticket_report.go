package domain

import (
	"fmt"
	"time"

	"support-ticket.com/internal/errmsgs"
)

type TicketReport struct {
	ID                  uint      `json:"id" gorm:"primaryKey"`
	ReportDate          time.Time `json:"report_date" gorm:"column:report_date;type:date;not null;uniqueIndex"`
	NewCount            int       `json:"new_count" gorm:"column:new_count;not null;default:0"`
	ResolvedCount       int       `json:"resolved_count" gorm:"column:resolved_count;not null;default:0"`
	CancelledCount      int       `json:"cancelled_count" gorm:"column:cancelled_count;not null;default:0"`
	OverdueCount        int       `json:"overdue_count" gorm:"column:overdue_count;not null;default:0"`
	AvgResolutionTime   int       `json:"avg_resolution_time" gorm:"column:avg_resolution_time;not null;default:0"`
	HighPriorityCount   int       `json:"high_priority_count" gorm:"column:high_priority_count;not null;default:0"`
	MediumPriorityCount int       `json:"medium_priority_count" gorm:"column:medium_priority_count;not null;default:0"`
	LowPriorityCount    int       `json:"low_priority_count" gorm:"column:low_priority_count;not null;default:0"`
	CreatedAt           time.Time `json:"created_at" gorm:"column:created_at;not null;autoCreateTime:milli"`
}

func (r *TicketReport) Validate() error {
	if r.ReportDate.IsZero() {
		return fmt.Errorf("%w: Report date is required", errmsgs.ErrInvalidInput)
	}
	if r.NewCount < 0 || r.ResolvedCount < 0 || r.CancelledCount < 0 {
		return fmt.Errorf("%w: Status counts cannot be negative", errmsgs.ErrInvalidInput)
	}
	if r.OverdueCount < 0 {
		return fmt.Errorf("%w: Overdue count cannot be negative", errmsgs.ErrInvalidInput)
	}
	if r.AvgResolutionTime < 0 {
		return fmt.Errorf("%w: Average resolution time cannot be negative", errmsgs.ErrInvalidInput)
	}
	if r.HighPriorityCount < 0 || r.MediumPriorityCount < 0 || r.LowPriorityCount < 0 {
		return fmt.Errorf("%w: Priority counts cannot be negative", errmsgs.ErrInvalidInput)
	}
	return nil
}
