package memory

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// representationBlobBudget is the oracle/fact dump cap. 80k left only ~15%
// headroom at observed LoCoMo volume; LME-scale subjects overflow.
const representationBlobBudget = 400000

// maxEnumerateAnswerItems bounds list answers. Counts keep the full typed set.
const maxEnumerateAnswerItems = 8

// RecallRequest is the product synthesis surface (master-plan W4 / program §14).
type RecallRequest struct {
	TenantID          string `json:"tenant_id"`
	SubjectID         string `json:"subject_id"`
	Query             string `json:"q"`
	Mode              string `json:"mode"` // context | enumerate | answer
	Vertical          string `json:"vertical,omitempty"`
	TopK              int    `json:"top_k,omitempty"`
	BudgetTokens      int    `json:"budget_tokens,omitempty"`
	IncludeHistorical bool   `json:"include_historical,omitempty"`
	View              string `json:"view,omitempty"` // current | historical | all
	AsOf              string `json:"as_of,omitempty"`
	IncludeProvenance bool   `json:"include_provenance,omitempty"`
	MaxEvidenceTokens int    `json:"max_evidence_tokens,omitempty"`
	CandidateLimit    int    `json:"candidate_limit,omitempty"`
	OracleMode        string `json:"oracle_mode,omitempty"` // reader | evidence | semantic | representation | retrieval | coverage | ""
}

// RecallItem is one enumerated value with evidence.
type RecallItem struct {
	Value      string   `json:"value"`
	Predicate  string   `json:"predicate,omitempty"`
	Evidence   []string `json:"evidence"`
	ObservedAt string   `json:"observed_at,omitempty"`
}

// RecallResponse is returned by POST /recall.
type RecallResponse struct {
	Mode         string         `json:"mode"`
	AnswerStatus string         `json:"answer_status,omitempty"`
	Intents      []string       `json:"intents,omitempty"`
	ContextBlock string         `json:"context_block,omitempty"`
	Answer       string         `json:"answer,omitempty"`
	Abstained    bool           `json:"abstained,omitempty"`
	Items        []RecallItem   `json:"items,omitempty"`
	Memories     []SearchResult `json:"memories"`
	Coverage     map[string]any `json:"coverage,omitempty"`
	Explain      map[string]any `json:"explain,omitempty"`
	Trace        *SearchTrace   `json:"trace,omitempty"`
}

