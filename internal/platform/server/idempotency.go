package server

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

const (
	// IdempotencyKeyHeader is the HTTP header for idempotency keys.
	IdempotencyKeyHeader = "Idempotency-Key"
	// Maximum idempotency key length to prevent abuse.
	maxIdempotencyKeyLength = 128
	// How long to keep idempotency records before cleanup.
	idempotencyTTL = 24 * time.Hour
)

// IdempotencyRecord stores the result of a previously processed request.
type IdempotencyRecord struct {
	StatusCode int
	Body       []byte
	CreatedAt  time.Time
}

// IdempotencyStore provides in-memory idempotency tracking.
// For production at scale, replace with Redis-backed implementation.
type IdempotencyStore struct {
	mu      sync.RWMutex
	records map[string]*IdempotencyRecord
}

// NewIdempotencyStore creates a new in-memory idempotency store.
func NewIdempotencyStore() *IdempotencyStore {
	store := &IdempotencyStore{
		records: make(map[string]*IdempotencyRecord),
	}
	go store.cleanupLoop()
	return store
}

// Get retrieves a stored idempotency record if it exists and hasn't expired.
func (s *IdempotencyStore) Get(key string) (*IdempotencyRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rec, exists := s.records[key]
	if !exists {
		return nil, false
	}
	if time.Since(rec.CreatedAt) > idempotencyTTL {
		return nil, false
	}
	return rec, true
}

// Set stores an idempotency record.
func (s *IdempotencyStore) Set(key string, rec *IdempotencyRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[key] = rec
}

func (s *IdempotencyStore) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, rec := range s.records {
			if now.Sub(rec.CreatedAt) > idempotencyTTL {
				delete(s.records, key)
			}
		}
		s.mu.Unlock()
	}
}

// idempotencyResponseWriter captures the response for storage.
type idempotencyResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
	written    bool
}

func (w *idempotencyResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *idempotencyResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.body = make([]byte, len(b))
		copy(w.body, b)
		w.written = true
	}
	return w.ResponseWriter.Write(b)
}

// Idempotency middleware enforces idempotent mutations.
// When a client sends an Idempotency-Key header on POST/PUT/PATCH requests,
// duplicate requests with the same key return the cached response.
func Idempotency(store *IdempotencyStore) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to mutation methods
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get(IdempotencyKeyHeader)
			if key == "" {
				// No idempotency key — process normally
				next.ServeHTTP(w, r)
				return
			}

			// Validate key length
			if len(key) > maxIdempotencyKeyLength {
				WriteError(w, http.StatusBadRequest, "idempotency key too long (max 128 characters)")
				return
			}

			// Scope the key to the tenant + endpoint for isolation
			tenantID := TenantIDFromContext(r.Context())
			scopedKey := hashIdempotencyKey(tenantID, r.Method, r.URL.Path, key)

			// Check if we've already processed this request
			if rec, exists := store.Get(scopedKey); exists {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Idempotent-Replayed", "true")
				w.WriteHeader(rec.StatusCode)
				w.Write(rec.Body) //nolint:errcheck
				return
			}

			// Process the request and capture the response
			crw := &idempotencyResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			next.ServeHTTP(crw, r)

			// Store the response for future replays (only for successful requests)
			if crw.statusCode >= 200 && crw.statusCode < 500 {
				store.Set(scopedKey, &IdempotencyRecord{
					StatusCode: crw.statusCode,
					Body:       crw.body,
					CreatedAt:  time.Now(),
				})
			}
		})
	}
}

func hashIdempotencyKey(tenantID, method, path, key string) string {
	h := sha256.Sum256([]byte(tenantID + ":" + method + ":" + path + ":" + key))
	return hex.EncodeToString(h[:])
}
