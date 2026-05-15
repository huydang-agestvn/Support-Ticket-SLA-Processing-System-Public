package common

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

type PaginationQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}
type PaginatedResult[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}

func (q *PaginationQuery) GetPage() int {
	if q.Page <= 0 {
		return DefaultPage
	}
	return q.Page
}

func (q *PaginationQuery) GetLimit() int {
	if q.Limit <= 0 {
		return DefaultLimit
	}
	if q.Limit > MaxLimit {
		return MaxLimit
	}
	return q.Limit
}

func (q *PaginationQuery) GetOffset() int {
	return (q.GetPage() - 1) * q.GetLimit()
}
