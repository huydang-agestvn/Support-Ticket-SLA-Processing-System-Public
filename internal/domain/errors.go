package domain

import "errors"

var (
    ErrTicketNotFound     = errors.New("ticket not found")
    ErrInvalidTransition  = errors.New("invalid status transition")
    ErrDuplicateEvent     = errors.New("duplicate event")
)