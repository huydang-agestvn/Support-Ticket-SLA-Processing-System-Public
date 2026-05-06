package domain

import "time"

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