// Recall assembles product-side context / enumeration / answer.
func (s *Service) Recall(ctx context.Context, req RecallRequest) (RecallResponse, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.SubjectID) == "" || strings.TrimSpace(req.Query) == "" {
		return RecallResponse{}, errors.New("tenant_id, subject_id, and q are required")
	}
	modeExplicit := strings.TrimSpace(req.Mode) != ""
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	topK := req.TopK
	if topK <= 0 {
		topK = 30
	}
	budget := req.BudgetTokens
	if req.MaxEvidenceTokens > 0 {
		budget = req.MaxEvidenceTokens
	} else if budget <= 0 {
		budget = 4000
	}

	intents := AnalyzeQueryIntents(req.Query)
	hist := req.IncludeHistorical || strings.EqualFold(req.View, "historical") || strings.EqualFold(req.View, "all") || WantsHistoricalRetrieval(intents)
	search, err := s.SearchOpt(ctx, req.TenantID, req.SubjectID, req.Vertical, "", req.Query, SearchOptions{
		IncludeHistorical: hist,
		IncludeEpisodes:   strings.EqualFold(req.View, "all"),
		Limit:             topK,
		CandidateLimit:    req.CandidateLimit,
	})
	if err != nil {
		return RecallResponse{}, err
	}

	plan := PlanQuery(req.Query, intents)
	if !modeExplicit {
		mode = plan.PreferredModeHint
		if mode == "" {
			mode = "context"
		}
	}
	out := RecallResponse{
		Mode:     mode,
		Memories: search.Results,
		Intents:  intents,
		Trace:    search.Trace,
		Explain: map[string]any{
			"top_k":              topK,
			"budget_tokens":      budget,
			"query_plan":         plan,
			"include_historical": hist,
		},
	}
	if req.View != "" {
		out.Explain["view"] = req.View
	}
	asOfTime, asOfOK := ParseAsOf(req.AsOf)
	if req.AsOf != "" {
		out.Explain["as_of"] = req.AsOf
		out.Explain["as_of_applied"] = asOfOK
		if !asOfOK {
			out.Explain["as_of_error"] = "unparseable"
		}
	}
	s.applyTemporalResolution(ctx, req, &out, asOfTime, asOfOK)
	temporalApplied := false
	if _, ok := out.Explain["temporal_applied"].(bool); ok {
		temporalApplied, _ = out.Explain["temporal_applied"].(bool)
	}
	if ta, _ := out.Explain["temporal_answer"].(string); ta != "" {
		temporalApplied = true
	}
	pkt := BuildEvidencePacket(plan, search.Results, out.Explain)
	bindPacketToTargets(&pkt, search.Results, req.Query, plan.CoverageTargets)
	var hopResults []HopResult
	// Typed hop executor: bind packet from hop joins when hops exist.
	if (plan.NeedsMultiHop || plan.NeedsEnumeration) && len(plan.Hops) > 0 {
		var byKey map[string]HopResult
		hopResults, byKey = s.executeTypedHops(ctx, req.TenantID, req.SubjectID, req.Vertical, hist, plan, topK, req.Query)
		bindPacketFromHopResults(&pkt, hopResults, byKey)
		out.Explain["hop_results"] = hopResults
		out.Explain["hop_join_proven"] = hopJoinProven(hopResults)
		// Merge hop memory hits into the search pool for synthesis context.
		extra := make([]SearchResult, 0)
		for _, hr := range hopResults {
			for i, id := range hr.MemoryIDs {
				content := ""
				if i < len(hr.Contents) {
					content = hr.Contents[i]
				} else if len(hr.Contents) > 0 {
					content = hr.Contents[0]
				}
				extra = append(extra, SearchResult{MemoryID: id, Content: content, Score: 0.85})
			}
		}
		if len(extra) > 0 {
			merged := mergePreferQueryCoverage(search.Results, extra, req.Query, topK)
			search.Results = merged
			out.Memories = merged
		}
		// Second pass only when join not yet proven — leftover query tokens
		// first, then unresolved hop keys. Entity-name probes are last.
		if plan.BudgetPasses >= 2 && !hopJoinProven(hopResults) {
			unc := uncoveredTargets(pkt)
			if len(unc) == 0 {
				unc = uncoveredQueryTokensFromResults(req.Query, search.Results)
			}
			if len(unc) > 0 {
				probe := nextHopProbe(plan, pkt)
				if tok := distinctiveProbeToken(unc); tok != "" {
					probe = tok
				}
				if probe != "" {
					if second, err := s.SearchOpt(ctx, req.TenantID, req.SubjectID, req.Vertical, "", probe, SearchOptions{
						IncludeHistorical: hist,
						Limit:             topK,
					}); err == nil && len(second.Results) > 0 {
						merged := mergePreferQueryCoverage(search.Results, second.Results, req.Query, topK)
						search.Results = merged
						out.Memories = merged
						hopResults2, byKey2 := s.executeTypedHops(ctx, req.TenantID, req.SubjectID, req.Vertical, hist, plan, topK, req.Query)
						pkt = BuildEvidencePacket(plan, merged, out.Explain)
						bindPacketFromHopResults(&pkt, hopResults2, byKey2)
						hopResults = hopResults2
						out.Explain["second_pass"] = map[string]any{
							"probe":           probe,
							"hit_count":       len(second.Results),
							"merged":          len(merged),
							"typed_hops":      plan.Hops,
							"hop_join_proven": hopJoinProven(hopResults2),
						}
						out.Explain["hop_results"] = hopResults2
						plan.Tools = append(plan.Tools, "second_pass")
					}
				}
			}
		}
	} else if plan.NeedsMultiHop && plan.BudgetPasses >= 2 {
		// Legacy lexical second pass when no typed hops were planned.
		if unc := uncoveredTargets(pkt); len(unc) > 0 {
			probe := nextHopProbe(plan, pkt)
			if tok := distinctiveProbeToken(unc); tok != "" {
				probe = tok
			}
			if probe != "" {
				if second, err := s.SearchOpt(ctx, req.TenantID, req.SubjectID, req.Vertical, "", probe, SearchOptions{
					IncludeHistorical: hist,
					Limit:             topK,
				}); err == nil && len(second.Results) > 0 {
					merged := mergePreferQueryCoverage(search.Results, second.Results, req.Query, topK)
					search.Results = merged
					out.Memories = merged
					pkt = BuildEvidencePacket(plan, merged, out.Explain)
					bindPacketToTargets(&pkt, merged, req.Query, plan.CoverageTargets)
					out.Explain["second_pass"] = map[string]any{
						"probe":      probe,
						"hit_count":  len(second.Results),
						"merged":     len(merged),
						"typed_hops": plan.Hops,
					}
					plan.Tools = append(plan.Tools, "second_pass")
				}
			}
		}
	}
	satisfied := packetCoverageSatisfied(plan, pkt)
	pkt.Coverage["satisfied"] = satisfied
	pkt.Coverage["hit_count"] = len(pkt.MemoryIDs)
	pkt.Coverage["item_count"] = len(pkt.Items)
	out.Explain["evidence_packet"] = pkt
	if out.Coverage == nil && pkt.Coverage != nil {
		out.Coverage = pkt.Coverage
	}
	oracle := strings.ToLower(strings.TrimSpace(req.OracleMode))
	if oracle != "" {
		out.Explain["oracle_mode"] = oracle
		switch oracle {
		case "evidence":
			if raw, ok := s.store.(RawEvidenceWriter); ok {
				rows, err := raw.ListEvidence(ctx, req.TenantID, req.SubjectID, topK)
				if err != nil {
					return RecallResponse{}, err
				}
				out.Explain["oracle_evidence_count"] = len(rows)
				out.ContextBlock = formatEvidenceOracle(rows, budget)
				out.AnswerStatus = AnswerSupported
				if len(rows) == 0 {
					out.AnswerStatus = AnswerNotFound
					out.Abstained = true
				}
				out.Coverage = map[string]any{"targets": 1, "satisfied": len(rows) > 0, "oracle": "evidence"}
				return out, nil
			}
			out.Explain["oracle_unsupported"] = true
			out.AnswerStatus = AnswerInsufficient
			out.Abstained = true
			out.Answer = "oracle_unsupported"
			return out, nil
		case "semantic", "representation":
			rep, err := s.collectRepresentationOracle(ctx, req, hist, topK)
			if err != nil {
				return RecallResponse{}, err
			}
			out.Explain["oracle_semantic"] = true
			out.Explain["oracle_representation"] = true
			out.Explain["oracle_atom_count"] = rep.atomCount
			out.Explain["oracle_fact_count"] = rep.factCount
			out.Explain["oracle_episode_count"] = rep.episodeCount
			out.Explain["oracle_memory_count"] = len(search.Results)
			out.Explain["oracle_fact_blob"] = rep.factBlob
			out.Explain["oracle_episode_blob"] = rep.episodeBlob
			if search.Trace != nil && search.Trace.RepresentationStatus != "" {
				out.Explain["oracle_representation_status"] = search.Trace.RepresentationStatus
			}
			present := rep.atomCount > 0 || rep.factCount > 0
			out.Coverage = map[string]any{"targets": 1, "satisfied": present, "oracle": oracle}
			if !present {
				out.AnswerStatus = AnswerNotFound
				out.Abstained = true
				out.Answer = "representation_absent"
				return out, nil
			}
			out.ContextBlock = assembleContextFromPacket(pkt, budget)
			if out.ContextBlock == "" {
				out.ContextBlock = rep.factBlob
			}
			out.AnswerStatus = AnswerSupported
			out.Answer = pickStructuredAnswer(req.Query, search.Results)
			if strings.TrimSpace(out.Answer) == "" {
				out.Answer = firstFactLine(rep.factBlob)
			}
			return out, nil
		case "retrieval":
			factN, epN := countSearchRepresentation(search.Results)
			out.Explain["oracle_retrieval"] = true
			out.Explain["oracle_memory_count"] = len(search.Results)
			out.Explain["oracle_fact_count"] = factN
			out.Explain["oracle_episode_count"] = epN
			retrieved := factN > 0 || (factN == 0 && epN > 0 && search.Trace != nil && search.Trace.RepresentationStatus == RepresentationEmpty)
			out.Coverage = map[string]any{"targets": 1, "satisfied": retrieved, "oracle": "retrieval"}
			if !retrieved {
				out.AnswerStatus = AnswerNotFound
				out.Abstained = true
				out.Answer = "retrieval_miss"
				return out, nil
			}
			out.ContextBlock = assembleContextFromPacket(pkt, budget)
			out.AnswerStatus = AnswerSupported
			out.Answer = pickStructuredAnswer(req.Query, search.Results)
			if strings.TrimSpace(out.Answer) == "" {
				out.Answer = firstFactLine(joinSearchFacts(search.Results))
			}
			return out, nil
		case "coverage":
			items := s.enumerateFromSearch(ctx, req, search.Results, hopResults)
			out.Items = items
			out.Explain["oracle_coverage"] = true
			out.Explain["oracle_item_count"] = len(items)
			proven := hopJoinProven(hopResults)
			if p, ok := pkt.Coverage["hop_join_proven"].(bool); ok && p {
				proven = true
			}
			slot := composeFromHopValues(hopResults) != "" || composeFromPacketStructuredValues(pkt, hopResults) != "" || pickStructuredAnswer(req.Query, search.Results) != ""
			satisfied := len(items) > 0 || proven || slot
			out.Coverage = map[string]any{"targets": 1, "satisfied": satisfied, "oracle": "coverage", "hop_join_proven": proven}
			if !satisfied {
				out.AnswerStatus = AnswerInsufficient
				out.Abstained = true
				out.Answer = "coverage_miss"
				return out, nil
			}
			out.ContextBlock = formatEnumerateContext(items)
			out.AnswerStatus = AnswerSupported
			out.Answer = out.ContextBlock
			return out, nil
		case "reader":
			out.Explain["oracle_reader"] = true
			mode = "answer"
			out.Mode = mode
		default:
			out.Explain["oracle_unsupported"] = true
			out.AnswerStatus = AnswerInsufficient
			out.Abstained = true
			out.Answer = "oracle_unsupported"
			return out, nil
		}
	}

	enumerated := false
	packetOK := packetCoverageSatisfied(plan, pkt)

	switch mode {
	case "enumerate":
		enumerated = true
		items := s.enumerateFromSearch(ctx, req, search.Results, hopResults)
		items = s.refineEnumeratedItems(ctx, req, items, hopResults)
		out.Items = items
		out.ContextBlock = formatEnumerateContext(items)
		out.Explain["item_count"] = len(items)
		out.AnswerStatus = AnswerSupported
		if len(items) == 0 {
			out.AnswerStatus = AnswerNotFound
			out.Abstained = true
		} else if looksSuperlativeQuery(req.Query) {
			if v := superlativeAnswer(items); v != "" {
				out.Answer = v
			}
		} else {
			vals := make([]string, 0, len(items))
			for _, it := range items {
				vals = append(vals, it.Value)
			}
			out.Answer = strings.Join(vals, ", ")
		}
		out.Coverage = map[string]any{
			"targets":   len(plan.CoverageTargets),
			"satisfied": len(items) > 0,
			"source":    "evidence_packet",
		}
	case "answer":
		out.ContextBlock = assembleContextFromPacket(pkt, budget)
		if (plan.NeedsEnumeration || looksListQuery(tokenize(req.Query)) || looksUnwindQuery(req.Query) || looksSuperlativeQuery(req.Query) || looksLocationListQuery(req.Query) || transferRecipient(req.Query) != "") && !looksConsequenceQuery(req.Query) && !looksWhereQuery(req.Query) {
			enumerated = true
			items := s.enumerateFromSearch(ctx, req, search.Results, hopResults)
			items = s.refineEnumeratedItems(ctx, req, items, hopResults)
			out.Items = items
			if len(items) > 0 {
				if looksSuperlativeQuery(req.Query) {
					if v := superlativeAnswer(items); v != "" {
						out.Answer = v
					}
				} else {
					vals := make([]string, 0, len(items))
					for _, it := range items {
						vals = append(vals, it.Value)
					}
					out.Answer = strings.Join(vals, ", ")
				}
				out.AnswerStatus = AnswerSupported
				if plan.NeedsMultiHop && !packetOK {
					out.AnswerStatus = AnswerPartiallySupported
				}
			}
		}
		if looksCountQuery(req.Query) {
			if !enumerated {
				enumerated = true
				out.Items = s.enumerateFromSearch(ctx, req, search.Results, hopResults)
			}
			out.Items = s.filterChildCountItems(ctx, req, out.Items, hopResults)
			if n := countAnswer(req.Query, out.Items); n != "" {
				out.Answer = n
				out.AnswerStatus = AnswerSupported
				out.Explain["count_answer"] = true
			}
		}
		if looksWhenEventQuery(req.Query) {
			if ans := s.dateAnswerFromHops(ctx, req, hopResults); ans != "" {
				out.Answer = ans
				out.AnswerStatus = AnswerSupported
				out.Explain["date_answer"] = true
			}
		}
		if looksWhereQuery(req.Query) {
			if looksLocationListQuery(req.Query) {
				loc := s.locationItemsFromEvidence(ctx, req, hopResults, nil)
				if len(loc) > 0 {
					vals := make([]string, 0, len(loc))
					for _, it := range loc {
						vals = append(vals, it.Value)
					}
					out.Items = loc
					out.Answer = strings.Join(vals, ", ")
					out.AnswerStatus = AnswerSupported
					out.Explain["where_answer"] = true
					enumerated = true
				}
			}
			if strings.TrimSpace(out.Answer) == "" {
				if ans := whereAnswerFromHops(req.Query, hopResults); ans != "" && !typedAnswerIsHopDump(ans) {
					out.Answer = ans
					out.AnswerStatus = AnswerSupported
					out.Explain["where_answer"] = true
				}
			}
		}
		if strings.TrimSpace(out.Answer) == "" && pkt.TemporalAnswer != "" && (plan.NeedsTemporal || temporalApplied) && !looksWhenEventQuery(req.Query) {
			out.Answer = pkt.TemporalAnswer
			out.AnswerStatus = AnswerSupported
			out.Coverage = map[string]any{"targets": 1, "satisfied": true, "source": "temporal_resolver"}
		}
		if strings.TrimSpace(out.Answer) == "" && looksPolarQuery(req.Query) {
			if ans := polarAnswerFromHops(req.Query, hopResults); ans != "" {
				out.Answer = ans
				out.AnswerStatus = AnswerSupported
				out.Explain["polar_answer"] = true
			}
		}
		if strings.TrimSpace(out.Answer) == "" && looksWhoQuery(req.Query) {
			if ans := whoAnswerFromHops(req.Query, hopResults); ans != "" {
				out.Answer = ans
				out.AnswerStatus = AnswerSupported
				out.Explain["who_answer"] = true
			}
		}
		blockSlotDump := looksLocationListQuery(req.Query) || looksPolarQuery(req.Query)
		if strings.TrimSpace(out.Answer) == "" && plan.NeedsMultiHop && hopComposeAllowed(req.Query) && !blockSlotDump {
			if composed := composeMultiHopAnswer(pkt); composed != "" {
				out.Answer = composed
				if packetOK {
					out.AnswerStatus = AnswerSupported
				} else {
					out.AnswerStatus = AnswerPartiallySupported
				}
				out.Explain["reader_source"] = "multihop_bridge_chain"
			}
		}
		if strings.TrimSpace(out.Answer) == "" && !blockSlotDump {
			if ans := pickStructuredAnswer(req.Query, search.Results); ans != "" {
				out.Answer = ans
				out.AnswerStatus = AnswerSupported
				out.Explain["structured_answer"] = true
			}
		}
		if strings.TrimSpace(out.Answer) == "" && !blockSlotDump {
			if ans := pickEpisodeFallback(req.Query, search.Results); ans != "" {
				out.Answer = ans
				out.AnswerStatus = AnswerSupported
				out.Explain["episode_fallback"] = true
			}
		}
		if strings.TrimSpace(out.Answer) == "" {
			out.Abstained = true
			out.Answer = "not in memory"
			out.AnswerStatus = AnswerInsufficient
		}
		if ans := ordinalNameFromPacket(req.Query, pkt); ans != "" {
			cur := strings.TrimSpace(out.Answer)
			if cur == "" || strings.EqualFold(cur, "not in memory") || typedAnswerIsHopDump(cur) || hopsAreIdentityOnly(hopResults) {
				out.Answer = ans
				out.Abstained = false
				out.AnswerStatus = AnswerSupported
				out.Explain["ordinal_name"] = true
			}
		}
		if plan.NeedsMultiHop && !packetOK && !out.Abstained {
			out.AnswerStatus = AnswerPartiallySupported
			out.Coverage = map[string]any{
				"targets":   len(plan.CoverageTargets),
				"satisfied": false,
				"source":    "evidence_packet",
			}
		} else if out.Coverage == nil {
			out.Coverage = map[string]any{
				"targets":   len(plan.CoverageTargets),
				"satisfied": !out.Abstained && packetOK,
				"source":    "evidence_packet",
			}
		}
	case "context":
		out.ContextBlock = assembleContextFromPacket(pkt, budget)
		if out.ContextBlock == "" && pkt.TemporalAnswer != "" {
			out.ContextBlock = pkt.TemporalAnswer
		}
		if out.ContextBlock == "" {
			out.AnswerStatus = AnswerNotFound
			out.Abstained = true
		} else {
			out.AnswerStatus = AnswerSupported
		}
		out.Coverage = map[string]any{
			"targets":   len(plan.CoverageTargets),
			"satisfied": !out.Abstained,
			"source":    "evidence_packet",
		}
	default:
		return RecallResponse{}, fmt.Errorf("unsupported recall mode %q", mode)
	}

	// Prefer temporal answer when synthesis was empty/weak.
	if temporalAnswer, _ := out.Explain["temporal_answer"].(string); temporalAnswer != "" {
		if mode == "answer" && (out.Abstained || out.AnswerStatus == AnswerPartiallySupported || strings.TrimSpace(out.Answer) == "" || strings.TrimSpace(out.Answer) == "not in memory") {
			out.Answer = temporalAnswer
			out.Abstained = false
			out.AnswerStatus = AnswerSupported
			out.Coverage = map[string]any{"targets": 1, "satisfied": true, "source": "temporal_resolver"}
			temporalApplied = true
		}
		if mode == "context" && out.ContextBlock == "" {
			out.ContextBlock = temporalAnswer
			out.Abstained = false
			out.AnswerStatus = AnswerSupported
			temporalApplied = true
		}
	}

	plan.Tools = ExecutedPlanTools(plan, temporalApplied, enumerated, out.Abstained)
	out.Explain["query_plan"] = plan
	out.Explain["tools_executed"] = plan.Tools
	out.Explain["reader_source"] = "evidence_packet"
	out.Explain["memory_count"] = len(search.Results)
	out.Explain["context_runes"] = utf8.RuneCountInString(out.ContextBlock)
	// Bounded hybrid LLM reader over the packet when composition is needed.
	// Enumerate mode must be reachable: the product harness sends list /
	// multi-evidence questions as mode=enumerate, and answer mode also
	// enumerates internally. Where/polar stay locked. Dates may be rewritten.
	// Typed counts, dual-entity lists, and enumerated lists that hybrid
	// does not shorten stay locked so dumps cannot replace a typed join.
	if mode == "answer" || mode == "enumerate" {
		hybrid := s.synthesizeHybridAnswer(ctx, req.Query, plan, pkt)
		if hybrid.Reason != "" {
			out.Explain["hybrid_reader_reason"] = hybrid.Reason
		}
		skipSlots := skipUnrelatedHopSlots(req.Query, hopResults, pkt)
		if skipSlots {
			out.Explain["hybrid_skipped_unrelated_slots"] = true
		}
		if hybrid.RawSnippet != "" {
			out.Explain["hybrid_reader_raw_prefix"] = hybrid.RawSnippet
		}
		if hybrid.Attempted {
			out.Explain["hybrid_reader_attempted"] = true
			if hybrid.ParseMode != "" {
				out.Explain["hybrid_reader_parse_mode"] = hybrid.ParseMode
			}
			if len(hybrid.SupportingIDs) > 0 {
				out.Explain["hybrid_supporting_memory_ids"] = hybrid.SupportingIDs
			}
			if len(hybrid.UnresolvedTargets) > 0 {
				out.Explain["hybrid_unresolved_targets"] = hybrid.UnresolvedTargets
			}
		}
		lockedDate := false
		lockedWhere := looksWhereQuery(req.Query) && out.Explain["where_answer"] == true
		lockedPolar := looksPolarQuery(req.Query) && out.Explain["polar_answer"] == true
		lockedCount := looksCountQuery(req.Query) && out.Explain["count_answer"] == true
		typedAnswer := strings.TrimSpace(out.Answer)
		typedItems := itemsFromCommaAnswer(typedAnswer)
		if len(typedItems) == 0 {
			typedItems = typedListItems(typedAnswer, out.Items)
		}
		typedN := len(typedItems)
		hybridN := len(itemsFromCommaAnswer(hybrid.Answer))
		extras := uncoveredHybridItemCount(typedItems, hybrid.Answer)
		lockedMHList := plan.NeedsMultiHop && len(hopQueryEntities(req.Query)) >= 2 && typedAnswer != "" && !strings.EqualFold(typedAnswer, "not in memory")
		lockedOrdinal := out.Explain["ordinal_name"] == true
		if typedAnswerIsHopDump(typedAnswer) || skipSlots {
			// Dual-entity SH questions often plan as MH and lock a slogan
			// dump or a leftover-unrelated short slot (Rocks) over hybrid.
			lockedWhere = false
			lockedMHList = false
		}
		// Keep typed multi-item lists when hybrid adds uncovered values as
		// another multi-item list, or expands a short typed list. A long
		// dump may still be replaced by a 1–2 item hybrid answer.
		lockedList := lockHybridListExtras(enumerated, typedN, hybridN, extras, typedAnswerIsHopDump(typedAnswer))
		if hybrid.Attempted {
			out.Explain["hybrid_pre_item_count"] = typedN
			out.Explain["hybrid_extra_item_count"] = extras
		}
		if hybrid.OK && (lockedWhere || lockedPolar || lockedCount || lockedMHList || lockedList || lockedOrdinal) {
			lock := "where"
			if lockedList {
				lock = "list"
			}
			if lockedMHList {
				lock = "mh_list"
			}
			if lockedPolar {
				lock = "polar"
			}
			if lockedCount {
				lock = "count"
			}
			if lockedOrdinal {
				lock = "ordinal"
			}
			out.Explain["hybrid_skipped_lock"] = lock
		}
		if hybrid.OK && !lockedDate && !lockedWhere && !lockedPolar && !lockedCount && !lockedMHList && !lockedList && !lockedOrdinal {
			out.Answer = strings.TrimSpace(hybrid.Answer)
			// Enumerated answers already have a typed list; hop-slot
			// grounding re-expands them into unrelated dumps. Unproven
			// search_fallback hops must not replace a hybrid answer either.
			if shouldGroundHybridToHops(req.Query, hopResults, pkt, enumerated) {
				grounded := groundToHopValues(hybrid.Answer, hopResults)
				out.Answer = grounded
				if composed := composeFromHopValues(hopResults); composed != "" && grounded == composed && grounded != strings.TrimSpace(hybrid.Answer) {
					out.Explain["hybrid_grounded_to_hops"] = true
				}
			} else if !enumerated && hopJoinProven(hopResults) && hopComposeAllowed(req.Query) && len(leftoverNonEntityQueryTokens(req.Query, hopResults)) > 0 {
				out.Explain["hybrid_skipped_uncovered_hop_ground"] = true
			}
			if mode == "enumerate" || enumerated {
				if items := itemsFromCommaAnswer(out.Answer); len(items) > 0 {
					out.Items = items
					out.Explain["item_count"] = len(items)
				}
			}
			out.Abstained = false
			out.AnswerStatus = hybridAnswerStatus(hybrid, plan, pkt, packetOK)
			out.Explain["reader_source"] = "hybrid_llm_packet"
			plan.Tools = append(plan.Tools, "hybrid_reader")
			out.Explain["query_plan"] = plan
			out.Explain["tools_executed"] = plan.Tools
		} else if hybrid.Abstain && !lockedDate && !lockedWhere && !lockedPolar && !lockedCount && !lockedMHList && !lockedList && !lockedOrdinal {
			leftover := leftoverNonEntityQueryTokens(req.Query, hopResults)
			canComposeHops := hopComposeAllowed(req.Query) && hopJoinProven(hopResults) &&
				!skipSlots &&
				(len(leftover) == 0 || hopsKeepTypedJoin(hopResults))
			typedKeep := strings.TrimSpace(out.Answer)
			keepTyped := typedKeep != "" && !strings.EqualFold(typedKeep, "not in memory") &&
				!typedAnswerIsHopDump(typedKeep) && !skipSlots && !hopsAreIdentityOnly(hopResults)
			if canComposeHops {
				if composed := composeFromHopValues(hopResults); hopComposeUsable(composed, hopResults) {
					out.Answer = composed
					out.Abstained = false
					out.AnswerStatus = AnswerSupported
					out.Explain["reader_source"] = "multihop_bridge_chain"
				} else if keepTyped {
					out.Abstained = false
					out.AnswerStatus = AnswerSupported
				} else {
					out.Abstained = true
					out.Answer = "not in memory"
					out.AnswerStatus = AnswerInsufficient
					out.Explain["reader_source"] = "hybrid_llm_packet"
				}
			} else if keepTyped {
				out.Abstained = false
				out.AnswerStatus = AnswerSupported
			} else if ta, _ := out.Explain["temporal_answer"].(string); strings.TrimSpace(ta) != "" && !strings.Contains(strings.ToLower(ta), "::") {
				out.Answer = ta
				out.Abstained = false
				out.AnswerStatus = AnswerSupported
				out.Explain["reader_source"] = "temporal_resolve"
			} else {
				out.Abstained = true
				out.Answer = "not in memory"
				out.AnswerStatus = AnswerInsufficient
				out.Explain["reader_source"] = "hybrid_llm_packet"
			}
		}
		if (out.Abstained || strings.EqualFold(strings.TrimSpace(out.Answer), "not in memory")) &&
			out.Explain["ordinal_name"] != true {
			if covering := leftoverCoveringSpecificAnswer(req.Query, hopResults, pkt); covering != "" {
				out.Answer = covering
				out.Abstained = false
				out.AnswerStatus = AnswerSupported
				out.Explain["reader_source"] = "leftover_packet_fact"
			}
		}
	}
	pkt.Plan = plan
	out.Explain["evidence_packet"] = pkt
	return out, nil
}

