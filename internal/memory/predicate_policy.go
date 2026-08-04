package memory

import "strings"

// TemporalPolicy for predicates (program §7.3) — not universal "latest wins".
type TemporalPolicy string

const (
	PolicySingleCurrentState  TemporalPolicy = "SINGLE_CURRENT_STATE"
	PolicyTemporalState       TemporalPolicy = "TEMPORAL_STATE"
	PolicyAppendOnlyEvent     TemporalPolicy = "APPEND_ONLY_EVENT"
	PolicyAccumulatingSet     TemporalPolicy = "ACCUMULATING_SET"
	PolicyScopedPreference    TemporalPolicy = "SCOPED_PREFERENCE"
	PolicyAuthorityResolved   TemporalPolicy = "AUTHORITY_RESOLVED_RULE"
	PolicyDerivedBelief       TemporalPolicy = "DERIVED_BELIEF"
)

// PredicatePolicy returns the temporal behavior for a predicate.
func PredicatePolicy(predicate string) TemporalPolicy {
	switch strings.TrimSpace(predicate) {
	case PredicateResidence, PredicateOrigin, PredicateRelationshipStatus, PredicateOccupation, PredicateIdentity:
		return PolicyTemporalState
	case PredicatePreference:
		return PolicyScopedPreference
	case PredicateActivity, PredicateMediaConsumed, PredicateEvent, PredicateFamilyMember:
		return PolicyAccumulatingSet
	case PredicateBelief:
		return PolicyDerivedBelief
	case PredicatePlan, PredicateSkill, PredicateAffiliation, PredicateHealth:
		return PolicyAccumulatingSet
	default:
		return PolicyAppendOnlyEvent
	}
}

// IsStatefulPredicate reports whether auto-supersession / latest-wins applies.
func IsStatefulPredicate(predicate string) bool {
	switch PredicatePolicy(predicate) {
	case PolicyTemporalState, PolicySingleCurrentState:
		return true
	default:
		return false
	}
}
