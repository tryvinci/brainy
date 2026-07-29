package memory

// Predicate taxonomy v1 (master-plan W2). Extensible per vertical pack.
// Product code must reference these constants — never benchmark answer keys.

const (
	PredicateIdentity           = "identity"
	PredicateRelationshipStatus = "relationship_status"
	PredicateOrigin             = "origin"
	PredicateResidence          = "residence"
	PredicateOccupation         = "occupation"
	PredicateEducation          = "education"
	PredicateFamilyMember       = "family_member"
	PredicateActivity           = "activity"
	PredicateActivityPurpose    = "activity_purpose"
	PredicateEvent              = "event"
	PredicateMediaConsumed      = "media_consumed"
	PredicatePreference         = "preference"
	PredicatePossession         = "possession"
	PredicateHealth             = "health"
	PredicatePlan               = "plan"
	PredicateBelief             = "belief"
	PredicateSkill              = "skill"
	PredicateAffiliation        = "affiliation"
	PredicateContactFact        = "contact_fact"
	PredicateMetric             = "metric"
)

// PilotPredicates are the five classes emitted by the deterministic typed-atom
// extractor in P1 (master-plan §9.3).
var PilotPredicates = []string{
	PredicateActivity,
	PredicateMediaConsumed,
	PredicateEvent,
	PredicateFamilyMember,
	PredicateOrigin,
}

// Atom is a typed (subject, predicate, value) memory fact with optional time.
type Atom struct {
	Subject    string
	Predicate  string
	Value      string
	Qualifier  string
	ObservedAt string
	ValidFrom  string
	ValidTo    string
	Provenance string
}
