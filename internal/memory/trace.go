package memory

import "strings"

// Failure taxonomy (program §16.2) — primary cause labels for eval/traces.
const (
	FailSourceMiss             = "SOURCE_MISS"
	FailWriteMiss              = "WRITE_MISS"
	FailRepresentationMiss     = "REPRESENTATION_MISS"
	FailEntityLinkMiss         = "ENTITY_LINK_MISS"
	FailRelationMiss           = "RELATION_MISS"
	FailRetrievalMiss          = "RETRIEVAL_MISS"
	FailEvidenceCoverageMiss   = "EVIDENCE_COVERAGE_MISS"
	FailTemporalResolutionMiss = "TEMPORAL_RESOLUTION_MISS"
	FailConflictResolutionMiss = "CONFLICT_RESOLUTION_MISS"
	FailPlanningMiss           = "PLANNING_MISS"
	FailProofMiss              = "PROOF_MISS"
	FailReaderMiss             = "READER_MISS"
	FailAbstentionMiss         = "ABSTENTION_MISS"
	FailJudgeMiss              = "JUDGE_MISS"
	FailHarnessError           = "HARNESS_ERROR"
)

// RepresentationStatus is Search's coverage of structured (non-episode) memory.
const (
	RepresentationEmpty    = "empty"
	RepresentationPartial  = "partial"
	RepresentationComplete = "complete"
)

// AnswerStatus values for POST /recall (program §12.11).
const (
	AnswerSupported          = "supported"
	AnswerPartiallySupported = "partially_supported"
	AnswerConflicted         = "conflicted"
	AnswerNotFound           = "not_found"
	AnswerInsufficient       = "insufficient_evidence"
	AnswerSuppressed         = "suppressed"
)

// QueryIntent labels from the deterministic planner (program §12.2).
const (
	IntentPointFact        = "point_fact"
	IntentCurrentState     = "current_state"
	IntentHistoricalState  = "historical_state"
	IntentTemporalSequence = "temporal_sequence"
	IntentEnumeration      = "enumeration"
	IntentAggregation      = "aggregation"
	IntentMultiHop         = "multi_hop"
	IntentPreference       = "preference"
	IntentProcedure        = "procedure"
	IntentOutcome          = "outcome"
	IntentBelief           = "belief"
	IntentProvenance       = "provenance"
	IntentAbstentionSens   = "abstention_sensitive"
)

// SearchTrace records retrieval diagnostics (program §15.4 / MEM-001).
type SearchTrace struct {
	CandidateOverfetch   int                `json:"candidate_overfetch,omitempty"`
	LexicalHits          int                `json:"lexical_hits,omitempty"`
	DenseAdmitted        int                `json:"dense_admitted,omitempty"`
	EntityHubAdmitted    int                `json:"entity_hub_admitted,omitempty"`
	AtomScanAdmitted     int                `json:"atom_scan_admitted,omitempty"`
	ListedSubject        bool               `json:"listed_subject,omitempty"`
	FusionV2             bool               `json:"fusion_v2,omitempty"`
	Intents              []string           `json:"intents,omitempty"`
	ChannelScores        map[string]float64 `json:"channel_scores,omitempty"`
	EpisodesDropped      int                `json:"episodes_dropped,omitempty"`
	EpisodeFallback      bool               `json:"episode_fallback,omitempty"`
	RepresentationStatus string             `json:"representation_status,omitempty"`
}

// AnalyzeQueryIntents is a deterministic intent classifier (Phase 4 starter).
func AnalyzeQueryIntents(query string) []string {
	tokens := tokenize(query)
	lower := strings.ToLower(query)
	out := make([]string, 0, 4)
	add := func(s string) {
		for _, e := range out {
			if e == s {
				return
			}
		}
		out = append(out, s)
	}
	if looksCountQuery(query) {
		add(IntentAggregation)
	} else if looksListQuery(tokens) {
		add(IntentEnumeration)
	}
	if looksPolarQuery(query) || looksMultiHopQuery(tokens) {
		add(IntentMultiHop)
	}
	if strings.Contains(lower, "when") || strings.Contains(lower, "before") ||
		strings.Contains(lower, "after") || strings.Contains(lower, "ago") ||
		strings.Contains(lower, "last year") || strings.Contains(lower, "yesterday") ||
		strings.Contains(lower, "how long") || strings.Contains(lower, "how old") {
		add(IntentTemporalSequence)
	}
	if strings.Contains(lower, "prefer") || strings.Contains(lower, "favorite") ||
		strings.Contains(lower, "like") || strings.Contains(lower, "dislike") {
		add(IntentPreference)
	}
	if strings.Contains(lower, "how do i") || strings.Contains(lower, "how to") ||
		strings.Contains(lower, "workflow") || strings.Contains(lower, "procedure") {
		add(IntentProcedure)
	}
	if strings.Contains(lower, "currently") || strings.Contains(lower, "right now") ||
		strings.Contains(lower, "current") {
		add(IntentCurrentState)
	}
	if strings.Contains(lower, "used to") || strings.Contains(lower, "previously") ||
		strings.Contains(lower, "before that") {
		add(IntentHistoricalState)
	}
	if len(out) == 0 {
		add(IntentPointFact)
	}
	return out
}

// WantsHistoricalRetrieval is true for before/after/when/used-to style intents.
func WantsHistoricalRetrieval(intents []string) bool {
	for _, intent := range intents {
		if intent == IntentHistoricalState || intent == IntentTemporalSequence {
			return true
		}
	}
	return false
}
