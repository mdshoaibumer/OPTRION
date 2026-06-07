package server

import (
	"encoding/json"
	"net/http"
)

// ErrorCode represents a structured API error code.
type ErrorCode struct {
	Code        string `json:"code"`
	HTTPStatus  int    `json:"http_status"`
	Description string `json:"description"`
}

// ErrorCodes is the catalog of all known API error codes.
var ErrorCodes = map[string]ErrorCode{
	// Authentication & Authorization
	"AUTH_MISSING_HEADER": {
		Code: "AUTH_MISSING_HEADER", HTTPStatus: 401,
		Description: "The Authorization header is missing from the request.",
	},
	"AUTH_INVALID_FORMAT": {
		Code: "AUTH_INVALID_FORMAT", HTTPStatus: 401,
		Description: "Authorization header must use 'Bearer <api-key>' format.",
	},
	"AUTH_INVALID_KEY": {
		Code: "AUTH_INVALID_KEY", HTTPStatus: 401,
		Description: "The provided API key is invalid or does not exist.",
	},
	"AUTH_KEY_REVOKED": {
		Code: "AUTH_KEY_REVOKED", HTTPStatus: 401,
		Description: "The API key has been revoked.",
	},
	"AUTH_KEY_EXPIRED": {
		Code: "AUTH_KEY_EXPIRED", HTTPStatus: 401,
		Description: "The API key has expired.",
	},
	"AUTH_KEY_ROTATED": {
		Code: "AUTH_KEY_ROTATED", HTTPStatus: 401,
		Description: "The API key has been rotated and the grace period has expired.",
	},
	"AUTH_LOCKOUT": {
		Code: "AUTH_LOCKOUT", HTTPStatus: 429,
		Description: "Too many failed authentication attempts. Try again after the Retry-After period.",
	},

	// Validation
	"VALIDATION_FAILED": {
		Code: "VALIDATION_FAILED", HTTPStatus: 400,
		Description: "One or more request fields failed validation.",
	},
	"INVALID_REQUEST_BODY": {
		Code: "INVALID_REQUEST_BODY", HTTPStatus: 400,
		Description: "The request body could not be parsed as valid JSON.",
	},
	"INVALID_API_VERSION": {
		Code: "INVALID_API_VERSION", HTTPStatus: 400,
		Description: "The requested API version is not supported.",
	},

	// Resource Errors
	"TENANT_NOT_FOUND": {
		Code: "TENANT_NOT_FOUND", HTTPStatus: 404,
		Description: "The specified tenant does not exist.",
	},
	"PRODUCT_NOT_FOUND": {
		Code: "PRODUCT_NOT_FOUND", HTTPStatus: 404,
		Description: "The specified product does not exist.",
	},
	"ENVIRONMENT_NOT_FOUND": {
		Code: "ENVIRONMENT_NOT_FOUND", HTTPStatus: 404,
		Description: "The specified environment does not exist.",
	},
	"COMPONENT_NOT_FOUND": {
		Code: "COMPONENT_NOT_FOUND", HTTPStatus: 404,
		Description: "The specified component does not exist.",
	},
	"INCIDENT_NOT_FOUND": {
		Code: "INCIDENT_NOT_FOUND", HTTPStatus: 404,
		Description: "The specified incident does not exist.",
	},
	"ALERT_NOT_FOUND": {
		Code: "ALERT_NOT_FOUND", HTTPStatus: 404,
		Description: "The specified alert does not exist.",
	},
	"API_KEY_NOT_FOUND": {
		Code: "API_KEY_NOT_FOUND", HTTPStatus: 404,
		Description: "The specified API key does not exist.",
	},

	// Conflict
	"SLUG_TAKEN": {
		Code: "SLUG_TAKEN", HTTPStatus: 409,
		Description: "The specified slug is already in use.",
	},
	"DUPLICATE_RESOURCE": {
		Code: "DUPLICATE_RESOURCE", HTTPStatus: 409,
		Description: "A resource with the same unique identifier already exists.",
	},

	// State Machine
	"INVALID_TRANSITION": {
		Code: "INVALID_TRANSITION", HTTPStatus: 422,
		Description: "The requested state transition is not allowed from the current state.",
	},
	"TENANT_INACTIVE": {
		Code: "TENANT_INACTIVE", HTTPStatus: 403,
		Description: "The tenant account is inactive. Contact support.",
	},

	// Rate Limiting
	"RATE_LIMIT_EXCEEDED": {
		Code: "RATE_LIMIT_EXCEEDED", HTTPStatus: 429,
		Description: "Too many requests. Slow down and retry after the Retry-After period.",
	},

	// Server Errors
	"INTERNAL_ERROR": {
		Code: "INTERNAL_ERROR", HTTPStatus: 500,
		Description: "An unexpected error occurred. Contact support with the X-Request-ID header value.",
	},
	"SERVICE_UNAVAILABLE": {
		Code: "SERVICE_UNAVAILABLE", HTTPStatus: 503,
		Description: "The service is temporarily unavailable. Retry with exponential backoff.",
	},
	"AI_PROVIDER_UNAVAILABLE": {
		Code: "AI_PROVIDER_UNAVAILABLE", HTTPStatus: 503,
		Description: "All AI providers are currently unavailable. Analysis will be retried.",
	},
}

// ErrorCodesHandler returns an HTTP handler that serves the error codes catalog.
func ErrorCodesHandler() http.HandlerFunc {
	codes := make([]ErrorCode, 0, len(ErrorCodes))
	for _, c := range ErrorCodes {
		codes = append(codes, c)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error_codes": codes,
			"total":       len(codes),
		})
	}
}
