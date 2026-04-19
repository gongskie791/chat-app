package util

import (
	"encoding/json"
	"net/http"
)

type Response[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    *T     `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
type ValidationErrorResponse struct {
	Success bool              `json:"success"`
	Error   string            `json:"error"`
	Fields  map[string]string `json:"fields"`
}

func JSON[T any](w http.ResponseWriter, status int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Success[T any](w http.ResponseWriter, status int, message string, data T) {
	JSON(w, status, Response[T]{
		Success: true,
		Message: message,
		Data:    &data,
	})
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, ErrorResponse{
		Success: false,
		Error:   message,
	})
}

type PaginatedResult[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Pages int64 `json:"pages"`
}

func ValidationFailed(w http.ResponseWriter, fields map[string]string) {
	JSON(w, http.StatusBadRequest, ValidationErrorResponse{
		Success: false,
		Error:   "validation failed",
		Fields:  fields,
	})
}