// listItemCount prefers structured items, then comma-split answer text.
func typedListItems(answer string, items []RecallItem) []RecallItem {
	out := make([]RecallItem, 0, len(items))
	for _, it := range items {
		if strings.TrimSpace(it.Value) != "" {
			out = append(out, it)
		}
	}
	if len(out) > 0 {
		return out
	}
	return itemsFromCommaAnswer(answer)
}

func itemCoveredByTyped(value string, typed []RecallItem) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return true
	}
	for _, t := range typed {
		tv := strings.ToLower(strings.TrimSpace(t.Value))
		if tv == "" {
			continue
		}
		if tv == v || strings.Contains(tv, v) || strings.Contains(v, tv) {
			return true
		}
	}
	return false
}

func uncoveredHybridItemCount(typed []RecallItem, hybrid string) int {
	n := 0
	for _, h := range itemsFromCommaAnswer(hybrid) {
		if !itemCoveredByTyped(h.Value, typed) {
			n++
		}
	}
	return n
}

// lockHybridListExtras keeps a typed list when hybrid injects uncovered
// values as another multi-item list, or expands a short typed list. A long
// dump may still be replaced by a 1–2 item hybrid answer.
func lockHybridListExtras(enumerated bool, typedN, hybridN, extras int, typedDump bool) bool {
	if !enumerated || typedN <= 0 {
		return false
	}
	if extras > 0 {
		return (typedN >= 3 && hybridN >= 3) || hybridN > typedN
	}
	// Hybrid dropped items from a real typed list (not a slogan dump).
	return !typedDump && typedN >= 3 && hybridN > 0 && hybridN < typedN
}

func shouldGroundHybridToHops(query string, hops []HopResult, pkt EvidencePacket, enumerated bool) bool {
	if enumerated || !hopComposeAllowed(query) || !hopJoinProven(hops) {
		return false
	}
	if skipUnrelatedHopSlots(query, hops, pkt) {
		return false
	}
	// Proven hops that never mention leftover distinctive query tokens are
	// slogan dumps (visa/travel identity, occupation lists). Do not replace
	// a covering hybrid answer with those slots.
	if len(leftoverNonEntityQueryTokens(query, hops)) > 0 {
		return false
	}
	return true
}

func listItemCount(answer string, items []RecallItem) int {
	n := 0
	for _, it := range items {
		if strings.TrimSpace(it.Value) != "" {
			n++
		}
	}
	if n > 0 {
		return n
	}
	return len(itemsFromCommaAnswer(answer))
}

func itemsFromCommaAnswer(answer string) []RecallItem {
	answer = strings.TrimSpace(answer)
	if answer == "" || strings.EqualFold(answer, "not in memory") {
		return nil
	}
	parts := strings.Split(answer, ",")
	out := make([]RecallItem, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		v := strings.TrimSpace(part)
		if v == "" {
			continue
		}
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, RecallItem{Value: v})
	}
	return out
}

// hybridAnswerStatus maps hybrid outcomes to truthful AnswerStatus values.
func hybridAnswerStatus(hybrid hybridReaderResult, plan QueryPlan, pkt EvidencePacket, packetOK bool) string {
	if hybrid.Abstain || strings.TrimSpace(hybrid.Answer) == "" {
		return AnswerInsufficient
	}
	if evidenceConflicted(pkt) {
		return AnswerConflicted
	}
	unresolved := len(hybrid.UnresolvedTargets) > 0
	if unc, ok := pkt.Coverage["uncovered"].([]string); ok && len(unc) > 0 {
		unresolved = true
	}
	hopMissing := false
	if plan.NeedsMultiHop && len(plan.Hops) > 0 {
		if proven, _ := pkt.Coverage["hop_join_proven"].(bool); !proven {
			hopMissing = true
		}
	}
	if unresolved || hopMissing || (plan.NeedsMultiHop && !packetOK) {
		return AnswerPartiallySupported
	}
	return AnswerSupported
}

func evidenceConflicted(pkt EvidencePacket) bool {
	// Same predicate with distinct values across packet items → conflicted.
	byPred := map[string]string{}
	for _, it := range pkt.Items {
		pred := strings.ToLower(strings.TrimSpace(it.Predicate))
		if pred == "" {
			continue
		}
		val := strings.ToLower(NormalizeText(it.Content))
		if val == "" {
			continue
		}
		if prev, ok := byPred[pred]; ok && prev != val && !strings.Contains(prev, val) && !strings.Contains(val, prev) {
			return true
		}
		byPred[pred] = val
	}
	return false
}

// applyTemporalResolution fills explain + optional temporal_answer from typed
// current-state / as-of / history reads when the store supports TemporalStore.
func (s *Service) applyTemporalResolution(ctx context.Context, req RecallRequest, out *RecallResponse, asOf time.Time, asOfOK bool) {
	ts, ok := s.store.(TemporalStore)
	if !ok {
		// CurrentStateStore alone still helps view=current.
		if cs, ok := s.store.(CurrentStateStore); ok {
			s.applyCurrentStateOnly(ctx, req, out, cs)
		}
		return
	}
	view := strings.ToLower(strings.TrimSpace(req.View))
	wantCurrent := view == "current" || view == "" && hasIntent(out.Intents, IntentCurrentState)
	wantHistory := req.IncludeHistorical || view == "historical" || view == "all" || hasIntent(out.Intents, IntentHistoricalState) || hasIntent(out.Intents, IntentTemporalSequence)
	preds := predicateHintsFromQuery(req.Query)
	// Also try entity-scoped keys when the query names people (multi-speaker).
	entities := nameLikeTokens(contentBearingTokens(tokenize(req.Query)))
	scoped := make([]string, 0, len(preds)*len(entities))
	for _, ent := range entities {
		for _, pred := range preds {
			scoped = append(scoped, statePredicateKey(ent, pred))
		}
	}
	preds = append(preds, scoped...)
	resolved := make([]map[string]any, 0, len(preds))
	var answerParts []string

	seenPred := map[string]struct{}{}
	for _, pred := range preds {
		if _, ok := seenPred[pred]; ok {
			continue
		}
		seenPred[pred] = struct{}{}
		entry := map[string]any{"predicate": pred}
		switch {
		case asOfOK && strings.EqualFold(view, "as_known_at"):
			id, val, found, _ := ts.GetStateAsKnownAt(ctx, req.TenantID, req.SubjectID, pred, asOf)
			entry["mode"] = "as_known_at"
			if found {
				entry["memory_id"] = id
				entry["value"] = val
				answerParts = append(answerParts, pred+": "+val)
			}
		case asOfOK:
			id, val, found, _ := ts.GetStateAsOf(ctx, req.TenantID, req.SubjectID, pred, asOf)
			entry["mode"] = "as_of"
			if found {
				entry["memory_id"] = id
				entry["value"] = val
				answerParts = append(answerParts, pred+": "+val)
			}
		case wantCurrent || view == "current":
			id, val, policy, found, _ := ts.GetCurrentState(ctx, req.TenantID, req.SubjectID, pred)
			entry["mode"] = "current"
			entry["policy"] = policy
			if found {
				entry["memory_id"] = id
				entry["value"] = val
				answerParts = append(answerParts, pred+": "+val)
			}
		}
		if wantHistory {
			hist, err := ts.ListStateHistory(ctx, req.TenantID, req.SubjectID, pred, 8)
			if err == nil && len(hist) > 0 {
				entry["history"] = hist
			}
		}
		if entry["value"] != nil || entry["history"] != nil {
			resolved = append(resolved, entry)
		}
	}
	if len(resolved) > 0 {
		out.Explain["temporal"] = resolved
		out.Explain["temporal_applied"] = true
	}
	if len(answerParts) > 0 {
		out.Explain["temporal_answer"] = strings.Join(answerParts, "; ")
	}
}

func (s *Service) applyCurrentStateOnly(ctx context.Context, req RecallRequest, out *RecallResponse, cs CurrentStateStore) {
	view := strings.ToLower(strings.TrimSpace(req.View))
	if view != "current" && !hasIntent(out.Intents, IntentCurrentState) {
		return
	}
	preds := predicateHintsFromQuery(req.Query)
	resolved := make([]map[string]any, 0, len(preds))
	var parts []string
	for _, pred := range preds {
		id, val, policy, found, _ := cs.GetCurrentState(ctx, req.TenantID, req.SubjectID, pred)
		if !found {
			continue
		}
		resolved = append(resolved, map[string]any{
			"predicate": pred, "memory_id": id, "value": val, "policy": policy, "mode": "current",
		})
		parts = append(parts, pred+": "+val)
	}
	if len(resolved) > 0 {
		out.Explain["temporal"] = resolved
		out.Explain["temporal_applied"] = true
		out.Explain["temporal_answer"] = strings.Join(parts, "; ")
	}
}

func hasIntent(intents []string, want string) bool {
	for _, i := range intents {
		if i == want {
			return true
		}
	}
	return false
}

