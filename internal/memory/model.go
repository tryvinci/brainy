package memory

import "time"

const (
	KindProfile    = "profile"
	KindPreference = "preference"
	KindFact       = "fact"

	StatusActive     = "active"
	StatusSuppressed = "suppressed"

	VerticalCore = "core"

	PrimitivePrinciple     = "principle"
	PrimitiveIdentityPrior = "identity_prior"
	PrimitiveEpisode       = "episode"
	PrimitivePattern       = "pattern"
	PrimitiveBelief        = "belief"
	PrimitiveOutcome       = "outcome"
	PrimitiveExperiment    = "experiment"
	PrimitiveTasteSignal   = "taste_signal"
	PrimitiveReflection    = "reflection"

	LifecycleActive        = "active"
	LifecycleDeprioritized = "deprioritized"
	LifecycleArchived      = "archived"
	LifecycleSuppressed    = "suppressed"
	LifecycleSuperseded    = "superseded"

	RawIngestStatusPending   = "pending"
	RawIngestStatusProcessed = "processed"
	RawIngestStatusFailed    = "failed"

	JobStatusPending    = "pending"
	JobStatusInProgress = "in_progress"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type IngestRequest struct {
	TenantID   string         `json:"tenant_id"`
	SubjectID  string         `json:"subject_id"`
	Vertical   string         `json:"vertical,omitempty"`
	Label      string         `json:"label,omitempty"`
	Scope      string         `json:"scope,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Messages   []Message      `json:"messages"`
	SourceType string         `json:"source_type"`
}

type CorrectionRequest struct {
	Content    string `json:"content"`
	SourceText string `json:"source_text,omitempty"`
}

type MemoryRecord struct {
	MemoryID          string
	TenantID          string
	SubjectID         string
	Vertical          string
	Primitive         string
	Label             string
	Scope             string
	Metadata          map[string]any
	LifecycleState    string
	Kind              string
	Content           string
	SourceText        string
	SourceType        string
	DedupeKey         string
	Status            string
	Confidence        float64
	ExtractionVersion string
	Explain           map[string]any
	Embedding         []float32
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CorrectedAt       *time.Time
	// ObservedAt is conversational event time (client metadata.observed_at or provider when).
	ObservedAt *time.Time
}

type IngestResult struct {
	IngestID string               `json:"ingest_id"`
	Accepted bool                 `json:"accepted"`
	Created  int                  `json:"created"`
	Updated  int                  `json:"updated"`
	Deduped  int                  `json:"deduped"`
	Memories []IngestResultMemory `json:"memories"`
}

type AsyncIngestResult struct {
	IngestID string `json:"ingest_id"`
	JobID    string `json:"job_id"`
	Accepted bool   `json:"accepted"`
}

type IngestResultMemory struct {
	MemoryID string `json:"memory_id"`
	Kind     string `json:"kind"`
	Content  string `json:"content"`
	Status   string `json:"status"`
}

type SearchResult struct {
	MemoryID   string         `json:"memory_id"`
	Kind       string         `json:"kind"`
	Content    string         `json:"content"`
	Score      float64        `json:"score"`
	ObservedAt *time.Time     `json:"observed_at,omitempty"`
	Explain    map[string]any `json:"explain"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

type MutationResult struct {
	MemoryID string `json:"memory_id"`
	Kind     string `json:"kind"`
	Content  string `json:"content"`
	Status   string `json:"status"`
}

type ExtractedMemory struct {
	Kind       string
	Content    string
	SourceText string
	Confidence float64
	Explain    map[string]any
	When       string // optional temporal slot from provider extract
	Duration   string
}

type ExtractionJob struct {
	JobID        string
	IngestID     string
	Request      IngestRequest
	Attempts     int
	MaxAttempts  int
	CreatedAt    time.Time
}

type EnqueueResult struct {
	IngestID  string `json:"ingest_id"`
	JobID     string `json:"job_id"`
	Accepted  bool   `json:"accepted"`
	Duplicate bool   `json:"duplicate"`
}
