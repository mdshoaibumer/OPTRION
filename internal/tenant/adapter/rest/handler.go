package rest

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/optrion/optrion/internal/platform/server"
	"github.com/optrion/optrion/internal/tenant/app"
	"github.com/optrion/optrion/internal/tenant/domain"
	"github.com/optrion/optrion/internal/tenant/port"
)

// Handler handles tenant-related HTTP requests.
type Handler struct {
	service *app.TenantService
	logger  *slog.Logger
}

// NewHandler creates a new tenant REST handler.
func NewHandler(service *app.TenantService, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers all tenant-related routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tenants", h.CreateTenant)
	mux.HandleFunc("GET /api/v1/tenants", h.ListTenants)
	mux.HandleFunc("GET /api/v1/tenants/{id}", h.GetTenant)
	mux.HandleFunc("POST /api/v1/products", h.CreateProduct)
	mux.HandleFunc("GET /api/v1/products", h.ListProducts)
	mux.HandleFunc("POST /api/v1/environments", h.CreateEnvironment)
	mux.HandleFunc("GET /api/v1/environments", h.ListEnvironments)
	mux.HandleFunc("POST /api/v1/components", h.RegisterComponent)
	mux.HandleFunc("GET /api/v1/components", h.ListComponents)
}

// RegisterAuthenticatedRoutes registers tenant routes wrapped with authentication middleware.
func (h *Handler) RegisterAuthenticatedRoutes(mux *http.ServeMux, authWrap func(http.Handler) http.Handler) {
	mux.Handle("POST /api/v1/tenants", authWrap(http.HandlerFunc(h.CreateTenant)))
	mux.Handle("GET /api/v1/tenants", authWrap(http.HandlerFunc(h.ListTenants)))
	mux.Handle("GET /api/v1/tenants/{id}", authWrap(http.HandlerFunc(h.GetTenant)))
	mux.Handle("POST /api/v1/products", authWrap(http.HandlerFunc(h.CreateProduct)))
	mux.Handle("GET /api/v1/products", authWrap(http.HandlerFunc(h.ListProducts)))
	mux.Handle("POST /api/v1/environments", authWrap(http.HandlerFunc(h.CreateEnvironment)))
	mux.Handle("GET /api/v1/environments", authWrap(http.HandlerFunc(h.ListEnvironments)))
	mux.Handle("POST /api/v1/components", authWrap(http.HandlerFunc(h.RegisterComponent)))
	mux.Handle("GET /api/v1/components", authWrap(http.HandlerFunc(h.ListComponents)))
	mux.Handle("GET /api/v1/audit-logs", authWrap(http.HandlerFunc(h.ListAuditLogs)))
}

// CreateTenant handles POST /api/v1/tenants
func (h *Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req CreateTenantRequest
	if err := server.ReadJSON(w, r, &req); err != nil {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "validation failed", Details: errs})
		return
	}

	tenant, err := h.service.CreateTenant(r.Context(), app.CreateTenantCmd{
		Name: req.Name,
		Slug: req.Slug,
		Plan: req.Plan,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	server.WriteJSON(w, http.StatusCreated, toTenantResponse(tenant))
}

// ListTenants handles GET /api/v1/tenants
func (h *Handler) ListTenants(w http.ResponseWriter, r *http.Request) {
	filter := port.TenantFilter{
		Limit:  parseIntQuery(r, "limit", 50),
		Offset: parseIntQuery(r, "offset", 0),
	}

	// Enforce tenant isolation: authenticated users can only see their own tenant
	if authTenantID := server.TenantIDFromContext(r.Context()); authTenantID != "" {
		filter.ID = &authTenantID
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.Status(status)
		filter.Status = &s
	}

	tenants, err := h.service.ListTenants(r.Context(), filter)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	resp := ListResponse[TenantResponse]{
		Data:   make([]TenantResponse, 0, len(tenants)),
		Count:  len(tenants),
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}
	for _, t := range tenants {
		resp.Data = append(resp.Data, toTenantResponse(t))
	}

	server.WriteJSON(w, http.StatusOK, resp)
}

// GetTenant handles GET /api/v1/tenants/{id}
func (h *Handler) GetTenant(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "tenant id is required"})
		return
	}

	tenant, err := h.service.GetTenant(r.Context(), id)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	server.WriteJSON(w, http.StatusOK, toTenantResponse(tenant))
}

