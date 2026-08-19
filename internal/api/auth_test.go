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

func TestAPIKeyMiddlewareAllowsRuntimeWithoutKey(t *testing.T) {
	ring := auth.ParseKeyRing("demo:sk_demo")
	handler := APIKeyMiddleware(ring, true)(NewRouter(memory.NewService(newMemoryStoreAdapter()), observability.NewMetrics()))
	req := httptest.NewRequest(http.MethodGet, "/runtime", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("runtime should be public like healthz, status=%d body=%s", rec.Code, rec.Body.String())
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

func TestAPIKeyMiddlewareWildcardTenantAllowsAny(t *testing.T) {
	ring := auth.ParseKeyRing("*:sk_bench")
	handler := APIKeyMiddleware(ring, true)(NewRouter(memory.NewService(newMemoryStoreAdapter()), observability.NewMetrics()))

	req := httptest.NewRequest(http.MethodGet, "/memories/search?tenant_id=opmem-t1&subject_id=u1&q=test", nil)
	req.Header.Set("Authorization", "Bearer sk_bench")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
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

func TestAPIKeyMiddlewareWithMaxBytesReturns413(t *testing.T) {
	ring := auth.ParseKeyRing("demo:sk_demo")
	base := NewRouter(memory.NewService(newMemoryStoreAdapter()), observability.NewMetrics())
	handler := MaxBytesMiddleware(64)(APIKeyMiddleware(ring, true)(base))

	body := `{"tenant_id":"demo","subject_id":"u1","source_type":"conversation","messages":[{"role":"user","content":"` + strings.Repeat("x", 4096) + `"}]}`
	req := httptest.NewRequest(http.MethodPost, "/ingest", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer sk_demo")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAPIKeyMiddlewareWithMaxBytesAllowsValidRequest(t *testing.T) {
	ring := auth.ParseKeyRing("demo:sk_demo")
	base := NewRouter(memory.NewService(newMemoryStoreAdapter()), observability.NewMetrics())
	handler := MaxBytesMiddleware(1 << 20)(APIKeyMiddleware(ring, true)(base))

	body := `{"tenant_id":"demo","subject_id":"u1","source_type":"conversation","messages":[{"role":"user","content":"I prefer concise answers."}]}`
	req := httptest.NewRequest(http.MethodPost, "/ingest", strings.NewReader(body))
	req.Header.Set("X-API-Key", "sk_demo")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
