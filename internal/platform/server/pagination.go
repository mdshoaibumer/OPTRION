package server

import (
	"net/http"
	"strconv"
)

// PaginationParams holds parsed pagination parameters.
type PaginationParams struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Cursor string `json:"cursor,omitempty"`
}

// PaginatedResponse wraps a list response with pagination metadata.
type PaginatedResponse struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// PaginationMeta contains pagination metadata for list responses.
type PaginationMeta struct {
	Total      int    `json:"total"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
}

const (
	// DefaultPageLimit is the default number of items per page.
	DefaultPageLimit = 50
	// MaxPageLimit is the maximum allowed items per page.
	MaxPageLimit = 200
)

// ParsePagination extracts and validates pagination parameters from a request.
func ParsePagination(r *http.Request) PaginationParams {
	limit := parseQueryInt(r, "limit", DefaultPageLimit)
	if limit <= 0 {
		limit = DefaultPageLimit
	}
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}

	offset := parseQueryInt(r, "offset", 0)
	if offset < 0 {
		offset = 0
	}

	cursor := r.URL.Query().Get("cursor")

	return PaginationParams{
		Limit:  limit,
		Offset: offset,
		Cursor: cursor,
	}
}

// NewPaginatedResponse creates a paginated response from data and pagination info.
func NewPaginatedResponse(data interface{}, total, limit, offset int) PaginatedResponse {
	return PaginatedResponse{
		Data: data,
		Pagination: PaginationMeta{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: offset+limit < total,
		},
	}
}

// parseQueryInt parses an integer query parameter with a default value.
func parseQueryInt(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}
