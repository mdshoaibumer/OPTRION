package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParsePagination_Defaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	params := ParsePagination(req)

	if params.Limit != DefaultPageLimit {
		t.Fatalf("expected default limit %d, got %d", DefaultPageLimit, params.Limit)
	}
	if params.Offset != 0 {
		t.Fatalf("expected default offset 0, got %d", params.Offset)
	}
}

func TestParsePagination_CustomValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items?limit=25&offset=50", nil)
	params := ParsePagination(req)

	if params.Limit != 25 {
		t.Fatalf("expected limit 25, got %d", params.Limit)
	}
	if params.Offset != 50 {
		t.Fatalf("expected offset 50, got %d", params.Offset)
	}
}

func TestParsePagination_EnforcesMaxLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items?limit=500", nil)
	params := ParsePagination(req)

	if params.Limit != MaxPageLimit {
		t.Fatalf("expected max limit %d, got %d", MaxPageLimit, params.Limit)
	}
}

func TestParsePagination_NegativeValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items?limit=-5&offset=-10", nil)
	params := ParsePagination(req)

	if params.Limit != DefaultPageLimit {
		t.Fatalf("expected default limit on negative, got %d", params.Limit)
	}
	if params.Offset != 0 {
		t.Fatalf("expected 0 offset on negative, got %d", params.Offset)
	}
}

func TestParsePagination_InvalidValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items?limit=abc&offset=xyz", nil)
	params := ParsePagination(req)

	if params.Limit != DefaultPageLimit {
		t.Fatalf("expected default limit on invalid, got %d", params.Limit)
	}
	if params.Offset != 0 {
		t.Fatalf("expected default offset on invalid, got %d", params.Offset)
	}
}

func TestParsePagination_Cursor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items?cursor=abc123", nil)
	params := ParsePagination(req)

	if params.Cursor != "abc123" {
		t.Fatalf("expected cursor abc123, got %s", params.Cursor)
	}
}

func TestNewPaginatedResponse_HasMore(t *testing.T) {
	resp := NewPaginatedResponse([]string{"a", "b"}, 100, 10, 0)

	if !resp.Pagination.HasMore {
		t.Fatal("expected has_more to be true when offset+limit < total")
	}
	if resp.Pagination.Total != 100 {
		t.Fatalf("expected total 100, got %d", resp.Pagination.Total)
	}
}

func TestNewPaginatedResponse_NoMore(t *testing.T) {
	resp := NewPaginatedResponse([]string{"a"}, 10, 10, 5)

	if resp.Pagination.HasMore {
		t.Fatal("expected has_more to be false when offset+limit >= total")
	}
}
