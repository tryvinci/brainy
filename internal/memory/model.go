package memory

import "time"

const (
	KindProfile    = "profile"
	KindPreference = "preference"
	KindFact       = "fact"

	StatusActive     = "active"
	StatusSuppressed = "suppressed"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type IngestRequest struct {
	TenantID   string    `json:"tenant_id"`
	SubjectID  string    `json:"subject_id"`
	Messages   []Message `json:"messages"`
	SourceType string    `json:"source_type"`
}

type CorrectionRequest struct {
	Content    string `json:"content"`
	SourceText string `json:"source_text,omitempty"`
}

type MemoryRecord struct {
	MemoryID          string
	TenantID          string
	SubjectID         string
	Kind              string
	Content           string
	SourceText        string
	SourceType        string
	DedupeKey         string
	Status            string
	Confidence        float64
	ExtractionVersion string
	Explain           map[string]any
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type IngestResult struct {
	IngestID  string               `json:"ingest_id"`
	Accepted  bool                 `json:"accepted"`
	Created   int                  `json:"created"`
	Updated   int                  `json:"updated"`
	Deduped   int                  `json:"deduped"`
	Memories  []IngestResultMemory `json:"memories"`
}

type IngestResultMemory struct {
	MemoryID string `json:"memory_id"`
	Kind     string `json:"kind"`
	Content  string `json:"content"`
	Status   string `json:"status"`
}

type SearchResult struct {
	MemoryID string         `json:"memory_id"`
	Kind     string         `json:"kind"`
	Content  string         `json:"content"`
	Score    float64        `json:"score"`
	Explain  map[string]any `json:"explain"`
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
}
