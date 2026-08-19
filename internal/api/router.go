package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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
	mux.HandleFunc("/runtime", router.handleRuntime)
	mux.HandleFunc("/metrics", router.handleMetrics)
	mux.HandleFunc("/ingest", router.handleIngest)
	mux.HandleFunc("/ingest/async", router.handleIngestAsync)
	mux.HandleFunc("/events", router.handleDomainEvent)
	mux.HandleFunc("/recall", router.handleRecall)
	mux.HandleFunc("/jobs/status", router.handleJobsStatus)
	mux.HandleFunc("/jobs/", router.handleJobByID)
	mux.HandleFunc("/memories/search", router.handleSearch)
	mux.HandleFunc("/memories/", router.handleMemoryAction)
	return mux
}

// MaxBytesMiddleware caps request bodies with http.MaxBytesReader. When the
// limit is exceeded, subsequent reads return *http.MaxBytesError, which the
// JSON decode helper maps to a stable 413 response. A maxBytes <= 0 disables
// the limit.
func MaxBytesMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxBytes > 0 && r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// decodeJSON decodes a JSON request body, returning 400 for malformed JSON and
// a stable 413 when the body exceeds the configured size limit.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request_too_large", "request body exceeds the maximum allowed size")
			return err
		}
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return err
	}
	return nil
}

func (r *Router) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (r *Router) handleRuntime(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, r.service.Runtime(req.Context()))
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
	if err := decodeJSON(w, req, &payload); err != nil {
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
	includeHistorical := queryTruthy(req.URL.Query().Get("include_historical"))
	limit := 0
	if raw := strings.TrimSpace(req.URL.Query().Get("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	result, err := r.service.SearchOpt(
		req.Context(),
		req.URL.Query().Get("tenant_id"),
		req.URL.Query().Get("subject_id"),
		req.URL.Query().Get("vertical"),
		req.URL.Query().Get("scope"),
		req.URL.Query().Get("q"),
		memory.SearchOptions{IncludeHistorical: includeHistorical, Limit: limit},
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
	case strings.HasSuffix(path, "/supersede"):
		r.handleSupersede(w, req, path)
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
	if err := decodeJSON(w, req, &payload); err != nil {
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

func (r *Router) handleSupersede(w http.ResponseWriter, req *http.Request, path string) {
	memoryID := memoryIDFromActionPath(path, "/supersede")
	if memoryID == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "memory id is required")
		return
	}

	var payload memory.SupersedeRequest
	if err := decodeJSON(w, req, &payload); err != nil {
		return
	}

	tenantID := req.URL.Query().Get("tenant_id")
	subjectID := req.URL.Query().Get("subject_id")
	result, err := r.service.Supersede(req.Context(), tenantID, subjectID, memoryID, payload)
	if err != nil {
		if errors.Is(err, memory.ErrMemoryNotFound) {
			writeError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		if errors.Is(err, memory.ErrMemoryConflict) {
			writeError(w, http.StatusConflict, "conflict", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (r *Router) handleJobsStatus(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	tenantID := req.URL.Query().Get("tenant_id")
	subjectID := req.URL.Query().Get("subject_id")
	counts, err := r.service.SubjectJobCounts(req.Context(), tenantID, subjectID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, counts)
}

func (r *Router) handleJobByID(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	jobID := strings.TrimPrefix(req.URL.Path, "/jobs/")
	jobID = strings.Trim(jobID, "/")
	if jobID == "" || jobID == "status" {
		writeError(w, http.StatusBadRequest, "bad_request", "job_id is required")
		return
	}
	info, ok, err := r.service.GetJob(req.Context(), jobID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "job not found")
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (r *Router) handleDomainEvent(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	var payload memory.DomainEventRequest
	if err := decodeJSON(w, req, &payload); err != nil {
		return
	}
	result, err := r.service.ApplyDomainEvent(req.Context(), payload)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (r *Router) handleRecall(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	var payload memory.RecallRequest
	if err := decodeJSON(w, req, &payload); err != nil {
		return
	}
	result, err := r.service.Recall(req.Context(), payload)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func memoryIDFromActionPath(path, suffix string) string {
	return strings.TrimSuffix(strings.TrimSuffix(path, suffix), "/")
}

func queryTruthy(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
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
