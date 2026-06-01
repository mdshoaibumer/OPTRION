package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/optrion/optrion/internal/platform/server"
	"github.com/optrion/optrion/internal/registration/app"
	"github.com/optrion/optrion/internal/registration/domain"
)

// RegisterRequest is the JSON request body for POST /api/v1/register.
type RegisterRequest struct {
	Tenant      TenantReg      `json:"tenant"`
	Product     ProductReg     `json:"product"`
	Environment EnvironmentReg `json:"environment"`
	Components  []ComponentReg `json:"components"`
}

type TenantReg struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	Plan string `json:"plan"`
}

type ProductReg struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type EnvironmentReg struct {
	Name string `json:"name"`
	Tier string `json:"tier"`
}

type ComponentReg struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
	Port        int    `json:"port"`
}

// RegisterResponse is the JSON response body for successful registration.
type RegisterResponse struct {
	TenantID      string   `json:"tenant_id"`
	ProductID     string   `json:"product_id"`
	EnvironmentID string   `json:"environment_id"`
	ComponentIDs  []string `json:"component_ids"`
	APIKey        string   `json:"api_key"`
	Endpoint      string   `json:"endpoint"`
	Message       string   `json:"message"`
}

// NewRegistrationHandler creates HTTP handlers for the registration API.
func NewRegistrationHandler(svc *app.RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{
		service: svc,
	}
}

// RegistrationHandler provides HTTP handlers for registration operations.
type RegistrationHandler struct {
	service *app.RegistrationService
}

// Register handles POST /api/v1/register.
func (h *RegistrationHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		server.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var reqBody RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		server.WriteError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	// Convert request to domain model
	domainReq := domain.RegistrationRequest{
		Tenant: domain.TenantRegistration{
			Name: reqBody.Tenant.Name,
			Slug: reqBody.Tenant.Slug,
			Plan: reqBody.Tenant.Plan,
		},
		Product: domain.ProductRegistration{
			Name:        reqBody.Product.Name,
			Slug:        reqBody.Product.Slug,
			Description: reqBody.Product.Description,
		},
		Environment: domain.EnvironmentRegistration{
			Name: reqBody.Environment.Name,
			Tier: reqBody.Environment.Tier,
		},
		Components: make([]domain.ComponentRegistration, len(reqBody.Components)),
	}

	for i, comp := range reqBody.Components {
		domainReq.Components[i] = domain.ComponentRegistration{
			Name:        comp.Name,
			Kind:        comp.Kind,
			Description: comp.Description,
			Endpoint:    comp.Endpoint,
			Port:        comp.Port,
		}
	}

	// Call service
	result, err := h.service.Register(r.Context(), domainReq)
	if err != nil {
		// Determine appropriate HTTP status code
		statusCode := http.StatusInternalServerError
		msg := "internal server error"
		if errors.Is(err, context.Canceled) {
			statusCode = http.StatusRequestTimeout
			msg = "request timeout"
		} else if isValidationError(err) {
			statusCode = http.StatusBadRequest
			msg = err.Error()
		}
		server.WriteError(w, statusCode, msg)
		return
	}

	// Convert response back to HTTP format
	resp := RegisterResponse{
		TenantID:      result.TenantID,
		ProductID:     result.ProductID,
		EnvironmentID: result.EnvironmentID,
		ComponentIDs:  result.ComponentIDs,
		APIKey:        result.APIKey,
		Endpoint:      result.Endpoint,
		Message:       result.Message,
	}

	server.WriteJSON(w, http.StatusCreated, resp)
}

// isValidationError checks if an error is a domain validation error (safe to expose to clients).
func isValidationError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "invalid") ||
		strings.Contains(msg, "required") ||
		strings.Contains(msg, "registration")
}
