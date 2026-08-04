package memory

import (
	"context"
	"strings"
	"time"
)

// StateHistoryRow is one assertion in a predicate's world-valid history.
type StateHistoryRow struct {
	MemoryID   string     `json:"memory_id"`
	Value      string     `json:"value"`
	ValidFrom  *time.Time `json:"valid_from,omitempty"`
	ValidTo    *time.Time `json:"valid_to,omitempty"`
	RecordedAt *time.Time `json:"recorded_at,omitempty"`
	RetiredAt  *time.Time `json:"retired_at,omitempty"`
}

// TemporalStore provides typed temporal reads over atoms + current_state.
type TemporalStore interface {
	CurrentStateStore
	GetStateAsOf(ctx context.Context, tenantID, subjectID, predicate string, asOf time.Time) (memoryID, value string, ok bool, err error)
	GetStateAsKnownAt(ctx context.Context, tenantID, subjectID, predicate string, asKnownAt time.Time) (memoryID, value string, ok bool, err error)
	ListStateHistory(ctx context.Context, tenantID, subjectID, predicate string, limit int) ([]StateHistoryRow, error)
}

// shouldReplaceCurrentState decides whether incoming may overwrite the projected
// current value. Late-arriving older facts must not blind-win.
func shouldReplaceCurrentState(incoming, existing MemoryRecord) bool {
	if existing.MemoryID == "" {
		return true
	}
	if sid := supersedesMemoryIDFromMetadata(incoming.Metadata); sid != "" && sid == existing.MemoryID {
		return true
	}
	if ak := assertionKindOf(incoming); strings.EqualFold(ak, "corrective") {
		return true
	}
	inT := worldValidTime(incoming)
	exT := worldValidTime(existing)
	switch {
	case inT != nil && exT != nil:
		if inT.After(*exT) {
			return true
		}
		if inT.Before(*exT) {
			return false
		}
	case inT != nil && exT == nil:
		return true
	case inT == nil && exT != nil:
		return false
	}
	if incoming.Confidence > existing.Confidence+1e-9 {
		return true
	}
	if incoming.Confidence+1e-9 < existing.Confidence {
		return false
	}
	return !incoming.CreatedAt.Before(existing.CreatedAt)
}

func assertionKindOf(record MemoryRecord) string {
	if record.Explain != nil {
		if v, ok := record.Explain["assertion_kind"].(string); ok {
			return strings.TrimSpace(v)
		}
	}
	if record.Metadata != nil {
		if v, ok := record.Metadata["assertion_kind"].(string); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func worldValidTime(record MemoryRecord) *time.Time {
	if record.ObservedAt != nil && !record.ObservedAt.IsZero() {
		t := record.ObservedAt.UTC()
		return &t
	}
	if !record.CreatedAt.IsZero() {
		t := record.CreatedAt.UTC()
		return &t
	}
	return nil
}

// ParseAsOf parses RFC3339 / date-only as_of strings.
func ParseAsOf(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.UTC(), true
	}
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		return t.UTC(), true
	}
	return time.Time{}, false
}

// predicateHintsFromQuery picks stateful predicates likely relevant to a query.
func predicateHintsFromQuery(query string) []string {
	lower := strings.ToLower(query)
	var out []string
	add := func(p string) {
		for _, e := range out {
			if e == p {
				return
			}
		}
		out = append(out, p)
	}
	switch {
	case strings.Contains(lower, "live") || strings.Contains(lower, "reside") || strings.Contains(lower, "city") || strings.Contains(lower, "moved"):
		add(PredicateResidence)
		add(PredicateOrigin)
	case strings.Contains(lower, "job") || strings.Contains(lower, "work") || strings.Contains(lower, "occupation") || strings.Contains(lower, "nurse") || strings.Contains(lower, "engineer"):
		add(PredicateOccupation)
	case strings.Contains(lower, "married") || strings.Contains(lower, "relationship") || strings.Contains(lower, "partner"):
		add(PredicateRelationshipStatus)
	case strings.Contains(lower, "name") || strings.Contains(lower, "who is") || strings.Contains(lower, "identity"):
		add(PredicateIdentity)
	}
	if len(out) == 0 {
		// Conservative defaults for "currently" questions.
		add(PredicateResidence)
		add(PredicateOccupation)
		add(PredicateRelationshipStatus)
		add(PredicateIdentity)
	}
	return out
}
