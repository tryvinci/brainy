package memory

import (
	"context"
	"strings"
	"time"
)

// MemoryRelation is a projection of an entity-valued atomic fact.
type MemoryRelation struct {
	TenantID   string
	SubjectID  string
	SrcEntity  string
	Relation   string
	DstEntity  string
	MemoryID   string
	ObservedAt *time.Time
}

// RelationIndexer persists Graphiti-style edges in Postgres (ADR-004).
type RelationIndexer interface {
	UpsertMemoryRelation(ctx context.Context, rel MemoryRelation) error
	ListRelationsFrom(ctx context.Context, tenantID, subjectID, srcEntity, relation string, limit int) ([]MemoryRelation, error)
}

// ProjectMemoryRelation turns an entity-valued atomic fact into a Postgres edge.
// Relations are a projection of the compiler, not a second extractor.
func ProjectMemoryRelation(record MemoryRecord) (MemoryRelation, bool) {
	pred, _ := record.Explain["predicate"].(string)
	val, _ := record.Explain["value_norm"].(string)
	if pred == "" && record.Metadata != nil {
		if p, ok := record.Metadata["predicate"].(string); ok {
			pred = p
		}
		if v, ok := record.Metadata["value_norm"].(string); ok {
			val = v
		}
	}
	pred = strings.TrimSpace(pred)
	val = strings.ToLower(strings.TrimSpace(val))
	if pred == "" || val == "" || utf8Len(val) < 3 {
		return MemoryRelation{}, false
	}
	switch pred {
	case PredicateOrigin, PredicateResidence, PredicateFamilyMember,
		PredicateActivity, PredicateMediaConsumed, PredicateOccupation,
		PredicateEducation, PredicatePlan, PredicateEvent, PredicateRelationshipStatus,
		PredicateIdentity, PredicatePreference:
	default:
		return MemoryRelation{}, false
	}
	src := ""
	if record.Explain != nil {
		if s, ok := record.Explain["subject"].(string); ok {
			src = strings.ToLower(strings.TrimSpace(s))
		}
	}
	if src == "" {
		ents := recordEntities(record)
		if len(ents) > 0 {
			src = ents[0]
		}
	}
	if src == "" {
		return MemoryRelation{}, false
	}
	return MemoryRelation{
		TenantID:   record.TenantID,
		SubjectID:  record.SubjectID,
		SrcEntity:  src,
		Relation:   pred,
		DstEntity:  val,
		MemoryID:   record.MemoryID,
		ObservedAt: record.ObservedAt,
	}, true
}

func utf8Len(s string) int {
	return len([]rune(s))
}
