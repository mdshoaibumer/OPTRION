package server

import (
	"context"
	"net/http"
)

// contextKey for RLS tenant ID (separate from auth tenant ID for clarity)
const ContextKeyRLSTenantID contextKey = "rls_tenant_id"

// RLSTenantIDFromContext extracts the RLS-enforced tenant ID from context.
func RLSTenantIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ContextKeyRLSTenantID).(string)
	return v
}

// TenantContextInjection middleware automatically populates the tenant_id
// from the authenticated context into a separate RLS key.
// This ensures handlers always have tenant_id available without requiring
// it as a query parameter — eliminating the cross-tenant leakage vector.
func TenantContextInjection() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := TenantIDFromContext(r.Context())
			if tenantID == "" {
				// Not authenticated — pass through (public routes)
				next.ServeHTTP(w, r)
				return
			}

			// Inject the tenant ID as the authoritative RLS tenant
			ctx := context.WithValue(r.Context(), ContextKeyRLSTenantID, tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