func (s *Service) enumerateFromSearch(ctx context.Context, req RecallRequest, results []SearchResult, hops []HopResult) []RecallItem {
	tokens := tokenize(req.Query)
	pred := predicateFromListQuery(tokens)
	if pred == "" {
		if hints := predicateHintsFromQuery(req.Query); len(hints) > 0 {
			pred = hints[0]
		}
	}
	seen := map[string]RecallItem{}
	order := make([]string, 0)
	entity := ""
	if ents := hopQueryEntities(req.Query); len(ents) > 0 {
		entity = ents[0]
	} else {
		entity = hopEntityName(nameLikeTokens(contentBearingTokens(tokens)))
	}
	for _, h := range hops {
		for _, d := range h.DependsOn {
			if d == "e_rel" && strings.TrimSpace(h.Entity) != "" {
				entity = h.Entity
			}
		}
	}

	add := func(value, predicate, memoryID, observed string) {
		v := strings.TrimSpace(value)
		if hasSlotTemplate(v) {
			if extracted, ok := slotValueFromMemoryContent(v); ok {
				v = extracted
			}
		}
		if v == "" || anaphoricSlotValue(v) || !validEnumeratedValue(v) {
			return
		}
		if predicate == PredicateFamilyMember || predicate == PredicatePreference {
			head, _, _ := strings.Cut(strings.ToLower(v), " ")
			if _, stop := preferenceHeadStop[head]; stop {
				return
			}
		}
		key := strings.ToLower(v)
		if existing, ok := seen[key]; ok {
			existing.Evidence = appendUnique(existing.Evidence, memoryID)
			seen[key] = existing
			return
		}
		seen[key] = RecallItem{
			Value:      v,
			Predicate:  predicate,
			Evidence:   []string{memoryID},
			ObservedAt: observed,
		}
		order = append(order, key)
	}

	join := len(hopQueryEntities(req.Query)) >= 2
	if join {
		shared := hopSharedSlotValues(hops)
		if looksCommunityQuery(req.Query) {
			have := map[string]struct{}{}
			for _, v := range shared {
				have[strings.ToLower(v)] = struct{}{}
			}
			addShared := func(v string) {
				key := strings.ToLower(strings.TrimSpace(v))
				if key == "" {
					return
				}
				if _, ok := have[key]; ok {
					return
				}
				have[key] = struct{}{}
				shared = append(shared, v)
			}
			for _, v := range intersectHopValuesByContainment(hops) {
				addShared(v)
			}
			for _, v := range hopValuesMentioningPartner(hops) {
				addShared(v)
			}
		}
		if len(shared) == 0 {
			shared = hopSharedContentValues(hops)
		}
		slotPred := pred
		for _, h := range hops {
			if h.Kind == "resolve_entity" || h.Source == "unresolved" || h.Source == "search_fallback" {
				continue
			}
			if hopUsefulForList(h.Predicate, pred) && strings.TrimSpace(h.Predicate) != "" {
				slotPred = firstNonEmpty(h.Predicate, pred)
				break
			}
		}
		for _, v := range shared {
			add(v, slotPred, hopMemoryIDForValue(hops, v), "")
		}
	} else {
		for _, h := range hops {
			if h.Kind == "resolve_entity" || h.Source == "unresolved" || h.Source == "search_fallback" {
				continue
			}
			if !hopUsefulForList(h.Predicate, pred) {
				continue
			}
			slotPred := firstNonEmpty(h.Predicate, pred)
			if len(h.Values) > 0 {
				for i, v := range h.Values {
					if hopValueIsAttendedEvent(v) || hopValueHasForeignPossessive(v, h.Entity, hops) {
						continue
					}
					id := ""
					if i < len(h.MemoryIDs) {
						id = h.MemoryIDs[i]
					} else if len(h.MemoryIDs) > 0 {
						id = h.MemoryIDs[0]
					}
					add(v, slotPred, id, "")
				}
				continue
			}
			if v := strings.TrimSpace(h.Value); v != "" && utf8.RuneCountInString(v) <= 80 {
				add(v, slotPred, firstID(h.MemoryIDs), "")
			}
		}
	}

	// Prefer indexed atoms when hops missed or to fill additional values.
	// Multi-entity joins must not refill the union from the first entity's atoms.
	preds := make([]string, 0, 2)
	if pred != "" {
		preds = append(preds, pred)
		switch pred {
		case PredicateFamilyMember:
			preds = append(preds, PredicatePreference)
		case PredicateOccupation:
			preds = append(preds, PredicateIdentity)
		case PredicatePossession:
			preds = append(preds, PredicateIdentity)
		case PredicateSkill:
			preds = append(preds, PredicateActivity)
		}
		if looksCommunityQuery(req.Query) && pred == PredicateActivity {
			preds = append(preds, PredicateAffiliation)
		}
	}
	// Hops already listed typed values. Refilling the atom index dumps every
	// activity/preference onto list answers that already had a slot.
	if !join && len(preds) > 0 && len(order) == 0 {
		if indexer, ok := s.store.(AtomIndexer); ok {
			ids := make([]string, 0, 40)
			seenID := map[string]struct{}{}
			for _, p := range preds {
				got, err := indexer.ListAtomMemoryIDs(ctx, req.TenantID, req.SubjectID, p, "", 40)
				if err != nil {
					continue
				}
				for _, id := range got {
					if _, ok := seenID[id]; ok {
						continue
					}
					seenID[id] = struct{}{}
					ids = append(ids, id)
				}
			}
			byID := map[string]SearchResult{}
			for _, r := range results {
				byID[r.MemoryID] = r
			}
			for _, id := range ids {
				rec, err := s.store.GetMemory(ctx, req.TenantID, req.SubjectID, id)
				content := ""
				obs := ""
				if err == nil {
					content = rec.Content
					if rec.ObservedAt != nil {
						obs = rec.ObservedAt.UTC().Format(time.RFC3339)
					}
					if entity != "" && !memoryMentionsEntity(rec, entity) {
						continue
					}
				} else if r, ok := byID[id]; ok {
					content = r.Content
					if r.ObservedAt != nil {
						obs = r.ObservedAt.UTC().Format(time.RFC3339)
					}
				}
				if content == "" {
					continue
				}
				val, ok := slotValueFromMemoryContent(content)
				if !ok {
					continue
				}
				add(val, pred, id, obs)
			}
		}
	}

	if !join && len(order) == 0 {
		for _, r := range results {
			if searchResultIsEpisode(r) {
				continue
			}
			if entity != "" && !strings.Contains(strings.ToLower(r.Content), strings.ToLower(entity)) {
				continue
			}
			recPred := searchResultPredicate(r)
			if pred != "" && recPred != "" && !hopUsefulForList(recPred, pred) {
				continue
			}
			obs := ""
			if r.ObservedAt != nil {
				obs = r.ObservedAt.UTC().Format(time.RFC3339)
			}
			val := structuredValueOf(r)
			if val == "" {
				continue
			}
			p := recPred
			if p == "" {
				p = pred
			}
			if p == "" {
				p = "fact"
			}
			add(val, p, r.MemoryID, obs)
		}
	}

	items := make([]RecallItem, 0, len(order))
	for _, key := range order {
		items = append(items, seen[key])
	}
	if looksLocationListQuery(req.Query) {
		return s.locationItemsFromEvidence(ctx, req, hops, items)
	}
	return items
}

func looksLocationListQuery(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if strings.Contains(q, "location") {
		return true
	}
	if !strings.Contains(q, "practice") {
		return false
	}
	return strings.Contains(q, " at") || strings.HasPrefix(q, "where ")
}

// practiceObjectTokens is the noun after practice/practices until at/in/near.
func practiceObjectTokens(query string) []string {
	lower := strings.ToLower(query)
	cue := ""
	idx := -1
	for _, c := range []string{" practices ", " practice ", " practicing ", " practised ", " practiced "} {
		i := strings.Index(lower, c)
		if i < 0 {
			continue
		}
		if idx < 0 || i < idx {
			idx = i
			cue = c
		}
	}
	if idx < 0 {
		return nil
	}
	rest := query[idx+len(cue):]
	restLower := strings.ToLower(rest)
	cut := len(rest)
	for _, stop := range []string{" at", " in", " near", " on", "?"} {
		if k := strings.Index(restLower, stop); k >= 0 && k < cut {
			cut = k
		}
	}
	toks := contentBearingTokens(tokenize(rest[:cut]))
	if len(toks) == 0 {
		return nil
	}
	return toks
}

func compositionalPracticePlace(content string, focus []string) string {
	if strings.TrimSpace(content) == "" || len(focus) == 0 {
		return ""
	}
	fields := strings.Fields(content)
	for _, obj := range focus {
		obj = strings.ToLower(strings.TrimSpace(obj))
		if obj == "" {
			continue
		}
		for i, raw := range fields {
			w := strings.ToLower(strings.Trim(raw, ".,;:!?\"'"))
			if w != obj || i+1 >= len(fields) {
				continue
			}
			nextRaw := fields[i+1]
			next := strings.ToLower(strings.Trim(nextRaw, ".,;:!?\"'"))
			if next == "" || next == obj || isQueryStopword(next) || next == "practice" || next == "practices" || next == "practicing" || next == "practised" || next == "practiced" {
				continue
			}
			if utf8.RuneCountInString(next) < 3 {
				continue
			}
			if strings.HasSuffix(next, "ing") {
				continue
			}
			if looksHopPerson(strings.Trim(nextRaw, ".,;:!?\"'")) {
				continue
			}
			if i == 0 || !strings.EqualFold(strings.Trim(fields[i-1], ".,;:!?\"'"), "the") {
				continue
			}
			return "the " + obj + " " + next
		}
	}
	return ""
}

func (s *Service) locationItemsFromEvidence(ctx context.Context, req RecallRequest, hops []HopResult, items []RecallItem) []RecallItem {
	seen := map[string]RecallItem{}
	order := make([]string, 0)
	focus := practiceObjectTokens(req.Query)
	add := func(p, id string) {
		p = strings.TrimSpace(p)
		if p == "" || anaphoricSlotValue(p) || !validEnumeratedValue(p) {
			return
		}
		if placeEqualsAny(p, focus) {
			return
		}
		key := strings.ToLower(p)
		if existing, ok := seen[key]; ok {
			existing.Evidence = appendUnique(existing.Evidence, id)
			seen[key] = existing
			return
		}
		ev := []string{}
		if id != "" {
			ev = []string{id}
		}
		seen[key] = RecallItem{
			Value:     titleCaseWords(p),
			Predicate: PredicateActivity,
			Evidence:  ev,
		}
		order = append(order, key)
	}
	addFrom := func(blob, id string, requireFocus bool) {
		if strings.TrimSpace(blob) == "" {
			return
		}
		if requireFocus && len(focus) > 0 && !itemHitsExclusion(blob, focus) {
			return
		}
		for _, p := range placesFromContent(blob) {
			add(p, id)
		}
		if p := compositionalPracticePlace(blob, focus); p != "" {
			add(p, id)
		}
	}
	scan := func(requireFocus bool) {
		for _, h := range hops {
			if h.Kind == "resolve_entity" || h.Source == "unresolved" {
				continue
			}
			if h.Source == "search_fallback" && !requireFocus {
				continue
			}
			if h.Predicate != "" && h.Predicate != PredicateActivity && h.Predicate != PredicateEvent && h.Predicate != PredicateResidence {
				continue
			}
			for i, c := range h.Contents {
				id := ""
				if i < len(h.MemoryIDs) {
					id = h.MemoryIDs[i]
				} else if len(h.MemoryIDs) > 0 {
					id = h.MemoryIDs[0]
				}
				addFrom(c, id, requireFocus)
			}
		}
		for _, it := range items {
			blob := s.itemEvidenceBlob(ctx, req, it, hops)
			id := ""
			if len(it.Evidence) > 0 {
				id = it.Evidence[0]
			}
			addFrom(blob, id, requireFocus)
		}
	}
	scan(len(focus) > 0)
	if len(order) == 0 && len(focus) > 0 {
		scan(false)
	}
	out := make([]RecallItem, 0, len(order))
	for _, key := range order {
		out = append(out, seen[key])
	}
	return out
}

func hopMemoryIDForValue(hops []HopResult, value string) string {
	want := strings.ToLower(strings.TrimSpace(value))
	if want == "" {
		return ""
	}
	for _, h := range hops {
		if h.Kind == "resolve_entity" || h.Source == "unresolved" || h.Source == "search_fallback" {
			continue
		}
		for i, v := range h.Values {
			if strings.ToLower(strings.TrimSpace(v)) != want {
				continue
			}
			if i < len(h.MemoryIDs) && h.MemoryIDs[i] != "" {
				return h.MemoryIDs[i]
			}
			if len(h.MemoryIDs) > 0 {
				return h.MemoryIDs[0]
			}
		}
		if strings.ToLower(strings.TrimSpace(h.Value)) == want && len(h.MemoryIDs) > 0 {
			return h.MemoryIDs[0]
		}
	}
	return ""
}

func hopUsefulForList(hopPred, listPred string) bool {
	if listPred == "" || hopPred == "" || strings.EqualFold(hopPred, listPred) {
		return true
	}
	switch listPred {
	case PredicateOccupation:
		return hopPred == PredicateIdentity || hopPred == PredicateEducation || hopPred == PredicatePlan
	case PredicateFamilyMember:
		return hopPred == PredicatePreference
	case PredicatePreference:
		return hopPred == PredicateFamilyMember || hopPred == PredicateActivity
	case PredicateActivity:
		return hopPred == PredicateEvent || hopPred == PredicatePreference || hopPred == PredicateSkill || hopPred == PredicateAffiliation
	case PredicatePossession:
		return hopPred == PredicateIdentity
	case PredicateSkill:
		return hopPred == PredicateActivity
	case PredicateHealth:
		return hopPred == PredicateEvent
	case PredicateAffiliation:
		return hopPred == PredicateEvent || hopPred == PredicateFamilyMember
	case PredicateEvent:
		return hopPred == PredicatePlan || hopPred == PredicateActivity
	case PredicatePlan:
		return hopPred == PredicateEvent
	}
	return false
}

func countHeadNoun(query string) string {
	lower := strings.ToLower(query)
	key := "how many "
	i := strings.Index(lower, key)
	if i < 0 {
		key = "how much "
		i = strings.Index(lower, key)
	}
	if i < 0 {
		return ""
	}
	rest := strings.TrimSpace(query[i+len(key):])
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return ""
	}
	return strings.Trim(strings.ToLower(fields[0]), "?,.!'\"")
}

func countAnswer(query string, items []RecallItem) string {
	if len(items) == 0 {
		return ""
	}
	head := countHeadNoun(query)
	if head == "times" || head == "time" {
		seen := map[string]struct{}{}
		n := 0
		for _, it := range items {
			if len(it.Evidence) == 0 {
				n++
				continue
			}
			for _, id := range it.Evidence {
				if id == "" {
					continue
				}
				if _, ok := seen[id]; ok {
					continue
				}
				seen[id] = struct{}{}
				n++
			}
		}
		if n == 0 {
			return strconv.Itoa(len(items))
		}
		return strconv.Itoa(n)
	}
	switch head {
	case "", "children", "kids", "people":
		return strconv.Itoa(len(items))
	}
	stem := strings.TrimSuffix(head, "es")
	stem = strings.TrimSuffix(stem, "s")
	if len(stem) < 3 {
		stem = head
	}
	n := 0
	for _, it := range items {
		v := strings.ToLower(it.Value)
		if strings.Contains(v, head) || strings.Contains(v, stem) {
			n++
		}
	}
	if n == 0 {
		return strconv.Itoa(len(items))
	}
	return strconv.Itoa(n)
}

