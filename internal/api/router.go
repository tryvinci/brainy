package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"brainy/internal/memory"
)

type Router struct {
	service *memory.Service
}

func NewRouter(service *memory.Service) http.Handler {
	router := &Router{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", router.handleHealth)
	mux.HandleFunc("/ingest", router.handleIngest)
	mux.HandleFunc("/memories/search", router.handleSearch)
	mux.HandleFunc("/memories/", router.handleMemoryAction)
	return mux
}

func (r *Router) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (r *Router) handleIngest(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload memory.IngestRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	result, err := r.service.Ingest(req.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (r *Router) handleSearch(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result, err := r.service.Search(
		req.Context(),
		req.URL.Query().Get("tenant_id"),
		req.URL.Query().Get("subject_id"),
		req.URL.Query().Get("q"),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (r *Router) handleMemoryAction(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/memories/")
	if !strings.HasSuffix(path, "/suppress") {
		http.NotFound(w, req)
		return
	}

	memoryID := strings.TrimSuffix(path, "/suppress")
	memoryID = strings.TrimSuffix(memoryID, "/")
	if memoryID == "" {
		http.Error(w, "memory id is required", http.StatusBadRequest)
		return
	}

	tenantID := req.URL.Query().Get("tenant_id")
	subjectID := req.URL.Query().Get("subject_id")
	if err := r.service.Suppress(req.Context(), tenantID, subjectID, memoryID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"memory_id": memoryID,
		"status":    memory.StatusSuppressed,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
