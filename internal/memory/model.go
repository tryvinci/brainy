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
	Role      string   `json:"role"`
	Content   string   `json:"content"`
	ImageURLs []string `json:"image_urls,omitempty"`
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

// SupersedeRequest replaces a memory with a new active record and marks the
// prior record lifecycle=superseded (ENG-86). Prefer this over in-place Correct
// when the product needs ADD-style lineage (supersedes_id).
type SupersedeRequest struct {
	Content    string `json:"content"`
	SourceText string `json:"source_text,omitempty"`
}

// DomainEventRequest batch-invalidates memories (campaign end, fact revision).
// Prefer explicit IDs when known; or Match to select by label/metadata (v2).
type DomainEventRequest struct {
	TenantID           string            `json:"tenant_id"`
	SubjectID          string            `json:"subject_id"`
	EventType          string            `json:"event_type"`
	SupersedeMemoryIDs []string          `json:"supersede_memory_ids,omitempty"`
	Match              *DomainEventMatch `json:"match,omitempty"`
	Metadata           map[string]any    `json:"metadata,omitempty"`
}

// DomainEventMatch selects memories to supersede without listing IDs.
type DomainEventMatch struct {
	Label    string            `json:"label,omitempty"`
	Kind     string            `json:"kind,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type DomainEventResult struct {
	EventType  string   `json:"event_type"`
	Superseded []string `json:"superseded"`
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
	// SupersedesID is the memory this record replaces (new → old). Empty if none.
	SupersedesID string
	// SupersededAt is when this record was marked lifecycle=superseded.
	SupersededAt *time.Time
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

// SearchOptions controls retrieval visibility. Default excludes superseded /
// archived / suppressed lifecycle states (production default).
type SearchOptions struct {
	IncludeHistorical bool // when true, include lifecycle=superseded rows
	// IncludeEpisodes keeps conversation_episode rows even when structured
	// coverage is complete. Off by default: facts/edges are recall-primary;
	// episodes stay as bounded fallback while representation_status is
	// empty or partial (R1c). view=all sets this so provenance remains visible.
	IncludeEpisodes bool
	// Limit caps returned results (0 = unlimited / caller truncates).
	Limit int
	// CandidateLimit is the explicit retrieval pool size before the context
	// token budget. 0 uses CandidateOverfetch(Limit), still capped at 200.
	CandidateLimit int
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Trace   *SearchTrace   `json:"trace,omitempty"`
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
	JobID       string
	IngestID    string
	Request     IngestRequest
	Attempts    int
	MaxAttempts int
	CreatedAt   time.Time
	// LeaseOwner is the per-claim fencing token written when the job is
	// claimed. Complete/Fail must present the same token; a reclaimed job
	// rejects the old claimant.
	LeaseOwner string
}

type EnqueueResult struct {
	IngestID  string `json:"ingest_id"`
	JobID     string `json:"job_id"`
	Accepted  bool   `json:"accepted"`
	Duplicate bool   `json:"duplicate"`
}
