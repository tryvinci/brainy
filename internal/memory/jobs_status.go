package memory

import "context"

type JobStatusInfo struct {
	JobID     string `json:"job_id"`
	IngestID  string `json:"ingest_id"`
	Status    string `json:"status"`
	TenantID  string `json:"tenant_id,omitempty"`
	SubjectID string `json:"subject_id,omitempty"`
	Reason    string `json:"failure_reason,omitempty"`
}

type SubjectJobCounts struct {
	Pending    int `json:"pending"`
	InProgress int `json:"in_progress"`
	Failed     int `json:"failed"`
	Completed  int `json:"completed"`
	Open       int `json:"open"`
}

// JobQuerier exposes extraction job readiness for harnesses and operators.
type JobQuerier interface {
	GetExtractionJob(ctx context.Context, jobID string) (JobStatusInfo, bool, error)
	CountSubjectJobs(ctx context.Context, tenantID, subjectID string) (SubjectJobCounts, error)
}
