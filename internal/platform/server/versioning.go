package server

import (
	"context"
	"net/http"
	"strings"
)

// APIVersion represents a supported API version.
type APIVersion string

const (
	// APIVersionV1 is the initial stable API version.
	APIVersionV1 APIVersion = "v1"
	// APIVersionV2 is reserved for future breaking changes.
	APIVersionV2 APIVersion = "v2"

	// CurrentAPIVersion is the latest supported version.
	CurrentAPIVersion = APIVersionV1

	// MinimumAPIVersion is the oldest supported version.
	MinimumAPIVersion = APIVersionV1
)

// SupportedVersions lists all currently supported API versions.
var SupportedVersions = map[APIVersion]bool{
	APIVersionV1: true,
}

// DeprecatedVersions lists versions that still work but emit deprecation warnings.
var DeprecatedVersions = map[APIVersion]string{
	// Example: APIVersionV1: "v1 is deprecated, migrate to v2 by 2027-01-01",
}

// APIVersionHeader is the header clients can use to specify API version.
const APIVersionHeader = "X-API-Version"

// APIVersionContextKey is the context key for the resolved API version.
const APIVersionContextKey contextKey = "api_version"

// APIVersionFromContext extracts the resolved API version from the request context.
func APIVersionFromContext(ctx context.Context) APIVersion {
	v, _ := ctx.Value(APIVersionContextKey).(APIVersion)
	if v == "" {
		return CurrentAPIVersion
	}
	return v
}

// VersionNegotiation middleware resolves the API version from the URL path or header.
// URL path takes precedence: /api/v1/... extracts "v1".
// Falls back to X-API-Version header, then defaults to current version.
// Rejects requests with unsupported versions.
func VersionNegotiation() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			version := resolveVersion(r)

			if !SupportedVersions[version] {
				WriteError(w, http.StatusBadRequest, "unsupported API version: "+string(version))
				return
			}

			// Add deprecation warning header if applicable
			if msg, deprecated := DeprecatedVersions[version]; deprecated {
				w.Header().Set("Deprecation", "true")
				w.Header().Set("Sunset", msg)
				w.Header().Set("X-API-Deprecation-Notice", msg)
			}

			// Inject version into context
			ctx := context.WithValue(r.Context(), APIVersionContextKey, version)

			// Set response header to confirm which version was used
			w.Header().Set(APIVersionHeader, string(version))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func resolveVersion(r *http.Request) APIVersion {
	// 1. Extract from URL path: /api/v1/tenants/...
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(pathParts) >= 2 && pathParts[0] == "api" {
		candidate := APIVersion(pathParts[1])
		if SupportedVersions[candidate] {
			return candidate
		}
	}

	// 2. Extract from header
	headerVersion := r.Header.Get(APIVersionHeader)
	if headerVersion != "" {
		return APIVersion(headerVersion)
	}

	// 3. Default to current
	return CurrentAPIVersion
}