func polarClaimTokens(query string) []string {
	skip := map[string]struct{}{
		"tried": {}, "try": {}, "teach": {}, "taught": {}, "learned": {}, "learn": {},
		"himself": {}, "herself": {}, "myself": {}, "play": {}, "playing": {},
		"how": {}, "kind": {},
	}
	ents := map[string]struct{}{}
	for _, e := range hopQueryEntities(query) {
		ents[strings.ToLower(e)] = struct{}{}
	}
	out := make([]string, 0, 4)
	for _, t := range contentBearingTokens(tokenize(query)) {
		if _, ok := skip[t]; ok {
			continue
		}
		if _, ok := ents[t]; ok {
			continue
		}
		if len(t) < 4 {
			continue
		}
		out = append(out, t)
	}
	return out
}

func polarAnswerFromHops(query string, hops []HopResult) string {
	if len(hopQueryEntities(query)) == 0 {
		return ""
	}
	claim := polarClaimTokens(query)
	if len(claim) == 0 {
		return ""
	}
	var b strings.Builder
	for _, h := range hops {
		if h.Kind == "resolve_entity" || h.Source == "unresolved" || h.Source == "search_fallback" {
			continue
		}
		for _, v := range h.Values {
			b.WriteString(" ")
			b.WriteString(strings.ToLower(v))
		}
		if h.Value != "" {
			b.WriteString(" ")
			b.WriteString(strings.ToLower(h.Value))
		}
		for _, c := range h.Contents {
			b.WriteString(" ")
			b.WriteString(strings.ToLower(c))
		}
	}
	blob := b.String()
	if strings.TrimSpace(blob) == "" {
		return ""
	}
	for _, t := range claim {
		if strings.Contains(blob, t) {
			return "Yes"
		}
	}
	return ""
}

func filterBesides(query string, items []RecallItem) []RecallItem {
	excl := besidesExclusionTokens(query)
	if len(excl) == 0 || len(items) == 0 {
		return items
	}
	out := make([]RecallItem, 0, len(items))
	for _, it := range items {
		if itemHitsExclusion(it.Value, excl) {
			continue
		}
		out = append(out, it)
	}
	return out
}

func besidesExclusionTokens(query string) []string {
	lower := strings.ToLower(query)
	i := strings.Index(lower, " besides ")
	if i < 0 {
		i = strings.Index(lower, " besides")
	}
	if i < 0 {
		return nil
	}
	rest := query
	if i+len(" besides ") <= len(query) && strings.HasPrefix(lower[i:], " besides ") {
		rest = query[i+len(" besides "):]
	} else if i+len(" besides") <= len(query) {
		rest = strings.TrimSpace(query[i+len(" besides"):])
	}
	return contentBearingTokens(tokenize(rest))
}

func itemHitsExclusion(value string, excl []string) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return false
	}
	vtoks := tokenize(value)
	vstems := map[string]struct{}{exclusionStem(v): {}}
	for _, vt := range vtoks {
		vstems[exclusionStem(vt)] = struct{}{}
		vstems[vt] = struct{}{}
	}
	for _, t := range excl {
		if len(t) < 4 {
			continue
		}
		if strings.Contains(v, t) || strings.HasPrefix(v, t) {
			return true
		}
		if _, ok := vstems[exclusionStem(t)]; ok && exclusionStem(t) != "" {
			return true
		}
		for _, vt := range vtoks {
			if tokensMatch(t, vt) {
				return true
			}
		}
	}
	return false
}

func exclusionStem(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	for _, suf := range []string{"ingly", "ing", "edly", "ed", "es"} {
		if strings.HasSuffix(t, suf) && len(t)-len(suf) >= 3 {
			t = t[:len(t)-len(suf)]
			break
		}
	}
	if strings.HasSuffix(t, "e") && len(t) > 3 {
		t = t[:len(t)-1]
	}
	if strings.HasSuffix(t, "s") && len(t) >= 4 {
		t = t[:len(t)-1]
	}
	return t
}

func capEnumerateItems(items []RecallItem) []RecallItem {
	if len(items) <= maxEnumerateAnswerItems {
		return items
	}
	return items[:maxEnumerateAnswerItems]
}

func (s *Service) refineEnumeratedItems(ctx context.Context, req RecallRequest, items []RecallItem, hops []HopResult) []RecallItem {
	items = filterBesides(req.Query, items)
	items = s.filterHopEvidence(ctx, req, items, hops)
	if !looksCountQuery(req.Query) {
		items = s.rankItemsByQuery(ctx, req, items, hops)
		items = capEnumerateItems(items)
	}
	return items
}

func (s *Service) rankItemsByQuery(ctx context.Context, req RecallRequest, items []RecallItem, hops []HopResult) []RecallItem {
	if len(items) < 2 {
		return items
	}
	type scored struct {
		it    RecallItem
		score int
		idx   int
	}
	rows := make([]scored, 0, len(items))
	for i, it := range items {
		blob := s.itemEvidenceBlob(ctx, req, it, hops)
		rows = append(rows, scored{it: it, score: itemQueryScore(req.Query, it.Value, blob), idx: i})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].score != rows[j].score {
			return rows[i].score > rows[j].score
		}
		return rows[i].idx < rows[j].idx
	})
	if len(negatedModifierTokens(req.Query)) > 0 {
		out := make([]RecallItem, 0, len(rows))
		for _, r := range rows {
			out = append(out, r.it)
		}
		return out
	}
	any := false
	for _, r := range rows {
		if r.score > 0 {
			any = true
			break
		}
	}
	out := make([]RecallItem, 0, len(rows))
	for _, r := range rows {
		if any && r.score == 0 {
			continue
		}
		out = append(out, r.it)
	}
	if len(out) == 0 {
		return items
	}
	return out
}

