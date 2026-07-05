package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"brainy/internal/auth"
	"brainy/internal/memory"
	"brainy/internal/observability"
)

func TestAPIKeyMiddlewareRejectsMissingKey(t *testing.T) {
	ring := auth.ParseKeyRing("demo:sk_demo")
	handler := APIKeyMiddleware(ring, true)(NewRouter(memory.NewService(newMemoryStoreAdapter()), observability.NewMetrics()))

	req := httptest.NewRequest(http.MethodGet, "/memories/search?tenant_id=demo&subject_id=u1&q=test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAPIKeyMiddlewareAllowsMatchingTenant(t *testing.T) {
	ring := auth.ParseKeyRing("demo:sk_demo")
	handler := APIKeyMiddleware(ring, true)(NewRouter(memory.NewService(newMemoryStoreAdapter()), observability.NewMetrics()))

	req := httptest.NewRequest(http.MethodGet, "/memories/search?tenant_id=demo&subject_id=u1&q=test", nil)
	req.Header.Set("Authorization", "Bearer sk_demo")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAPIKeyMiddlewareRejectsTenantMismatch(t *testing.T) {
	ring := auth.ParseKeyRing("demo:sk_demo")
	handler := APIKeyMiddleware(ring, true)(NewRouter(memory.NewService(newMemoryStoreAdapter()), observability.NewMetrics()))

	req := httptest.NewRequest(http.MethodGet, "/memories/search?tenant_id=other&subject_id=u1&q=test", nil)
	req.Header.Set("X-API-Key", "sk_demo")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAPIKeyMiddlewareSkipsHealthz(t *testing.T) {
	ring := auth.ParseKeyRing("demo:sk_demo")
	handler := APIKeyMiddleware(ring, true)(NewRouter(memory.NewService(newMemoryStoreAdapter()), observability.NewMetrics()))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "ok") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
