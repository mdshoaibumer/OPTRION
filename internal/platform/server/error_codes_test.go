package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorCodes_AllHaveRequiredFields(t *testing.T) {
	for key, code := range ErrorCodes {
		if code.Code == "" {
			t.Errorf("error code %s has empty Code field", key)
		}
		if code.HTTPStatus == 0 {
			t.Errorf("error code %s has zero HTTPStatus", key)
		}
		if code.Description == "" {
			t.Errorf("error code %s has empty Description", key)
		}
		if code.Code != key {
			t.Errorf("error code key %s doesn't match Code field %s", key, code.Code)
		}
	}
}

func TestErrorCodes_ValidHTTPStatus(t *testing.T) {
	for key, code := range ErrorCodes {
		if code.HTTPStatus < 400 || code.HTTPStatus > 599 {
			t.Errorf("error code %s has unexpected HTTP status %d", key, code.HTTPStatus)
		}
	}
}

func TestErrorCodesHandler_ReturnsAllCodes(t *testing.T) {
	handler := ErrorCodesHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/error-codes", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		ErrorCodes []ErrorCode `json:"error_codes"`
		Total      int         `json:"total"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Total != len(ErrorCodes) {
		t.Fatalf("expected %d error codes, got %d", len(ErrorCodes), resp.Total)
	}
}
