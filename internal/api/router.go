package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"brainy/internal/memory"
	"brainy/internal/observability"
)

type Router struct {
	service *memory.Service
	metrics *observability.Metrics
}

func NewRouter(service *memory.Service, metrics *observability.Metrics) http.Handler {
	router := &Router{service: service, metrics: metrics}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", router.handleHealth)
	mux.HandleFunc("/metrics", router.handleMetrics)
	mux.HandleFunc("/ingest", router.handleIngest)
	mux.HandleFunc("/ingest/async", router.handleIngestAsync)
	mux.HandleFunc("/memories/search", router.handleSearch)
	mux.HandleFunc("/memories/", router.handleMemoryAction)
	return mux
}

func (r *Router) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (r *Router) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = w.Write([]byte(r.metrics.Prometheus()))
}

func (r *Router) handleIngest(w http.ResponseWriter, req *http.Request) {
	r.handleIngestRequest(w, req, false)
}

func (r *Router) handleIngestAsync(w http.ResponseWriter, req *http.Request) {
	r.handleIngestRequest(w, req, true)
}

func (r *Router) handleIngestRequest(w http.ResponseWriter, req *http.Request, async bool) {
	if req.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	var payload memory.IngestRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}

	start := time.Now()
	var (
		result any
		status = http.StatusOK
		err    error
	)
	if async {
		result, err = r.service.IngestAsync(req.Context(), payload)
		status = http.StatusAccepted
	} else {
		result, err = r.service.Ingest(req.Context(), payload)
	}
	r.metrics.RecordIngest(time.Since(start), err != nil)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	writeJSON(w, status, result)
}

func (r *Router) handleSearch(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	start := time.Now()
	result, err := r.service.Search(
		req.Context(),
		req.URL.Query().Get("tenant_id"),
		req.URL.Query().Get("subject_id"),
		req.URL.Query().Get("vertical"),
		req.URL.Query().Get("q"),
	)
	r.metrics.RecordSearch(time.Since(start), err != nil)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (r *Router) handleMemoryAction(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/memories/")
	switch {
	case strings.HasSuffix(path, "/suppress"):
		r.handleSuppress(w, req, path)
		return
	case strings.HasSuffix(path, "/correct"):
		r.handleCorrect(w, req, path)
		return
	default:
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}
}

func (r *Router) handleSuppress(w http.ResponseWriter, req *http.Request, path string) {
	memoryID := memoryIDFromActionPath(path, "/suppress")
	if memoryID == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "memory id is required")
		return
	}

	tenantID := req.URL.Query().Get("tenant_id")
	subjectID := req.URL.Query().Get("subject_id")
	if err := r.service.Suppress(req.Context(), tenantID, subjectID, memoryID); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"memory_id": memoryID,
		"status":    memory.StatusSuppressed,
	})
}

func (r *Router) handleCorrect(w http.ResponseWriter, req *http.Request, path string) {
	memoryID := memoryIDFromActionPath(path, "/correct")
	if memoryID == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "memory id is required")
		return
	}

	var payload memory.CorrectionRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}

	tenantID := req.URL.Query().Get("tenant_id")
	subjectID := req.URL.Query().Get("subject_id")
	result, err := r.service.Correct(req.Context(), tenantID, subjectID, memoryID, payload)
	if err != nil {
		if errors.Is(err, memory.ErrMemoryConflict) {
			writeError(w, http.StatusConflict, "conflict", err.Error())
			return
		}
		if errors.Is(err, memory.ErrMemoryNotFound) {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func memoryIDFromActionPath(path, suffix string) string {
	return strings.TrimSuffix(strings.TrimSuffix(path, suffix), "/")
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
