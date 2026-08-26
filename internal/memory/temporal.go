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

// TemporalScore is a [0,1] fusion channel for world-valid history vs current state.
// It is not a recency tie-break. Superseded rows score high only on historical intent.
func TemporalScore(record MemoryRecord, intents []string, includeHistorical bool) float64 {
	hist := includeHistorical || WantsHistoricalRetrieval(intents)
	current := false
	for _, intent := range intents {
		if intent == IntentCurrentState {
			current = true
		}
	}
	superseded := record.LifecycleState == LifecycleSuperseded
	memType := memoryTypeOf(record)
	switch {
	case hist && superseded:
		return 1.0
	case current && superseded:
		return 0
	case current && !superseded && memType == "state":
		return 0.7
	case hist && memType == "state":
		return 0.55
	case hist && record.ObservedAt != nil:
		return 0.45
	default:
		return 0
	}
}

func memoryTypeOf(record MemoryRecord) string {
	if record.Metadata != nil {
		if v, ok := record.Metadata["memory_type"].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	if record.Primitive == PrimitiveEpisode {
		return "episode"
	}
	pred := ""
	if record.Metadata != nil {
		pred, _ = record.Metadata["predicate"].(string)
	}
	if IsStatefulPredicate(pred) {
		return "state"
	}
	return "event"
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
	case looksConsequenceQuery(query):
		add(PredicateHealth)
		add(PredicatePreference)
	case queryHasToken(query, "visit", "visited", "travel", "travels", "traveled"):
		add(PredicateActivity)
		add(PredicateEvent)
		add(PredicateOrigin)
	case strings.Contains(lower, "moved") || strings.Contains(lower, "country") ||
		(strings.Contains(lower, "where") && strings.Contains(lower, "from")):
		add(PredicateOrigin)
		add(PredicateResidence)
	case strings.Contains(lower, "live") || strings.Contains(lower, "reside") || strings.Contains(lower, "city"):
		add(PredicateResidence)
		add(PredicateOrigin)
	case looksWhereQuery(query):
		add(PredicateActivity)
		add(PredicateEvent)
		add(PredicateResidence)
	case strings.Contains(lower, "job") || strings.Contains(lower, "work") || strings.Contains(lower, "occupation") || strings.Contains(lower, "career") || strings.Contains(lower, "pursue") || strings.Contains(lower, "educat"):
		add(PredicateOccupation)
		add(PredicateIdentity)
		add(PredicateEducation)
		add(PredicatePlan)
	case strings.Contains(lower, "married") || strings.Contains(lower, "relationship") ||
		(strings.Contains(lower, "partner") && !queryHasToken(query, "dogs", "dog", "pets", "pet")) ||
		strings.Contains(lower, "single"):
		add(PredicateRelationshipStatus)
	case queryHasToken(query, "trick", "tricks", "instrument", "instruments") ||
		queryHasToken(query, "teach", "taught"):
		add(PredicateSkill)
		add(PredicateActivity)
	case queryHasToken(query, "pet", "pets", "dog", "dogs"):
		add(PredicatePossession)
	case queryHasToken(query, "meal", "meals", "food", "snack", "snacks") ||
		queryHasToken(query, "given", "gave", "give") || strings.Contains(lower, "suggest"):
		add(PredicatePreference)
		add(PredicateActivity)
	case strings.Contains(lower, "injur") || queryHasToken(query, "health"):
		add(PredicateHealth)
		add(PredicateEvent)
	case (queryHasToken(query, "support", "supports", "supported") ||
		queryHasToken(query, "tell", "told")) && looksPlaceOrPersonSlot(query):
		add(PredicateFamilyMember)
	case strings.Contains(lower, "organization") || strings.Contains(lower, "beneficiar"):
		add(PredicateAffiliation)
	case strings.Contains(lower, "community") || strings.Contains(lower, "participat"):
		add(PredicateActivity)
		add(PredicateAffiliation)
		add(PredicateIdentity)
	case strings.Contains(lower, "journey") || queryHasToken(query, "changes", "change"):
		add(PredicateIdentity)
		add(PredicateActivity)
	case strings.Contains(lower, "name") || strings.Contains(lower, "who is") || strings.Contains(lower, "identity"):
		add(PredicateIdentity)
	case looksTriedPolarQuery(query):
		add(PredicateActivity)
		add(PredicatePreference)
	case strings.Contains(lower, "activit") || strings.Contains(lower, "hobby") || strings.Contains(lower, "hobbies") || strings.Contains(lower, "camp") || strings.Contains(lower, "unwind") || strings.Contains(lower, "relax") || strings.Contains(lower, "workshop") || strings.Contains(lower, "do to") ||
		queryHasToken(query, "tried", "try"):
		add(PredicateActivity)
		add(PredicateEvent)
	case strings.Contains(lower, "book") || strings.Contains(lower, "read") || strings.Contains(lower, "library"):
		add(PredicateMediaConsumed)
	case queryHasToken(query, "item", "items") &&
		(queryHasToken(query, "child", "childhood") || strings.Contains(lower, "as a child")):
		add(PredicatePossession)
	case strings.Contains(lower, "kid") || strings.Contains(lower, "child"):
		add(PredicateFamilyMember)
		add(PredicatePreference)
	case strings.Contains(lower, "when did") || strings.Contains(lower, "when was") || strings.Contains(lower, "when is"):
		add(PredicateEvent)
		add(PredicateActivity)
		add(PredicatePlan)
	case strings.Contains(lower, "plan") || strings.Contains(lower, "going"):
		add(PredicatePlan)
		add(PredicateEvent)
	case strings.Contains(lower, "research"):
		add(PredicatePlan)
	case strings.Contains(lower, "like") || strings.Contains(lower, "prefer") ||
		strings.Contains(lower, "favorite") || strings.Contains(lower, "enjoy") ||
		strings.Contains(lower, "love"):
		add(PredicatePreference)
		add(PredicateActivity)
	case strings.Contains(lower, "where") || strings.Contains(lower, "location") ||
		strings.Contains(lower, "practice"):
		add(PredicateActivity)
		add(PredicateEvent)
		add(PredicateResidence)
	case strings.Contains(lower, "own") || strings.Contains(lower, "possess") ||
		strings.Contains(lower, "bought") || strings.Contains(lower, "collectible") ||
		(strings.Contains(lower, "items") && (strings.Contains(lower, "bought") || strings.Contains(lower, "made") || strings.Contains(lower, "have") || strings.Contains(lower, "has"))) ||
		(strings.Contains(lower, "how many") && (strings.Contains(lower, "have") || strings.Contains(lower, "has") || strings.Contains(lower, "own"))):
		add(PredicatePossession)
	}
	if len(out) == 0 && (strings.Contains(lower, "currently") || strings.Contains(lower, "right now") || strings.Contains(lower, "current")) {
		// Conservative defaults for "currently" questions only.
		add(PredicateResidence)
		add(PredicateOccupation)
		add(PredicateRelationshipStatus)
		add(PredicateIdentity)
	}
	return out
}

func queryHasToken(query string, want ...string) bool {
	set := map[string]struct{}{}
	for _, w := range want {
		set[strings.ToLower(w)] = struct{}{}
	}
	for _, t := range tokenize(query) {
		t = strings.Trim(t, "'\"")
		t = strings.TrimSuffix(t, "'s")
		if _, ok := set[t]; ok {
			return true
		}
	}
	return false
}
