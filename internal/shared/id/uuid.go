package id

import "github.com/google/uuid"

// New generates a new UUID v7 (time-sortable, globally unique).
func New() string {
	return uuid.Must(uuid.NewV7()).String()
}

// IsValid checks whether the given string is a valid UUID.
func IsValid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
