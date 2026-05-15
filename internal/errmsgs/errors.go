package errmsgs

import "errors"

// Common errors
var (
	ErrNotFound                = errors.New("resource not found")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrInvalidInput            = errors.New("invalid input")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrConflict                = errors.New("conflict")
	ErrInternal                = errors.New("internal server error")
	ErrValidation              = errors.New("ticket validation failed")
	ErrTicketNotFound          = errors.New("ticket not found")
	ErrEmptyBody               = errors.New("request body is empty")
	ErrEmptyBatch              = errors.New("batch is empty")
	ErrBatchTooLarge           = errors.New("batch size exceeds maximum allowed")
	ErrInvalidFlowTicket       = errors.New("invalid flow ticket")
)
