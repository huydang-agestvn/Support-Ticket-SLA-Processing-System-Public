package domain

import "time"

type TicketReport struct {
    id int
    report_date time.Time
    new_count int
    resolved_count int
    cancelled_count int
    overdue_count int
    avg_resolution_time int
    high_priority_count int
    medium_priority_count int
    low_priority_count int
    created_at time.Time
}