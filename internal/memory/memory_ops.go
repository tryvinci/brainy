package memory

import (
	"context"
	"strings"
)

// Mem0-style memory update events (classic extract merge ops).
const (
	MemoryEventAdd    = "ADD"
	MemoryEventUpdate = "UPDATE"
	MemoryEventDelete = "DELETE"
	MemoryEventNone   = "NONE"
)

func normalizeMemoryEvent(raw string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case MemoryEventAdd, "CREATE", "INSERT":
		return MemoryEventAdd
	case MemoryEventUpdate, "CORRECT", "REPLACE":
		return MemoryEventUpdate
	case MemoryEventDelete, "REMOVE", "SUPPRESS":
		return MemoryEventDelete
	case MemoryEventNone, "SKIP", "NOOP":
		return MemoryEventNone
	default:
		return ""
	}
}

func MemoryEventOf(extracted ExtractedMemory) string {
	if extracted.Explain == nil {
		return MemoryEventAdd
	}
	if v, ok := extracted.Explain["memory_event"].(string); ok {
		if ev := normalizeMemoryEvent(v); ev != "" {
			return ev
		}
	}
	return MemoryEventAdd
}

func TargetMemoryIDOf(extracted ExtractedMemory) string {
	if extracted.Explain == nil {
		return ""
	}
	if v, ok := extracted.Explain["target_memory_id"].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

// PrepareExtractedForPersist mutates UPDATE extracts so BuildMemoryRecord
// carries supersedes_memory_id lineage. Returns false when the extract should
// not be upserted (NONE, or DELETE handled separately).
func PrepareExtractedForPersist(extracted *ExtractedMemory) (persist bool) {
	if extracted == nil {
		return false
	}
	switch MemoryEventOf(*extracted) {
	case MemoryEventNone, MemoryEventDelete:
		return false
	case MemoryEventUpdate:
		if tid := TargetMemoryIDOf(*extracted); tid != "" {
			if extracted.Explain == nil {
				extracted.Explain = map[string]any{}
			}
			extracted.Explain["assertion_kind"] = "corrective"
			// Stash on a side channel BuildMemoryRecord already understands via
			// metadata merge: copy into a synthetic metadata field after build.
			extracted.Explain["supersedes_memory_id"] = tid
		}
		return true
	default:
		return true
	}
}

// ApplyDeleteMemoryEvent suppresses the targeted prior memory.
func ApplyDeleteMemoryEvent(ctx context.Context, store Store, tenantID, subjectID string, extracted ExtractedMemory) error {
	if store == nil || MemoryEventOf(extracted) != MemoryEventDelete {
		return nil
	}
	tid := TargetMemoryIDOf(extracted)
	if tid == "" {
		return nil
	}
	return store.SuppressMemory(ctx, tenantID, subjectID, tid)
}

// ApplyIngestSupersession marks prior memory superseded when the new record
// declares supersedes_memory_id / SupersedesID (shared by sync + async).
func ApplyIngestSupersession(ctx context.Context, store Store, record MemoryRecord) error {
	if store == nil {
		return nil
	}
	priorID := supersedesMemoryIDFromMetadata(record.Metadata)
	if priorID == "" {
		priorID = strings.TrimSpace(record.SupersedesID)
	}
	if priorID == "" || priorID == record.MemoryID {
		return nil
	}
	return store.MarkSuperseded(ctx, record.TenantID, record.SubjectID, priorID)
}
