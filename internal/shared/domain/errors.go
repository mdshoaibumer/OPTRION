package domain

import "fmt"

// ErrNotFound is returned when a requested resource does not exist.
type ErrNotFound struct {
	Entity string
	ID     string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Entity, e.ID)
}

// ErrValidation is returned when input validation fails.
type ErrValidation struct {
	Field   string
	Message string
}

func (e ErrValidation) Error() string {
	return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Message)
}

// ErrConflict is returned when a resource already exists or version conflict occurs.
type ErrConflict struct {
	Entity  string
	Message string
}

func (e ErrConflict) Error() string {
	return fmt.Sprintf("conflict on %s: %s", e.Entity, e.Message)
}
