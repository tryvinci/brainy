package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"brainy/internal/auth"
)

type tenantContextKey struct{}

func TenantFromContext(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(tenantContextKey{}).(string)
	return tenantID, ok
}

func withTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, tenantID)
}

func APIKeyMiddleware(ring *auth.KeyRing, require bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !require && (ring == nil || !ring.Enabled()) {
				next.ServeHTTP(w, r)
				return
			}
			if r.URL.Path == "/healthz" || r.URL.Path == "/runtime" {
				next.ServeHTTP(w, r)
				return
			}

			key := apiKeyFromRequest(r)
			tenantID, ok := ring.TenantForKey(key)
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized", "invalid or missing API key")
				return
			}

			// Wildcard tenant "*" authenticates without binding tenant_id
			// (staging / OpMem benchmarks that synthesize many tenants).
			requestTenant := tenantIDFromRequest(r)
			if tenantID != "*" && requestTenant != "" && requestTenant != tenantID {
				writeError(w, http.StatusForbidden, "forbidden", "tenant_id does not match API key")
				return
			}

			bound := tenantID
			if tenantID == "*" && requestTenant != "" {
				bound = requestTenant
			}
			next.ServeHTTP(w, r.WithContext(withTenant(r.Context(), bound)))
		})
	}
}

func apiKeyFromRequest(r *http.Request) string {
	if header := r.Header.Get("Authorization"); header != "" {
		if token, ok := strings.CutPrefix(header, "Bearer "); ok {
			return strings.TrimSpace(token)
		}
	}
	return strings.TrimSpace(r.Header.Get("X-API-Key"))
}

func tenantIDFromRequest(r *http.Request) string {
	if tenantID := r.URL.Query().Get("tenant_id"); tenantID != "" {
		return tenantID
	}
	if r.Method != http.MethodPost || r.Body == nil {
		return ""
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return ""
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	var payload struct {
		TenantID string `json:"tenant_id"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	return payload.TenantID
}