func itemQueryScore(query, value, blob string) int {
	toks := contentBearingTokens(tokenize(query))
	skip := map[string]struct{}{
		"enjoy": {}, "enjoys": {}, "enjoyed": {},
		"like": {}, "likes": {}, "liked": {},
		"love": {}, "loves": {}, "loved": {},
		"prefer": {}, "prefers": {}, "preferred": {},
		"hobby": {}, "hobbies": {},
		"activity": {}, "activities": {},
		"item": {}, "items": {},
		"instrument": {}, "instruments": {},
	}
	if looksInstrumentQuery(query) {
		skip["play"] = struct{}{}
		skip["plays"] = struct{}{}
		skip["playing"] = struct{}{}
		skip["practice"] = struct{}{}
		skip["practices"] = struct{}{}
		skip["practicing"] = struct{}{}
	}
	for _, e := range hopQueryEntities(query) {
		el := strings.ToLower(e)
		skip[el] = struct{}{}
		skip[el+"'s"] = struct{}{}
		skip[el+"’s"] = struct{}{}
	}
	hay := strings.ToLower(strings.TrimSpace(value + " " + blob))
	if hay == "" {
		return 0
	}
	n := 0
	seen := map[string]struct{}{}
	for _, t := range toks {
		t = strings.ToLower(t)
		if len(t) < 4 {
			continue
		}
		if _, ok := skip[t]; ok {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		if itemHitsExclusion(hay, []string{t}) {
			n++
		}
	}
	if strings.Contains(strings.ToLower(query), "names") && itemHitsExclusion(hay, []string{"named"}) {
		n += 2
	}
	if len(childhoodClauseTokens(query)) > 0 && (valueHasChildCue(value, blob) || itemHitsExclusion(hay, []string{"child", "childhood"})) {
		n += 2
	}
	if looksUnwindQuery(query) && unwindEvidenceHit(hay) {
		n += 2
	}
	if looksInstrumentQuery(query) && itemHitsExclusion(hay, []string{"play", "plays", "playing", "practice", "practices", "practicing"}) {
		n += 2
	}
	if looksTrickQuery(query) && itemHitsExclusion(hay, []string{"trick", "tricks"}) {
		n += 2
	}
	return n
}

func nameCueTokens(query string) []string {
	q := strings.ToLower(query)
	if !strings.Contains(q, "names") && !strings.Contains(q, "name") {
		return nil
	}
	if looksWhenEventQuery(query) || looksWhereQuery(query) {
		return nil
	}
	return []string{"named"}
}

func childhoodClauseTokens(query string) []string {
	q := strings.ToLower(query)
	if strings.Contains(q, "childhood") || strings.Contains(q, " as a child") || strings.Contains(q, " as a kid") {
		return []string{"child", "childhood"}
	}
	return nil
}

func superlativeAnswer(items []RecallItem) string {
	if len(items) == 0 {
		return ""
	}
	best := items[0]
	bestN := len(items[0].Evidence)
	if bestN == 0 {
		bestN = 1
	}
	for _, it := range items[1:] {
		n := len(it.Evidence)
		if n == 0 {
			n = 1
		}
		if n > bestN {
			best = it
			bestN = n
		}
	}
	return best.Value
}

func whoAnswerFromHops(query string, hops []HopResult) string {
	if queryHasToken(query, "organization", "organizations") ||
		strings.Contains(strings.ToLower(query), "beneficiar") {
		return composeFromHopValues(hops)
	}
	skip := map[string]struct{}{}
	for _, e := range hopQueryEntities(query) {
		skip[strings.ToLower(e)] = struct{}{}
	}
	if len(skip) == 0 {
		return ""
	}
	seen := map[string]struct{}{}
	names := make([]string, 0, 4)
	addPerson := func(n string) {
		n = strings.TrimSpace(n)
		if n == "" {
			return
		}
		if i := strings.IndexAny(n, " ,."); i > 0 {
			n = strings.TrimSpace(n[:i])
		}
		if !looksHopPerson(n) {
			return
		}
		key := strings.ToLower(n)
		if _, ok := skip[key]; ok {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		names = append(names, n)
	}
	addPhrase := func(n string) {
		n = strings.TrimSpace(n)
		if n == "" || !looksSupporterGroup(n) {
			return
		}
		key := strings.ToLower(n)
		if _, ok := skip[key]; ok {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		names = append(names, titleCaseWords(n))
	}
	for _, h := range hops {
		if h.Kind == "resolve_entity" || h.Source == "unresolved" || h.Source == "search_fallback" {
			continue
		}
		for _, v := range h.Values {
			addPerson(v)
			addPhrase(v)
		}
		addPerson(h.Value)
		addPhrase(h.Value)
		for _, c := range h.Contents {
			for _, w := range strings.Fields(c) {
				addPerson(strings.Trim(w, "?,.!'\"()"))
			}
		}
	}
	if len(names) == 0 {
		return composeFromHopValues(hops)
	}
	return strings.Join(names, ", ")
}

func looksSupporterGroup(v string) bool {
	lower := strings.ToLower(v)
	for _, g := range []string{"friend", "team", "colleague", "coworker", "classmate"} {
		if strings.Contains(lower, g) {
			return true
		}
	}
	return false
}

func (s *Service) dateAnswerFromHops(ctx context.Context, req RecallRequest, hops []HopResult) string {
	year := queryCalendarYear(req.Query)
	focus := whenEventFocusTokens(req.Query)
	type hit struct {
		t     time.Time
		score int
	}
	var hits []hit
	for _, h := range hops {
		if h.Kind == "resolve_entity" || h.Source == "unresolved" || h.Source == "search_fallback" {
			continue
		}
		for i, id := range h.MemoryIDs {
			content := ""
			if i < len(h.Contents) {
				content = h.Contents[i]
			} else if len(h.Contents) > 0 {
				content = h.Contents[0]
			}
			var rec MemoryRecord
			if id != "" {
				if got, err := s.store.GetMemory(ctx, req.TenantID, req.SubjectID, id); err == nil {
					rec = got
					if content == "" {
						content = rec.Content
					}
				}
			}
			blob := strings.ToLower(strings.TrimSpace(content + " " + h.Value))
			for _, v := range h.Values {
				blob += " " + strings.ToLower(v)
			}
			t := eventTimeFromRecord(rec, content)
			if t == nil {
				continue
			}
			if year != 0 && t.Year() != year {
				continue
			}
			hits = append(hits, hit{t: t.UTC(), score: focusHitScore(blob, focus)})
		}
	}
	if len(hits) == 0 {
		return ""
	}
	best := hits[0].score
	for _, h := range hits[1:] {
		if h.score > best {
			best = h.score
		}
	}
	var pick *hit
	for i := range hits {
		h := hits[i]
		if best > 0 && h.score < best {
			continue
		}
		if pick == nil || h.t.Before(pick.t) {
			pick = &hits[i]
		}
	}
	if pick == nil {
		return ""
	}
	return pick.t.Format("2 January 2006")
}

func eventTimeFromRecord(rec MemoryRecord, content string) *time.Time {
	if rec.ObservedAt != nil && !rec.ObservedAt.IsZero() {
		t := rec.ObservedAt.UTC()
		return &t
	}
	return parseDateFromText(firstNonEmpty(content, rec.Content))
}

func parseDateFromText(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if t := parseFlexibleTime(s); t != nil && t.Year() > 1900 {
		return t
	}
	fields := strings.Fields(s)
	for n := 5; n >= 2; n-- {
		for i := 0; i+n <= len(fields); i++ {
			chunk := strings.Trim(strings.Join(fields[i:i+n], " "), "()[];,")
			if t := parseFlexibleTime(chunk); t != nil && t.Year() > 1900 {
				return t
			}
		}
	}
	return nil
}

func queryCalendarYear(query string) int {
	for _, tok := range tokenize(query) {
		if len(tok) != 4 {
			continue
		}
		digits := true
		for _, r := range tok {
			if r < '0' || r > '9' {
				digits = false
				break
			}
		}
		if !digits {
			continue
		}
		y, err := strconv.Atoi(tok)
		if err == nil && y >= 1900 && y <= 2100 {
			return y
		}
	}
	return 0
}

func whenEventFocusTokens(query string) []string {
	skip := map[string]struct{}{
		"get": {}, "got": {}, "gone": {}, "go": {}, "going": {},
	}
	for _, e := range hopQueryEntities(query) {
		skip[strings.ToLower(e)] = struct{}{}
	}
	if y := queryCalendarYear(query); y != 0 {
		skip[strconv.Itoa(y)] = struct{}{}
	}
	out := make([]string, 0)
	for _, t := range contentBearingTokens(tokenize(query)) {
		t = strings.ToLower(t)
		if _, ok := skip[t]; ok {
			continue
		}
		if looksLikeDateToken(t) {
			continue
		}
		out = append(out, t)
	}
	return out
}

func focusHitScore(blob string, focus []string) int {
	n := 0
	for _, t := range focus {
		if len(t) < 4 {
			continue
		}
		if strings.Contains(blob, t) {
			n++
			continue
		}
		stem := exclusionStem(t)
		hit := stem != "" && strings.Contains(blob, stem)
		if !hit {
			for _, w := range tokenize(blob) {
				if tokensMatch(t, w) || (stem != "" && stem == exclusionStem(w)) {
					hit = true
					break
				}
			}
		}
		if hit {
			n++
		}
	}
	return n
}

func (s *Service) filterHopEvidence(ctx context.Context, req RecallRequest, items []RecallItem, hops []HopResult) []RecallItem {
	if recip := transferRecipient(req.Query); recip != "" {
		items = s.filterItemsByMention(ctx, req, items, hops, strings.ToLower(recip))
		for i := range items {
			items[i].Value = trimRecipientSuffix(items[i].Value, recip)
		}
	}
	if toks := afterClauseTokens(req.Query); len(toks) > 0 {
		items = s.filterItemsByTokens(ctx, req, items, hops, toks)
	}
	head := listHeadModifierTokens(req.Query)
	group := groupCompanionTokens(req.Query)
	if len(head) > 0 && len(group) > 0 {
		items = s.filterItemsByTokenGroupsPreferIntersect(ctx, req, items, hops, head, group)
	} else {
		if len(group) > 0 {
			items = s.filterItemsByTokens(ctx, req, items, hops, group)
		}
		if len(head) > 0 {
			pos := make([]string, 0, len(head))
			for _, t := range head {
				if unNegationPositive(t) == "" {
					pos = append(pos, t)
				}
			}
			if len(pos) > 0 {
				items = s.filterItemsByTokens(ctx, req, items, hops, pos)
			}
		}
	}
	if toks := forClauseTokens(req.Query); len(toks) > 0 {
		items = s.filterItemsByTokens(ctx, req, items, hops, toks)
	}
	if toks := inCommunityTokens(req.Query); len(toks) > 0 {
		items = s.filterItemsByTokens(ctx, req, items, hops, toks)
	}
	if toks := duringClauseTokens(req.Query); len(toks) > 0 {
		items = s.filterItemsByTokens(ctx, req, items, hops, toks)
	}
	if toks := negatedModifierTokens(req.Query); len(toks) > 0 {
		items = s.filterItemsByNegatedModifier(ctx, req, items, hops, toks)
	}
	if toks := practiceObjectTokens(req.Query); len(toks) > 0 {
		items = s.filterItemsByTokens(ctx, req, items, hops, toks)
	}
	if len(childhoodClauseTokens(req.Query)) > 0 {
		kept := make([]RecallItem, 0, len(items))
		for _, it := range items {
			blob := s.itemEvidenceBlob(ctx, req, it, hops)
			if valueHasChildCue(it.Value, blob) || itemHitsExclusion(it.Value+" "+blob, []string{"child", "childhood"}) {
				kept = append(kept, it)
			}
		}
		if len(kept) > 0 {
			items = kept
		}
	}
	return items
}

func afterClauseTokens(query string) []string {
	if looksWhenEventQuery(query) {
		return nil
	}
	lower := strings.ToLower(query)
	i := strings.Index(lower, " after ")
	if i < 0 {
		return nil
	}
	rest := strings.TrimSpace(query[i+len(" after "):])
	toks := contentBearingTokens(tokenize(rest))
	if len(toks) == 0 {
		return nil
	}
	return toks
}

func groupCompanionTokens(query string) []string {
	if personAfterCue(query, "with") != "" {
		return nil
	}
	group := map[string]struct{}{
		"colleagues": {}, "colleague": {},
		"coworkers": {}, "coworker": {},
		"teammates": {}, "teammate": {},
		"classmates": {}, "classmate": {},
		"friends": {}, "friend": {},
	}
	fields := strings.Fields(query)
	for i, raw := range fields {
		if !strings.EqualFold(strings.Trim(raw, "?,.!\""), "with") || i+1 >= len(fields) {
			continue
		}
		j := i + 1
		next := strings.ToLower(strings.Trim(fields[j], "?,.!\"'"))
		switch next {
		case "his", "her", "their", "the":
			if j+1 >= len(fields) {
				continue
			}
			j++
			next = strings.ToLower(strings.Trim(fields[j], "?,.!\"'"))
		}
		if _, ok := group[next]; ok {
			return []string{next}
		}
	}
	return nil
}

func forClauseTokens(query string) []string {
	if looksWhenEventQuery(query) || looksWhereQuery(query) || transferRecipient(query) != "" {
		return nil
	}
	lower := strings.ToLower(query)
	i := strings.Index(lower, " for ")
	if i < 0 {
		return nil
	}
	rest := strings.TrimSpace(query[i+len(" for "):])
	toks := contentBearingTokens(tokenize(rest))
	if len(toks) == 0 {
		return nil
	}
	return toks
}

// inCommunityTokens pulls a named group from "in the X community".
// Unnamed "in the community" yields nothing so existing community lists stay intact.
func inCommunityTokens(query string) []string {
	lower := strings.ToLower(strings.TrimSpace(query))
	lower = strings.TrimRight(lower, "?,.!")
	if !strings.HasSuffix(lower, " community") {
		return nil
	}
	i := strings.LastIndex(lower, " community")
	if i < 0 {
		return nil
	}
	prefix := lower[:i]
	j := strings.LastIndex(prefix, " in ")
	if j < 0 {
		return nil
	}
	mid := strings.TrimSpace(prefix[j+len(" in "):])
	toks := contentBearingTokens(tokenize(mid))
	if len(toks) == 0 {
		return nil
	}
	return toks
}

func duringClauseTokens(query string) []string {
	if looksWhenEventQuery(query) || looksWhereQuery(query) {
		return nil
	}
	lower := strings.ToLower(query)
	i := strings.Index(lower, " during ")
	if i < 0 {
		return nil
	}
	rest := strings.TrimSpace(query[i+len(" during "):])
	toks := contentBearingTokens(tokenize(rest))
	out := make([]string, 0, len(toks))
	for _, t := range toks {
		switch strings.ToLower(t) {
		case "journey", "journeys", "time", "period", "life", "year", "years":
			continue
		}
		out = append(out, t)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// listHeadModifierTokens pulls the adjective/noun immediately before a list
// head (outdoor activities, sports collectible). Soft-filter only: if no item
// hits, the original list is kept. "similar" is a join cue, not evidence.
func listHeadModifierTokens(query string) []string {
	heads := map[string]struct{}{
		"activities": {}, "activity": {},
		"collectible": {}, "collectibles": {},
		"snacks": {}, "snack": {},
	}
	skip := map[string]struct{}{
		"kind": {}, "type": {}, "some": {}, "any": {},
		"which": {}, "what": {}, "similar": {},
		// Domain/join cue, not an evidence adjective like "outdoor".
		"community": {},
	}
	fields := strings.Fields(query)
	for i, raw := range fields {
		w := strings.ToLower(strings.Trim(raw, "?,.!\"'"))
		if _, ok := heads[w]; !ok || i == 0 {
			continue
		}
		prevRaw := strings.Trim(fields[i-1], "?,.!\"'")
		prev := strings.ToLower(prevRaw)
		if isQueryStopword(prev) {
			continue
		}
		if _, ok := skip[prev]; ok {
			continue
		}
		if looksHopPerson(prevRaw) {
			continue
		}
		if len(prev) < 4 {
			continue
		}
		return []string{prev}
	}
	return nil
}

// negatedModifierTokens returns un- adjectives in the query (unhealthy →
// drop items whose evidence has the positive form only). Not a food list.
func negatedModifierTokens(query string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, 1)
	add := func(t string) {
		t = strings.ToLower(strings.TrimSpace(t))
		if unNegationPositive(t) == "" {
			return
		}
		if _, ok := seen[t]; ok {
			return
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	for _, t := range listHeadModifierTokens(query) {
		add(t)
	}
	for _, t := range contentBearingTokens(tokenize(query)) {
		add(t)
	}
	return out
}

func unNegationPositive(tok string) string {
	tok = strings.ToLower(strings.TrimSpace(tok))
	if !strings.HasPrefix(tok, "un") || len(tok) < 6 {
		return ""
	}
	switch tok {
	case "unless", "until", "under", "unique", "university", "understand",
		"unknown", "union", "united", "unusual", "uniform", "universe",
		"uncle", "uncover", "unfold", "undo", "unit", "units",
		"unwind", "unwinds", "unwinding", "unwound":
		return ""
	}
	rest := tok[2:]
	if len(rest) < 4 {
		return ""
	}
	for _, r := range rest {
		if r < 'a' || r > 'z' {
			return ""
		}
	}
	return rest
}

func itemContradictsNegation(blob, unTok string) bool {
	pos := unNegationPositive(unTok)
	if pos == "" {
		return false
	}
	hasUn := itemHitsExclusion(blob, []string{unTok})
	hasPos := itemHitsExclusion(blob, polarityPositiveForms(pos))
	return hasPos && !hasUn
}

func polarityPositiveForms(pos string) []string {
	pos = strings.ToLower(strings.TrimSpace(pos))
	if pos == "" {
		return nil
	}
	out := []string{pos}
	if strings.HasSuffix(pos, "y") && len(pos) > 4 {
		stem := pos[:len(pos)-1] + "i"
		out = append(out, stem+"er", stem+"est")
	} else {
		out = append(out, pos+"er", pos+"est")
	}
	return out
}

func (s *Service) filterItemsByNegatedModifier(ctx context.Context, req RecallRequest, items []RecallItem, hops []HopResult, toks []string) []RecallItem {
	if len(toks) == 0 || len(items) == 0 {
		return items
	}
	out := make([]RecallItem, 0, len(items))
	for _, it := range items {
		blob := it.Value + " " + s.itemEvidenceBlob(ctx, req, it, hops)
		drop := false
		for _, t := range toks {
			if itemContradictsNegation(blob, t) {
				drop = true
				break
			}
		}
		if !drop {
			out = append(out, it)
		}
	}
	if len(out) == 0 {
		return items
	}
	return out
}

func whereAnswerFromHops(query string, hops []HopResult) string {
	seen := map[string]struct{}{}
	places := make([]string, 0, 2)
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" || anaphoricSlotValue(p) {
			return
		}
		key := strings.ToLower(p)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		places = append(places, titleCaseWords(p))
	}
	locToks := locativeLeftoverTokens(query, hops)
	for _, h := range hops {
		if h.Kind == "resolve_entity" || h.Source == "unresolved" || h.Source == "search_fallback" {
			continue
		}
		for _, c := range h.Contents {
			if len(locToks) > 0 && !contentCoversAnyQueryToken(c, locToks) {
				continue
			}
			for _, p := range placesFromContent(c) {
				add(p)
			}
		}
		if h.Value != "" {
			if len(locToks) == 0 || contentCoversAnyQueryToken(h.Value, locToks) {
				for _, p := range placesFromContent(h.Value) {
					add(p)
				}
			}
		}
	}
	if len(places) == 0 {
		return ""
	}
	return strings.Join(places, ", ")
}

func placeFromContent(content string) string {
	places := placesFromContent(content)
	if len(places) == 0 {
		return ""
	}
	return places[len(places)-1]
}

func placesFromContent(content string) []string {
	content = strings.TrimSpace(stripTrailingStamp(content))
	content = strings.ReplaceAll(content, "`", "'")
	if content == "" {
		return nil
	}
	lower := strings.ToLower(content)
	seen := map[string]struct{}{}
	out := make([]string, 0, 4)
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" || anaphoricSlotValue(p) {
			return
		}
		key := strings.ToLower(p)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, p)
	}
	for _, prep := range []string{" in ", " at ", " near ", " around ", " on "} {
		start := 0
		for {
			i := strings.Index(lower[start:], prep)
			if i < 0 {
				break
			}
			i += start
			rest := content[i+len(prep):]
			if placePrepRestIsDate(rest) {
				start = i + len(prep)
				continue
			}
			for _, part := range splitPlaceList(rest) {
				if cand := cleanPlaceCandidate(part); cand != "" {
					add(cand)
				}
			}
			start = i + len(prep)
		}
	}
	return out
}

func splitPlaceList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if j := strings.IndexAny(s, ".([;!?"); j >= 0 {
		s = strings.TrimSpace(s[:j])
	}
	raw := strings.Split(s, ",")
	parts := make([]string, 0, len(raw))
	for _, p := range raw {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		low := strings.ToLower(p)
		if strings.HasPrefix(low, "and ") {
			p = strings.TrimSpace(p[4:])
		}
		if i := strings.Index(strings.ToLower(p), " and "); i > 0 {
			parts = append(parts, strings.TrimSpace(p[:i]), strings.TrimSpace(p[i+5:]))
			continue
		}
		parts = append(parts, p)
	}
	return parts
}