// CreateProduct handles POST /api/v1/products
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := server.ReadJSON(w, r, &req); err != nil {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "validation failed", Details: errs})
		return
	}

	// Auto-inject tenant_id from auth context — never trust client-supplied tenant_id
	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
		return
	}

	product, err := h.service.CreateProduct(r.Context(), app.CreateProductCmd{
		TenantID:    tenantID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	server.WriteJSON(w, http.StatusCreated, toProductResponse(product))
}

// ListProducts handles GET /api/v1/products
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	// Auto-inject tenant_id from authenticated context (eliminates cross-tenant leakage)
	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
		return
	}

	filter := port.ProductFilter{
		Limit:  parseIntQuery(r, "limit", 50),
		Offset: parseIntQuery(r, "offset", 0),
	}

	products, err := h.service.ListProducts(r.Context(), tenantID, filter)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	resp := ListResponse[ProductResponse]{
		Data:   make([]ProductResponse, 0, len(products)),
		Count:  len(products),
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}
	for _, p := range products {
		resp.Data = append(resp.Data, toProductResponse(p))
	}

	server.WriteJSON(w, http.StatusOK, resp)
}

// CreateEnvironment handles POST /api/v1/environments
func (h *Handler) CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	var req CreateEnvironmentRequest
	if err := server.ReadJSON(w, r, &req); err != nil {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "validation failed", Details: errs})
		return
	}

	// Auto-inject tenant_id from auth context
	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
		return
	}

	env, err := h.service.CreateEnvironment(r.Context(), app.CreateEnvironmentCmd{
		TenantID:  tenantID,
		ProductID: req.ProductID,
		Name:      req.Name,
		Slug:      req.Slug,
		Tier:      req.Tier,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	server.WriteJSON(w, http.StatusCreated, toEnvironmentResponse(env))
}

// ListEnvironments handles GET /api/v1/environments?product_id=...
func (h *Handler) ListEnvironments(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("product_id")
	if productID == "" {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "product_id query parameter is required"})
		return
	}

	filter := port.EnvironmentFilter{
		Limit:  parseIntQuery(r, "limit", 50),
		Offset: parseIntQuery(r, "offset", 0),
	}

	envs, err := h.service.ListEnvironments(r.Context(), productID, filter)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	resp := ListResponse[EnvironmentResponse]{
		Data:   make([]EnvironmentResponse, 0, len(envs)),
		Count:  len(envs),
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}
	for _, e := range envs {
		resp.Data = append(resp.Data, toEnvironmentResponse(e))
	}

	server.WriteJSON(w, http.StatusOK, resp)
}

// RegisterComponent handles POST /api/v1/components
func (h *Handler) RegisterComponent(w http.ResponseWriter, r *http.Request) {
	var req RegisterComponentRequest
	if err := server.ReadJSON(w, r, &req); err != nil {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "validation failed", Details: errs})
		return
	}

	// Auto-inject tenant_id from auth context
	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
		return
	}

	comp, err := h.service.RegisterComponent(r.Context(), app.RegisterComponentCmd{
		TenantID:      tenantID,
		ProductID:     req.ProductID,
		EnvironmentID: req.EnvironmentID,
		Name:          req.Name,
		Slug:          req.Slug,
		Kind:          req.Kind,
		EndpointURL:   req.EndpointURL,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	server.WriteJSON(w, http.StatusCreated, toComponentResponse(comp))
}

// ListComponents handles GET /api/v1/components?environment_id=...
func (h *Handler) ListComponents(w http.ResponseWriter, r *http.Request) {
	environmentID := r.URL.Query().Get("environment_id")
	if environmentID == "" {
		server.WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "environment_id query parameter is required"})
		return
	}

	filter := port.ComponentFilter{
		Limit:  parseIntQuery(r, "limit", 50),
		Offset: parseIntQuery(r, "offset", 0),
	}

	comps, err := h.service.ListComponents(r.Context(), environmentID, filter)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	resp := ListResponse[ComponentResponse]{
		Data:   make([]ComponentResponse, 0, len(comps)),
		Count:  len(comps),
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}
	for _, c := range comps {
		resp.Data = append(resp.Data, toComponentResponse(c))
	}

	server.WriteJSON(w, http.StatusOK, resp)
}

