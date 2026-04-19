package util

import "github.com/google/uuid"

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

type NotFoundError struct {
	Entity string
	ID     uuid.UUID
}

func (e *NotFoundError) Error() string {
	return e.Entity + " with id " + e.ID.String() + " not found"
}