func cleanPlaceCandidate(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "`", "'"))
	if s == "" {
		return ""
	}
	if j := strings.IndexAny(s, ".([;!?"); j >= 0 {
		s = strings.TrimSpace(s[:j])
	}
	fields := strings.Fields(s)
	skipLead := map[string]struct{}{
		"his": {}, "her": {}, "their": {}, "its": {}, "my": {}, "our": {},
		"the": {}, "a": {}, "an": {}, "this": {}, "that": {}, "these": {}, "those": {},
	}
	for len(fields) > 0 {
		w := strings.ToLower(strings.Trim(fields[0], ",'\"'"))
		if _, ok := skipLead[w]; ok {
			fields = fields[1:]
			continue
		}
		break
	}
	if len(fields) == 0 {
		return ""
	}
	first := strings.ToLower(strings.Trim(fields[0], ",'\"'"))
	switch first {
	case "order", "common", "fact", "addition":
		return ""
	}
	if looksPlaceDateToken(first) {
		return ""
	}
	stopAt := map[string]struct{}{
		"last": {}, "during": {}, "when": {}, "after": {}, "before": {},
		"because": {}, "with": {}, "while": {}, "where": {}, "and": {},
		"which": {}, "that": {}, "who": {}, "whose": {}, "whom": {},
		"in": {}, "at": {}, "on": {}, "near": {}, "around": {},
	}
	keep := make([]string, 0, 4)
	for i, f := range fields {
		w := strings.ToLower(strings.Trim(f, ",'\"'"))
		if _, ok := stopAt[w]; ok && i > 0 {
			break
		}
		if i > 0 && looksHopPerson(titleCaseWords(strings.Trim(f, ",'\"'"))) && i+1 < len(fields) && looksPlaceClauseVerb(fields[i+1]) {
			break
		}
		keep = append(keep, strings.Trim(f, ",'\"'"))
		if len(keep) >= 4 {
			break
		}
	}
	if len(keep) == 0 {
		return ""
	}
	if strings.EqualFold(keep[0], "the") && len(keep) == 1 {
		return ""
	}
	if len(keep) == 1 && strings.HasSuffix(strings.ToLower(keep[0]), "ing") {
		return ""
	}
	out := strings.Join(keep, " ")
	if anaphoricSlotValue(out) {
		return ""
	}
	return out
}

func placePrepRestIsDate(s string) bool {
	fields := strings.Fields(strings.TrimSpace(s))
	if len(fields) == 0 {
		return false
	}
	return looksPlaceDateToken(fields[0])
}

func looksPlaceDateToken(w string) bool {
	w = strings.ToLower(strings.Trim(w, ",.'\"'"))
	if w == "" {
		return false
	}
	r := []rune(w)
	if r[0] >= '0' && r[0] <= '9' {
		return true
	}
	switch w {
	case "january", "february", "march", "april", "may", "june", "july",
		"august", "september", "october", "november", "december",
		"jan", "feb", "mar", "apr", "jun", "jul", "aug", "sep", "sept",
		"oct", "nov", "dec",
		"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday":
		return true
	}
	return false
}

func looksPlaceClauseVerb(w string) bool {
	switch strings.ToLower(strings.Trim(w, ",.'\"'")) {
	case "met", "meets", "meet", "went", "goes", "go", "did", "does", "do",
		"was", "were", "is", "are", "attended", "attends", "visited", "visits",
		"saw", "sees", "told", "said", "practices", "practice", "tried":
		return true
	}
	return false
}

func placeEqualsAny(place string, toks []string) bool {
	p := strings.ToLower(strings.TrimSpace(place))
	if p == "" {
		return false
	}
	for _, t := range toks {
		if p == strings.ToLower(strings.TrimSpace(t)) {
			return true
		}
	}
	return false
}

func (s *Service) filterChildCountItems(ctx context.Context, req RecallRequest, items []RecallItem, hops []HopResult) []RecallItem {
	head := countHeadNoun(req.Query)
	switch head {
	case "children", "kids", "child":
	default:
		return items
	}
	if len(items) == 0 {
		return items
	}
	out := make([]RecallItem, 0, len(items))
	for _, it := range items {
		blob := s.itemEvidenceBlob(ctx, req, it, hops)
		if valueHasChildCue(it.Value, blob) {
			out = append(out, it)
		}
	}
	return out
}

func valueHasChildCue(value, blob string) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	b := strings.ToLower(blob)
	if v == "" || b == "" {
		return false
	}
	childCues := []string{"son", "daughter", "child", "kid", "children", "kids"}
	partnerCues := []string{"partner", "spouse", "wife", "husband"}
	for _, c := range childCues {
		if v == c || strings.HasPrefix(v, c+" ") {
			return true
		}
	}
	vidx := strings.Index(b, v)
	if vidx < 0 {
		return queryHasToken(b, childCues...) && !queryHasToken(b, partnerCues...)
	}
	childDist := minCueDist(b, vidx, childCues)
	partnerDist := minCueDist(b, vidx, partnerCues)
	if childDist < 0 {
		return false
	}
	if partnerDist < 0 {
		return true
	}
	return childDist <= partnerDist
}

func minCueDist(blob string, at int, cues []string) int {
	best := -1
	for _, c := range cues {
		if c == "" {
			continue
		}
		searchFrom := 0
		for {
			i := strings.Index(blob[searchFrom:], c)
			if i < 0 {
				break
			}
			i += searchFrom
			end := i + len(c)
			ok := (i == 0 || !isTokenChar(blob[i-1])) && (end >= len(blob) || !isTokenChar(blob[end]))
			if ok {
				d := i - at
				if d < 0 {
					d = -d
				}
				if best < 0 || d < best {
					best = d
				}
			}
			searchFrom = i + 1
			if searchFrom >= len(blob) {
				break
			}
		}
	}
	return best
}

func isTokenChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func trimRecipientSuffix(v, recip string) string {
	v = strings.TrimSpace(v)
	if recip == "" {
		return v
	}
	suf := " to " + recip
	if len(v) > len(suf) && strings.EqualFold(v[len(v)-len(suf):], suf) {
		return strings.TrimSpace(v[:len(v)-len(suf)])
	}
	return v
}

func (s *Service) itemEvidenceBlob(ctx context.Context, req RecallRequest, it RecallItem, hops []HopResult) string {
	var b strings.Builder
	b.WriteString(it.Value)
	seen := map[string]struct{}{}
	for _, id := range it.Evidence {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		for _, h := range hops {
			for i, hid := range h.MemoryIDs {
				if hid != id {
					continue
				}
				if i < len(h.Contents) {
					b.WriteByte(' ')
					b.WriteString(h.Contents[i])
				}
			}
		}
		if rec, err := s.store.GetMemory(ctx, req.TenantID, req.SubjectID, id); err == nil {
			b.WriteByte(' ')
			b.WriteString(rec.Content)
		}
	}
	return strings.ToLower(b.String())
}

func (s *Service) filterItemsByMention(ctx context.Context, req RecallRequest, items []RecallItem, hops []HopResult, mention string) []RecallItem {
	if mention == "" || len(items) == 0 {
		return items
	}
	out := make([]RecallItem, 0, len(items))
	for _, it := range items {
		if strings.Contains(s.itemEvidenceBlob(ctx, req, it, hops), mention) {
			out = append(out, it)
		}
	}
	if len(out) == 0 {
		return items
	}
	return out
}

func (s *Service) filterItemsByTokens(ctx context.Context, req RecallRequest, items []RecallItem, hops []HopResult, toks []string) []RecallItem {
	if len(toks) == 0 || len(items) == 0 {
		return items
	}
	out := make([]RecallItem, 0, len(items))
	for _, it := range items {
		blob := s.itemEvidenceBlob(ctx, req, it, hops)
		if itemHitsExclusion(it.Value, toks) || itemHitsExclusion(blob, toks) {
			out = append(out, it)
		}
	}
	if len(out) == 0 {
		return items
	}
	return out
}

// filterItemsByTokenGroupsPreferIntersect keeps items that hit both cue
// groups when any exist; otherwise list-head hits; otherwise companion
// hits; otherwise the original list. Sequential companion-then-head
// filtering would keep a colleague indoor singleton and drop outdoor
// evidence that never said "colleagues".
func (s *Service) filterItemsByTokenGroupsPreferIntersect(ctx context.Context, req RecallRequest, items []RecallItem, hops []HopResult, head, group []string) []RecallItem {
	if len(items) == 0 {
		return items
	}
	posHead := make([]string, 0, len(head))
	for _, t := range head {
		if unNegationPositive(t) == "" {
			posHead = append(posHead, t)
		}
	}
	if len(posHead) == 0 {
		return s.filterItemsByTokens(ctx, req, items, hops, group)
	}
	if len(group) == 0 {
		return s.filterItemsByTokens(ctx, req, items, hops, posHead)
	}
	var both, headHits, groupHits []RecallItem
	for _, it := range items {
		blob := it.Value + " " + s.itemEvidenceBlob(ctx, req, it, hops)
		headHit := itemHitsExclusion(it.Value, posHead) || itemHitsExclusion(blob, posHead)
		groupHit := itemHitsExclusion(it.Value, group) || itemHitsExclusion(blob, group)
		switch {
		case headHit && groupHit:
			both = append(both, it)
		case headHit:
			headHits = append(headHits, it)
		case groupHit:
			groupHits = append(groupHits, it)
		}
	}
	if len(both) > 0 {
		return both
	}
	if len(headHits) > 0 {
		return headHits
	}
	if len(groupHits) > 0 {
		return groupHits
	}
	return items
}

func memoryMentionsEntity(rec MemoryRecord, entity string) bool {
	if strings.EqualFold(entitySubjectOf(rec), entity) {
		return true
	}
	return strings.Contains(strings.ToLower(rec.Content), strings.ToLower(entity))
}

func validEnumeratedValue(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" || utf8.RuneCountInString(v) < 3 || utf8.RuneCountInString(v) > 80 {
		return false
	}
	if strings.HasSuffix(v, "?") {
		return false
	}
	return true
}

type representationOracleCounts struct {
	atomCount    int
	factCount    int
	episodeCount int
	factBlob     string
	episodeBlob  string
}

func (s *Service) collectRepresentationOracle(ctx context.Context, req RecallRequest, includeSuperseded bool, topK int) (representationOracleCounts, error) {
	var out representationOracleCounts
	if indexer, ok := s.store.(AtomIndexer); ok {
		ids, err := indexer.ListAtomMemoryIDs(ctx, req.TenantID, req.SubjectID, "", "", topK)
		if err != nil {
			return out, err
		}
		out.atomCount = len(ids)
	}
	listed, err := s.store.ListMemories(ctx, req.TenantID, req.SubjectID, includeSuperseded)
	if err != nil {
		return out, err
	}
	facts := make([]string, 0, len(listed))
	episodes := make([]string, 0, len(listed))
	for _, rec := range listed {
		if rec.Status != "" && rec.Status != StatusActive {
			continue
		}
		if IsProvenanceEpisode(rec) {
			out.episodeCount++
			episodes = append(episodes, rec.Content)
			continue
		}
		out.factCount++
		facts = append(facts, rec.Content)
	}
	out.factBlob = joinBounded(facts, representationBlobBudget)
	out.episodeBlob = joinBounded(episodes, representationBlobBudget)
	return out, nil
}

func joinBounded(parts []string, budget int) string {
	var b strings.Builder
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		need := len(p)
		if b.Len() > 0 {
			need++
		}
		if b.Len()+need > budget {
			break
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(p)
	}
	return b.String()
}

func countSearchRepresentation(results []SearchResult) (facts, episodes int) {
	for _, r := range results {
		prim, _ := r.Explain["primitive"].(string)
		if prim == PrimitiveEpisode {
			episodes++
			continue
		}
		facts++
	}
	return facts, episodes
}

func formatEvidenceOracle(rows []map[string]any, budgetTokens int) string {
	if budgetTokens <= 0 {
		budgetTokens = 4000
	}
	budget := budgetTokens * 4
	var b strings.Builder
	for _, row := range rows {
		content, _ := row["content"].(string)
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}
		line := "- " + content + "\n"
		if b.Len()+len(line) > budget {
			break
		}
		b.WriteString(line)
	}
	return b.String()
}

func valueFromMemoryContent(content string) string {
	if v, ok := slotValueFromMemoryContent(content); ok {
		return v
	}
	v := speakerPrefixRe.ReplaceAllString(content, "")
	return strings.TrimSpace(v)
}

func slotValueFromMemoryContent(content string) (string, bool) {
	content = stripTrailingStamp(content)
	if m := quotedTitleRE.FindStringSubmatch(content); m != nil {
		title := NormalizeText(firstNonEmpty(m[1], m[2], m[3], m[4]))
		if utf8.RuneCountInString(title) >= 4 && len(strings.Fields(title)) >= 2 && !looksBrokenQuotedTitle(title) {
			return title, true
		}
	}
	if m := visibleTextBlockRE.FindStringSubmatch(content); m != nil {
		if title, ok := titleFromVisibleText(m[1]); ok {
			return title, true
		}
	}
	stripped := visibleTextBlockRE.ReplaceAllString(content, " ")
	lower := strings.ToLower(stripped)
	for _, sep := range []string{
		" participates in ", " participated in ", " enjoys ", " likes ", " liked ", " loves ", " moved from ", " is from ", " lives in ", " kids like ",
		" read \"", " has done ", " plans career in ", " plans career for ",
		" researched ", " unwinds via ", " works as ", " realized that ", " is a ", " is ",
		" owns ", " owned ", " bought ", " named ", " is named ", " plays ", " played ",
		" tried ", " injured ", " practices ", " practiced ",
		" supports ", " told ", " visited ", " given ", " gave ", " suggested ",
	} {
		if i := strings.Index(lower, sep); i >= 0 {
			if sep == " is " && (titleLikeCopula(content) || !identityCopulaSubject(stripped[:i])) {
				continue
			}
			v := strings.TrimSpace(stripped[i+len(sep):])
			v = strings.Trim(v, "\"")
			if j := strings.IndexAny(v, ".(["); j > 0 {
				v = strings.TrimSpace(v[:j])
			}
			if sep == " has done " || sep == " practices " || sep == " practiced " {
				if _, place, ok := strings.Cut(v, " at "); ok {
					v = strings.TrimSpace(place)
				}
			}
			if sep == " told " || sep == " given " {
				if head, _, ok := strings.Cut(v, " about "); ok {
					v = strings.TrimSpace(head)
				}
			}
			if strings.TrimSpace(v) == "" {
				return "", false
			}
			return v, true
		}
	}
	return "", false
}

func hasSlotTemplate(v string) bool {
	low := strings.ToLower(v)
	for _, sep := range []string{
		" participates in ", " participated in ", " enjoys ", " likes ", " liked ", " loves ", " moved from ", " is from ", " lives in ", " kids like ",
		" read \"", " has done ", " plans career in ", " plans career for ",
		" researched ", " unwinds via ", " works as ", " realized that ", " is a ",
		" owns ", " owned ", " bought ", " named ", " is named ", " plays ", " played ",
		" tried ", " injured ", " practices ", " practiced ",
		" supports ", " told ", " visited ", " given ", " gave ", " suggested ",
	} {
		if strings.Contains(low, sep) {
			return true
		}
	}
	return false
}

