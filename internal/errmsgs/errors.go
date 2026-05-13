package errmsgs

import "errors"

// Common errors
var (
	ErrNotFound                = errors.New("Resource not found")
	ErrInvalidStatusTransition = errors.New("Invalid status transition")
	ErrInvalidInput            = errors.New("Invalid input")
	ErrUnauthorized            = errors.New("Unauthorized")
	ErrConflict                = errors.New("Conflict")
	ErrInternal                = errors.New("Internal server error")
	ErrValidation              = errors.New("Ticket validation failed")
	ErrTicketNotFound          = errors.New("Ticket not found")
	ErrEmptyBody               = errors.New("Request body is empty")
	ErrEmptyBatch              = errors.New("Batch is empty")
	ErrBatchTooLarge		   = errors.New("Batch size exceeds maximum allowed")
)