// ListAuditLogs handles GET /api/v1/audit-logs
func (h *Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := server.TenantIDFromContext(r.Context())
	if tenantID == "" {
		server.WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
		return
	}

	filter := port.AuditFilter{
		Limit:  parseIntQuery(r, "limit", 50),
		Offset: parseIntQuery(r, "offset", 0),
	}

	if action := r.URL.Query().Get("action"); action != "" {
		filter.Action = &action
	}
	if entityType := r.URL.Query().Get("entity_type"); entityType != "" {
		filter.EntityType = &entityType
	}

	events, total, err := h.service.ListAuditEvents(r.Context(), tenantID, filter)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	resp := make([]AuditEventResponse, 0, len(events))
	for _, e := range events {
		resp = append(resp, toAuditEventResponse(e))
	}

	server.WriteJSON(w, http.StatusOK, server.NewPaginatedResponse(resp, total, filter.Limit, filter.Offset))
}

// --- Error Handling ---

func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.As(err, &domain.ErrTenantNotFound{}):
		server.WriteJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
	case errors.As(err, &domain.ErrProductNotFound{}):
		server.WriteJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
	case errors.As(err, &domain.ErrEnvironmentNotFound{}):
		server.WriteJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
	case errors.As(err, &domain.ErrComponentNotFound{}):
		server.WriteJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
	case errors.As(err, &domain.ErrTenantSlugTaken{}):
		server.WriteJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
	case errors.As(err, &domain.ErrProductSlugTaken{}):
		server.WriteJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
	case errors.As(err, &domain.ErrEnvironmentSlugTaken{}):
		server.WriteJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
	case errors.As(err, &domain.ErrComponentSlugTaken{}):
		server.WriteJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
	case errors.As(err, &domain.ErrTenantInactive{}):
		server.WriteJSON(w, http.StatusForbidden, ErrorResponse{Error: err.Error()})
	default:
		h.logger.ErrorContext(r.Context(), "unhandled error", "error", err)
		server.WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}

// --- Mappers ---

func toTenantResponse(t *domain.Tenant) TenantResponse {
	return TenantResponse{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug.String(),
		Plan:      string(t.Plan),
		Status:    string(t.Status),
		Settings:  t.Settings,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func toProductResponse(p *domain.Product) ProductResponse {
	return ProductResponse{
		ID:          p.ID,
		TenantID:    p.TenantID,
		Name:        p.Name,
		Slug:        p.Slug.String(),
		Description: p.Description,
		Status:      string(p.Status),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toEnvironmentResponse(e *domain.Environment) EnvironmentResponse {
	return EnvironmentResponse{
		ID:        e.ID,
		TenantID:  e.TenantID,
		ProductID: e.ProductID,
		Name:      e.Name,
		Slug:      e.Slug.String(),
		Tier:      string(e.Tier),
		Status:    string(e.Status),
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func toComponentResponse(c *domain.Component) ComponentResponse {
	return ComponentResponse{
		ID:            c.ID,
		TenantID:      c.TenantID,
		ProductID:     c.ProductID,
		EnvironmentID: c.EnvironmentID,
		Name:          c.Name,
		Slug:          c.Slug.String(),
		Kind:          string(c.Kind),
		EndpointURL:   c.EndpointURL,
		Status:        string(c.Status),
		Metadata:      c.Metadata,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

func toAuditEventResponse(e *domain.AuditEvent) AuditEventResponse {
	return AuditEventResponse{
		ID:         e.ID,
		TenantID:   e.TenantID,
		ActorID:    e.ActorID,
		Action:     e.Action,
		EntityType: e.EntityType,
		EntityID:   e.EntityID,
		Payload:    e.Payload,
		OccurredAt: e.OccurredAt,
	}
}

// --- Helpers ---

func parseIntQuery(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}