func titleLikeCopula(content string) bool {
	words := strings.Fields(strings.TrimSpace(content))
	if len(words) < 3 {
		return false
	}
	caps := 0
	for _, w := range words {
		if len(w) > 0 && w[0] >= 'A' && w[0] <= 'Z' {
			caps++
		}
	}
	return caps >= 2
}

// identityCopulaSubject is true when the left-hand side of " is " looks like
// a short entity ("Riley is single"), not a clause ("realized that X is Y").
func identityCopulaSubject(left string) bool {
	left = strings.TrimSpace(speakerPrefixRe.ReplaceAllString(left, ""))
	words := strings.Fields(left)
	if len(words) == 0 || len(words) > 3 {
		return false
	}
	lower := " " + strings.ToLower(left) + " "
	for _, cue := range []string{" that ", " because ", " when ", " after ", " before ", " who ", " which ", " what "} {
		if strings.Contains(lower, cue) {
			return false
		}
	}
	return true
}

func assembleContextBlock(results []SearchResult, budgetTokens int) string {
	if budgetTokens <= 0 {
		budgetTokens = 4000
	}
	// Rough rune budget ≈ 4 chars/token.
	budget := budgetTokens * 4
	var b strings.Builder
	seen := map[string]struct{}{}
	for _, r := range results {
		content := strings.TrimSpace(r.Content)
		if content == "" || strings.HasSuffix(content, "?") {
			continue
		}
		key := strings.ToLower(content)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		line := "- " + content
		if r.ObservedAt != nil {
			line += " [event_time=" + r.ObservedAt.UTC().Format(time.RFC3339) + "]"
		}
		line += "\n"
		if b.Len()+len(line) > budget {
			break
		}
		b.WriteString(line)
	}
	return strings.TrimSpace(b.String())
}

func assembleContextFromPacket(pkt EvidencePacket, budgetTokens int) string {
	if budgetTokens <= 0 {
		budgetTokens = 4000
	}
	budget := budgetTokens * 4
	var b strings.Builder
	seen := map[string]struct{}{}
	if pkt.TemporalAnswer != "" {
		line := "- " + pkt.TemporalAnswer + "\n"
		b.WriteString(line)
		seen[strings.ToLower(pkt.TemporalAnswer)] = struct{}{}
	}
	for _, it := range pkt.ProofChain {
		content := strings.TrimSpace(it.Content)
		if it.Value != "" && !looksTitleCaseSlogan(it.Value) {
			content = strings.TrimSpace(it.Value)
		}
		if content == "" || strings.HasSuffix(content, "?") || looksTitleCaseSlogan(content) {
			continue
		}
		key := strings.ToLower(content)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		line := "- " + content + "\n"
		if b.Len()+len(line) > budget {
			break
		}
		b.WriteString(line)
	}
	for _, content := range pkt.Contents {
		content = strings.TrimSpace(content)
		if content == "" || strings.HasSuffix(content, "?") {
			continue
		}
		key := strings.ToLower(content)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		line := "- " + content + "\n"
		if b.Len()+len(line) > budget {
			break
		}
		b.WriteString(line)
	}
	return strings.TrimSpace(b.String())
}

func searchResultIsEpisode(r SearchResult) bool {
	if r.Explain == nil {
		return false
	}
	p, _ := r.Explain["primitive"].(string)
	return p == PrimitiveEpisode
}

func searchResultPredicate(r SearchResult) string {
	if r.Explain == nil {
		return ""
	}
	p, _ := r.Explain["predicate"].(string)
	return strings.TrimSpace(p)
}

func searchResultValueNorm(r SearchResult) string {
	if r.Explain == nil {
		return ""
	}
	v, _ := r.Explain["value_norm"].(string)
	return strings.TrimSpace(v)
}

func looksTitleCaseSlogan(s string) bool {
	words := strings.Fields(strings.TrimSpace(s))
	if len(words) < 4 {
		return false
	}
	for _, w := range words {
		r := w[0]
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func structuredValueOf(r SearchResult) string {
	if looksTitleCaseSlogan(r.Content) {
		return ""
	}
	if v := searchResultValueNorm(r); v != "" && !anaphoricSlotValue(v) {
		return v
	}
	if val, ok := slotValueFromMemoryContent(r.Content); ok && !anaphoricSlotValue(val) {
		return val
	}
	return ""
}

func pickStructuredAnswer(query string, results []SearchResult) string {
	preds := predicateHintsFromQuery(query)
	predSet := map[string]struct{}{}
	for _, p := range preds {
		predSet[p] = struct{}{}
	}
	qtoks := map[string]struct{}{}
	for _, t := range contentBearingTokens(tokenize(query)) {
		if len(t) > 2 {
			qtoks[t] = struct{}{}
		}
	}
	var matched, overlapped, rest []string
	seen := map[string]struct{}{}
	add := func(dst *[]string, v string) {
		v = strings.TrimSpace(v)
		if v == "" || looksTitleCaseSlogan(v) {
			return
		}
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		*dst = append(*dst, v)
	}
	for _, r := range results {
		v := structuredValueOf(r)
		if v == "" || looksLikeQueryNameEcho(query, v) {
			continue
		}
		recPred := searchResultPredicate(r)
		if recPred == "" {
			for _, h := range predicateHintsFromQuery(r.Content) {
				if recPred == "" {
					recPred = h
				}
				if _, ok := predSet[h]; ok {
					recPred = h
					break
				}
			}
		}
		if len(predSet) > 0 && recPred != "" {
			if _, ok := predSet[recPred]; ok {
				add(&matched, v)
				continue
			}
			if !hopUsefulForList(recPred, firstHint(preds)) {
				continue
			}
		}
		blob := strings.ToLower(v + " " + r.Content)
		hit := false
		for tok := range qtoks {
			if strings.Contains(blob, tok) {
				hit = true
				break
			}
		}
		if hit {
			add(&overlapped, v)
		} else {
			add(&rest, v)
		}
	}
	switch {
	case len(matched) == 1:
		return matched[0]
	case len(matched) > 1 && len(matched) <= 6:
		return strings.Join(matched, ", ")
	case len(overlapped) == 1:
		return overlapped[0]
	case len(overlapped) > 1 && len(overlapped) <= 4:
		return strings.Join(overlapped, ", ")
	case len(rest) == 1:
		return rest[0]
	default:
		return ""
	}
}

func firstHint(preds []string) string {
	if len(preds) == 0 {
		return ""
	}
	return preds[0]
}

func pickEpisodeFallback(query string, results []SearchResult) string {
	qtoks := map[string]struct{}{}
	for _, t := range contentBearingTokens(tokenize(query)) {
		if len(t) > 2 {
			qtoks[t] = struct{}{}
		}
	}
	if len(qtoks) == 0 {
		return ""
	}
	best := ""
	bestHits := 0
	for _, r := range results {
		if searchResultPredicate(r) != "" && searchResultValueNorm(r) != "" {
			continue
		}
		prim := ""
		if r.Explain != nil {
			prim, _ = r.Explain["primitive"].(string)
		}
		if prim != "" && prim != PrimitiveEpisode {
			continue
		}
		c := strings.TrimSpace(r.Content)
		if c == "" || strings.HasSuffix(c, "?") || looksTitleCaseSlogan(c) || looksLikeQueryNameEcho(query, c) {
			continue
		}
		hits := 0
		lower := strings.ToLower(c)
		for tok := range qtoks {
			if strings.Contains(lower, tok) {
				hits++
			}
		}
		if hits > bestHits {
			bestHits = hits
			best = c
		}
	}
	if bestHits < 2 {
		return ""
	}
	return best
}

func packetContentLines(pkt EvidencePacket) []string {
	out := make([]string, 0, 16)
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	for _, it := range pkt.ContextEvidence {
		add(it.Content)
	}
	for _, it := range pkt.ProofChain {
		add(it.Content)
	}
	for _, it := range pkt.Items {
		add(it.Content)
	}
	for _, c := range pkt.Contents {
		add(c)
	}
	return out
}

func queryNameOrdinal(query string) int {
	if !queryHasToken(query, "name", "named") {
		return 0
	}
	q := strings.ToLower(query)
	ords := []struct {
		needle string
		n      int
	}{
		{"first ", 1}, {"1st ", 1},
		{"second ", 2}, {"2nd ", 2},
		{"third ", 3}, {"3rd ", 3},
		{"fourth ", 4}, {"4th ", 4},
	}
	for _, o := range ords {
		if strings.Contains(q, o.needle) {
			return o.n
		}
	}
	return 0
}

func namedInstanceFromContent(content string) string {
	lower := strings.ToLower(content)
	idx := strings.Index(lower, " named ")
	rest := ""
	if idx >= 0 {
		rest = strings.TrimSpace(content[idx+len(" named "):])
	} else {
		idx = strings.Index(lower, " name is ")
		if idx < 0 {
			return ""
		}
		rest = strings.TrimSpace(content[idx+len(" name is "):])
	}
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return ""
	}
	name := strings.Trim(fields[0], ".,;:!?\"'")
	if name == "" || utf8Len(name) < 2 {
		return ""
	}
	r := name[0]
	if r < 'A' || r > 'Z' {
		return ""
	}
	return name
}

func ordinalNameFromPacket(query string, pkt EvidencePacket) string {
	ord := queryNameOrdinal(query)
	if ord <= 0 {
		return ""
	}
	ents := map[string]struct{}{}
	for _, e := range hopQueryEntities(query) {
		ents[strings.ToLower(e)] = struct{}{}
	}
	type named struct {
		name string
		t    time.Time
		has  bool
	}
	found := make([]named, 0, 4)
	seen := map[string]struct{}{}
	for _, line := range packetContentLines(pkt) {
		name := namedInstanceFromContent(line)
		if name == "" {
			continue
		}
		if _, ok := ents[strings.ToLower(name)]; ok {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		item := named{name: name}
		if dt := parseDateFromText(line); dt != nil {
			item.t = *dt
			item.has = true
		}
		found = append(found, item)
	}
	if len(found) < ord {
		return ""
	}
	for i := 0; i < len(found); i++ {
		for j := i + 1; j < len(found); j++ {
			li, lj := found[i], found[j]
			swap := false
			switch {
			case li.has && lj.has:
				swap = lj.t.Before(li.t)
			case lj.has && !li.has:
				swap = true
			}
			if swap {
				found[i], found[j] = found[j], found[i]
			}
		}
	}
	return found[ord-1].name
}

func looksChatTurnLine(s string) bool {
	t := strings.TrimSpace(s)
	if t == "" || strings.HasSuffix(t, "?") {
		return true
	}
	lower := strings.ToLower(t)
	if strings.HasPrefix(lower, "oh,") || strings.HasPrefix(lower, "oh ") {
		return true
	}
	if i := strings.Index(t, ": "); i > 0 && i < 24 {
		head := strings.TrimSpace(t[:i])
		if head != "" && !strings.Contains(head, " ") && consecutiveProperNouns(head) >= 1 {
			return true
		}
	}
	return false
}

func looksCodedEventToken(s string) bool {
	for _, w := range strings.Fields(s) {
		w = strings.Trim(w, ".,;()[]\"'")
		if strings.Contains(w, ":") && utf8Len(w) >= 4 {
			return true
		}
	}
	return false
}

func leftoverCoveringSpecificAnswer(query string, hops []HopResult, pkt EvidencePacket) string {
	if looksWhereQuery(query) || looksPolarQuery(query) || hopsKeepTypedJoin(hops) {
		return ""
	}
	leftover := leftoverNonEntityQueryTokens(query, hops)
	if len(leftover) == 0 {
		leftover = distinctiveQueryTokens(tokenize(query))
	}
	rare := make([]string, 0, len(leftover))
	for _, tok := range leftover {
		if utf8Len(tok) >= 6 {
			rare = append(rare, tok)
		}
	}
	if len(rare) == 0 {
		return ""
	}
	best := ""
	bestScore := 0
	for _, line := range packetContentLines(pkt) {
		if looksChatTurnLine(line) || looksTitleCaseSlogan(line) {
			continue
		}
		if !contentCoversAnyQueryToken(line, rare) {
			continue
		}
		score := 1
		if looksCodedEventToken(line) {
			score += 3
		}
		if looksSpecificPlaceOrNameLine(line) || looksHyphenatedEventLine(line) {
			score += 2
		}
		hits := 0
		for _, tok := range rare {
			if contentCoversQueryToken(line, tok) {
				hits++
			}
		}
		score += hits
		if score > bestScore {
			bestScore = score
			best = line
		}
	}
	if bestScore < 2 {
		return ""
	}
	return best
}

func looksLikeQueryNameEcho(query, value string) bool {
	vv := strings.ToLower(strings.TrimSpace(value))
	if vv == "" {
		return true
	}
	if strings.ContainsAny(vv, " \t") {
		return false
	}
	return strings.Contains(strings.ToLower(query), vv)
}

func firstFactLine(blob string) string {
	for _, line := range strings.Split(blob, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasSuffix(line, "?") || looksTitleCaseSlogan(line) {
			continue
		}
		return line
	}
	return ""
}

func joinSearchFacts(results []SearchResult) string {
	parts := make([]string, 0, len(results))
	for _, r := range results {
		if searchResultIsEpisode(r) {
			continue
		}
		c := strings.TrimSpace(r.Content)
		if c == "" {
			continue
		}
		parts = append(parts, c)
	}
	return joinBounded(parts, representationBlobBudget)
}

func formatEnumerateContext(items []RecallItem) string {
	if len(items) == 0 {
		return ""
	}
	parts := make([]string, 0, len(items))
	for _, it := range items {
		parts = append(parts, it.Value)
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

func appendUnique(ids []string, id string) []string {
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append(ids, id)
}
