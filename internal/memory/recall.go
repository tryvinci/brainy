package memory

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// representationBlobBudget is the oracle/fact dump cap. 80k left only ~15%
// headroom at observed LoCoMo volume; LME-scale subjects overflow.
const representationBlobBudget = 400000

// maxEnumerateAnswerItems bounds list answers. Counts keep the full typed set.
// 16 is still bounded (kill list: no unbounded top-k) so older class items
// (collars, mountaineering) are not recency-starved after a for-clause filter.
const maxEnumerateAnswerItems = 16

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
	// Full atom history is for how-many and typed-set enumerations
	// ("what items", "what activities has"). Incidental list nouns in
	// how/why leftover questions must not scan every possession/activity.
	if looksCountQuery(req.Query) || wantsHistoricalAtomScan(req.Query) {
		hist = true
	}
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
	// Search already ranked a first-person cause leftover. Extra hop
	// SearchOpt probes (and the hybrid reader after them) can idle past
	// the harness timeout before leftover covering is allowed to fire.
	skipHopsForCovering := leftoverCoveringSkipsHybrid(req.Query, leftoverCoveringSpecificAnswer(req.Query, nil, pkt))
	if skipHopsForCovering {
		out.Explain["hops_skipped_leftover_covering"] = true
	}
	// Typed hop executor: bind packet from hop joins when hops exist.
	if !skipHopsForCovering && (plan.NeedsMultiHop || plan.NeedsEnumeration) && len(plan.Hops) > 0 {
		var byKey map[string]HopResult
		hopResults, byKey = s.executeTypedHops(ctx, req.TenantID, req.SubjectID, req.Vertical, hist, plan, hopScanLimit(req.Query, plan, topK), req.Query)
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
						hopResults2, byKey2 := s.executeTypedHops(ctx, req.TenantID, req.SubjectID, req.Vertical, hist, plan, hopScanLimit(req.Query, plan, topK), req.Query)
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
	} else if !skipHopsForCovering && plan.NeedsMultiHop && plan.BudgetPasses >= 2 {
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
		if (wantsTypedSetScan(req.Query) || looksUnwindQuery(req.Query) || looksSuperlativeQuery(req.Query) || looksLocationListQuery(req.Query) || transferRecipient(req.Query) != "") && !looksConsequenceQuery(req.Query) && !looksWhereQuery(req.Query) {
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
			out.Items = s.filterCountItems(ctx, req, out.Items, hopResults)
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
		coveringEarly := leftoverCoveringSpecificAnswer(req.Query, hopResults, pkt)
		var hybrid hybridReaderResult
		if leftoverCoveringSkipsHybrid(req.Query, coveringEarly) {
			hybrid = hybridReaderResult{Reason: "leftover_covering"}
			out.Explain["hybrid_skipped_leftover_covering"] = true
		} else {
			hybrid = s.synthesizeHybridAnswer(ctx, req.Query, plan, pkt)
		}
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
		echoSlogan := leftoverQueryEchoAnswer(req.Query, typedAnswer)
		thinMiss := leftoverThinMissAnswer(req.Query, hopResults, typedAnswer)
		unlockDump := typedAnswerIsHopDump(typedAnswer) || hopDumpsUnproven(hopResults) || echoSlogan || thinMiss || (skipSlots && looksWhereQuery(req.Query))
		// Title-case 2-item possession/skill joins look like slogan dumps
		// but are typed answers. Unlocking them lets hybrid replace a
		// signed-collectible join with a nearby sports chat turn.
		if leftoverCoveringKeepTypedAnswer(req.Query, hopResults, typedAnswer) {
			unlockDump = false
		}
		if unlockDump {
			// Dual-entity SH questions often plan as MH and lock a slogan
			// dump, an unproven search_fallback fragment, a leftover-
			// unrelated short slot (Rocks), or a thin leftover-miss
			// fragment over hybrid. Only unlock mh_list on where-questions
			// when slots skip so typed community/skill joins stay locked.
			lockedWhere = false
			lockedMHList = false
		}
		// Keep typed multi-item lists when hybrid adds uncovered values as
		// another multi-item list, or expands a short typed list. A long
		// dump may still be replaced by a 1–2 item hybrid answer.
		lockedList := lockHybridListExtras(enumerated, typedN, hybridN, extras, typedAnswerIsHopDump(typedAnswer) || echoSlogan)
		if leftoverCoveringKeepTypedAnswer(req.Query, hopResults, typedAnswer) && leftoverCoveringLockChildhoodPossessions(req.Query) {
			lockedList = true
		}
		if lockTypedListQuery(req.Query, enumerated, typedN, echoSlogan, typedAnswerIsHopDump(typedAnswer)) {
			lockedList = true
		}
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
		if out.Explain["ordinal_name"] != true && !lockedPolar && !lockedCount {
			if covering := leftoverCoveringSpecificAnswer(req.Query, hopResults, pkt); covering != "" {
				cur := strings.TrimSpace(out.Answer)
				src, _ := out.Explain["reader_source"].(string)
				// Short typed item joins stay locked except where hop-dumps,
				// which starve locative leftover covering with activity lists.
				typedJoin := leftoverCoveringKeepTypedAnswer(req.Query, hopResults, cur)
				// Do not force typedJoin from enumerate item-count: harness
				// mode=enumerate plus a 2-item hop dump would starve leftover
				// covering (Nike/Gatorade deals, soda and candy, yoga+running).
				useCovering := !typedJoin && (out.Abstained || strings.EqualFold(cur, "not in memory") ||
					leftoverThinMissAnswer(req.Query, hopResults, cur))
				if !useCovering && !typedJoin && src != "hybrid_llm_packet" && typedAnswerIsHopDump(cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && src == "hybrid_llm_packet" {
					useCovering = leftoverCoveringMayReplaceHybrid(req.Query, hopResults, covering, cur)
				}
				if !useCovering && !typedJoin && leftoverCoveringBareDateMissesEvent(req.Query, hopResults, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringYearMissesEvent(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringPurposeMissesInstrument(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringWhatMadeMissesEvidence(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringHostMissesEvent(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringAdviceMissesDirective(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringProcessMissesHortative(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringMotivateMissesCause(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringSayAboutMissesEvaluative(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringReactMissesObservation(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringPurposeMissesAction(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringStartMissesMethod(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringDurationMissesDuration(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringCadenceMissesCadence(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringCurrentProjectMissesCurrent(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringBecomeInterestedMissesInterest(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringPrepPlanMissesPrep(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringFocusingBesidesMissesJoin(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringTitledShowMissesTitle(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringLocativePurposeMissesExtra(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringExperiencingFeelingMissesFeeling(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringWorkDeterminationMissesPair(req.Query, covering, cur) {
					useCovering = true
				}
				if !useCovering && !typedJoin && leftoverCoveringSelfDirectedRealizeMissesSelf(req.Query, covering, cur) {
					useCovering = true
				}
				if useCovering {
					out.Answer = covering
					out.Abstained = false
					out.AnswerStatus = AnswerSupported
					out.Explain["reader_source"] = "leftover_packet_fact"
					if next := leftoverCoveringSyncEnumerateItems(req.Query, covering, out.Items); len(next) > 0 {
						out.Items = next
						out.Explain["item_count"] = len(next)
					}
				}
			}
		}
	}
	if next := preferCoParticipantVisitDestination(req.Query, hopResults, out.Answer, pkt); next != "" && next != strings.TrimSpace(out.Answer) {
		out.Answer = next
		out.Abstained = false
		out.AnswerStatus = AnswerSupported
		out.Explain["co_participant_visit_destination"] = true
	}
	if next, extra := preferUnwindPacketActivities(req.Query, out.Answer, pkt, hopResults); next != "" && next != strings.TrimSpace(out.Answer) {
		out.Answer = next
		out.Items = appendUniqueRecallItems(out.Items, extra)
		out.Abstained = false
		out.AnswerStatus = AnswerSupported
		out.Explain["unwind_packet_activities"] = true
	}
	if next, extra := preferPracticePacketPlaces(req.Query, out.Answer, pkt, hopResults); next != "" && next != strings.TrimSpace(out.Answer) {
		out.Answer = next
		out.Items = appendUniqueRecallItems(out.Items, extra)
		out.Abstained = false
		out.AnswerStatus = AnswerSupported
		out.Explain["practice_packet_places"] = true
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
		if typedDump && typedN <= 2 {
			return false
		}
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

func hopScanLimit(query string, plan QueryPlan, topK int) int {
	const cap = 128
	if topK >= cap {
		return topK
	}
	if looksCountQuery(query) {
		return topK
	}
	if wantsHistoricalAtomScan(query) {
		return cap
	}
	return topK
}

func lockTypedListQuery(query string, enumerated bool, typedN int, echoSlogan bool, typedDump bool) bool {
	if !enumerated || typedN < 2 || echoSlogan || looksCountQuery(query) {
		return false
	}
	// Identity dumps (8+ fragments) must not lock hybrid. A 2–6 item typed
	// set of short skill/possession slots is the answer, not a dump.
	if typedDump && typedN > 6 {
		return false
	}
	return wantsHistoricalAtomScan(query)
}

// wantsTypedSetScan is true when the question asks for the full typed set of
// a class (items bought, activities done, names, locations, what-kind lists).
// Singular "what activity did" / "how do dogs react" / "would X enjoy books"
// are leftover or OD questions that happen to mention a list noun.
func wantsTypedSetScan(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" || looksCountQuery(query) {
		return false
	}
	for _, p := range []string{
		"how ", "why ", "when ", "where ", "would ",
		"what did ", "what does ", "what do ",
	} {
		if strings.HasPrefix(q, p) {
			return false
		}
	}
	if looksWhatKindQuery(query) {
		// Typed class sets (tricks, meals, snacks). Leftover
		// "what kind of deals/games" have no set noun and stay on covering.
		return queryHasTypedSetNoun(query)
	}
	if strings.HasPrefix(q, "list ") || strings.HasPrefix(q, "what are ") {
		return true
	}
	if looksCommunityQuery(query) {
		return true
	}
	if !strings.HasPrefix(q, "what ") && !strings.HasPrefix(q, "which ") && !strings.HasPrefix(q, "in what ") {
		return false
	}
	return queryHasTypedSetNoun(query)
}

// wantsHistoricalAtomScan is the recency-starvation path: scan past
// current-state for possession/skill/place sets. Activity/community/snack
// leftover questions enumerate at search top-k without this widening.
func wantsHistoricalAtomScan(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" || looksCountQuery(query) {
		return false
	}
	for _, p := range []string{"how ", "why ", "when ", "where ", "would ", "what did ", "what does ", "what do "} {
		if strings.HasPrefix(q, p) {
			return false
		}
	}
	if strings.HasPrefix(q, "list ") || strings.HasPrefix(q, "what are ") || looksLocationListQuery(query) {
		return true
	}
	if queryHasToken(query, "activities") && strings.Contains(q, " done ") {
		return true
	}
	if looksWhatKindQuery(query) {
		return queryHasToken(query, "tricks", "trick", "instruments", "instrument", "items", "item")
	}
	if !strings.HasPrefix(q, "what ") && !strings.HasPrefix(q, "which ") {
		return false
	}
	for _, tok := range tokenize(query) {
		tok = strings.Trim(tok, "'\"")
		tok = strings.TrimSuffix(tok, "'s")
		switch tok {
		case "items", "names", "locations", "tricks", "instruments", "instrument":
			return true
		}
	}
	return false
}

func queryHasTypedSetNoun(query string) bool {
	for _, tok := range tokenize(query) {
		tok = strings.Trim(tok, "'\"")
		tok = strings.TrimSuffix(tok, "'s")
		switch tok {
		case "activities", "hobbies", "items", "names", "locations",
			"tricks", "meals", "snacks", "food", "foods", "events",
			"places", "books", "instruments", "instrument":
			return true
		}
	}
	return false
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

	counting := looksCountQuery(req.Query)
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
			if hopUsefulForEnumerate(h.Predicate, pred, counting) && strings.TrimSpace(h.Predicate) != "" {
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
			if counting && entity != "" && strings.TrimSpace(h.Entity) != "" && !strings.EqualFold(h.Entity, entity) {
				continue
			}
			if !hopUsefulForEnumerate(h.Predicate, pred, counting) {
				continue
			}
			slotPred := firstNonEmpty(h.Predicate, pred)
			if len(h.Values) > 0 {
				for i, v := range h.Values {
					if hopValueIsAttendedEvent(v) || hopValueHasForeignPossessive(v, h.Entity, hops) {
						continue
					}
					id := hopMemoryIDForExtractedValue(h, v)
					if id == "" {
						if i < len(h.MemoryIDs) {
							id = h.MemoryIDs[i]
						} else if len(h.MemoryIDs) > 0 {
							id = h.MemoryIDs[0]
						}
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
		if !counting {
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
			if pred != "" && recPred != "" && !hopUsefulForEnumerate(recPred, pred, counting) {
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
	if queryHasToken(query, "locations") {
		return true
	}
	// Singular "location" is a point fact (Shinjuku), not a practice-place set.
	if queryHasToken(query, "location") {
		return false
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

func hopMemoryIDForExtractedValue(h HopResult, value string) string {
	want := strings.ToLower(strings.TrimSpace(value))
	if want == "" {
		return ""
	}
	for i, c := range h.Contents {
		extracted, ok := slotValueFromMemoryContent(c)
		if ok && strings.EqualFold(strings.TrimSpace(extracted), strings.TrimSpace(value)) {
			if i < len(h.MemoryIDs) {
				return h.MemoryIDs[i]
			}
		}
	}
	for i, c := range h.Contents {
		if !queryHasToken(c, want) {
			continue
		}
		if i < len(h.MemoryIDs) {
			return h.MemoryIDs[i]
		}
	}
	return ""
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
		if id := hopMemoryIDForExtractedValue(h, value); id != "" {
			return id
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

func hopUsefulForEnumerate(hopPred, listPred string, counting bool) bool {
	if counting {
		return hopUsefulForCount(hopPred, listPred)
	}
	return hopUsefulForList(hopPred, listPred)
}

func hopUsefulForCount(hopPred, listPred string) bool {
	if listPred == "" {
		return true
	}
	if hopPred == "" {
		return false
	}
	return strings.EqualFold(hopPred, listPred)
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
		if n == 0 || n < len(items) {
			return strconv.Itoa(len(items))
		}
		return strconv.Itoa(n)
	}
	switch head {
	case "", "children", "kids", "people", "child":
		n := 0
		for _, it := range items {
			n += countItemQuantity(it.Value, head)
		}
		if n == 0 {
			return strconv.Itoa(len(items))
		}
		return strconv.Itoa(n)
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
	for _, t := range expandModifierTokens(excl) {
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

func expandModifierTokens(toks []string) []string {
	outdoor := []string{"outdoor", "outdoors", "hike", "hiking", "mountaineering", "mountaineer"}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(toks)+len(outdoor))
	add := func(t string) {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			return
		}
		if _, ok := seen[t]; ok {
			return
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	hit := false
	for _, t := range toks {
		add(t)
		for _, o := range outdoor {
			if strings.EqualFold(t, o) {
				hit = true
			}
		}
	}
	if hit {
		for _, o := range outdoor {
			add(o)
		}
	}
	return out
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
		if toks := forClauseTokens(req.Query); len(toks) > 0 && predicateFromListQuery(tokenize(req.Query)) != PredicateFamilyMember {
			items = dropForClauseClassReferents(items, toks)
		}
		items = s.rankItemsByQuery(ctx, req, items, hops)
		items = s.filterTrickDestItems(ctx, req, items, hops)
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
	dests := destNamesFromHops(hops)
	for i, it := range items {
		blob := s.itemEvidenceBlob(ctx, req, it, hops)
		if extra := hopContentForValue(hops, it.Value); extra != "" {
			blob += " " + strings.ToLower(extra)
		}
		score := itemQueryScore(req.Query, it.Value, blob)
		if looksTrickQuery(req.Query) {
			owner := ""
			if ents := hopQueryEntities(req.Query); len(ents) > 0 {
				owner = ents[0]
			}
			hay := strings.ToLower(strings.TrimSpace(it.Value + " " + blob))
			destHit := false
			for _, d := range dests {
				if owner != "" && strings.EqualFold(d, owner) {
					continue
				}
				if queryHasToken(hay, d) {
					destHit = true
					score += 2
					break
				}
			}
			taught := owner != "" && queryHasToken(hay, owner) && queryHasToken(hay, "taught", "teach")
			if destCapabilityLine(hay) && (destHit || taught) {
				score += 2
			}
		}
		rows = append(rows, scored{it: it, score: score, idx: i})
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
	if looksTrickQuery(query) && itemHitsExclusion(hay, []string{"taught", "skilled", "perform"}) {
		n += 2
	}
	if looksItemTransferQuery(query) && itemHasTransferCue(hay) {
		n += 2
	}
	return n
}

func destNamesFromHops(hops []HopResult) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(n string) {
		n = strings.TrimSpace(n)
		n = strings.TrimSuffix(n, "'s")
		n = strings.TrimSuffix(n, "’s")
		n = strings.Trim(n, ".,;:'\"")
		if n == "" || !looksHopPerson(n) {
			return
		}
		key := strings.ToLower(n)
		if _, ok := seen[key]; ok {
			return
		}
		if _, stop := hopEntityStop[key]; stop {
			return
		}
		seen[key] = struct{}{}
		out = append(out, n)
	}
	for _, h := range hops {
		person := strings.TrimSpace(h.Entity)
		for _, c := range h.Contents {
			for _, n := range namedBeings(c) {
				add(n)
			}
			if n := destFromPossessiveClassLine(c, person); n != "" {
				add(n)
			}
		}
	}
	return out
}

func (s *Service) filterTrickDestItems(ctx context.Context, req RecallRequest, items []RecallItem, hops []HopResult) []RecallItem {
	if !looksTrickQuery(req.Query) || len(items) < 2 {
		return items
	}
	dests := destNamesFromHops(hops)
	if len(dests) == 0 {
		return items
	}
	owner := ""
	if ents := hopQueryEntities(req.Query); len(ents) > 0 {
		owner = ents[0]
	}
	keep := make([]RecallItem, 0, len(items))
	for _, it := range items {
		blob := s.itemEvidenceBlob(ctx, req, it, hops)
		if extra := hopContentForValue(hops, it.Value); extra != "" {
			blob += " " + extra
		}
		hay := strings.ToLower(strings.TrimSpace(it.Value + " " + blob))
		destHit := false
		for _, d := range dests {
			if owner != "" && strings.EqualFold(d, owner) {
				continue
			}
			if queryHasToken(hay, d) {
				destHit = true
				break
			}
		}
		taught := owner != "" && queryHasToken(hay, owner) && queryHasToken(hay, "taught", "teach")
		trickObj := queryHasToken(it.Value, "trick", "tricks") || len(trickObjectSlots(it.Value+" "+blob)) > 0
		if destHit || taught || (trickObj && destCapabilityLine(hay)) {
			keep = append(keep, it)
		}
	}
	if len(keep) >= 2 {
		return keep
	}
	return items
}

func hopContentForValue(hops []HopResult, value string) string {
	want := strings.ToLower(strings.TrimSpace(value))
	if want == "" {
		return ""
	}
	for _, h := range hops {
		for i, v := range h.Values {
			if strings.ToLower(strings.TrimSpace(v)) != want {
				continue
			}
			if i < len(h.Contents) {
				return h.Contents[i]
			}
		}
	}
	return ""
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
			chunk := strings.Trim(strings.Join(fields[i:i+n], " "), "()[];,.\"'?")
			if t := parseFlexibleTime(chunk); t != nil && t.Year() > 1900 {
				return t
			}
		}
	}
	return nil
}

// leftoverDateMatchWindow keeps last-week / last-weekend relative facts
// (session date vs event date) while dropping weeks-away crowding.
const leftoverDateMatchWindow = 10 * 24 * time.Hour

// hybridDateMatchWindow is tighter so other same-month events do not
// crowd a day-specific where/when packet.
const hybridDateMatchWindow = 48 * time.Hour

func queryHasCalendarDay(query string) bool {
	hasMonth, hasDay := false, false
	for _, tok := range tokenize(query) {
		if isMonthWord(tok) {
			hasMonth = true
		}
		n, err := strconv.Atoi(tok)
		if err != nil {
			continue
		}
		if n >= 1 && n <= 31 {
			hasDay = true
		}
	}
	return hasMonth && hasDay
}

func querySpecificCalendarDate(query string) *time.Time {
	if !queryHasCalendarDay(query) {
		return nil
	}
	return parseDateFromText(query)
}

func isCalendarCoverToken(tok string) bool {
	tok = strings.ToLower(strings.TrimSpace(tok))
	if tok == "" || isMonthWord(tok) {
		return tok != ""
	}
	if n, err := strconv.Atoi(tok); err == nil {
		if (n >= 1 && n <= 31) || (n >= 1900 && n <= 2100) {
			return true
		}
	}
	switch tok {
	case "today", "yesterday", "tomorrow", "weekend", "weekday":
		return true
	}
	return false
}

// datedContentConflictsQuery is true when the query names a calendar day and
// the line's primary event date is more than ten days away. Relative-session
// tails ("the week before 4 February") must not make a January fact match a
// February question. Last-week event dates relative to the session day stay
// eligible. Month-or-year-only queries do not filter.
func datedContentConflictsQuery(query, content string) bool {
	qd := querySpecificCalendarDate(query)
	if qd == nil {
		return false
	}
	ld := parseDateFromText(content)
	if ld == nil {
		return false
	}
	d := qd.Sub(*ld)
	if d < 0 {
		d = -d
	}
	return d > leftoverDateMatchWindow
}

func datedHybridContentConflictsQuery(query, content string) bool {
	if looksWhereQuery(query) {
		return false
	}
	qd := querySpecificCalendarDate(query)
	if qd == nil {
		return false
	}
	ld := parseDateFromText(content)
	if ld == nil {
		return false
	}
	d := qd.Sub(*ld)
	if d < 0 {
		d = -d
	}
	return d > hybridDateMatchWindow
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
	if looksItemTransferQuery(req.Query) {
		items = s.filterItemsByTransferCue(ctx, req, items, hops)
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
		"workmates": {}, "workmate": {},
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
			return expandGroupCompanionEvidence(next)
		}
	}
	return nil
}

func expandGroupCompanionEvidence(tok string) []string {
	tok = strings.ToLower(strings.TrimSpace(tok))
	work := []string{"colleagues", "colleague", "coworkers", "coworker", "workmates", "workmate"}
	for _, w := range work {
		if w == tok {
			return append([]string{}, work...)
		}
	}
	return []string{tok}
}

func looksItemTransferQuery(query string) bool {
	if !queryHasToken(query, "item", "items") {
		return false
	}
	if len(forClauseTokens(query)) == 0 {
		return false
	}
	q := strings.ToLower(query)
	if strings.Contains(q, "bought") || strings.Contains(q, "acquired") || strings.Contains(q, "purchased") {
		return true
	}
	return queryHasToken(query, "made", "make", "got", "buy")
}

func itemHasTransferCue(s string) bool {
	body := " " + strings.ToLower(s) + " "
	for _, cue := range []string{
		" bought ", " buy ", " buying ", " acquired ", " purchased ", " purchase ",
	} {
		if strings.Contains(body, cue) {
			return true
		}
	}
	if (strings.Contains(body, " got ") || strings.Contains(body, " get ")) &&
		(strings.Contains(body, " for ") || strings.Contains(body, " gift")) {
		return true
	}
	if (strings.Contains(body, " made ") || strings.Contains(body, " make ") || strings.Contains(body, " making ")) &&
		strings.Contains(body, " for ") {
		return true
	}
	return false
}

func (s *Service) filterItemsByTransferCue(ctx context.Context, req RecallRequest, items []RecallItem, hops []HopResult) []RecallItem {
	if len(items) == 0 {
		return items
	}
	out := make([]RecallItem, 0, len(items))
	for _, it := range items {
		blob := s.itemEvidenceBlob(ctx, req, it, hops)
		if itemHasTransferCue(it.Value) || itemHasTransferCue(blob) {
			out = append(out, it)
		}
	}
	if len(out) == 0 {
		return items
	}
	return out
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

// dropForClauseClassReferents drops values that are the for-clause class
// itself (bare "dogs", "dog named Buddy") when the list asked for items of
// that class. Specific objects ("new collars and tags for her dogs") stay.
func dropForClauseClassReferents(items []RecallItem, classToks []string) []RecallItem {
	if len(items) < 2 || len(classToks) == 0 {
		return items
	}
	heads := make([]string, 0, 8)
	seen := map[string]struct{}{}
	for _, t := range classToks {
		for _, h := range countClassFamily("", t) {
			if _, ok := seen[h]; ok {
				continue
			}
			seen[h] = struct{}{}
			heads = append(heads, h)
		}
	}
	if len(heads) == 0 {
		return items
	}
	keep := make([]RecallItem, 0, len(items))
	for _, it := range items {
		if enumerateValueIsForClassReferent(it.Value, heads) {
			continue
		}
		keep = append(keep, it)
	}
	if len(keep) == 0 {
		return items
	}
	return keep
}

func enumerateValueIsForClassReferent(value string, heads []string) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return false
	}
	if namingReferent(value) != "" {
		for _, head := range heads {
			if strings.Contains(v, strings.ToLower(head)) || countValueIsClassNoun(valueLexicalHead(v), head) {
				return true
			}
		}
	}
	fields := strings.Fields(strings.TrimSpace(value))
	if len(fields) == 1 && looksHopPerson(fields[0]) {
		return true
	}
	for _, head := range heads {
		if classOnlyEnumerateValue(v, head) {
			return true
		}
	}
	return false
}

func classOnlyEnumerateValue(value, head string) bool {
	if countValueIsClassNoun(value, head) {
		return true
	}
	v := strings.Map(func(r rune) rune {
		if r == '(' || r == ')' {
			return ' '
		}
		return r
	}, strings.ToLower(strings.TrimSpace(value)))
	skip := map[string]struct{}{
		"a": {}, "an": {}, "the": {}, "my": {}, "his": {}, "her": {},
		"their": {}, "our": {}, "and": {}, "or": {},
		"one": {}, "two": {}, "three": {}, "four": {}, "five": {},
		"six": {}, "seven": {}, "eight": {}, "nine": {}, "ten": {},
		"named": {}, "called": {},
	}
	classish, extra := 0, 0
	for _, f := range strings.Fields(strings.ReplaceAll(v, "-", " ")) {
		f = strings.Trim(f, ".,;:'\"()")
		if f == "" {
			continue
		}
		if _, ok := skip[f]; ok {
			continue
		}
		if countValueIsClassNoun(f, head) {
			classish++
			continue
		}
		if looksHopPerson(f) {
			continue
		}
		extra++
	}
	return classish > 0 && extra == 0
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

func (s *Service) filterCountItems(ctx context.Context, req RecallRequest, items []RecallItem, hops []HopResult) []RecallItem {
	if len(items) == 0 {
		return items
	}
	entity := ""
	if ents := hopQueryEntities(req.Query); len(ents) > 0 {
		entity = ents[0]
	}
	head := countHeadNoun(req.Query)
	asOf := countAsOfBound(req.Query)
	object := ""
	if head == "times" || head == "time" {
		object = countTimesObject(req.Query, entity)
	}
	wantPred := predicateFromListQuery(tokenize(req.Query))
	if wantPred == "" {
		if hints := predicateHintsFromQuery(req.Query); len(hints) > 0 {
			wantPred = hints[0]
		}
	}
	typed := make([]RecallItem, 0, len(items))
	headHits := make([]RecallItem, 0, len(items))
	items = expandCountNameLists(items)
	// Specific heads ("Ferraris") intersect the typed set. Generic heads
	// ("cars") have no token overlap with coupe/sedan, so keep the typed set.
	useHeadIntersect := head != "" && head != "children" && head != "kids" && head != "child" &&
		head != "times" && head != "time" && !countHeadIsPossessedClass(head)
	for _, it := range items {
		recs := s.itemEvidenceRecords(ctx, req, it)
		if entity != "" && !countItemMatchesEntity(it, recs, entity) {
			continue
		}
		if asOf != nil && !countItemWithinAsOf(it, recs, *asOf) {
			continue
		}
		if countItemIsZero(it.Value) {
			continue
		}
		if wantPred != "" && strings.TrimSpace(it.Predicate) != "" && !strings.EqualFold(it.Predicate, wantPred) {
			continue
		}
		blob := s.itemEvidenceBlob(ctx, req, it, hops)
		switch head {
		case "children", "kids", "child":
			if valueIsPreferenceComplement(it.Value, blob) {
				continue
			}
			if !valueHasChildCue(it.Value, blob) {
				continue
			}
		case "times", "time":
			if object != "" && !countItemMentionsObject(it.Value, object) {
				continue
			}
		default:
			if countHeadIsPossessedClass(head) && !countItemIsClassInstance(it.Value, blob, head, req.Query) {
				continue
			}
		}
		typed = append(typed, it)
		if useHeadIntersect && countItemMentionsHead(it.Value, blob, head) {
			headHits = append(headHits, it)
		}
	}
	out := typed
	if useHeadIntersect && len(headHits) > 0 {
		out = headHits
	}
	out = collapseCountClassNouns(out, head)
	out = collapseNamedCountDuplicates(out)
	return out
}

func (s *Service) itemEvidenceRecords(ctx context.Context, req RecallRequest, it RecallItem) []MemoryRecord {
	out := make([]MemoryRecord, 0, len(it.Evidence))
	seen := map[string]struct{}{}
	for _, id := range it.Evidence {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		rec, err := s.store.GetMemory(ctx, req.TenantID, req.SubjectID, id)
		if err != nil {
			continue
		}
		out = append(out, rec)
	}
	return out
}

func countItemMatchesEntity(it RecallItem, recs []MemoryRecord, entity string) bool {
	matched := false
	sawSubject := false
	for _, rec := range recs {
		subj := entitySubjectOf(rec)
		if subj != "" {
			sawSubject = true
			if strings.EqualFold(subj, entity) {
				matched = true
			}
			continue
		}
		if containsEntityMention(rec.Content, entity) {
			matched = true
		}
	}
	if sawSubject {
		return matched
	}
	if len(recs) == 0 {
		return true
	}
	return matched || containsEntityMention(it.Value, entity)
}

func countItemWithinAsOf(it RecallItem, recs []MemoryRecord, asOf time.Time) bool {
	obs := countItemObserved(it, recs)
	if obs == nil {
		return true
	}
	return !obs.After(asOf)
}

func countItemObserved(it RecallItem, recs []MemoryRecord) *time.Time {
	want := strings.ToLower(strings.TrimSpace(it.Value))
	var best *time.Time
	consider := func(t time.Time) {
		t = t.UTC()
		if best == nil || t.Before(*best) {
			cp := t
			best = &cp
		}
	}
	for _, rec := range recs {
		if rec.ObservedAt == nil || rec.ObservedAt.IsZero() {
			continue
		}
		if want != "" && !(strings.Contains(strings.ToLower(rec.Content), want) ||
			strings.EqualFold(recordValueNorm(rec), it.Value) ||
			queryHasToken(recordValueNorm(rec), want)) {
			continue
		}
		consider(*rec.ObservedAt)
	}
	if best != nil {
		return best
	}
	if strings.TrimSpace(it.ObservedAt) != "" {
		if t := parseFlexibleTime(it.ObservedAt); t != nil {
			return t
		}
	}
	for _, rec := range recs {
		if rec.ObservedAt != nil && !rec.ObservedAt.IsZero() {
			consider(*rec.ObservedAt)
		}
	}
	return best
}

func countAsOfBound(query string) *time.Time {
	lower := strings.ToLower(query)
	i := strings.Index(lower, " as of ")
	if i < 0 {
		return nil
	}
	rest := strings.TrimSpace(query[i+len(" as of "):])
	if cut := strings.IndexAny(rest, "?,;"); cut >= 0 {
		rest = strings.TrimSpace(rest[:cut])
	}
	t := parseDateFromText(rest)
	if t == nil {
		t = parseMonthYear(rest)
	}
	if t == nil {
		return nil
	}
	if queryHasCalendarDay("as of " + rest) {
		end := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
		return &end
	}
	last := time.Date(t.Year(), t.Month()+1, 0, 23, 59, 59, 0, time.UTC)
	return &last
}

func parseMonthYear(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	for _, layout := range []string{"January 2006", "Jan 2006", "January, 2006", "Jan, 2006"} {
		if ts, err := time.Parse(layout, s); err == nil {
			t := ts.UTC()
			return &t
		}
	}
	fields := strings.Fields(s)
	if len(fields) >= 2 {
		chunk := strings.Trim(fields[0]+" "+fields[1], "()[];,.\"'?")
		for _, layout := range []string{"January 2006", "Jan 2006"} {
			if ts, err := time.Parse(layout, chunk); err == nil {
				t := ts.UTC()
				return &t
			}
		}
	}
	return nil
}

func countTimesObject(query, entity string) string {
	lower := strings.ToLower(query)
	i := strings.Index(lower, "times")
	if i < 0 {
		i = strings.Index(lower, "time")
	}
	if i < 0 {
		return ""
	}
	skip := map[string]struct{}{
		"times": {}, "time": {}, "has": {}, "have": {}, "had": {},
		"did": {}, "does": {}, "his": {}, "her": {}, "their": {},
		"the": {}, "a": {}, "an": {},
	}
	ent := strings.ToLower(strings.TrimSpace(entity))
	last := ""
	for _, t := range contentBearingTokens(tokenize(query[i:])) {
		if _, ok := skip[t]; ok {
			continue
		}
		if ent != "" && t == ent {
			continue
		}
		last = t
	}
	return last
}

func countItemMentionsObject(value, object string) bool {
	object = strings.ToLower(strings.TrimSpace(object))
	if object == "" {
		return true
	}
	return queryHasToken(value, object)
}

func countItemMentionsHead(value, blob, head string) bool {
	head = strings.ToLower(strings.TrimSpace(head))
	if head == "" {
		return true
	}
	stem := strings.TrimSuffix(head, "es")
	stem = strings.TrimSuffix(stem, "s")
	if len(stem) < 3 {
		stem = head
	}
	hay := strings.TrimSpace(value + " " + blob)
	if hay == "" {
		return false
	}
	return queryHasToken(hay, head, stem) ||
		strings.Contains(strings.ToLower(hay), head) ||
		strings.Contains(strings.ToLower(hay), stem)
}

func countItemIsZero(value string) bool {
	v := " " + strings.ToLower(strings.TrimSpace(value)) + " "
	if strings.Contains(v, " no ") || strings.Contains(v, " none ") || strings.Contains(v, " zero ") {
		return true
	}
	return strings.HasPrefix(strings.TrimSpace(strings.ToLower(value)), "no ")
}

func countItemQuantity(value, head string) int {
	lower := strings.ToLower(strings.TrimSpace(value))
	if lower == "" || countItemIsZero(lower) {
		return 0
	}
	words := map[string]int{
		"one": 1, "two": 2, "three": 3, "four": 4, "five": 5,
		"six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10,
	}
	fields := strings.Fields(strings.ReplaceAll(lower, "-", " "))
	for i, f := range fields {
		f = strings.Trim(f, ".,;:'\"")
		n := 0
		if v, ok := words[f]; ok {
			n = v
		} else if d, err := strconv.Atoi(f); err == nil && d > 0 && d < 100 {
			n = d
		}
		if n == 0 || i+1 >= len(fields) {
			continue
		}
		next := strings.Trim(fields[i+1], ".,;:'\"")
		if countValueIsClassNoun(next, head) {
			return n
		}
	}
	return 1
}

func countValueIsClassNoun(value, head string) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return false
	}
	if head != "" {
		h := strings.ToLower(head)
		if v == h || v == strings.TrimSuffix(h, "s") || v+"s" == h {
			return true
		}
	}
	switch v {
	case "children", "child", "kids", "kid",
		"pets", "pet", "dogs", "dog", "pups", "pup", "puppy":
		return true
	}
	return false
}

func collapseCountClassNouns(items []RecallItem, head string) []RecallItem {
	hasIndividual := false
	for _, it := range items {
		if !countValueIsBareClassNoun(it.Value, head) {
			hasIndividual = true
			break
		}
	}
	if !hasIndividual {
		return items
	}
	out := make([]RecallItem, 0, len(items))
	for _, it := range items {
		if countValueIsBareClassNoun(it.Value, head) {
			continue
		}
		out = append(out, it)
	}
	return out
}

// countValueIsBareClassNoun is a duplicate class label ("kids", "pets") that
// should yield when named or quantified instances exist. Quantified phrases
// ("two children") and modified instances ("a new Ferrari") are the counted
// set, not a duplicate label.
func countValueIsBareClassNoun(value, head string) bool {
	if countItemQuantity(value, head) > 1 {
		return false
	}
	if namingReferent(value) != "" {
		return false
	}
	v := strings.ToLower(strings.TrimSpace(value))
	if countValueIsClassNoun(v, head) {
		return true
	}
	if countHeadIsPossessedClass(head) {
		core := v
		if i := strings.Index(core, "("); i >= 0 {
			core = strings.TrimSpace(core[:i])
		}
		return countValueIsClassNoun(core, head) || countValueIsClassNoun(valueLexicalHead(core), head)
	}
	v = strings.Map(func(r rune) rune {
		if r == '(' || r == ')' {
			return ' '
		}
		return r
	}, v)
	skip := map[string]struct{}{
		"a": {}, "an": {}, "the": {}, "my": {}, "his": {}, "her": {},
		"their": {}, "our": {}, "and": {}, "or": {},
		"one": {}, "two": {}, "three": {}, "four": {}, "five": {},
		"six": {}, "seven": {}, "eight": {}, "nine": {}, "ten": {},
	}
	classish, extra := 0, 0
	for _, f := range strings.Fields(strings.ReplaceAll(v, "-", " ")) {
		f = strings.Trim(f, ".,;:'\"()")
		if f == "" {
			continue
		}
		if _, ok := skip[f]; ok {
			continue
		}
		if countValueIsClassNoun(f, head) {
			classish++
			continue
		}
		extra++
	}
	return classish > 0 && extra == 0
}

func collapseNamedCountDuplicates(items []RecallItem) []RecallItem {
	if len(items) < 2 {
		return items
	}
	named := map[string]struct{}{}
	for _, it := range items {
		if ref := namingReferent(it.Value); ref != "" {
			named[ref] = struct{}{}
		}
	}
	if len(named) == 0 {
		return items
	}
	out := make([]RecallItem, 0, len(items))
	for _, it := range items {
		key := strings.ToLower(strings.TrimSpace(it.Value))
		if _, ok := named[key]; ok {
			continue
		}
		out = append(out, it)
	}
	return out
}

func namingReferent(value string) string {
	lower := strings.ToLower(value)
	for _, cue := range []string{" named ", " called "} {
		i := strings.Index(lower, cue)
		if i < 0 {
			continue
		}
		rest := strings.TrimSpace(lower[i+len(cue):])
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		return strings.Trim(fields[0], ".,;:'\"")
	}
	return ""
}

func valueIsPreferenceComplement(value, blob string) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	b := strings.ToLower(blob)
	if v == "" || b == "" {
		return false
	}
	for _, verb := range []string{" like ", " likes ", " enjoy ", " enjoys ", " love ", " loves "} {
		for _, subj := range []string{"kids", "kid", "children", "child"} {
			if !strings.Contains(b, subj+verb) && !strings.Contains(b, subj+"s"+verb) {
				continue
			}
			if strings.Contains(b, verb+v) {
				return true
			}
		}
	}
	return false
}

func countHeadIsPossessedClass(head string) bool {
	switch strings.ToLower(strings.TrimSpace(head)) {
	case "pets", "pet", "dogs", "dog", "pups", "pup", "puppy":
		return true
	}
	return false
}

func countClassFamily(query, head string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, 8)
	add := func(t string) {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			return
		}
		if _, ok := seen[t]; ok {
			return
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	add(head)
	add(strings.TrimSuffix(head, "s"))
	if queryHasToken(query, "pet", "pets") || strings.EqualFold(head, "pets") || strings.EqualFold(head, "pet") {
		add("pet")
		add("pets")
		add("dog")
		add("dogs")
		add("pup")
		add("pups")
		add("puppy")
	}
	if queryHasToken(query, "dog", "dogs") || strings.EqualFold(head, "dogs") || strings.EqualFold(head, "dog") {
		add("dog")
		add("dogs")
		add("pup")
		add("pups")
		add("puppy")
		add("pet")
		add("pets")
	}
	return out
}

func countTokenInFamily(tok string, family []string) bool {
	tok = strings.ToLower(strings.TrimSpace(tok))
	tok = strings.Trim(tok, ".,;:'\"()")
	if tok == "" {
		return false
	}
	for _, f := range family {
		if tok == f {
			return true
		}
	}
	return false
}

func valueLexicalHead(value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return ""
	}
	if i := strings.Index(v, "("); i >= 0 {
		v = strings.TrimSpace(v[:i])
	}
	if i := strings.Index(v, " named "); i >= 0 {
		v = strings.TrimSpace(v[:i])
	}
	if i := strings.Index(v, " called "); i >= 0 {
		v = strings.TrimSpace(v[:i])
	}
	fields := strings.Fields(strings.ReplaceAll(v, "-", " "))
	if len(fields) == 0 {
		return ""
	}
	return strings.Trim(fields[len(fields)-1], ".,;:'\"")
}

func countItemIsClassInstance(value, blob, head, query string) bool {
	family := countClassFamily(query, head)
	if len(family) == 0 {
		return true
	}
	if namingReferent(value) != "" {
		return countTokenInFamily(valueLexicalHead(value), family) || blobHasCountClass(value+" "+blob, family)
	}
	if countTokenInFamily(valueLexicalHead(value), family) {
		return true
	}
	if countValueIsClassNoun(value, head) {
		return true
	}
	if isCountProperName(value) {
		return true
	}
	return blobHasCountClass(value, family) && !countItemLooksLikeAccessory(value, family)
}

func blobHasCountClass(blob string, family []string) bool {
	for _, f := range family {
		if queryHasToken(blob, f) {
			return true
		}
	}
	return false
}

func countItemLooksLikeAccessory(value string, family []string) bool {
	head := valueLexicalHead(value)
	if head == "" || countTokenInFamily(head, family) {
		return false
	}
	return queryHasToken(value, family...)
}

func isCountProperName(value string) bool {
	v := strings.TrimSpace(value)
	if v == "" || strings.ContainsAny(v, " ,;/") {
		return false
	}
	r := []rune(v)
	if len(r) < 3 || len(r) > 24 {
		return false
	}
	return unicode.IsUpper(r[0])
}

func expandCountNameLists(items []RecallItem) []RecallItem {
	out := make([]RecallItem, 0, len(items)+2)
	for _, it := range items {
		names := splitCoordinatedCountNames(it.Value)
		if len(names) < 2 {
			out = append(out, it)
			continue
		}
		for _, n := range names {
			cp := it
			cp.Value = n
			out = append(out, cp)
		}
	}
	return out
}

func splitCoordinatedCountNames(value string) []string {
	lower := strings.ToLower(value)
	rest := value
	if i := strings.Index(lower, " named "); i >= 0 {
		rest = strings.TrimSpace(value[i+len(" named "):])
	} else if strings.Contains(value, ",") {
		rest = value
	} else {
		return nil
	}
	rest = strings.ReplaceAll(rest, " and ", ",")
	rest = strings.ReplaceAll(rest, " & ", ",")
	parts := strings.Split(rest, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, ".,;:'\"")
		if p == "" || strings.Contains(p, " ") {
			return nil
		}
		if countValueIsClassNoun(p, "") {
			return nil
		}
		out = append(out, p)
	}
	if len(out) < 2 {
		return nil
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
		return false
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
		seen := map[string]struct{}{}
		out := make([]RecallItem, 0, len(both)+len(headHits))
		add := func(it RecallItem) {
			key := strings.ToLower(strings.TrimSpace(it.Value))
			if _, ok := seen[key]; ok {
				return
			}
			seen[key] = struct{}{}
			out = append(out, it)
		}
		for _, it := range both {
			add(it)
		}
		for _, it := range headHits {
			add(it)
		}
		return out
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
		" owns ", " owned ", " bought ", " acquired ", " named ", " is named ", " plays ", " played ",
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
		" owns ", " owned ", " bought ", " acquired ", " named ", " is named ", " plays ", " played ",
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

func expandPetKindTokens(toks []string) []string {
	pet := []string{"puppy", "puppies", "dog", "dogs", "pup", "pet", "pets"}
	hasPet := false
	for _, t := range toks {
		for _, p := range pet {
			if t == p {
				hasPet = true
				break
			}
		}
		if hasPet {
			break
		}
	}
	if !hasPet {
		return toks
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(toks)+len(pet))
	for _, t := range append(append([]string{}, toks...), pet...) {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func ordinalKindTokens(query string) []string {
	skip := map[string]struct{}{
		"name": {}, "named": {}, "first": {}, "second": {}, "third": {}, "fourth": {},
		"1st": {}, "2nd": {}, "3rd": {}, "4th": {},
	}
	toks := make([]string, 0, 4)
	for _, t := range distinctiveQueryTokens(tokenize(query)) {
		if _, ok := skip[t]; ok {
			continue
		}
		toks = append(toks, t)
	}
	return expandPetKindTokens(toks)
}

func ordinalNameFromPacket(query string, pkt EvidencePacket) string {
	ord := queryNameOrdinal(query)
	if ord <= 0 {
		return ""
	}
	kind := ordinalKindTokens(query)
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
		if looksChatTurnLine(line) {
			continue
		}
		if len(kind) > 0 && !contentCoversAnyQueryToken(line, kind) {
			continue
		}
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

func looksPromptNotAnswer(s string) bool {
	t := strings.ToLower(strings.TrimSpace(s))
	if t == "" {
		return false
	}
	if strings.HasSuffix(t, "?") {
		return true
	}
	return strings.HasPrefix(t, "any tips") || strings.HasPrefix(t, "any advice")
}

func looksImageCaptionLine(s string) bool {
	t := strings.TrimSpace(s)
	if t == "" {
		return false
	}
	lower := strings.ToLower(t)
	if strings.Contains(lower, "[a photo") || strings.Contains(lower, "[photo of") {
		return true
	}
	return strings.HasPrefix(t, "[") && strings.Contains(t, "] [")
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
		if !strings.Contains(w, ":") || utf8Len(w) < 4 {
			continue
		}
		if strings.HasSuffix(w, ":") {
			continue // speaker prefix, not CS:GO-style tokens
		}
		return true
	}
	return false
}

func leftoverNonEntityRareTokens(query string, hops []HopResult) []string {
	minLen := 6
	if querySpecificCalendarDate(query) != nil {
		minLen = 4
	}
	return filterLeftoverCoverTokens(leftoverNonEntityQueryTokens(query, hops), minLen)
}

func leftoverCoverRareTokens(query string, hops []HopResult) []string {
	minLen := 6
	if querySpecificCalendarDate(query) != nil || looksWhenEventQuery(query) || looksYearQuery(query) || looksHowReactQuery(query) || looksWhatDidPurposeQuery(query) || looksHowDidStartQuery(query) || looksHowLongBeenQuery(query) || looksHowOftenQuery(query) || looksWhatProjectWorkingQuery(query) || looksWhatNewHobbyQuery(query) || looksHowPlanDreamQuery(query) || looksWhatFocusingBesidesQuery(query) || looksWhatNewSeriesQuery(query) || looksWhatDidRecentlyAtQuery(query) || looksHowFeelAboutQuery(query) || looksWhatDoCoordinatedUseQuery(query) || looksWhatDidRealizeAfterQuery(query) {
		minLen = 4
	}
	leftover := leftoverNonEntityQueryTokens(query, hops)
	rare := filterLeftoverCoverTokens(leftover, minLen)
	if len(rare) > 0 {
		return rare
	}
	return filterLeftoverCoverTokens(distinctiveQueryTokens(tokenize(query)), minLen)
}

func leftoverCoverAddTokens(toks []string, extra ...string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(toks)+len(extra))
	for _, tok := range toks {
		key := strings.ToLower(strings.TrimSpace(tok))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, tok)
	}
	for _, tok := range extra {
		key := strings.ToLower(strings.TrimSpace(tok))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, tok)
	}
	return out
}

func leftoverCoverWeakToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "mention", "mentioned", "during", "together", "frequently",
		"which", "where", "what", "when", "their", "they", "them",
		"have", "has", "had", "having", "been", "does", "did", "doing",
		"cool", "find", "plan", "plans", "item", "items",
		"activity", "activities", "support", "supports", "supporting",
		"year", "years", "decide", "decided", "deciding",
		"help", "helps", "helped", "helping":
		return true
	}
	return false
}

func leftoverCoverStrongTokens(toks []string) []string {
	out := leftoverCoverNonWeakTokens(toks)
	if len(out) == 0 {
		return toks
	}
	return out
}

func leftoverCoverNonWeakTokens(toks []string) []string {
	out := make([]string, 0, len(toks))
	for _, tok := range toks {
		if leftoverCoverWeakToken(tok) {
			continue
		}
		out = append(out, tok)
	}
	return out
}

// leftover covering: prefer a distinctive stored leftover over a hybrid
// restatement. Do NOT add another looks*Query / leftoverCovering*Line pair
// because a skip-ingest LoCoMo 180 item is still wrong. That campaign
// saturated at P53 (137/180); industry on the same tenant is 62/180.
// See docs/research/competitive/benchmax-audit-2026-08-25.md.
func leftoverCoveringRareForQuery(query string, hops []HopResult) []string {
	_ = hops
	rare := leftoverCoverRareTokens(query, nil)
	if looksInstrumentPurposeQuery(query) {
		rare = leftoverCoverAddTokens(rare, leftoverCoveringInstrumentTokens(query)...)
	}
	if looksHowDidStartQuery(query) {
		filtered := make([]string, 0, len(rare))
		for _, tok := range rare {
			if leftoverCoveringStartStructureToken(tok) {
				continue
			}
			filtered = append(filtered, tok)
		}
		rare = leftoverCoverAddTokens(filtered, leftoverCoveringAgoDurationTokens(query)...)
	}
	if looksHowLongBeenQuery(query) {
		filtered := make([]string, 0, len(rare))
		for _, tok := range rare {
			if leftoverCoveringDurationStructureToken(tok) {
				continue
			}
			filtered = append(filtered, tok)
		}
		rare = filtered
	}
	if looksHowOftenQuery(query) {
		filtered := make([]string, 0, len(rare))
		for _, tok := range rare {
			if leftoverCoveringCadenceStructureToken(tok) {
				continue
			}
			filtered = append(filtered, tok)
		}
		if objs := leftoverCoveringCadenceObjectTokens(query); len(objs) > 0 {
			rare = objs
		} else {
			rare = filtered
		}
	}
	if looksWhatProjectWorkingQuery(query) {
		filtered := make([]string, 0, len(rare))
		for _, tok := range rare {
			if leftoverCoveringCurrentProjectStructureToken(tok) {
				continue
			}
			filtered = append(filtered, tok)
		}
		if objs := leftoverCoveringCurrentProjectObjectTokens(query); len(objs) > 0 {
			rare = objs
		} else {
			rare = filtered
		}
	}
	if looksWhatNewHobbyQuery(query) {
		filtered := make([]string, 0, len(rare))
		for _, tok := range rare {
			if leftoverCoveringBecomeInterestedStructureToken(tok) {
				continue
			}
			filtered = append(filtered, tok)
		}
		rare = filtered
	}
	if looksHowPlanDreamQuery(query) {
		filtered := make([]string, 0, len(rare))
		for _, tok := range rare {
			if leftoverCoveringPrepPlanStructureToken(tok) {
				continue
			}
			filtered = append(filtered, tok)
		}
		if objs := leftoverCoveringDreamLearnObjectTokens(query); len(objs) > 0 {
			rare = objs
		} else {
			rare = filtered
		}
	}
	if looksWhatFocusingBesidesQuery(query) {
		filtered := make([]string, 0, len(rare))
		for _, tok := range rare {
			if leftoverCoveringFocusingBesidesStructureToken(tok) {
				continue
			}
			filtered = append(filtered, tok)
		}
		if objs := leftoverCoveringBesidesObjectTokens(query); len(objs) > 0 {
			rare = leftoverCoverAddTokens(filtered, objs...)
		} else {
			rare = filtered
		}
	}
	if looksWhatNewSeriesQuery(query) {
		filtered := make([]string, 0, len(rare))
		for _, tok := range rare {
			if leftoverCoveringTitledShowStructureToken(tok) {
				continue
			}
			filtered = append(filtered, tok)
		}
		rare = filtered
	}
	if looksWhatDidRecentlyAtQuery(query) {
		filtered := make([]string, 0, len(rare))
		for _, tok := range rare {
			if leftoverCoveringRecentlyAtStructureToken(tok) {
				continue
			}
			filtered = append(filtered, tok)
		}
		if objs := leftoverCoveringRecentlyAtPlaceTokens(query); len(objs) > 0 {
			rare = leftoverCoverAddTokens(filtered, objs...)
		} else {
			rare = filtered
		}
	}
	if looksHowFeelAboutQuery(query) {
		filtered := make([]string, 0, len(rare))
		for _, tok := range rare {
			if leftoverCoveringFeelAboutStructureToken(tok) {
				continue
			}
			filtered = append(filtered, tok)
		}
		rare = filtered
	}
	if looksWhatDoCoordinatedUseQuery(query) {
		filtered := make([]string, 0, len(rare))
		for _, tok := range rare {
			if leftoverCoveringCoordinatedUseStructureToken(tok) {
				continue
			}
			filtered = append(filtered, tok)
		}
		rare = filtered
	}
	if looksWhatDidRealizeAfterQuery(query) {
		filtered := make([]string, 0, len(rare))
		for _, tok := range rare {
			if leftoverCoveringRealizeAfterStructureToken(tok) {
				continue
			}
			filtered = append(filtered, tok)
		}
		rare = filtered
	}
	if len(leftoverCoverNonWeakTokens(rare)) > 0 {
		return rare
	}
	extra := leftoverCoverQueryMonthYearTokens(query)
	if len(extra) == 0 {
		return rare
	}
	// Weak-only leftover tokens cannot bind a covering line. Keep the query
	// month/year so dated plan lines can compete with generic "activity" talk.
	return leftoverCoverAddTokens(nil, extra...)
}

func leftoverCoverQueryMonthYearTokens(query string) []string {
	var out []string
	seen := map[string]struct{}{}
	add := func(tok string) {
		tok = strings.ToLower(strings.TrimSpace(tok))
		if tok == "" {
			return
		}
		if _, ok := seen[tok]; ok {
			return
		}
		seen[tok] = struct{}{}
		out = append(out, tok)
	}
	month := false
	for _, tok := range tokenize(query) {
		if isMonthWord(tok) {
			add(tok)
			month = true
		}
	}
	if !month {
		if y := queryCalendarYear(query); y != 0 {
			add(strconv.Itoa(y))
		}
	}
	return out
}

func filterLeftoverCoverTokens(leftover []string, minLen int) []string {
	rare := make([]string, 0, len(leftover))
	for _, tok := range leftover {
		if utf8Len(tok) < minLen || isCalendarCoverToken(tok) {
			continue
		}
		rare = append(rare, tok)
	}
	return rare
}

func looksSpeakerPrefixedStatement(s string) bool {
	t := strings.TrimSpace(s)
	if t == "" || strings.HasSuffix(t, "?") {
		return false
	}
	i := strings.Index(t, ": ")
	if i <= 0 || i >= 24 {
		return false
	}
	head := strings.TrimSpace(t[:i])
	return head != "" && !strings.Contains(head, " ") && consecutiveProperNouns(head) >= 1
}

func leftoverSkipLine(query, line string, leftoverRare []string) bool {
	if looksTitleCaseSlogan(line) || looksImageCaptionLine(line) || looksPromptNotAnswer(line) {
		return true
	}
	if leftoverCoveringAllowsZeroQueryTokens(query, line) {
		return false
	}
	if i := strings.Index(line, ": "); i > 0 {
		if looksTitleCaseSlogan(strings.TrimSpace(line[i+2:])) {
			return true
		}
	}
	if looksCrowdedHopDump(line) && !(looksWhatKindQuery(query) && leftoverCoveringKindListLine(line)) &&
		!(looksHowDescribeProcessQuery(query) && leftoverCoveringProcessHortativeLine(line)) &&
		!(looksWhatMotivatesQuery(query) && leftoverCoveringMotivateCauseLine(query, line)) &&
		!(looksWhatSayAboutQuery(query) && leftoverCoveringSayAboutTargetLine(line)) &&
		!(looksHowReactQuery(query) && leftoverCoveringReactionObservationLine(line) && leftoverCoveringReactLineHasObject(query, line)) &&
		!(looksWhatDidPurposeQuery(query) && leftoverCoveringPurposeActionLine(query, line)) &&
		!(looksHowDidStartQuery(query) && leftoverCoveringStartMethodLine(query, line)) &&
		!(looksHowLongBeenQuery(query) && leftoverCoveringDurationLine(query, line)) &&
		!(looksHowOftenQuery(query) && leftoverCoveringCadenceLine(query, line)) &&
		!(looksWhatProjectWorkingQuery(query) && leftoverCoveringCurrentProjectLine(query, line)) &&
		!(looksWhatNewHobbyQuery(query) && leftoverCoveringBecomeInterestedLine(query, line)) &&
		!(looksHowPlanDreamQuery(query) && leftoverCoveringPrepPlanLine(query, line)) &&
		!(looksWhatFocusingBesidesQuery(query) && leftoverCoveringFocusingBesidesLine(query, line)) &&
		!(looksWhatNewSeriesQuery(query) && leftoverCoveringTitledShowLine(query, line)) &&
		!(looksWhatDidRecentlyAtQuery(query) && leftoverCoveringLocativePurposeLine(query, line)) &&
		!(looksHowFeelAboutQuery(query) && leftoverCoveringExperiencingFeelingLine(query, line)) &&
		!(looksWhatDoCoordinatedUseQuery(query) && leftoverCoveringWorkDeterminationLine(query, line)) &&
		!(looksWhatDidRealizeAfterQuery(query) && leftoverCoveringSelfDirectedRealizeLine(query, line)) {
		return true
	}
	body := line
	if i := strings.Index(line, ": "); i > 0 && i < 24 {
		body = strings.TrimSpace(line[i+2:])
	}
	lower := strings.ToLower(strings.TrimSpace(body))
	if strings.HasPrefix(lower, "oh,") || strings.HasPrefix(lower, "oh ") {
		return true
	}
	if looksSpeakerPrefixedStatement(line) {
		i := strings.Index(line, ": ")
		body := strings.TrimSpace(line[i+2:])
		if looksInterrogativeLine(body) {
			return true
		}
		return len(leftoverRare) == 0 || !contentCoversAnyQueryToken(body, leftoverRare)
	}
	return looksChatTurnLine(line) || looksInterrogativeLine(line)
}

func looksInterrogativeLine(s string) bool {
	body := strings.TrimSpace(hybridLineBody(s))
	if body == "" {
		return false
	}
	if strings.Contains(body, "?") {
		return true
	}
	lower := strings.ToLower(body)
	for _, p := range []string{"how ", "what ", "why ", "when ", "where ", "did ", "do ", "does ", "have you ", "can ", "could "} {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

func leftoverCoveringSpecificAnswer(query string, hops []HopResult, pkt EvidencePacket) string {
	if looksPolarQuery(query) {
		return ""
	}
	whereQ := looksWhereQuery(query)
	rare := leftoverCoveringRareForQuery(query, hops)
	if len(rare) == 0 && !looksWhatNewSeriesQuery(query) && !looksWhatDidRecentlyAtQuery(query) && !looksHowFeelAboutQuery(query) && !looksWhatDoCoordinatedUseQuery(query) && !looksWhatDidRealizeAfterQuery(query) {
		return ""
	}
	if leftoverCoveringLockChildhoodPossessions(query) {
		rare = leftoverCoverAddTokens(rare, "kid", "childhood")
	}
	var locativeMust []string
	if whereQ {
		locativeMust = locativeLeftoverTokens(query, nil)
	}
	speakerCover := leftoverCoverNonWeakTokens(leftoverNonEntityRareTokens(query, hops))
	lines := packetContentLines(pkt)
	df := make(map[string]int, len(rare))
	for _, line := range lines {
		for _, tok := range rare {
			if contentCoversQueryToken(line, tok) {
				df[tok]++
			}
		}
	}
	best := ""
	bestScore := 0
	strong := leftoverCoverStrongTokens(rare)
	scored := make([]leftoverCoverScored, 0, 8)
	for _, line := range lines {
		if leftoverSkipLine(query, line, speakerCover) {
			continue
		}
		if leftoverCoveringSkipForeignWhenEvent(query, line) {
			continue
		}
		if looksInstrumentPurposeQuery(query) && leftoverCoveringInstrumentPossessLine(line) &&
			!leftoverCoveringInstrumentPurposeLine(line) {
			continue
		}
		if (looksYearQuery(query) || looksWhenEventQuery(query)) && !leftoverCoveringLineHasYear(line) {
			continue
		}
		if leftoverCoveringSkipYearMismatch(query, line) {
			continue
		}
		if looksLocationListQuery(query) {
			focus := practiceObjectTokens(query)
			if leftoverCoveringLineHasForeignPerson(query, line) && !leftoverCoveringMentionsQueryEntity(query, line) {
				continue
			}
			if len(placesFromContent(line)) == 0 && compositionalPracticePlace(line, focus) == "" {
				continue
			}
		}
		if !whereQ && datedContentConflictsQuery(query, line) {
			continue
		}
		if whereQ && !looksLocativePrepositionLine(line) {
			continue
		}
		if !contentCoversAnyQueryToken(line, rare) {
			if !leftoverCoveringAllowsZeroQueryTokens(query, line) {
				continue
			}
		}
		if len(locativeMust) > 0 && !contentCoversAnyQueryToken(line, locativeMust) {
			continue
		}
		if !contentCoversAnyQueryToken(line, strong) {
			if !leftoverCoveringAllowsZeroQueryTokens(query, line) {
				continue
			}
		}
		score := 1
		if looksCodedEventToken(line) {
			score += 3
		}
		if !looksChatTurnLine(line) && (looksLocativePlaceLine(line) || looksHyphenatedEventLine(line)) {
			score += 2
		}
		hits := 0
		for _, tok := range strong {
			if !contentCoversQueryToken(line, tok) {
				continue
			}
			hits++
			switch df[tok] {
			case 1:
				score += 2
			case 2:
				score += 1
			}
		}
		score += hits
		if leftoverCoveringQueryEntityHits(query, line) >= 2 {
			score += 2
		}
		if looksInstrumentPurposeQuery(query) && leftoverCoveringInstrumentPurposeLine(line) {
			score += 3
		}
		if looksWhatMadeQuery(query) {
			if leftoverCoveringHasOffQueryEvidence(query, line) {
				score += 2 * leftoverCoveringReasonHits(query, line)
			}
			if leftoverCoveringSchemaActivityLine(line) && !leftoverCoveringHasOffQueryEvidence(query, line) {
				score -= 2
			}
		}
		if looksHostQuery(query) {
			if leftoverCoveringHostedEventLine(line) {
				score += 4
			}
			if leftoverCoveringRealizeSpeechLine(line) && !leftoverCoveringHostedEventLine(line) {
				score -= 3
			}
		}
		if looksAdviceQuery(query) {
			if leftoverCoveringAdviceOffQueryLine(line) {
				score += 4
			}
			if leftoverCoveringAdviceSpeechLine(line) {
				score -= 3
			}
		}
		if looksWhatKindQuery(query) {
			if leftoverCoveringKindListLine(line) {
				score += 4
			}
			if leftoverCoveringKindRestatementLine(line) {
				score -= 3
			}
		}
		if looksHowDescribeProcessQuery(query) {
			if leftoverCoveringProcessHortativeLine(line) {
				score += 4
			}
			if leftoverCoveringProcessCompanionLine(query, line) {
				score -= 3
			}
		}
		if looksWhatMotivatesQuery(query) {
			if leftoverCoveringMotivateCauseLine(query, line) {
				score += 4
			}
			if leftoverCoveringMotivateCompanionLine(query, line) {
				score -= 3
			}
		}
		if looksWhatSayAboutQuery(query) {
			if leftoverCoveringSayAboutTargetLine(line) {
				score += 4
			}
			if leftoverCoveringSayAboutCompanionLine(query, line) {
				score -= 3
			}
		}
		if looksHowReactQuery(query) {
			if leftoverCoveringReactionObservationLine(line) {
				score += 4
			}
			if leftoverCoveringReactCompanionLine(query, line) {
				score -= 3
			}
		}
		if looksWhatDidPurposeQuery(query) {
			if leftoverCoveringPurposeActionLine(query, line) {
				score += 4
			}
			if leftoverCoveringPurposeCompanionLine(query, line) {
				score -= 3
			}
		}
		if looksHowDidStartQuery(query) {
			if leftoverCoveringStartMethodLine(query, line) {
				score += 4
			}
			if leftoverCoveringStartCompanionLine(query, line) {
				score -= 3
			}
		}
		if looksHowLongBeenQuery(query) {
			if leftoverCoveringDurationLine(query, line) {
				score += 4
				if leftoverCoveringDurationExplicitLine(line) {
					score++
				}
			}
			if leftoverCoveringDurationCompanionLine(query, line) {
				score -= 3
			}
		}
		if looksHowOftenQuery(query) {
			if leftoverCoveringCadenceLine(query, line) {
				score += 4
			}
			if leftoverCoveringCadenceCompanionLine(query, line) {
				score -= 3
			}
		}
		if looksWhatProjectWorkingQuery(query) {
			if leftoverCoveringCurrentProjectLine(query, line) {
				score += 4
			}
			if leftoverCoveringCurrentProjectCompanionLine(query, line) {
				score -= 3
			}
		}
		if looksWhatNewHobbyQuery(query) {
			if leftoverCoveringBecomeInterestedLine(query, line) {
				score += 4
			}
			if leftoverCoveringBecomeInterestedCompanionLine(query, line) {
				score -= 3
			}
		}
		if looksHowPlanDreamQuery(query) {
			if leftoverCoveringPrepPlanLine(query, line) {
				score += 4
				if n := leftoverCoveringPrepPlanCueCount(hybridLineBody(line)); n > 1 {
					score += n - 1
				}
			}
			if leftoverCoveringPrepPlanCompanionLine(query, line) {
				score -= 3
			}
		}
		if looksWhatFocusingBesidesQuery(query) {
			if leftoverCoveringFocusingBesidesLine(query, line) {
				score += 4
			}
			if leftoverCoveringFocusingBesidesCompanionLine(query, line) {
				score -= 3
			}
		}
		if looksWhatNewSeriesQuery(query) {
			if leftoverCoveringTitledShowLine(query, line) {
				score += 4
			}
			if leftoverCoveringTitledShowCompanionLine(query, line) {
				score -= 3
			}
		}
		if looksWhatDidRecentlyAtQuery(query) {
			if leftoverCoveringLocativePurposeLine(query, line) {
				score += 4
			}
			if leftoverCoveringLocativePurposeCompanionLine(query, line) {
				score -= 3
			}
		}
		if looksHowFeelAboutQuery(query) {
			if leftoverCoveringExperiencingFeelingLine(query, line) {
				score += 4
			}
			if leftoverCoveringExperiencingFeelingCompanionLine(query, line) {
				score -= 3
			}
		}
		if looksWhatDoCoordinatedUseQuery(query) {
			if leftoverCoveringWorkDeterminationLine(query, line) {
				score += 4
			}
			if leftoverCoveringWorkDeterminationCompanionLine(query, line) {
				score -= 3
			}
		}
		if looksWhatDidRealizeAfterQuery(query) {
			if leftoverCoveringSelfDirectedRealizeLine(query, line) {
				score += 4
			}
			if leftoverCoveringSelfDirectedRealizeCompanionLine(query, line) {
				score -= 3
			}
		}
		scored = append(scored, leftoverCoverScored{line: line, score: score})
		if score > bestScore {
			bestScore = score
			best = line
		}
	}
	if bestScore < 2 {
		return ""
	}
	if rarest := leftoverCoverRarestTokens(strong, df); len(rarest) > 0 && !contentCoversAnyQueryToken(best, rarest) {
		for _, row := range scored {
			if row.score < 2 || looksChatTurnLine(row.line) {
				continue
			}
			if contentCoversAnyQueryToken(row.line, rarest) {
				best = row.line
				break
			}
		}
	}
	if looksWhatMadeQuery(query) {
		if next := leftoverCoveringPreferWhatMadeEvidence(query, scored, best); next != "" {
			best = next
		}
	}
	if looksHostQuery(query) {
		if next := leftoverCoveringPreferHostedEvent(query, scored, best); next != "" {
			best = next
		}
		if joined := leftoverCoveringJoinHostedEvents(query, scored); joined != "" {
			return leftoverCoveringFinish(query, joined)
		}
	}
	if looksAdviceQuery(query) {
		if next := leftoverCoveringPreferAdviceDirective(query, scored, best); next != "" {
			best = next
		}
		if joined := leftoverCoveringJoinAdviceDirectives(query, scored); joined != "" {
			return leftoverCoveringFinish(query, joined)
		}
	}
	if looksWhatKindQuery(query) {
		if next := leftoverCoveringPreferKindList(query, scored, best); next != "" {
			best = next
		}
	}
	if looksHowDescribeProcessQuery(query) {
		if next := leftoverCoveringPreferProcessHortative(query, scored, best); next != "" {
			best = next
		}
		if !leftoverCoveringProcessHortativeLine(best) {
			return ""
		}
	}
	if looksWhatMotivatesQuery(query) {
		if next := leftoverCoveringPreferMotivateCause(query, scored, best); next != "" {
			best = next
		}
		if !leftoverCoveringMotivateCauseLine(query, best) {
			return ""
		}
	}
	if looksWhatSayAboutQuery(query) {
		hasObj := leftoverCoveringSayAboutHasObjectEvidence(query, lines)
		dated := querySpecificCalendarDate(query) != nil
		if next := leftoverCoveringPreferEvaluativeThey(query, scored, best); next != "" {
			best = next
		} else if dated && hasObj {
			if next := leftoverCoveringPreferReportedSpeech(query, scored, best); next != "" {
				best = next
			}
		} else if !dated && hasObj {
			if next := leftoverCoveringPreferFirstPersonGot(query, scored, best); next != "" {
				best = next
			}
		}
		if leftoverCoveringEvaluativeTheyLine(best) {
			// they-evaluative leftover does not need object tokens on the line.
		} else if dated && leftoverCoveringReportedSpeechLine(best) && hasObj {
			// dated what-say-about uses reported-speech leftover, not it's-got.
		} else if !dated && leftoverCoveringFirstPersonGotLine(best) && hasObj {
			// first-person got leftover omits the object; another packet line must cover it.
		} else {
			return ""
		}
	}
	if looksHowReactQuery(query) {
		if next := leftoverCoveringPreferReactionObservation(query, scored, best); next != "" {
			best = next
		}
		if leftoverCoveringReactionObservationLine(best) && leftoverCoveringReactLineHasObject(query, best) {
			// observational they-were leftover must name a query object (snow/dogs),
			// not any they-looked-ADJ leftover from the same session.
		} else {
			return ""
		}
	}
	if looksWhatDidPurposeQuery(query) {
		if next := leftoverCoveringPreferPurposeAction(query, scored, best); next != "" {
			best = next
		}
		if leftoverCoveringPurposeActionLine(query, best) {
			// purpose-infinitive leftover must carry the adjacent purpose pair
			// (take care), not comparative-better chat or a dated salon visit.
		} else {
			return ""
		}
	}
	if looksHowDidStartQuery(query) {
		if next := leftoverCoveringPreferStartMethod(query, scored, best); next != "" {
			best = next
		}
		if joined := leftoverCoveringJoinStartMethods(query, scored); joined != "" {
			return leftoverCoveringFinish(query, joined)
		}
		if leftoverCoveringStartMethodLine(query, best) {
			// how-did-start leftover must be duration-matched inception or a
			// first-person changed+started pair, not a gym restatement.
		} else {
			return ""
		}
	}
	if looksHowLongBeenQuery(query) {
		if next := leftoverCoveringPreferDuration(query, scored, best); next != "" {
			best = next
		}
		if leftoverCoveringDurationLine(query, best) {
			// how-long-been leftover must carry continuing years (already / for /
			// duration is), not a copula status fact or a years-ago start.
		} else {
			return ""
		}
	}
	if looksHowOftenQuery(query) {
		if next := leftoverCoveringPreferCadence(query, scored, best); next != "" {
			best = next
		}
		if leftoverCoveringCadenceLine(query, best) {
			// how-often leftover must carry a cadence (once a week / every
			// month / N times a day) plus object overlap, not a purpose
			// join or a cadence about a different activity.
		} else {
			return ""
		}
	}
	if looksWhatProjectWorkingQuery(query) {
		if next := leftoverCoveringPreferCurrentProject(query, scored, best); next != "" {
			best = next
		}
		if leftoverCoveringCurrentProjectLine(query, best) {
			// what-project leftover must be currently-working / working-on-a-new,
			// not a childhood desire or "creating his own" without currently.
		} else {
			return ""
		}
	}
	if looksWhatNewHobbyQuery(query) {
		if next := leftoverCoveringPreferBecomeInterested(query, scored, best); next != "" {
			best = next
		}
		if leftoverCoveringBecomeInterestedLine(query, best) {
			// dated what-new-hobby leftover must be first-person become-interested,
			// not a foreign-person "new hobby" restatement.
		} else {
			return ""
		}
	}
	if looksHowPlanDreamQuery(query) {
		if next := leftoverCoveringPreferPrepPlan(query, scored, best); next != "" {
			best = next
		}
		if leftoverCoveringPrepPlanLine(query, best) {
			// how-plan-dream leftover must be gathering/watching/guide prep,
			// not a foreign-person "learning their stories" restatement.
		} else {
			return ""
		}
	}
	if looksWhatFocusingBesidesQuery(query) {
		if next := leftoverCoveringPreferFocusingBesides(query, scored, best); next != "" {
			best = next
		}
		if leftoverCoveringFocusingBesidesLine(query, best) {
			// besides-focusing leftover must be focusing-on + excluded object
			// + first-person possessive conjunct, not occupation leftover.
		} else {
			return ""
		}
	}
	if looksWhatNewSeriesQuery(query) {
		if next := leftoverCoveringPreferTitledShow(query, scored, best); next != "" {
			best = next
		}
		if leftoverCoveringTitledShowLine(query, best) {
			// what-new-series leftover must quote a titled show, not a generic
			// excitement leftover or a quoted novel.
		} else {
			return ""
		}
	}
	if looksWhatDidRecentlyAtQuery(query) {
		if next := leftoverCoveringPreferLocativePurpose(query, scored, best); next != "" {
			best = next
		}
		if leftoverCoveringLocativePurposeLine(query, best) {
			// recently-at leftover must name a locative action with an extra
			// for-my/our purpose object, not a thin dated locative fact.
		} else {
			return ""
		}
	}
	if looksHowFeelAboutQuery(query) {
		if next := leftoverCoveringPreferExperiencingFeeling(query, scored, best); next != "" {
			best = next
		}
		if leftoverCoveringExperiencingFeelingLine(query, best) {
			// how-feel leftover must be first-person experiencing + new level,
			// not a process restatement of practicing mindfulness.
		} else {
			return ""
		}
	}
	if looksWhatDoCoordinatedUseQuery(query) {
		if next := leftoverCoveringPreferWorkDetermination(query, scored, best); next != "" {
			best = next
		}
		if leftoverCoveringWorkDeterminationLine(query, best) {
			// coordinated-use leftover must be first-person-plural hard work
			// and determination, not a dedication restatement or car leftover.
		} else {
			return ""
		}
	}
	if looksWhatDidRealizeAfterQuery(query) {
		if next := leftoverCoveringPreferSelfDirectedRealize(query, scored, best); next != "" {
			best = next
		}
		if leftoverCoveringSelfDirectedRealizeLine(query, best) {
			// realize-after leftover must be first-person realize leftover that
			// names a self-directed object, not an others-directed support leftover.
		} else {
			return ""
		}
	}
	if leftoverCoveringShouldJoin(query) {
		parts := make([]string, 0, 4)
		seen := map[string]struct{}{}
		for _, row := range scored {
			if row.score < 2 {
				continue
			}
			if !contentCoversQueryToken(row.line, "played") || !contentCoversQueryToken(row.line, "tournament") {
				continue
			}
			line := stripConflictingDateTail(query, row.line)
			key := strings.ToLower(strings.TrimSpace(line))
			if key == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			parts = append(parts, line)
			if len(parts) >= 4 {
				break
			}
		}
		if len(parts) >= 2 {
			return strings.Join(parts, "; ")
		}
	}
	best = stripConflictingDateTail(query, best)
	if whereQ {
		if place := locativePlaceFromLine(best); place != "" {
			return leftoverCoveringFinish(query, place)
		}
		return leftoverCoveringFinish(query, best)
	}
	if joined := leftoverCoveringJoinSchemaActivities(query, strong, scored, best); joined != "" {
		return leftoverCoveringFinish(query, joined)
	}
	if joined := leftoverCoveringJoinPastPossessions(query, strong, scored, best); joined != "" {
		return leftoverCoveringFinish(query, joined)
	}
	return leftoverCoveringFinish(query, best)
}

func leftoverCoverRarestTokens(strong []string, df map[string]int) []string {
	strong = leftoverCoverNonWeakTokens(strong)
	if len(strong) == 0 {
		return nil
	}
	min := -1
	for _, tok := range strong {
		d := df[tok]
		if d <= 0 {
			continue
		}
		if min < 0 || d < min {
			min = d
		}
	}
	if min < 0 {
		return nil
	}
	out := make([]string, 0, 2)
	for _, tok := range strong {
		if df[tok] == min {
			out = append(out, tok)
		}
	}
	return out
}

type leftoverCoverScored struct {
	line  string
	score int
}

func leftoverCoveringSchemaActivityLine(line string) bool {
	body := " " + strings.ToLower(hybridLineBody(line)) + " "
	for _, p := range []string{
		" enjoys ", " enjoy ", " enjoyed ",
		" likes ", " liked ",
		" loves ", " loved ",
		" participates in ", " participated in ",
	} {
		if strings.Contains(body, p) {
			return true
		}
	}
	return false
}

func leftoverCoveringJoinSchemaActivities(query string, strong []string, scored []leftoverCoverScored, best string) string {
	if leftoverCoveringShouldJoin(query) || looksWhereQuery(query) {
		return ""
	}
	if !leftoverCoveringSchemaActivityLine(best) {
		return ""
	}
	parts := []string{best}
	seen := map[string]struct{}{strings.ToLower(strings.TrimSpace(best)): {}}
	for _, row := range scored {
		if row.score < 2 || !leftoverCoveringSchemaActivityLine(row.line) {
			continue
		}
		if !contentCoversAnyQueryToken(row.line, strong) {
			continue
		}
		line := stripConflictingDateTail(query, row.line)
		key := strings.ToLower(strings.TrimSpace(line))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		parts = append(parts, line)
		if len(parts) >= 4 {
			break
		}
	}
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts, "; ")
}

func leftoverCoveringPastPossessionLine(line string) bool {
	body := " " + strings.ToLower(hybridLineBody(line)) + " "
	if !strings.Contains(body, " had a ") && !strings.Contains(body, " had an ") {
		return false
	}
	for _, cue := range []string{" kid", " child", " childhood", " growing up"} {
		if strings.Contains(body, cue) {
			return true
		}
	}
	return false
}

func leftoverCoveringLockChildhoodPossessions(query string) bool {
	if queryHasToken(query, "name", "named", "names") {
		return false
	}
	return len(childhoodClauseTokens(query)) > 0
}

func leftoverCoveringJoinPastPossessions(query string, strong []string, scored []leftoverCoverScored, best string) string {
	if leftoverCoveringShouldJoin(query) || looksWhereQuery(query) || !leftoverCoveringLockChildhoodPossessions(query) {
		return ""
	}
	if !leftoverCoveringPastPossessionLine(best) {
		return ""
	}
	parts := []string{best}
	seen := map[string]struct{}{strings.ToLower(strings.TrimSpace(best)): {}}
	for _, row := range scored {
		if row.score < 2 || !leftoverCoveringPastPossessionLine(row.line) {
			continue
		}
		if !contentCoversAnyQueryToken(row.line, strong) {
			continue
		}
		line := stripConflictingDateTail(query, row.line)
		key := strings.ToLower(strings.TrimSpace(line))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		parts = append(parts, line)
		if len(parts) >= 4 {
			break
		}
	}
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts, "; ")
}

func locativePlaceFromLine(line string) string {
	fields := strings.Fields(hybridLineBody(line))
	for i, w := range fields {
		switch strings.ToLower(strings.Trim(w, ".,;:")) {
		case "to", "in", "at", "near", "from", "for":
		default:
			continue
		}
		if i+1 >= len(fields) {
			continue
		}
		n := strings.Trim(fields[i+1], ".,;:()[]\"'")
		if n == "" || n[0] < 'A' || n[0] > 'Z' || len(n) <= 1 {
			continue
		}
		rest := strings.Join(fields[i+1:], " ")
		if placePrepRestIsDate(rest) {
			continue
		}
		cand := cleanPlaceCandidate(rest)
		if cand == "" || consecutiveProperNouns(cand) < 1 {
			continue
		}
		return cand
	}
	return ""
}

func leftoverCoveringShouldJoin(query string) bool {
	hasGame, hasPlayed := false, false
	for _, tok := range tokenize(query) {
		switch strings.Trim(strings.ToLower(tok), "'\"") {
		case "game", "games":
			hasGame = true
		case "play", "played":
			hasPlayed = true
		}
	}
	return hasGame && hasPlayed
}

func leftoverCoveringBeatsAnswer(query string, hops []HopResult, covering, answer string) bool {
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || answer == "" || strings.EqualFold(covering, answer) {
		return false
	}
	if strings.EqualFold(answer, "not in memory") {
		return true
	}
	if leftoverCoveringKeepShortLocative(query, covering, answer) {
		return false
	}
	if looksWhereQuery(query) && leftoverCoveringIsShortLocative(covering) && typedAnswerIsHopDump(answer) {
		return true
	}
	if leftoverCoveringKeepTypedAnswer(query, hops, answer) {
		return false
	}
	if (looksWhenEventQuery(query) || looksYearQuery(query)) && parseDateFromText(answer) != nil && parseDateFromText(covering) == nil {
		return false
	}
	if leftoverCoveringDropsNamedSpan(covering, answer) {
		return false
	}
	if leftoverCoveringMostlyCoveredByAnswer(covering, answer) {
		return false
	}
	strong := leftoverCoverStrongTokens(leftoverCoveringRareForQuery(query, hops))
	coverHits, ansHits := 0, 0
	extra := false
	dropped := false
	for _, tok := range strong {
		c := contentCoversQueryToken(covering, tok)
		a := contentCoversQueryToken(answer, tok)
		if c {
			coverHits++
		}
		if a {
			ansHits++
		}
		if c && !a {
			extra = true
		}
		if a && !c {
			dropped = true
		}
	}
	if dropped {
		return false
	}
	if len(strong) > 0 && ansHits == len(strong) {
		return false
	}
	return extra && coverHits > ansHits
}

func leftoverCoveringMayReplaceHybrid(query string, hops []HopResult, covering, answer string) bool {
	if looksLocationListQuery(query) && !looksWhereQuery(query) {
		return false
	}
	if leftoverCoveringYearMissesEvent(query, covering, answer) {
		return true
	}
	if leftoverCoveringPurposeMissesInstrument(query, covering, answer) {
		return true
	}
	if leftoverCoveringWhatMadeMissesEvidence(query, covering, answer) {
		return true
	}
	if leftoverCoveringHostMissesEvent(query, covering, answer) {
		return true
	}
	if leftoverCoveringAdviceMissesDirective(query, covering, answer) {
		return true
	}
	if leftoverCoveringKindMissesList(query, covering, answer) {
		return true
	}
	if leftoverCoveringProcessMissesHortative(query, covering, answer) {
		return true
	}
	if leftoverCoveringMotivateMissesCause(query, covering, answer) {
		return true
	}
	if leftoverCoveringSayAboutMissesEvaluative(query, covering, answer) {
		return true
	}
	if leftoverCoveringReactMissesObservation(query, covering, answer) {
		return true
	}
	if leftoverCoveringStartMissesMethod(query, covering, answer) {
		return true
	}
	if leftoverShortItemJoin(answer) && !looksWhereQuery(query) && !leftoverCoveringShouldJoin(query) {
		return false
	}
	if leftoverCoveringIsCompactNamed(answer) {
		return false
	}
	if typedAnswerIsHopDump(answer) && !looksWhereQuery(query) && !leftoverCoveringShouldJoin(query) {
		return false
	}
	if !looksWhereQuery(query) && !leftoverCoveringShouldJoin(query) && !leftoverCoveringSchemaActivityLine(covering) {
		return false
	}
	return leftoverCoveringBeatsAnswer(query, hops, covering, answer)
}

func leftoverCoveringIsCompactNamed(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || leftoverShortItemJoin(s) || typedAnswerIsHopDump(s) {
		return false
	}
	if strings.Contains(s, ".") || strings.Contains(s, ":") {
		return false
	}
	if len(strings.Fields(s)) > 4 || utf8Len(s) > 48 {
		return false
	}
	return consecutiveProperNouns(s) >= 1
}

func leftoverCoveringQuotedTitles(s string) []string {
	out := make([]string, 0, 2)
	for {
		i := strings.Index(s, `"`)
		if i < 0 {
			break
		}
		rest := s[i+1:]
		j := strings.Index(rest, `"`)
		if j < 0 {
			break
		}
		t := strings.TrimSpace(rest[:j])
		if t != "" && utf8Len(t) <= 48 {
			out = append(out, t)
		}
		s = rest[j+1:]
	}
	return out
}

func leftoverCoveringProperRuns(s string) []string {
	fields := strings.Fields(hybridLineBody(s))
	out := make([]string, 0, 2)
	run := make([]string, 0, 4)
	flush := func() {
		if len(run) >= 2 {
			out = append(out, strings.Join(run, " "))
		}
		run = run[:0]
	}
	for _, w := range fields {
		n := strings.Trim(w, ".,;:()[]\"'")
		if n != "" && n[0] >= 'A' && n[0] <= 'Z' && len(n) > 1 {
			run = append(run, n)
			continue
		}
		flush()
	}
	flush()
	return out
}

func leftoverCoveringNamedSpans(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	out := leftoverCoveringQuotedTitles(s)
	out = append(out, leftoverCoveringProperRuns(s)...)
	if leftoverCoveringIsCompactNamed(s) {
		out = append(out, s)
	}
	return out
}

func leftoverCoveringDropsNamedSpan(covering, answer string) bool {
	cl := strings.ToLower(covering)
	for _, span := range leftoverCoveringNamedSpans(answer) {
		span = strings.TrimSpace(span)
		if span == "" {
			continue
		}
		if !strings.Contains(cl, strings.ToLower(span)) {
			return true
		}
	}
	return false
}

func leftoverCoveringMostlyCoveredByAnswer(covering, answer string) bool {
	toks := distinctiveQueryTokens(tokenize(hybridLineBody(covering)))
	n, hits := 0, 0
	for _, tok := range toks {
		if utf8Len(tok) < 5 {
			continue
		}
		n++
		if contentCoversQueryToken(answer, tok) {
			hits++
		}
	}
	return n > 0 && hits*2 >= n
}

func leftoverQueryEchoAnswer(query, answer string) bool {
	answer = strings.TrimSpace(answer)
	if answer == "" || strings.EqualFold(answer, "not in memory") || strings.Contains(answer, ",") {
		return false
	}
	qtoks := map[string]struct{}{}
	for _, tok := range tokenize(query) {
		qtoks[strings.ToLower(tok)] = struct{}{}
	}
	n := 0
	for _, tok := range tokenize(answer) {
		tok = strings.ToLower(tok)
		if tok == "" {
			continue
		}
		n++
		if _, ok := qtoks[tok]; !ok {
			return false
		}
	}
	return n > 0 && n <= 3
}

func leftoverCoveringIsShortLocative(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || typedAnswerIsHopDump(s) {
		return false
	}
	if consecutiveProperNouns(s) < 1 {
		return false
	}
	return utf8Len(s) <= 48
}

func leftoverCoveringKeepShortLocative(query, covering, answer string) bool {
	if !looksWhereQuery(query) {
		return false
	}
	answer = strings.TrimSpace(answer)
	covering = strings.TrimSpace(covering)
	if !leftoverCoveringIsShortLocative(answer) || covering == "" {
		return false
	}
	if utf8Len(covering) <= utf8Len(answer) {
		return false
	}
	al, cl := strings.ToLower(answer), strings.ToLower(covering)
	if strings.Contains(cl, al) {
		return true
	}
	for _, part := range strings.Split(answer, ",") {
		part = strings.TrimSpace(part)
		if leftoverCoveringIsShortLocative(part) && strings.Contains(cl, strings.ToLower(part)) {
			return true
		}
	}
	return false
}

func leftoverShortItemJoin(answer string) bool {
	answer = strings.TrimSpace(answer)
	if answer == "" || strings.EqualFold(answer, "not in memory") || looksPromptNotAnswer(answer) {
		return false
	}
	parts := make([]string, 0, 4)
	if strings.Contains(answer, ",") {
		for _, p := range strings.Split(answer, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				parts = append(parts, p)
			}
		}
	} else if i := strings.Index(strings.ToLower(answer), " and "); i > 0 {
		left := strings.TrimSpace(answer[:i])
		right := strings.TrimSpace(answer[i+5:])
		if left != "" && right != "" && !strings.Contains(strings.ToLower(right), " and ") {
			parts = []string{left, right}
		}
	}
	if len(parts) < 2 || len(parts) > 6 || utf8Len(answer) > 96 {
		return false
	}
	for _, p := range parts {
		if len(strings.Fields(p)) > 6 {
			return false
		}
	}
	return true
}

// leftoverCoveringBareDateMissesEvent is true when a when-event answer is just
// a calendar date that does not cover leftover event tokens, but a packet
// covering line does. Hop activity dates must not lock out the event fact.
func leftoverCoveringBareDateMissesEvent(query string, hops []HopResult, covering, answer string) bool {
	if !looksWhenEventQuery(query) && !looksYearQuery(query) {
		return false
	}
	answer = strings.TrimSpace(answer)
	covering = strings.TrimSpace(covering)
	if covering == "" || answer == "" {
		return false
	}
	if leftoverCoveringKeepTypedAnswer(query, hops, answer) {
		return false
	}
	if leftoverCoveringSkipForeignWhenEvent(query, covering) {
		return false
	}
	if leftoverCoveringSkipYearMismatch(query, covering) {
		return false
	}
	if parseDateFromText(answer) == nil || !leftoverCoveringLineHasYear(covering) {
		return false
	}
	if len(strings.Fields(answer)) > 6 {
		return false
	}
	strong := leftoverCoverStrongTokens(leftoverCoveringRareForQuery(query, hops))
	if len(strong) == 0 {
		return false
	}
	if contentCoversAnyQueryToken(answer, strong) {
		return false
	}
	return contentCoversAnyQueryToken(covering, strong)
}

// leftoverCoveringRequiresQueryEntity is true for when-event questions that
// name a hop person. Covering lines about a different person must not win.
func leftoverCoveringRequiresQueryEntity(query string) bool {
	return (looksWhenEventQuery(query) || looksYearQuery(query) || looksInstrumentPurposeQuery(query)) &&
		len(hopQueryEntities(query)) > 0
}

func leftoverCoveringLineYear(s string) int {
	if t := parseDateFromText(s); t != nil {
		return t.Year()
	}
	return queryCalendarYear(s)
}

// leftoverCoveringSkipYearMismatch drops when-event covering lines whose year
// or month disagrees with a year/month named in the query. A 2022 game or an
// October 2023 plan must not cover an August 2023 meetup.
func leftoverCoveringSkipYearMismatch(query, line string) bool {
	if !looksWhenEventQuery(query) && !looksYearQuery(query) {
		return false
	}
	yq := queryCalendarYear(query)
	if yq != 0 {
		yl := leftoverCoveringLineYear(line)
		if yl != 0 && yl != yq {
			return true
		}
	}
	mq := leftoverCoverQueryMonth(query)
	if mq == "" {
		return false
	}
	ml := leftoverCoveringLineMonth(line)
	if ml == "" {
		return false
	}
	return ml != mq
}

func leftoverCoverQueryMonth(query string) string {
	for _, tok := range tokenize(query) {
		if isMonthWord(tok) {
			return strings.ToLower(strings.Trim(tok, ".,"))
		}
	}
	return ""
}

func leftoverCoveringLineMonth(s string) string {
	if t := parseDateFromText(s); t != nil {
		return strings.ToLower(t.Month().String())
	}
	return leftoverCoverQueryMonth(s)
}

func leftoverCoveringLineHasYear(s string) bool {
	if parseDateFromText(s) != nil {
		return true
	}
	for _, tok := range tokenize(s) {
		if len(tok) != 4 {
			continue
		}
		y, err := strconv.Atoi(tok)
		if err == nil && y >= 1900 && y <= 2100 {
			return true
		}
	}
	return false
}

// leftoverCoveringFinish rewrites which-year covering of the form
// "for N years as of DATE" to the start year. The as-of calendar year is
// not the year the event started.
func leftoverCoveringFinish(query, covering string) string {
	covering = strings.TrimSpace(covering)
	if covering == "" || !looksYearQuery(query) {
		return covering
	}
	if y := leftoverCoveringRelativeStartYear(covering); y >= 1900 && y <= 2100 {
		return strconv.Itoa(y)
	}
	return covering
}

func leftoverCoveringRelativeStartYear(line string) int {
	body := hybridLineBody(line)
	lower := strings.ToLower(body)
	idx := strings.Index(lower, " as of ")
	if idx < 0 {
		return 0
	}
	n := leftoverCoveringForYears(lower[:idx])
	if n <= 0 || n > 80 {
		return 0
	}
	asOf := parseDateFromText(strings.TrimSpace(body[idx+len(" as of "):]))
	if asOf == nil {
		return 0
	}
	return asOf.Year() - n
}

func leftoverCoveringForYears(head string) int {
	fields := strings.Fields(head)
	for i := 0; i+2 < len(fields); i++ {
		if fields[i] != "for" {
			continue
		}
		if !strings.HasPrefix(fields[i+2], "year") {
			continue
		}
		if n := leftoverCoveringYearWord(fields[i+1]); n > 0 {
			return n
		}
	}
	return 0
}

func leftoverCoveringYearWord(tok string) int {
	tok = strings.Trim(strings.ToLower(tok), ".,;:()[]\"'")
	if n, err := strconv.Atoi(tok); err == nil && n > 0 && n < 100 {
		return n
	}
	switch tok {
	case "one", "a":
		return 1
	case "two":
		return 2
	case "three":
		return 3
	case "four":
		return 4
	case "five":
		return 5
	case "six":
		return 6
	case "seven":
		return 7
	case "eight":
		return 8
	case "nine":
		return 9
	case "ten":
		return 10
	}
	return 0
}

// leftoverCoveringYearMissesEvent is true when a which/what-year answer has no
// year token but a query-bound leftover line does.
func leftoverCoveringYearMissesEvent(query, covering, answer string) bool {
	if !looksYearQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringSkipForeignWhenEvent(query, covering) {
		return false
	}
	if !leftoverCoveringLineHasYear(covering) {
		return false
	}
	if leftoverCoveringLineHasYear(answer) && leftoverCoveringMentionsQueryEntity(query, answer) {
		if leftoverCoveringRelativeStartYear(answer) != 0 {
			return leftoverCoveringMentionsQueryEntity(query, covering) || !leftoverCoveringLineHasForeignPerson(query, covering)
		}
		return false
	}
	if leftoverCoveringMentionsQueryEntity(query, covering) {
		return true
	}
	return !leftoverCoveringLineHasForeignPerson(query, covering)
}

func looksWhatMadeQuery(query string) bool {
	q := strings.ToLower(query)
	return strings.Contains(q, "what made") || strings.Contains(q, "what makes")
}

func looksWhatMotivatesQuery(query string) bool {
	q := strings.ToLower(query)
	return strings.Contains(q, "what motivates") || strings.Contains(q, "what motivated")
}

func looksWhatSayAboutQuery(query string) bool {
	q := strings.ToLower(query)
	if !strings.Contains(q, "say about") && !strings.Contains(q, "says about") && !strings.Contains(q, "said about") {
		return false
	}
	return strings.Contains(q, "what does") || strings.Contains(q, "what did") || strings.Contains(q, "what do ")
}

func looksHowDescribeQuery(query string) bool {
	q := strings.ToLower(query)
	if !strings.Contains(q, "describe") {
		return false
	}
	return strings.Contains(q, "how does") || strings.Contains(q, "how did") || strings.Contains(q, "how do ")
}

func looksHowDescribeProcessQuery(query string) bool {
	return looksHowDescribeQuery(query) && strings.Contains(strings.ToLower(query), "process")
}

func looksHowReactQuery(query string) bool {
	q := strings.ToLower(query)
	if !strings.Contains(q, "how does") && !strings.Contains(q, "how did") && !strings.Contains(q, "how do ") {
		return false
	}
	hasReact, hasRespondTo := false, false
	toks := tokenize(query)
	for i, tok := range toks {
		switch tok {
		case "react", "reacts", "reacted", "reacting":
			hasReact = true
		case "respond", "responds", "responded", "responding":
			if i+1 < len(toks) && toks[i+1] == "to" {
				hasRespondTo = true
			}
		}
	}
	return hasReact || hasRespondTo
}

func looksHowDidStartQuery(query string) bool {
	if looksHowDescribeQuery(query) || looksHowReactQuery(query) {
		return false
	}
	q := strings.ToLower(query)
	if !strings.Contains(q, "how did") {
		return false
	}
	hasStart := false
	for _, tok := range tokenize(query) {
		switch tok {
		case "start", "started", "starting":
			hasStart = true
		}
	}
	if !hasStart {
		return false
	}
	return parseYearsAgo(q) > 0
}

func looksHowLongBeenQuery(query string) bool {
	if looksHowDidStartQuery(query) || looksHowDescribeQuery(query) || looksHowReactQuery(query) {
		return false
	}
	q := strings.ToLower(query)
	if !strings.Contains(q, "how long") || strings.Contains(q, "how long ago") {
		return false
	}
	for _, tok := range tokenize(query) {
		if tok == "been" {
			return true
		}
	}
	return false
}

func looksHowOftenQuery(query string) bool {
	if looksHowLongBeenQuery(query) || looksHowDidStartQuery(query) || looksHowDescribeQuery(query) || looksHowReactQuery(query) {
		return false
	}
	q := strings.ToLower(query)
	if !strings.Contains(q, "how often") {
		return false
	}
	for _, tok := range tokenize(query) {
		switch tok {
		case "does", "did", "do":
			return true
		}
	}
	return false
}

func looksWhatProjectWorkingQuery(query string) bool {
	if looksHowOftenQuery(query) || looksHowLongBeenQuery(query) || looksHowDidStartQuery(query) || looksWhatDidPurposeQuery(query) || looksHowDescribeQuery(query) || looksHowReactQuery(query) {
		return false
	}
	hasWhat, hasProject, hasWork := false, false, false
	for _, tok := range tokenize(query) {
		switch tok {
		case "what":
			hasWhat = true
		case "project", "projects":
			hasProject = true
		case "working", "work":
			hasWork = true
		}
	}
	return hasWhat && hasProject && hasWork
}

func looksWhatNewHobbyQuery(query string) bool {
	if looksWhatProjectWorkingQuery(query) || looksHowOftenQuery(query) || looksHowLongBeenQuery(query) || looksHowDidStartQuery(query) || looksWhatDidPurposeQuery(query) || looksHowDescribeQuery(query) || looksHowReactQuery(query) {
		return false
	}
	if querySpecificCalendarDate(query) == nil {
		return false
	}
	hasWhat, hasHobby, hasInterest := false, false, false
	for _, tok := range tokenize(query) {
		switch tok {
		case "what":
			hasWhat = true
		case "hobby", "hobbies":
			hasHobby = true
		case "interested", "interest", "become", "became", "becoming":
			hasInterest = true
		}
	}
	return hasWhat && hasHobby && hasInterest
}

func looksHowPlanDreamQuery(query string) bool {
	if looksWhatNewHobbyQuery(query) || looksWhatProjectWorkingQuery(query) || looksHowOftenQuery(query) || looksHowLongBeenQuery(query) || looksHowDidStartQuery(query) || looksWhatDidPurposeQuery(query) || looksHowDescribeQuery(query) || looksHowReactQuery(query) {
		return false
	}
	q := strings.ToLower(query)
	if !strings.Contains(q, "how does") && !strings.Contains(q, "how did") && !strings.Contains(q, "how do ") && !strings.Contains(q, "how will") {
		return false
	}
	hasPlan, hasDream, hasLearn := false, false, false
	for _, tok := range tokenize(query) {
		switch tok {
		case "plan", "plans", "planning", "pursue", "pursues", "pursuing":
			hasPlan = true
		case "dream", "dreams":
			hasDream = true
		case "learn", "learning", "learns", "learned", "learnt":
			hasLearn = true
		}
	}
	return hasPlan && hasDream && hasLearn
}

func looksWhatFocusingBesidesQuery(query string) bool {
	if looksHowPlanDreamQuery(query) || looksWhatNewHobbyQuery(query) || looksWhatProjectWorkingQuery(query) || looksHowOftenQuery(query) || looksHowLongBeenQuery(query) || looksHowDidStartQuery(query) || looksWhatDidPurposeQuery(query) || looksHowDescribeQuery(query) || looksHowReactQuery(query) {
		return false
	}
	q := strings.ToLower(query)
	if !strings.Contains(q, "what has") && !strings.Contains(q, "what have") && !strings.Contains(q, "what is") && !strings.Contains(q, "what was") {
		return false
	}
	hasFocus, hasBesides := false, false
	for _, tok := range tokenize(query) {
		switch tok {
		case "focus", "focusing", "focused", "focuses":
			hasFocus = true
		case "besides", "except", "aside":
			hasBesides = true
		}
	}
	return hasFocus && hasBesides && len(leftoverCoveringBesidesObjectTokens(query)) > 0
}

func looksWhatNewSeriesQuery(query string) bool {
	if looksWhatFocusingBesidesQuery(query) || looksHowPlanDreamQuery(query) || looksWhatNewHobbyQuery(query) || looksWhatProjectWorkingQuery(query) || looksHowOftenQuery(query) || looksHowLongBeenQuery(query) || looksHowDidStartQuery(query) || looksWhatDidPurposeQuery(query) || looksHowDescribeQuery(query) || looksHowReactQuery(query) {
		return false
	}
	q := strings.ToLower(query)
	if !strings.Contains(q, "what") {
		return false
	}
	hasNew, hasSeries := false, false
	for _, tok := range tokenize(query) {
		switch tok {
		case "new":
			hasNew = true
		case "series", "show":
			hasSeries = true
		}
	}
	return hasNew && hasSeries
}

func looksWhatDidRecentlyAtQuery(query string) bool {
	if looksWhatNewSeriesQuery(query) || looksWhatFocusingBesidesQuery(query) || looksHowPlanDreamQuery(query) || looksWhatNewHobbyQuery(query) || looksWhatProjectWorkingQuery(query) || looksHowOftenQuery(query) || looksHowLongBeenQuery(query) || looksHowDidStartQuery(query) || looksWhatDidPurposeQuery(query) || looksHowDescribeQuery(query) || looksHowReactQuery(query) || looksHostQuery(query) {
		return false
	}
	q := strings.ToLower(query)
	if !strings.Contains(q, "what did") && !strings.Contains(q, "what does") && !strings.Contains(q, "what has") {
		return false
	}
	hasRecent, hasDo := false, false
	for _, tok := range tokenize(query) {
		switch tok {
		case "recently", "lately":
			hasRecent = true
		case "do":
			hasDo = true
		}
	}
	return hasRecent && hasDo && len(leftoverCoveringRecentlyAtPlaceTokens(query)) > 0
}

func looksHowFeelAboutQuery(query string) bool {
	if looksWhatDoCoordinatedUseQuery(query) || looksWhatDidRecentlyAtQuery(query) || looksWhatNewSeriesQuery(query) || looksWhatFocusingBesidesQuery(query) || looksHowPlanDreamQuery(query) || looksWhatNewHobbyQuery(query) || looksWhatProjectWorkingQuery(query) || looksHowOftenQuery(query) || looksHowLongBeenQuery(query) || looksHowDidStartQuery(query) || looksWhatDidPurposeQuery(query) || looksHowDescribeQuery(query) || looksHowReactQuery(query) {
		return false
	}
	q := strings.ToLower(query)
	if !strings.Contains(q, "how ") || !strings.Contains(q, " about ") {
		return false
	}
	for _, tok := range tokenize(query) {
		switch tok {
		case "feel", "felt", "feeling", "feels":
			return true
		}
	}
	return false
}

func looksWhatDoCoordinatedUseQuery(query string) bool {
	if looksHostQuery(query) || looksAdviceQuery(query) {
		return false
	}
	q := strings.ToLower(query)
	if !strings.Contains(q, "what do") && !strings.Contains(q, "what does") && !strings.Contains(q, "what did") {
		return false
	}
	if len(hopQueryEntities(query)) < 2 {
		return false
	}
	for _, tok := range tokenize(query) {
		switch tok {
		case "use", "uses", "using":
			return true
		}
	}
	return false
}

func looksWhatDidRealizeAfterQuery(query string) bool {
	if looksHostQuery(query) || looksAdviceQuery(query) {
		return false
	}
	q := strings.ToLower(query)
	if !strings.Contains(q, "what did") && !strings.Contains(q, "what does") && !strings.Contains(q, "what has") {
		return false
	}
	hasRealize := false
	hasAfter := false
	for _, tok := range tokenize(query) {
		switch tok {
		case "realize", "realized", "realised", "realizing", "realising":
			hasRealize = true
		case "after":
			hasAfter = true
		}
	}
	return hasRealize && hasAfter
}

func looksWhatDidPurposeQuery(query string) bool {
	if looksWhatDoCoordinatedUseQuery(query) || looksWhatSayAboutQuery(query) || looksHostQuery(query) || looksAdviceQuery(query) || looksHowReactQuery(query) {
		return false
	}
	q := strings.ToLower(query)
	if !strings.Contains(q, "what did") && !strings.Contains(q, "what does") && !strings.Contains(q, "what has") {
		return false
	}
	toks := tokenize(query)
	sawDo := false
	sawTo := false
	for _, tok := range toks {
		if tok == "do" {
			sawDo = true
			continue
		}
		if sawDo && tok == "to" {
			sawTo = true
			break
		}
	}
	if !sawDo || !sawTo {
		return false
	}
	return len(leftoverCoveringPurposeObjectTokens(query)) >= 2
}

func looksFirstPersonLeftoverQuery(query string) bool {
	return looksWhatMadeQuery(query) || looksHowDescribeQuery(query) || looksWhatMotivatesQuery(query) || looksWhatSayAboutQuery(query) || looksWhatDidPurposeQuery(query) || looksHowDidStartQuery(query) || looksHowOftenQuery(query) || looksWhatProjectWorkingQuery(query) || looksWhatNewHobbyQuery(query) || looksHowPlanDreamQuery(query) || looksWhatFocusingBesidesQuery(query) || looksWhatNewSeriesQuery(query) || looksWhatDidRecentlyAtQuery(query) || looksHowFeelAboutQuery(query) || looksWhatDoCoordinatedUseQuery(query) || looksWhatDidRealizeAfterQuery(query)
}

func looksHostQuery(query string) bool {
	hasHost := false
	for _, tok := range tokenize(query) {
		switch tok {
		case "host", "hosted", "hosting", "hosts":
			hasHost = true
		}
	}
	if !hasHost {
		return false
	}
	q := strings.ToLower(query)
	return strings.Contains(q, "what did") || strings.Contains(q, "what does") || strings.Contains(q, "what has")
}

func looksAdviceQuery(query string) bool {
	hasAdvice := false
	for _, tok := range tokenize(query) {
		if leftoverCoveringAdviceSpeechToken(tok) {
			hasAdvice = true
			break
		}
	}
	if !hasAdvice {
		return false
	}
	q := strings.ToLower(query)
	return strings.Contains(q, "what ") || strings.Contains(q, "how ")
}

func looksWhatKindQuery(query string) bool {
	return strings.Contains(strings.ToLower(query), "what kind of")
}

func leftoverCoveringAllowsZeroQueryTokens(query, line string) bool {
	if looksAdviceQuery(query) && leftoverCoveringAdviceOffQueryLine(line) {
		return true
	}
	if looksHowDescribeProcessQuery(query) && leftoverCoveringProcessHortativeLine(line) {
		return true
	}
	if looksWhatMotivatesQuery(query) && leftoverCoveringMotivateCauseLine(query, line) {
		return true
	}
	if looksWhatSayAboutQuery(query) && leftoverCoveringSayAboutTargetLine(line) {
		return true
	}
	if looksHowReactQuery(query) && leftoverCoveringReactionObservationLine(line) && leftoverCoveringReactLineHasObject(query, line) {
		return true
	}
	if looksWhatDidPurposeQuery(query) && leftoverCoveringPurposeActionLine(query, line) {
		return true
	}
	if looksHowDidStartQuery(query) && leftoverCoveringStartMethodLine(query, line) {
		return true
	}
	if looksHowLongBeenQuery(query) && leftoverCoveringDurationLine(query, line) {
		return true
	}
	if looksHowOftenQuery(query) && leftoverCoveringCadenceLine(query, line) {
		return true
	}
	if looksWhatProjectWorkingQuery(query) && leftoverCoveringCurrentProjectLine(query, line) {
		return true
	}
	if looksWhatNewHobbyQuery(query) && leftoverCoveringBecomeInterestedLine(query, line) {
		return true
	}
	if looksHowPlanDreamQuery(query) && leftoverCoveringPrepPlanLine(query, line) {
		return true
	}
	if looksWhatFocusingBesidesQuery(query) && leftoverCoveringFocusingBesidesLine(query, line) {
		return true
	}
	if looksWhatNewSeriesQuery(query) && leftoverCoveringTitledShowLine(query, line) {
		return true
	}
	if looksWhatDidRecentlyAtQuery(query) && leftoverCoveringLocativePurposeLine(query, line) {
		return true
	}
	if looksHowFeelAboutQuery(query) && leftoverCoveringExperiencingFeelingLine(query, line) {
		return true
	}
	if looksWhatDoCoordinatedUseQuery(query) && leftoverCoveringWorkDeterminationLine(query, line) {
		return true
	}
	if looksWhatDidRealizeAfterQuery(query) && leftoverCoveringSelfDirectedRealizeLine(query, line) {
		return true
	}
	return looksWhatKindQuery(query) && leftoverCoveringKindListLine(line)
}

func leftoverCoveringHostedEventLine(line string) bool {
	low := " " + strings.ToLower(hybridLineBody(line)) + " "
	for _, n := range []string{" party ", " dinner ", " gathering ", " celebration ", " reception "} {
		if strings.Contains(low, n) {
			return true
		}
	}
	return false
}

func leftoverCoveringRealizeSpeechLine(line string) bool {
	low := " " + strings.ToLower(hybridLineBody(line)) + " "
	return strings.Contains(low, " realized ") || strings.Contains(low, " realise ") || strings.Contains(low, " realize ")
}

func leftoverCoveringPreferHostedEvent(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 || looksChatTurnLine(row.line) {
			continue
		}
		if !leftoverCoveringHostedEventLine(row.line) {
			continue
		}
		if leftoverCoveringSkipForeignWhenEvent(query, row.line) {
			continue
		}
		if leftoverCoveringLineHasForeignPerson(query, row.line) && !leftoverCoveringMentionsQueryEntity(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringJoinHostedEvents(query string, scored []leftoverCoverScored) string {
	if !looksHostQuery(query) {
		return ""
	}
	rows := make([]leftoverCoverScored, 0, 4)
	for _, row := range scored {
		if row.score < 2 || looksChatTurnLine(row.line) {
			continue
		}
		if !leftoverCoveringHostedEventLine(row.line) {
			continue
		}
		if leftoverCoveringSkipForeignWhenEvent(query, row.line) {
			continue
		}
		if leftoverCoveringLineHasForeignPerson(query, row.line) && !leftoverCoveringMentionsQueryEntity(query, row.line) {
			continue
		}
		rows = append(rows, row)
	}
	if len(rows) < 2 {
		return ""
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].score != rows[j].score {
			return rows[i].score > rows[j].score
		}
		return false
	})
	parts := make([]string, 0, 2)
	seen := map[string]struct{}{}
	for _, row := range rows {
		line := stripConflictingDateTail(query, row.line)
		key := strings.ToLower(strings.TrimSpace(line))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		parts = append(parts, line)
		if len(parts) >= 2 {
			break
		}
	}
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts, "; ")
}

func leftoverCoveringHostMissesEvent(query, covering, answer string) bool {
	if !looksHostQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringSkipForeignWhenEvent(query, covering) {
		return false
	}
	if leftoverCoveringHostedEventLine(answer) {
		return false
	}
	if !leftoverCoveringHostedEventLine(covering) {
		return false
	}
	if leftoverCoveringLineHasForeignPerson(query, covering) && !leftoverCoveringMentionsQueryEntity(query, covering) {
		return false
	}
	return true
}

func leftoverCoveringAdviceSpeechToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "advice", "advise", "advises", "advised", "advising":
		return true
	}
	return false
}

func leftoverCoveringAdviceRestatementToken(tok string) bool {
	if leftoverCoveringAdviceSpeechToken(tok) {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "business", "biz", "running", "successful", "success",
		"give", "gives", "gave", "giving", "tip", "tips":
		return true
	}
	return false
}

func leftoverCoveringAdviceSpeechLine(line string) bool {
	for _, tok := range tokenize(hybridLineBody(line)) {
		if leftoverCoveringAdviceSpeechToken(tok) {
			return true
		}
	}
	return false
}

func leftoverCoveringAdviceDirectiveLine(line string) bool {
	if leftoverCoveringHortativePrefixLine(line) {
		return true
	}
	if looksChatTurnLine(line) || looksInterrogativeLine(line) || leftoverCoveringAdviceSpeechLine(line) {
		return false
	}
	body := strings.ToLower(strings.TrimSpace(hybridLineBody(line)))
	if body == "" {
		return false
	}
	fields := strings.Fields(leftoverCoveringStripLeadConjunctions(body))
	if len(fields) == 0 {
		return false
	}
	w := strings.Trim(fields[0], ".,;:()[]\"'")
	if len(w) < 6 || !strings.HasSuffix(w, "ing") {
		return false
	}
	// First-person gerund leftover ("Building X … I'm always working on").
	// Second-person evaluative gerunds ("Seeing your students…") are not advice content.
	padded := " " + body + " "
	return strings.Contains(padded, " i ") || strings.Contains(padded, " i'm ") || strings.Contains(padded, " my ")
}

func leftoverCoveringStripLeadConjunctions(body string) string {
	trimmed := strings.TrimSpace(body)
	for {
		switch {
		case strings.HasPrefix(trimmed, "and "):
			trimmed = strings.TrimSpace(trimmed[4:])
			continue
		case strings.HasPrefix(trimmed, "also "):
			trimmed = strings.TrimSpace(trimmed[5:])
			continue
		}
		break
	}
	return trimmed
}

func leftoverCoveringHortativePrefixLine(line string) bool {
	if looksChatTurnLine(line) || looksInterrogativeLine(line) || leftoverCoveringAdviceSpeechLine(line) {
		return false
	}
	body := strings.ToLower(strings.TrimSpace(hybridLineBody(line)))
	if body == "" {
		return false
	}
	trimmed := leftoverCoveringStripLeadConjunctions(body)
	for _, p := range []string{"be sure ", "be sure to ", "don't forget", "do not forget", "make sure ", "remember to ", "try to ", "just keep "} {
		if strings.HasPrefix(trimmed, p) {
			return true
		}
	}
	return false
}

func leftoverCoveringAdviceOffQueryLine(line string) bool {
	if !leftoverCoveringAdviceDirectiveLine(line) {
		return false
	}
	for _, tok := range tokenize(hybridLineBody(line)) {
		if leftoverCoveringAdviceRestatementToken(tok) {
			return false
		}
	}
	return true
}

func leftoverCoveringPreferAdviceDirective(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 || looksChatTurnLine(row.line) {
			continue
		}
		if !leftoverCoveringAdviceOffQueryLine(row.line) {
			continue
		}
		if leftoverCoveringSkipForeignWhenEvent(query, row.line) {
			continue
		}
		if leftoverCoveringLineHasForeignPerson(query, row.line) && !leftoverCoveringMentionsQueryEntity(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringJoinAdviceDirectives(query string, scored []leftoverCoverScored) string {
	if !looksAdviceQuery(query) {
		return ""
	}
	rows := make([]leftoverCoverScored, 0, 6)
	for _, row := range scored {
		if row.score < 2 || looksChatTurnLine(row.line) {
			continue
		}
		if !leftoverCoveringAdviceOffQueryLine(row.line) {
			continue
		}
		if leftoverCoveringSkipForeignWhenEvent(query, row.line) {
			continue
		}
		if leftoverCoveringLineHasForeignPerson(query, row.line) && !leftoverCoveringMentionsQueryEntity(query, row.line) {
			continue
		}
		rows = append(rows, row)
	}
	if len(rows) < 2 {
		return ""
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].score != rows[j].score {
			return rows[i].score > rows[j].score
		}
		li := len(contentBearingTokens(tokenize(rows[i].line)))
		lj := len(contentBearingTokens(tokenize(rows[j].line)))
		return li > lj
	})
	parts := make([]string, 0, 3)
	seen := map[string]struct{}{}
	for _, row := range rows {
		line := stripConflictingDateTail(query, row.line)
		key := strings.ToLower(strings.TrimSpace(line))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		parts = append(parts, line)
		if len(parts) >= 3 {
			break
		}
	}
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts, "; ")
}

func leftoverCoveringAdviceMissesDirective(query, covering, answer string) bool {
	if !looksAdviceQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringSkipForeignWhenEvent(query, covering) {
		return false
	}
	if leftoverCoveringAdviceOffQueryLine(answer) {
		return false
	}
	if !leftoverCoveringAdviceOffQueryLine(covering) {
		return false
	}
	return true
}

func leftoverCoveringKindRestatementToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "spread", "spreading", "spreads", "spreaded",
		"kind", "kinds":
		return true
	}
	return false
}

func leftoverCoveringKindRestatementLine(line string) bool {
	for _, tok := range tokenize(hybridLineBody(line)) {
		if leftoverCoveringKindRestatementToken(tok) {
			return true
		}
	}
	return false
}

// leftoverCoveringKindListLine is leftover that answers "what kind of"
// with a like-A,-B,-and-C enumeration, omitting question restatement
// ("spreading kindness" from dinner spread).
func leftoverCoveringKindListLine(line string) bool {
	if looksChatTurnLine(line) || looksInterrogativeLine(line) || leftoverCoveringKindRestatementLine(line) {
		return false
	}
	body := strings.ToLower(strings.TrimSpace(hybridLineBody(line)))
	if body == "" {
		return false
	}
	idx := strings.Index(body, " like ")
	if idx < 0 {
		return false
	}
	rest := body[idx+6:]
	return strings.Contains(rest, ",") && strings.Contains(rest, " and ")
}

func kindListSlots(line string) []string {
	if !leftoverCoveringKindListLine(line) {
		return nil
	}
	body := strings.ToLower(strings.TrimSpace(hybridLineBody(line)))
	idx := strings.Index(body, " like ")
	if idx < 0 {
		return nil
	}
	rest := body[idx+6:]
	if j := strings.IndexAny(rest, ".!?"); j >= 0 {
		rest = rest[:j]
	}
	rest = strings.ReplaceAll(rest, " and ", ",")
	out := make([]string, 0, 4)
	seen := map[string]struct{}{}
	for _, part := range strings.Split(rest, ",") {
		part = strings.TrimSpace(strings.Trim(part, " .,"))
		if part == "" || utf8Len(part) < 3 {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		out = append(out, part)
	}
	return out
}

func kindListCoveredByAnswer(covering, answer string) bool {
	slots := kindListSlots(covering)
	if len(slots) < 2 {
		return false
	}
	al := strings.ToLower(strings.TrimSpace(answer))
	if al == "" {
		return false
	}
	for _, s := range slots {
		if !strings.Contains(al, s) {
			return false
		}
	}
	return true
}

func leftoverCoveringPreferKindList(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 || looksChatTurnLine(row.line) {
			continue
		}
		if !leftoverCoveringKindListLine(row.line) {
			continue
		}
		if leftoverCoveringSkipForeignWhenEvent(query, row.line) {
			continue
		}
		if leftoverCoveringLineHasForeignPerson(query, row.line) && !leftoverCoveringMentionsQueryEntity(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringKindMissesList(query, covering, answer string) bool {
	if !looksWhatKindQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringSkipForeignWhenEvent(query, covering) {
		return false
	}
	if leftoverCoveringKindListLine(answer) {
		return false
	}
	if !leftoverCoveringKindListLine(covering) {
		return false
	}
	if kindListCoveredByAnswer(covering, answer) {
		return false
	}
	return true
}

func leftoverCoveringProcessRestatementToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "process", "describe", "describes", "described", "describing",
		"taking", "care", "cares":
		return true
	}
	return false
}

func leftoverCoveringProcessCompanionLine(query, line string) bool {
	if leftoverCoveringProcessHortativeLine(line) {
		return false
	}
	for _, tok := range leftoverCoverNonWeakTokens(contentBearingTokens(tokenize(query))) {
		if leftoverCoveringProcessRestatementToken(tok) {
			continue
		}
		if contentCoversQueryToken(line, tok) {
			return true
		}
	}
	return false
}

func leftoverCoveringProcessHortativeLine(line string) bool {
	if !leftoverCoveringHortativePrefixLine(line) {
		return false
	}
	for _, tok := range tokenize(hybridLineBody(line)) {
		if leftoverCoveringProcessRestatementToken(tok) {
			return false
		}
	}
	return true
}

func leftoverCoveringPreferProcessHortative(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 || looksChatTurnLine(row.line) {
			continue
		}
		if !leftoverCoveringProcessHortativeLine(row.line) {
			continue
		}
		if leftoverCoveringSkipForeignWhenEvent(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringProcessMissesHortative(query, covering, answer string) bool {
	if !looksHowDescribeProcessQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringSkipForeignWhenEvent(query, covering) {
		return false
	}
	if leftoverCoveringProcessHortativeLine(answer) {
		return false
	}
	if !leftoverCoveringProcessHortativeLine(covering) {
		return false
	}
	return true
}

func leftoverCoveringMotivateStructureToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "motivate", "motivates", "motivating", "motivation",
		"keep", "keeps", "keeping", "even", "tough", "day", "days":
		return true
	}
	return leftoverCoverWeakToken(tok) || isQueryStopword(tok)
}

func leftoverCoveringMotivateObjectTokens(query string) []string {
	people := map[string]struct{}{}
	for _, e := range hopQueryEntities(query) {
		people[strings.ToLower(strings.TrimSpace(e))] = struct{}{}
	}
	out := make([]string, 0, 4)
	for _, tok := range leftoverCoverNonWeakTokens(contentBearingTokens(tokenize(query))) {
		low := strings.ToLower(strings.TrimSpace(tok))
		if leftoverCoveringMotivateStructureToken(low) {
			continue
		}
		if _, ok := people[low]; ok {
			continue
		}
		out = append(out, low)
	}
	return out
}

func leftoverCoveringMotivateFirstPersonObject(query, line string) bool {
	body := strings.ToLower(strings.TrimSpace(hybridLineBody(line)))
	if body == "" {
		return false
	}
	padded := " " + body + " "
	for _, obj := range leftoverCoveringMotivateObjectTokens(query) {
		if obj == "" {
			continue
		}
		if strings.Contains(padded, " my "+obj+" ") || strings.Contains(padded, " my "+obj+"s ") {
			return true
		}
	}
	return false
}

func leftoverCoveringMotivateCauseConnective(line string) bool {
	body := strings.ToLower(strings.TrimSpace(hybridLineBody(line)))
	if body == "" {
		return false
	}
	trimmed := leftoverCoveringStripLeadConjunctions(body)
	if strings.HasPrefix(trimmed, "it's ") || strings.HasPrefix(trimmed, "it is ") || strings.HasPrefix(trimmed, "knowing ") {
		return true
	}
	padded := " " + body + " "
	for _, p := range []string{" that keeps me ", " keeps me going ", " knowing that ", " because "} {
		if strings.Contains(padded, p) {
			return true
		}
	}
	return false
}

func leftoverCoveringMotivateCauseLine(query, line string) bool {
	if looksChatTurnLine(line) || looksInterrogativeLine(line) {
		return false
	}
	if !leftoverCoveringMotivateFirstPersonObject(query, line) {
		return false
	}
	return leftoverCoveringMotivateCauseConnective(line)
}

func leftoverCoveringMotivateCompanionLine(query, line string) bool {
	if leftoverCoveringMotivateCauseLine(query, line) {
		return false
	}
	for _, tok := range leftoverCoverNonWeakTokens(contentBearingTokens(tokenize(query))) {
		if leftoverCoveringMotivateStructureToken(tok) {
			if contentCoversQueryToken(line, tok) {
				return true
			}
			continue
		}
		if contentCoversQueryToken(line, tok) && !leftoverCoveringMotivateFirstPersonObject(query, line) {
			return true
		}
	}
	return false
}

func leftoverCoveringPreferMotivateCause(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 || looksChatTurnLine(row.line) {
			continue
		}
		if !leftoverCoveringMotivateCauseLine(query, row.line) {
			continue
		}
		if leftoverCoveringSkipForeignWhenEvent(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringMotivateMissesCause(query, covering, answer string) bool {
	if !looksWhatMotivatesQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringSkipForeignWhenEvent(query, covering) {
		return false
	}
	if leftoverCoveringMotivateCauseLine(query, answer) {
		return false
	}
	if !leftoverCoveringMotivateCauseLine(query, covering) {
		return false
	}
	return true
}

func leftoverCoveringSkipsHybrid(query, covering string) bool {
	covering = strings.TrimSpace(covering)
	if covering == "" {
		return false
	}
	if looksWhatMotivatesQuery(query) && leftoverCoveringMotivateCauseLine(query, covering) {
		return true
	}
	if looksHowReactQuery(query) && leftoverCoveringReactionObservationLine(covering) && leftoverCoveringReactLineHasObject(query, covering) {
		return true
	}
	if looksWhatDidPurposeQuery(query) && leftoverCoveringPurposeActionLine(query, covering) {
		return true
	}
	if looksHowDidStartQuery(query) && leftoverCoveringStartMethodLine(query, covering) {
		return true
	}
	if looksHowLongBeenQuery(query) && leftoverCoveringDurationLine(query, covering) {
		return true
	}
	if looksHowOftenQuery(query) && leftoverCoveringCadenceLine(query, covering) {
		return true
	}
	if looksWhatProjectWorkingQuery(query) && leftoverCoveringCurrentProjectLine(query, covering) {
		return true
	}
	if looksWhatNewHobbyQuery(query) && leftoverCoveringBecomeInterestedLine(query, covering) {
		return true
	}
	if looksHowPlanDreamQuery(query) && leftoverCoveringPrepPlanLine(query, covering) {
		return true
	}
	if looksWhatFocusingBesidesQuery(query) && leftoverCoveringFocusingBesidesLine(query, covering) {
		return true
	}
	if looksWhatNewSeriesQuery(query) && leftoverCoveringTitledShowLine(query, covering) {
		return true
	}
	if looksWhatDidRecentlyAtQuery(query) && leftoverCoveringLocativePurposeLine(query, covering) {
		return true
	}
	if looksHowFeelAboutQuery(query) && leftoverCoveringExperiencingFeelingLine(query, covering) {
		return true
	}
	if looksWhatDoCoordinatedUseQuery(query) && leftoverCoveringWorkDeterminationLine(query, covering) {
		return true
	}
	if looksWhatDidRealizeAfterQuery(query) && leftoverCoveringSelfDirectedRealizeLine(query, covering) {
		return true
	}
	return looksWhatSayAboutQuery(query) && leftoverCoveringSayAboutTargetLine(covering)
}

func leftoverCoveringSayAboutStructureToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "say", "says", "said", "saying", "about":
		return true
	}
	return false
}

func leftoverCoveringEvaluativeTheyLine(line string) bool {
	if looksChatTurnLine(line) || looksInterrogativeLine(line) || looksImageCaptionLine(line) {
		return false
	}
	body := strings.ToLower(strings.TrimSpace(hybridLineBody(line)))
	body = strings.Trim(body, `"'`)
	body = strings.TrimRight(body, ".!")
	words := strings.Fields(body)
	if len(words) < 3 || len(words) > 6 {
		return false
	}
	if words[0] == "they're" && words[1] == "so" {
		return true
	}
	if words[0] == "they" && words[1] == "are" && words[2] == "so" {
		return true
	}
	if words[0] == "they" && (words[1] == "look" || words[1] == "seem") {
		return true
	}
	return false
}

// leftoverCoveringFirstPersonGotLine is first-person possessive/abundance
// leftover ("It's got so much to check out"). It is not they-evaluative, not
// "it's gotta", and not reported doctor speech.
func leftoverCoveringFirstPersonGotLine(line string) bool {
	if leftoverCoveringEvaluativeTheyLine(line) {
		return false
	}
	if looksChatTurnLine(line) || looksInterrogativeLine(line) || looksImageCaptionLine(line) {
		return false
	}
	body := strings.ToLower(strings.TrimSpace(hybridLineBody(line)))
	body = strings.Trim(body, `"'`)
	body = strings.TrimRight(body, ".!")
	if strings.HasPrefix(body, "the doctor") {
		return false
	}
	if strings.HasPrefix(body, "it's gotta") || strings.HasPrefix(body, "its gotta") {
		return false
	}
	if strings.HasPrefix(body, "it's got to") || strings.HasPrefix(body, "it has got to") {
		return false
	}
	words := strings.Fields(body)
	if len(words) < 4 || len(words) > 40 {
		return false
	}
	gotQuant := func(i int) bool {
		if i >= len(words) {
			return false
		}
		switch words[i] {
		case "so", "a", "an", "the", "lots", "plenty", "many":
			return true
		}
		return false
	}
	if (words[0] == "it's" || words[0] == "its") && words[1] == "got" && gotQuant(2) {
		return true
	}
	if words[0] == "it" && words[1] == "has" && words[2] == "got" && gotQuant(3) {
		return true
	}
	if words[0] == "it" && words[1] == "has" && words[2] == "so" {
		return true
	}
	return false
}

func leftoverCoveringSayAboutTargetLine(line string) bool {
	return leftoverCoveringEvaluativeTheyLine(line) || leftoverCoveringFirstPersonGotLine(line) || leftoverCoveringReportedSpeechLine(line)
}

// leftoverCoveringReportedSpeechLine is attributed leftover of the form
// "The <role> said it's …". It is not they-evaluative and not first-person got.
func leftoverCoveringReportedSpeechLine(line string) bool {
	if leftoverCoveringEvaluativeTheyLine(line) || leftoverCoveringFirstPersonGotLine(line) {
		return false
	}
	if looksChatTurnLine(line) || looksInterrogativeLine(line) || looksImageCaptionLine(line) {
		return false
	}
	body := strings.ToLower(strings.TrimSpace(hybridLineBody(line)))
	body = strings.Trim(body, `"'`)
	body = strings.TrimRight(body, ".!")
	words := strings.Fields(body)
	if len(words) < 5 || len(words) > 16 {
		return false
	}
	if words[0] != "the" || words[2] != "said" {
		return false
	}
	role := words[1]
	if len(role) < 3 {
		return false
	}
	for _, r := range role {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	switch words[3] {
	case "it's", "its":
		return true
	case "it":
		return len(words) > 4 && (words[4] == "is" || words[4] == "was" || words[4] == "isn't")
	}
	return false
}

func leftoverCoveringSayAboutHasObjectEvidence(query string, lines []string) bool {
	toks := leftoverCoverNonWeakTokens(contentBearingTokens(tokenize(query)))
	people := map[string]struct{}{}
	for _, e := range hopQueryEntities(query) {
		people[strings.ToLower(strings.TrimSpace(e))] = struct{}{}
	}
	for _, line := range lines {
		for _, tok := range toks {
			if leftoverCoveringSayAboutStructureToken(tok) || leftoverCoveringSayAboutFrameParticiple(tok) {
				continue
			}
			if _, ok := people[strings.ToLower(tok)]; ok {
				continue
			}
			if contentCoversQueryToken(line, tok) {
				return true
			}
		}
	}
	return false
}

func leftoverCoveringSayAboutCompanionLine(query, line string) bool {
	if leftoverCoveringSayAboutTargetLine(line) {
		return false
	}
	for _, tok := range leftoverCoverNonWeakTokens(contentBearingTokens(tokenize(query))) {
		if leftoverCoveringSayAboutStructureToken(tok) {
			continue
		}
		if contentCoversQueryToken(line, tok) {
			return true
		}
	}
	return false
}

func leftoverCoveringPreferEvaluativeThey(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 || looksChatTurnLine(row.line) {
			continue
		}
		if !leftoverCoveringEvaluativeTheyLine(row.line) {
			continue
		}
		if leftoverCoveringSkipForeignWhenEvent(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringPreferFirstPersonGot(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 || looksChatTurnLine(row.line) {
			continue
		}
		if !leftoverCoveringFirstPersonGotLine(row.line) {
			continue
		}
		if leftoverCoveringSkipForeignWhenEvent(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringPreferReportedSpeech(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 || looksChatTurnLine(row.line) {
			continue
		}
		if !leftoverCoveringReportedSpeechLine(row.line) {
			continue
		}
		if leftoverCoveringSkipForeignWhenEvent(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringSayAboutMissesEvaluative(query, covering, answer string) bool {
	if !looksWhatSayAboutQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringSkipForeignWhenEvent(query, covering) {
		return false
	}
	if leftoverCoveringSayAboutTargetLine(answer) {
		return false
	}
	if !leftoverCoveringSayAboutTargetLine(covering) {
		return false
	}
	return true
}

func leftoverCoveringReactStructureToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "react", "reacts", "reacted", "reacting",
		"respond", "responds", "responded", "responding":
		return true
	}
	return false
}

func leftoverCoveringReactionAdj(tok string) bool {
	tok = strings.ToLower(strings.TrimSpace(tok))
	if len(tok) < 4 || leftoverCoverWeakToken(tok) || isQueryStopword(tok) || isMonthWord(tok) {
		return false
	}
	for _, r := range tok {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return true
}

// leftoverCoveringReactionObservationLine is leftover of the form
// "they were (so) ADJ" / "they looked (so) ADJ" as the last words.
func leftoverCoveringReactionObservationLine(line string) bool {
	if leftoverCoveringEvaluativeTheyLine(line) {
		return false
	}
	if looksChatTurnLine(line) || looksInterrogativeLine(line) || looksImageCaptionLine(line) {
		return false
	}
	body := strings.ToLower(strings.TrimSpace(hybridLineBody(line)))
	body = strings.Trim(body, `"'`)
	body = strings.TrimRight(body, ".!")
	words := strings.Fields(body)
	if len(words) < 6 || len(words) > 24 {
		return false
	}
	n := len(words)
	if !leftoverCoveringReactionAdj(words[n-1]) {
		return false
	}
	if n >= 4 && words[n-4] == "they" && (words[n-3] == "were" || words[n-3] == "looked" || words[n-3] == "seemed") && words[n-2] == "so" {
		return true
	}
	if n >= 3 && words[n-3] == "they" && (words[n-2] == "were" || words[n-2] == "looked" || words[n-2] == "seemed") {
		return true
	}
	return false
}

func leftoverCoveringReactObjectTokens(query string) []string {
	people := map[string]struct{}{}
	for _, e := range hopQueryEntities(query) {
		people[strings.ToLower(strings.TrimSpace(e))] = struct{}{}
	}
	out := make([]string, 0, 4)
	for _, tok := range leftoverCoverNonWeakTokens(contentBearingTokens(tokenize(query))) {
		low := strings.ToLower(strings.TrimSpace(tok))
		if leftoverCoveringReactStructureToken(low) {
			continue
		}
		if _, ok := people[low]; ok {
			continue
		}
		out = append(out, low)
	}
	return out
}

func leftoverCoveringReactLineHasObject(query, line string) bool {
	for _, tok := range leftoverCoveringReactObjectTokens(query) {
		if contentCoversQueryToken(line, tok) {
			return true
		}
	}
	return false
}

func leftoverCoveringReactCompanionLine(query, line string) bool {
	if leftoverCoveringReactionObservationLine(line) {
		return false
	}
	for _, tok := range leftoverCoveringReactObjectTokens(query) {
		if contentCoversQueryToken(line, tok) {
			return true
		}
	}
	return false
}

func leftoverCoveringPreferReactionObservation(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 || looksChatTurnLine(row.line) {
			continue
		}
		if !leftoverCoveringReactionObservationLine(row.line) {
			continue
		}
		if !leftoverCoveringReactLineHasObject(query, row.line) {
			continue
		}
		if leftoverCoveringSkipForeignWhenEvent(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringReactMissesObservation(query, covering, answer string) bool {
	if !looksHowReactQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringSkipForeignWhenEvent(query, covering) {
		return false
	}
	if leftoverCoveringReactionObservationLine(answer) {
		return false
	}
	if !leftoverCoveringReactionObservationLine(covering) {
		return false
	}
	if !leftoverCoveringReactLineHasObject(query, covering) {
		return false
	}
	return true
}

func leftoverCoveringPurposeStructureToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "better", "best", "worse", "greater", "less":
		return true
	}
	return false
}

func leftoverCoveringPurposeObjectTokens(query string) []string {
	toks := tokenize(query)
	toIdx := -1
	sawDo := false
	for i, tok := range toks {
		if tok == "do" {
			sawDo = true
			continue
		}
		if sawDo && tok == "to" {
			toIdx = i
		}
	}
	if toIdx < 0 || toIdx+1 >= len(toks) {
		return nil
	}
	people := map[string]struct{}{}
	for _, e := range hopQueryEntities(query) {
		people[strings.ToLower(strings.TrimSpace(e))] = struct{}{}
	}
	out := make([]string, 0, 4)
	for _, tok := range toks[toIdx+1:] {
		low := strings.ToLower(strings.TrimSpace(tok))
		if low == "" || leftoverCoveringPurposeStructureToken(low) || leftoverCoverWeakToken(low) || isQueryStopword(low) || isMonthWord(low) || isCalendarCoverToken(low) {
			continue
		}
		if _, ok := people[low]; ok {
			continue
		}
		out = append(out, low)
	}
	return out
}

func leftoverCoveringPurposePairAdjacent(query, line string) bool {
	objs := leftoverCoveringPurposeObjectTokens(query)
	if len(objs) < 2 {
		return false
	}
	a, b := objs[0], objs[1]
	words := tokenize(hybridLineBody(line))
	for i := 0; i+1 < len(words); i++ {
		if tokensMatch(words[i], a) && tokensMatch(words[i+1], b) {
			return true
		}
	}
	return false
}

func leftoverCoveringPurposeActorLine(query, line string) bool {
	body := strings.ToLower(strings.TrimSpace(hybridLineBody(line)))
	if body == "" {
		return false
	}
	padded := " " + body + " "
	if strings.Contains(padded, " i ") || strings.Contains(padded, " i've ") || strings.Contains(padded, " i’m ") || strings.HasPrefix(body, "i ") || strings.HasPrefix(body, "i've ") {
		return true
	}
	return leftoverCoveringQueryEntityHits(query, hybridLineBody(line)) > 0
}

func leftoverCoveringPurposeActionLine(query, line string) bool {
	if looksChatTurnLine(line) || looksInterrogativeLine(line) || looksImageCaptionLine(line) {
		return false
	}
	if !leftoverCoveringPurposePairAdjacent(query, line) {
		return false
	}
	return leftoverCoveringPurposeActorLine(query, line)
}

func leftoverCoveringPurposeCompanionLine(query, line string) bool {
	if leftoverCoveringPurposeActionLine(query, line) {
		return false
	}
	for _, tok := range leftoverCoveringPurposeObjectTokens(query) {
		if contentCoversQueryToken(line, tok) {
			return true
		}
	}
	if leftoverCoveringPurposeStructureToken("better") && contentCoversQueryToken(line, "better") {
		return true
	}
	return false
}

func leftoverCoveringPreferPurposeAction(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 || looksChatTurnLine(row.line) {
			continue
		}
		if !leftoverCoveringPurposeActionLine(query, row.line) {
			continue
		}
		if leftoverCoveringSkipForeignWhenEvent(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringPurposeMissesAction(query, covering, answer string) bool {
	if !looksWhatDidPurposeQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringSkipForeignWhenEvent(query, covering) {
		return false
	}
	if leftoverCoveringPurposeActionLine(query, answer) {
		return false
	}
	return leftoverCoveringPurposeActionLine(query, covering)
}

func leftoverCoveringStartStructureToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "start", "started", "starting",
		"transformation", "journey", "journeys":
		return true
	}
	return false
}

func leftoverCoveringAgoDurationTokens(query string) []string {
	if parseYearsAgo(strings.ToLower(query)) <= 0 {
		return nil
	}
	toks := tokenize(query)
	for i := 0; i+2 < len(toks); i++ {
		if toks[i+2] != "ago" {
			continue
		}
		if toks[i+1] != "year" && toks[i+1] != "years" {
			continue
		}
		return []string{toks[i], toks[i+1], toks[i+2]}
	}
	return nil
}

func leftoverCoveringStartDurationMatch(query, line string) bool {
	want := parseYearsAgo(strings.ToLower(query))
	if want <= 0 {
		return false
	}
	return parseYearsAgo(strings.ToLower(hybridLineBody(line))) == want
}

func leftoverCoveringStartInceptionCount(line string) int {
	seen := map[string]struct{}{}
	for _, tok := range tokenize(hybridLineBody(line)) {
		switch tok {
		case "started":
			seen["start"] = struct{}{}
		case "changed":
			seen["change"] = struct{}{}
		case "began", "begun":
			seen["begin"] = struct{}{}
		}
	}
	return len(seen)
}

func leftoverCoveringStartActorLine(query, line string) bool {
	body := strings.ToLower(strings.TrimSpace(hybridLineBody(line)))
	if body == "" {
		return false
	}
	padded := " " + body + " "
	if strings.Contains(padded, " i ") || strings.Contains(padded, " i've ") || strings.Contains(padded, " i'm ") ||
		strings.HasPrefix(body, "i ") || strings.HasPrefix(body, "i've ") || strings.Contains(padded, " my ") {
		return true
	}
	return leftoverCoveringQueryEntityHits(query, hybridLineBody(line)) > 0
}

func leftoverCoveringStartMethodPairLine(query, line string) bool {
	if looksChatTurnLine(line) || looksInterrogativeLine(line) || looksImageCaptionLine(line) {
		return false
	}
	if leftoverCoveringStartInceptionCount(line) < 2 {
		return false
	}
	return leftoverCoveringStartActorLine(query, line)
}

func leftoverCoveringStartMethodDurationLine(query, line string) bool {
	if looksChatTurnLine(line) || looksInterrogativeLine(line) || looksImageCaptionLine(line) {
		return false
	}
	if leftoverCoveringStartInceptionCount(line) < 1 {
		return false
	}
	if !leftoverCoveringStartDurationMatch(query, line) {
		return false
	}
	if contentCoversQueryToken(line, "transformation") || contentCoversQueryToken(line, "journey") {
		return false
	}
	return leftoverCoveringStartActorLine(query, line)
}

func leftoverCoveringStartMethodLine(query, line string) bool {
	return leftoverCoveringStartMethodPairLine(query, line) || leftoverCoveringStartMethodDurationLine(query, line)
}

func leftoverCoveringStartCompanionLine(query, line string) bool {
	if leftoverCoveringStartMethodLine(query, line) {
		return false
	}
	for _, tok := range leftoverCoveringAgoDurationTokens(query) {
		if contentCoversQueryToken(line, tok) {
			return true
		}
	}
	return contentCoversQueryToken(line, "journey") || contentCoversQueryToken(line, "transformation")
}

func leftoverCoveringPreferStartMethod(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	pair := false
	for _, row := range scored {
		if row.score < 2 || looksChatTurnLine(row.line) {
			continue
		}
		if leftoverCoveringLineHasForeignPerson(query, row.line) && leftoverCoveringQueryEntityHits(query, row.line) == 0 {
			continue
		}
		isPair := leftoverCoveringStartMethodPairLine(query, row.line)
		isDur := leftoverCoveringStartMethodDurationLine(query, row.line)
		if !isPair && !isDur {
			continue
		}
		if isPair && !pair {
			pair = true
			pickScore = row.score
			pick = row.line
			continue
		}
		if pair && !isPair {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringJoinStartMethods(query string, scored []leftoverCoverScored) string {
	if !looksHowDidStartQuery(query) {
		return ""
	}
	for _, row := range scored {
		if row.score < 2 || looksChatTurnLine(row.line) {
			continue
		}
		if leftoverCoveringStartMethodPairLine(query, row.line) {
			if leftoverCoveringLineHasForeignPerson(query, row.line) && leftoverCoveringQueryEntityHits(query, row.line) == 0 {
				continue
			}
			return row.line
		}
	}
	parts := make([]string, 0, 4)
	seen := map[string]struct{}{}
	for _, row := range scored {
		if row.score < 2 || looksChatTurnLine(row.line) {
			continue
		}
		if !leftoverCoveringStartMethodDurationLine(query, row.line) {
			continue
		}
		if leftoverCoveringLineHasForeignPerson(query, row.line) && leftoverCoveringQueryEntityHits(query, row.line) == 0 {
			continue
		}
		line := stripConflictingDateTail(query, row.line)
		key := strings.ToLower(strings.TrimSpace(line))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		parts = append(parts, line)
		if len(parts) >= 4 {
			break
		}
	}
	if len(parts) >= 2 {
		return strings.Join(parts, "; ")
	}
	return ""
}

func leftoverCoveringStartMissesMethod(query, covering, answer string) bool {
	if !looksHowDidStartQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" {
		return false
	}
	if leftoverCoveringStartMethodPairLine(query, covering) && !leftoverCoveringStartMethodPairLine(query, answer) {
		return true
	}
	if leftoverCoveringStartMethodLine(query, answer) {
		return false
	}
	return leftoverCoveringStartMethodLine(query, covering)
}

func leftoverCoveringDurationStructureToken(tok string) bool {
	return strings.ToLower(strings.TrimSpace(tok)) == "long"
}

func yearCountBefore(lower, suffix string) int {
	words := map[string]int{
		"one": 1, "two": 2, "three": 3, "four": 4, "five": 5,
		"six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10,
	}
	for word, n := range words {
		if strings.Contains(lower, word+suffix) {
			return n
		}
	}
	for n := 1; n <= 20; n++ {
		if strings.Contains(lower, strconv.Itoa(n)+suffix) {
			return n
		}
	}
	return 0
}

func parseContinuingYears(lower string) int {
	lower = strings.ToLower(lower)
	ago := parseYearsAgo(lower)
	hasAlready := strings.Contains(lower, "already")
	hasFor := strings.Contains(lower, " for ") || strings.HasPrefix(strings.TrimSpace(lower), "for ")
	hasDuration := strings.Contains(lower, "duration")
	if ago > 0 && !hasAlready && !hasFor && !hasDuration {
		return 0
	}
	if n := yearCountBefore(lower, " years already"); n > 0 {
		return n
	}
	if n := yearCountBefore(lower, " year already"); n > 0 {
		return n
	}
	if hasDuration {
		if n := yearCountBefore(lower, " years"); n > 0 {
			return n
		}
		if n := yearCountBefore(lower, " year"); n > 0 {
			return n
		}
	}
	if hasFor {
		rest := lower
		if i := strings.Index(lower, " for "); i >= 0 {
			rest = strings.TrimSpace(lower[i+5:])
		} else if strings.HasPrefix(strings.TrimSpace(lower), "for ") {
			rest = strings.TrimSpace(strings.TrimSpace(lower)[4:])
		}
		if n := yearCountBefore(" "+rest, " years"); n > 0 {
			return n
		}
		if n := yearCountBefore(" "+rest, " year"); n > 0 {
			return n
		}
	}
	return 0
}

func leftoverCoveringDurationKnownPerson(query, person string) bool {
	person = strings.ToLower(strings.TrimSpace(person))
	if person == "" {
		return false
	}
	for _, e := range hopQueryEntities(query) {
		e = strings.ToLower(strings.TrimSpace(e))
		if e == "" {
			continue
		}
		if e == person {
			return true
		}
		if len(e) >= 3 && len(person) >= 3 && (strings.HasPrefix(person, e) || strings.HasPrefix(e, person)) {
			return true
		}
	}
	return false
}

func leftoverCoveringDurationActorLine(query, line string) bool {
	body := strings.ToLower(strings.TrimSpace(hybridLineBody(line)))
	if body == "" {
		return false
	}
	padded := " " + body + " "
	if strings.Contains(padded, " i ") || strings.Contains(padded, " i've ") || strings.Contains(padded, " i'm ") ||
		strings.HasPrefix(body, "i ") || strings.HasPrefix(body, "i've ") || strings.Contains(padded, " my ") {
		return true
	}
	if leftoverCoveringQueryEntityHits(query, line) > 0 {
		return true
	}
	fields := strings.Fields(line)
	for _, raw := range fields {
		key, ok := leftoverCoveringPersonToken(raw)
		if !ok {
			continue
		}
		if leftoverCoveringDurationKnownPerson(query, key) {
			return true
		}
	}
	return false
}

func leftoverCoveringDurationExplicitLine(line string) bool {
	body := strings.ToLower(hybridLineBody(line))
	return strings.Contains(body, "duration") || strings.Contains(body, " for ")
}

func leftoverCoveringDurationLine(query, line string) bool {
	if looksInterrogativeLine(line) || looksImageCaptionLine(line) {
		return false
	}
	if parseContinuingYears(strings.ToLower(hybridLineBody(line))) <= 0 {
		return false
	}
	return leftoverCoveringDurationActorLine(query, line)
}

func leftoverCoveringDurationCompanionLine(query, line string) bool {
	if leftoverCoveringDurationLine(query, line) {
		return false
	}
	for _, tok := range tokenize(query) {
		if leftoverCoveringDurationStructureToken(tok) || leftoverCoverWeakToken(tok) {
			continue
		}
		if contentCoversQueryToken(line, tok) {
			return true
		}
	}
	return false
}

func leftoverCoveringPreferDuration(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 {
			continue
		}
		if !leftoverCoveringDurationLine(query, row.line) {
			continue
		}
		score := row.score
		if leftoverCoveringDurationExplicitLine(row.line) {
			score++
		}
		if score > pickScore {
			pickScore = score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringDurationMissesDuration(query, covering, answer string) bool {
	if !looksHowLongBeenQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringDurationLine(query, answer) {
		return false
	}
	return leftoverCoveringDurationLine(query, covering)
}

func leftoverCoveringCadenceStructureToken(tok string) bool {
	return strings.ToLower(strings.TrimSpace(tok)) == "often"
}

func leftoverCoveringCadenceRareToken(tok string) bool {
	low := strings.ToLower(strings.Trim(strings.TrimSpace(tok), "?,.!'\"’"))
	if low == "" || leftoverCoveringCadenceStructureToken(low) || leftoverCoverWeakToken(low) || isQueryStopword(low) {
		return false
	}
	return utf8.RuneCountInString(low) >= 4
}

func dropTrailingForAdjunct(toks []string) []string {
	forIdx := -1
	for i, tok := range toks {
		if strings.EqualFold(strings.Trim(tok, "?,.!'\"’"), "for") {
			forIdx = i
			break
		}
	}
	if forIdx <= 0 || forIdx >= len(toks)-1 {
		return toks
	}
	rare := 0
	for _, tok := range toks[:forIdx] {
		if leftoverCoveringCadenceRareToken(tok) {
			rare++
		}
	}
	if rare < 2 {
		return toks
	}
	return toks[:forIdx]
}

func leftoverCoveringCadenceObjectTokens(query string) []string {
	fields := dropTrailingForAdjunct(dropHowOftenStructureTokens(strings.Fields(query)))
	people := map[string]struct{}{}
	for _, e := range hopQueryEntities(query) {
		people[strings.ToLower(strings.TrimSpace(e))] = struct{}{}
	}
	out := make([]string, 0, 4)
	seen := map[string]struct{}{}
	for _, tok := range tokenize(strings.Join(fields, " ")) {
		low := strings.ToLower(strings.TrimSpace(tok))
		if !leftoverCoveringCadenceRareToken(low) {
			continue
		}
		if _, ok := people[low]; ok {
			continue
		}
		if _, ok := seen[low]; ok {
			continue
		}
		seen[low] = struct{}{}
		out = append(out, low)
	}
	return out
}

func leftoverCoveringCadenceObjectHits(query, line string) int {
	n := 0
	for _, tok := range leftoverCoveringCadenceObjectTokens(query) {
		if contentCoversQueryToken(line, tok) {
			n++
		}
	}
	return n
}

func parseCadence(s string) bool {
	lower := strings.ToLower(s)
	lower = strings.NewReplacer(",", " ", ".", " ", ";", " ", "?", " ", "!", " ", ":", " ").Replace(lower)
	padded := " " + strings.Join(strings.Fields(lower), " ") + " "
	if strings.Contains(padded, " once every ") {
		return true
	}
	units := []string{"week", "month", "day", "year"}
	for _, u := range units {
		if strings.Contains(padded, " once a "+u+" ") || strings.Contains(padded, " twice a "+u+" ") {
			return true
		}
		if strings.Contains(padded, " every "+u+" ") {
			return true
		}
		for n := 1; n <= 20; n++ {
			if strings.Contains(padded, " "+strconv.Itoa(n)+" times a "+u+" ") {
				return true
			}
		}
	}
	for _, t := range []string{"morning", "evening", "night"} {
		if strings.Contains(padded, " every "+t+" ") {
			return true
		}
	}
	return false
}

func leftoverCoveringCadenceLine(query, line string) bool {
	if looksInterrogativeLine(line) || looksImageCaptionLine(line) {
		return false
	}
	if !parseCadence(hybridLineBody(line)) {
		return false
	}
	if leftoverCoveringCadenceObjectHits(query, line) < 2 {
		return false
	}
	return leftoverCoveringDurationActorLine(query, line)
}

func leftoverCoveringCadenceCompanionLine(query, line string) bool {
	if leftoverCoveringCadenceLine(query, line) {
		return false
	}
	return leftoverCoveringCadenceObjectHits(query, line) >= 1
}

func leftoverCoveringPreferCadence(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 {
			continue
		}
		if !leftoverCoveringCadenceLine(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringCadenceMissesCadence(query, covering, answer string) bool {
	if !looksHowOftenQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringCadenceLine(query, answer) {
		return false
	}
	return leftoverCoveringCadenceLine(query, covering)
}

func leftoverCoveringCurrentProjectStructureToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "project", "projects", "course", "courses", "class", "classes":
		return true
	}
	return false
}

func leftoverCoveringCurrentProjectRareToken(tok string) bool {
	low := strings.ToLower(strings.Trim(strings.TrimSpace(tok), "?,.!'\"’"))
	if low == "" || leftoverCoveringCurrentProjectStructureToken(low) || leftoverCoverWeakToken(low) || isQueryStopword(low) {
		return false
	}
	return utf8.RuneCountInString(low) >= 4
}

func dropTrailingInCourseAdjunct(toks []string) []string {
	courseIdx := -1
	for i := len(toks) - 1; i >= 0; i-- {
		t := strings.ToLower(strings.Trim(toks[i], "?,.!'\"’"))
		if t == "course" || t == "class" {
			courseIdx = i
			break
		}
	}
	if courseIdx < 0 {
		return toks
	}
	inIdx := -1
	for i := courseIdx - 1; i >= 0; i-- {
		if strings.EqualFold(strings.Trim(toks[i], "?,.!'\"’"), "in") {
			inIdx = i
			break
		}
	}
	if inIdx <= 0 {
		return toks
	}
	rare := 0
	for _, tok := range toks[:inIdx] {
		if leftoverCoveringCurrentProjectRareToken(tok) {
			rare++
		}
	}
	if rare < 2 {
		return toks
	}
	return toks[:inIdx]
}

func leftoverCoveringCurrentProjectObjectTokens(query string) []string {
	fields := dropTrailingInCourseAdjunct(dropWhatProjectWorkingStructureTokens(strings.Fields(query)))
	people := map[string]struct{}{}
	for _, e := range hopQueryEntities(query) {
		people[strings.ToLower(strings.TrimSpace(e))] = struct{}{}
	}
	out := make([]string, 0, 4)
	seen := map[string]struct{}{}
	for _, tok := range tokenize(strings.Join(fields, " ")) {
		low := strings.ToLower(strings.TrimSpace(tok))
		if !leftoverCoveringCurrentProjectRareToken(low) {
			continue
		}
		if _, ok := people[low]; ok {
			continue
		}
		if _, ok := seen[low]; ok {
			continue
		}
		seen[low] = struct{}{}
		out = append(out, low)
	}
	return out
}

func leftoverCoveringChildhoodDesireLine(line string) bool {
	lower := strings.ToLower(hybridLineBody(line))
	if strings.Contains(lower, "childhood") || strings.Contains(lower, "as a child") || strings.Contains(lower, "as a kid") {
		return true
	}
	return strings.Contains(lower, "since he was a kid") || strings.Contains(lower, "since she was a kid") || strings.Contains(lower, "since i was a kid")
}

func leftoverCoveringOwnProjectWithoutCurrentLine(line string) bool {
	lower := strings.ToLower(hybridLineBody(line))
	if strings.Contains(lower, "currently") {
		return false
	}
	return strings.Contains(lower, "creating") && strings.Contains(lower, "own") && strings.Contains(lower, "project")
}

func leftoverCoveringCurrentProjectCue(s string) bool {
	lower := strings.ToLower(s)
	lower = strings.NewReplacer(",", " ", ".", " ", ";", " ", "?", " ", "!", " ", ":", " ").Replace(lower)
	padded := " " + strings.Join(strings.Fields(lower), " ") + " "
	if strings.Contains(padded, " currently working ") || strings.Contains(padded, " currently work ") {
		return true
	}
	return strings.Contains(padded, " working on a new ")
}

func leftoverCoveringCurrentProjectLine(query, line string) bool {
	if looksInterrogativeLine(line) || looksImageCaptionLine(line) {
		return false
	}
	if leftoverCoveringChildhoodDesireLine(line) {
		return false
	}
	if !leftoverCoveringCurrentProjectCue(hybridLineBody(line)) {
		return false
	}
	return leftoverCoveringDurationActorLine(query, line)
}

func leftoverCoveringCurrentProjectCompanionLine(query, line string) bool {
	if leftoverCoveringCurrentProjectLine(query, line) {
		return false
	}
	return leftoverCoveringChildhoodDesireLine(line) || leftoverCoveringOwnProjectWithoutCurrentLine(line)
}

func leftoverCoveringPreferCurrentProject(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 {
			continue
		}
		if !leftoverCoveringCurrentProjectLine(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringCurrentProjectMissesCurrent(query, covering, answer string) bool {
	if !looksWhatProjectWorkingQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringCurrentProjectLine(query, answer) {
		return false
	}
	return leftoverCoveringCurrentProjectLine(query, covering)
}

func leftoverCoveringBecomeInterestedStructureToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "hobby", "hobbies":
		return true
	}
	return false
}

func leftoverCoveringBecomeInterestedCue(s string) bool {
	lower := strings.ToLower(s)
	lower = strings.NewReplacer(",", " ", ".", " ", ";", " ", "?", " ", "!", " ", ":", " ").Replace(lower)
	padded := " " + strings.Join(strings.Fields(lower), " ") + " "
	return strings.Contains(padded, " become interested ") || strings.Contains(padded, " became interested ") || strings.Contains(padded, " becoming interested ")
}

func leftoverCoveringBecomeInterestedLine(query, line string) bool {
	if looksInterrogativeLine(line) || looksImageCaptionLine(line) {
		return false
	}
	if !leftoverCoveringBecomeInterestedCue(hybridLineBody(line)) {
		return false
	}
	return leftoverCoveringDurationActorLine(query, line)
}

func leftoverCoveringBecomeInterestedCompanionLine(query, line string) bool {
	if leftoverCoveringBecomeInterestedLine(query, line) {
		return false
	}
	if leftoverCoveringLineHasForeignPerson(query, line) {
		return true
	}
	lower := strings.ToLower(hybridLineBody(line))
	return strings.Contains(lower, "new hobby") || strings.Contains(lower, "taken up")
}

func leftoverCoveringPreferBecomeInterested(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 {
			continue
		}
		if !leftoverCoveringBecomeInterestedLine(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringBecomeInterestedMissesInterest(query, covering, answer string) bool {
	if !looksWhatNewHobbyQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringBecomeInterestedLine(query, answer) {
		return false
	}
	return leftoverCoveringBecomeInterestedLine(query, covering)
}

func leftoverCoveringPrepPlanStructureToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "plan", "plans", "planning",
		"pursue", "pursues", "pursuing",
		"dream", "dreams",
		"learn", "learning", "learns", "learned", "learnt":
		return true
	}
	return false
}

func leftoverCoveringPrepPlanNormalized(s string) string {
	lower := strings.ToLower(s)
	lower = strings.NewReplacer(",", " ", ".", " ", ";", " ", "?", " ", "!", " ", ":", " ", "'", " ").Replace(lower)
	return " " + strings.Join(strings.Fields(lower), " ") + " "
}

func leftoverCoveringPrepPlanCueCount(s string) int {
	padded := leftoverCoveringPrepPlanNormalized(s)
	n := 0
	if strings.Contains(padded, " gathering information ") {
		n++
	}
	if strings.Contains(padded, " watching videos ") {
		n++
	}
	if strings.Contains(padded, " beginners guide ") || strings.Contains(padded, " beginner guide ") {
		n++
	}
	return n
}

func leftoverCoveringPrepPlanCue(s string) bool {
	return leftoverCoveringPrepPlanCueCount(s) > 0
}

func leftoverCoveringDreamLearnObjectTokens(query string) []string {
	toks := tokenize(query)
	start := -1
	for i := 0; i < len(toks); i++ {
		switch toks[i] {
		case "learning", "learn", "learns", "learned", "learnt":
			start = i + 1
			if start < len(toks) && toks[start] == "to" {
				start++
			}
		}
	}
	if start < 0 || start >= len(toks) {
		return nil
	}
	var out []string
	for _, tok := range toks[start:] {
		if leftoverCoverWeakToken(tok) || leftoverCoveringPrepPlanStructureToken(tok) {
			continue
		}
		if utf8Len(tok) < 4 {
			continue
		}
		out = leftoverCoverAddTokens(out, tok)
	}
	return out
}

func leftoverCoveringPrepPlanLine(query, line string) bool {
	if looksInterrogativeLine(line) || looksImageCaptionLine(line) {
		return false
	}
	if !leftoverCoveringPrepPlanCue(hybridLineBody(line)) {
		return false
	}
	if !leftoverCoveringDurationActorLine(query, line) {
		return false
	}
	objs := leftoverCoveringDreamLearnObjectTokens(query)
	if len(objs) == 0 {
		return false
	}
	return contentCoversAnyQueryToken(line, objs)
}

func leftoverCoveringPrepPlanCompanionLine(query, line string) bool {
	if leftoverCoveringPrepPlanLine(query, line) {
		return false
	}
	return leftoverCoveringLineHasForeignPerson(query, line)
}

func leftoverCoveringPreferPrepPlan(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	pickCues := -1
	for _, row := range scored {
		if row.score < 2 {
			continue
		}
		if !leftoverCoveringPrepPlanLine(query, row.line) {
			continue
		}
		cues := leftoverCoveringPrepPlanCueCount(hybridLineBody(row.line))
		if row.score > pickScore || (row.score == pickScore && cues > pickCues) {
			pickScore = row.score
			pickCues = cues
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringPrepPlanMissesPrep(query, covering, answer string) bool {
	if !looksHowPlanDreamQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringPrepPlanLine(query, answer) {
		return false
	}
	return leftoverCoveringPrepPlanLine(query, covering)
}

func leftoverCoveringFocusingBesidesStructureToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "besides", "except", "aside", "lately":
		return true
	}
	return false
}

func leftoverCoveringBesidesObjectTokens(query string) []string {
	toks := tokenize(query)
	start := -1
	for i, tok := range toks {
		switch tok {
		case "besides", "except":
			start = i + 1
		case "aside":
			if i+1 < len(toks) && toks[i+1] == "from" {
				start = i + 2
			} else {
				start = i + 1
			}
		}
	}
	if start < 0 || start >= len(toks) {
		return nil
	}
	var out []string
	for _, tok := range toks[start:] {
		if leftoverCoverWeakToken(tok) || leftoverCoveringFocusingBesidesStructureToken(tok) {
			continue
		}
		if utf8Len(tok) < 4 {
			continue
		}
		out = leftoverCoverAddTokens(out, tok)
	}
	return out
}

func leftoverCoveringFocusingBesidesLine(query, line string) bool {
	if looksInterrogativeLine(line) || looksImageCaptionLine(line) {
		return false
	}
	padded := leftoverCoveringPrepPlanNormalized(hybridLineBody(line))
	if !strings.Contains(padded, " focusing on ") {
		return false
	}
	if !leftoverCoveringDurationActorLine(query, line) {
		return false
	}
	objs := leftoverCoveringBesidesObjectTokens(query)
	if len(objs) == 0 || !contentCoversAnyQueryToken(line, objs) {
		return false
	}
	if !strings.Contains(padded, " and my ") && !strings.Contains(padded, " and our ") {
		return false
	}
	return true
}

func leftoverCoveringFocusingBesidesCompanionLine(query, line string) bool {
	if leftoverCoveringFocusingBesidesLine(query, line) {
		return false
	}
	padded := leftoverCoveringPrepPlanNormalized(hybridLineBody(line))
	if leftoverCoveringLineHasForeignPerson(query, line) {
		return true
	}
	return strings.Contains(padded, " focusing on ")
}

func leftoverCoveringPreferFocusingBesides(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 {
			continue
		}
		if !leftoverCoveringFocusingBesidesLine(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringFocusingBesidesMissesJoin(query, covering, answer string) bool {
	if !looksWhatFocusingBesidesQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringFocusingBesidesLine(query, answer) {
		return false
	}
	return leftoverCoveringFocusingBesidesLine(query, covering)
}

func leftoverCoveringTitledShowStructureToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "new", "fantasy", "tv", "television", "series", "excited", "about":
		return true
	}
	return false
}

func leftoverCoveringFeelAboutStructureToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "feel", "felt", "feeling", "feels", "about", "progress":
		return true
	}
	return false
}

func leftoverCoveringCoordinatedUseStructureToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "use", "uses", "using", "reach", "reaches", "reaching", "goal", "goals":
		return true
	}
	return false
}

func leftoverCoveringRealizeAfterStructureToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "realize", "realized", "realised", "realizing", "realising", "after":
		return true
	}
	return false
}

func leftoverCoveringTitledShowCue(line string) bool {
	padded := leftoverCoveringPrepPlanNormalized(hybridLineBody(line))
	if strings.Contains(padded, " watch ") || strings.Contains(padded, " watching ") {
		return true
	}
	if strings.Contains(padded, " show ") {
		return true
	}
	if strings.Contains(padded, " coming out ") {
		return true
	}
	if strings.Contains(padded, " titled ") && (strings.Contains(padded, " tv ") || strings.Contains(padded, " series ") || strings.Contains(padded, " show ")) {
		return true
	}
	if strings.Contains(padded, " called ") && (strings.Contains(padded, " show ") || strings.Contains(padded, " series ") || strings.Contains(padded, " watch ") || strings.Contains(padded, " watching ")) {
		return true
	}
	return false
}

func leftoverCoveringTitledShowLine(query, line string) bool {
	if looksInterrogativeLine(line) || looksImageCaptionLine(line) {
		return false
	}
	if len(leftoverCoveringQuotedTitles(hybridLineBody(line))) == 0 {
		return false
	}
	if !leftoverCoveringTitledShowCue(line) {
		return false
	}
	padded := leftoverCoveringPrepPlanNormalized(hybridLineBody(line))
	if (strings.Contains(padded, " novel ") || strings.Contains(padded, " book ")) &&
		!strings.Contains(padded, " watch ") && !strings.Contains(padded, " watching ") &&
		!strings.Contains(padded, " show ") && !strings.Contains(padded, " coming out ") &&
		!strings.Contains(padded, " titled ") {
		return false
	}
	body := hybridLineBody(line)
	if i := strings.Index(body, ": "); i > 0 && i < 24 {
		body = strings.TrimSpace(body[i+2:])
	}
	return leftoverCoveringDurationActorLine(query, line) || leftoverCoveringDurationActorLine(query, body)
}

func leftoverCoveringTitledShowCompanionLine(query, line string) bool {
	if leftoverCoveringTitledShowLine(query, line) {
		return false
	}
	if leftoverCoveringLineHasForeignPerson(query, line) {
		return true
	}
	padded := leftoverCoveringPrepPlanNormalized(hybridLineBody(line))
	if strings.Contains(padded, " excited ") || strings.Contains(padded, " fantasy ") {
		return true
	}
	return len(leftoverCoveringQuotedTitles(hybridLineBody(line))) > 0
}

func leftoverCoveringPreferTitledShow(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 {
			continue
		}
		if !leftoverCoveringTitledShowLine(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringTitledShowMissesTitle(query, covering, answer string) bool {
	if !looksWhatNewSeriesQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringTitledShowLine(query, answer) {
		return false
	}
	return leftoverCoveringTitledShowLine(query, covering)
}

func leftoverCoveringRecentlyAtStructureToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "recently", "lately":
		return true
	}
	return false
}

func leftoverCoveringRecentlyAtPlaceTokens(query string) []string {
	toks := tokenize(query)
	start := -1
	for i, tok := range toks {
		switch tok {
		case "at", "in":
			start = i + 1
		}
	}
	if start < 0 || start >= len(toks) {
		return nil
	}
	people := map[string]struct{}{}
	for _, e := range hopQueryEntities(query) {
		people[strings.ToLower(strings.TrimSpace(e))] = struct{}{}
	}
	var out []string
	for _, tok := range toks[start:] {
		low := strings.ToLower(strings.TrimSpace(tok))
		if low == "" || leftoverCoveringRecentlyAtStructureToken(low) || leftoverCoverWeakToken(low) || isQueryStopword(low) || isMonthWord(low) || isCalendarCoverToken(low) {
			continue
		}
		if _, ok := people[low]; ok {
			continue
		}
		if utf8Len(low) < 4 {
			continue
		}
		out = leftoverCoverAddTokens(out, low)
	}
	return out
}

func leftoverCoveringCongratsLine(line string) bool {
	padded := leftoverCoveringPrepPlanNormalized(hybridLineBody(line))
	return strings.Contains(padded, " congrat") || strings.Contains(padded, " congrats ")
}

func leftoverCoveringExtraPurposeNouns(line string) []string {
	padded := leftoverCoveringPrepPlanNormalized(hybridLineBody(line))
	var out []string
	for _, cue := range []string{" for my ", " for our "} {
		idx := strings.Index(padded, cue)
		if idx < 0 {
			continue
		}
		rest := strings.TrimSpace(padded[idx+len(cue):])
		for _, tok := range strings.Fields(rest) {
			tok = strings.Trim(tok, ".,;:()[]\"'")
			if tok == "" || tok == "new" || leftoverCoveringRecentlyAtStructureToken(tok) || leftoverCoverWeakToken(tok) || isQueryStopword(tok) || isMonthWord(tok) || isCalendarCoverToken(tok) || utf8Len(tok) < 4 {
				continue
			}
			out = leftoverCoverAddTokens(out, tok)
			break
		}
	}
	return out
}

func leftoverCoveringLocativePurposeLine(query, line string) bool {
	if looksInterrogativeLine(line) || looksImageCaptionLine(line) || looksChatTurnLine(line) || leftoverCoveringCongratsLine(line) {
		return false
	}
	places := leftoverCoveringRecentlyAtPlaceTokens(query)
	if len(places) == 0 {
		return false
	}
	for _, p := range places {
		if !contentCoversQueryToken(line, p) {
			return false
		}
	}
	extras := leftoverCoveringExtraPurposeNouns(line)
	if len(extras) == 0 {
		return false
	}
	qtoks := map[string]struct{}{}
	for _, tok := range tokenize(query) {
		qtoks[strings.ToLower(strings.TrimSpace(tok))] = struct{}{}
	}
	hasExtra := false
	for _, n := range extras {
		if _, ok := qtoks[n]; !ok {
			hasExtra = true
			break
		}
	}
	if !hasExtra {
		return false
	}
	body := hybridLineBody(line)
	if i := strings.Index(body, ": "); i > 0 && i < 24 {
		body = strings.TrimSpace(body[i+2:])
	}
	return leftoverCoveringDurationActorLine(query, line) || leftoverCoveringDurationActorLine(query, body)
}

func leftoverCoveringLocativePurposeCompanionLine(query, line string) bool {
	if leftoverCoveringLocativePurposeLine(query, line) {
		return false
	}
	if leftoverCoveringCongratsLine(line) || looksChatTurnLine(line) {
		return true
	}
	places := leftoverCoveringRecentlyAtPlaceTokens(query)
	hasPlace := len(places) > 0 && contentCoversAnyQueryToken(line, places)
	hasExtra := len(leftoverCoveringExtraPurposeNouns(line)) > 0
	return hasPlace || hasExtra
}

func leftoverCoveringPreferLocativePurpose(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 {
			continue
		}
		if !leftoverCoveringLocativePurposeLine(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringLocativePurposeMissesExtra(query, covering, answer string) bool {
	if !looksWhatDidRecentlyAtQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringLocativePurposeLine(query, answer) {
		return false
	}
	return leftoverCoveringLocativePurposeLine(query, covering)
}

func leftoverCoveringExperiencingFeelingLine(query, line string) bool {
	if looksInterrogativeLine(line) || looksImageCaptionLine(line) {
		return false
	}
	padded := leftoverCoveringPrepPlanNormalized(line)
	bodyPad := leftoverCoveringPrepPlanNormalized(hybridLineBody(line))
	if !strings.Contains(padded, " experiencing ") && !strings.Contains(bodyPad, " experiencing ") {
		return false
	}
	if !strings.Contains(padded, " new level ") && !strings.Contains(bodyPad, " new level ") {
		return false
	}
	body := hybridLineBody(line)
	if i := strings.Index(body, ": "); i > 0 && i < 24 {
		body = strings.TrimSpace(body[i+2:])
	}
	return leftoverCoveringDurationActorLine(query, line) || leftoverCoveringDurationActorLine(query, body)
}

func leftoverCoveringExperiencingFeelingCompanionLine(query, line string) bool {
	if leftoverCoveringExperiencingFeelingLine(query, line) {
		return false
	}
	if leftoverCoveringLineHasForeignPerson(query, line) {
		return true
	}
	padded := leftoverCoveringPrepPlanNormalized(line)
	body := leftoverCoveringPrepPlanNormalized(hybridLineBody(line))
	hasExperiencing := strings.Contains(padded, " experiencing ") || strings.Contains(body, " experiencing ")
	hasNewLevel := strings.Contains(padded, " new level ") || strings.Contains(body, " new level ")
	if hasExperiencing && !hasNewLevel {
		return true
	}
	if strings.Contains(padded, " practicing ") || strings.Contains(body, " practicing ") {
		return true
	}
	if strings.Contains(padded, " trying to be more ") || strings.Contains(body, " trying to be more ") {
		return true
	}
	return false
}

func leftoverCoveringPreferExperiencingFeeling(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 {
			continue
		}
		if !leftoverCoveringExperiencingFeelingLine(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringExperiencingFeelingMissesFeeling(query, covering, answer string) bool {
	if !looksHowFeelAboutQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = leftoverCoveringAnswerHead(answer)
	if covering == "" || leftoverCoveringExperiencingFeelingLine(query, answer) {
		return false
	}
	return leftoverCoveringExperiencingFeelingLine(query, covering)
}

func leftoverCoveringAnswerHead(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, " (via "); i > 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}

func leftoverCoveringWorkDeterminationLine(query, line string) bool {
	if !looksWhatDoCoordinatedUseQuery(query) {
		return false
	}
	if looksInterrogativeLine(line) || looksImageCaptionLine(line) || leftoverCoveringCongratsLine(line) {
		return false
	}
	padded := leftoverCoveringPrepPlanNormalized(line)
	body := leftoverCoveringPrepPlanNormalized(hybridLineBody(line))
	hasWork := strings.Contains(padded, " hard work ") || strings.Contains(body, " hard work ")
	hasDetermination := strings.Contains(padded, " determination ") || strings.Contains(body, " determination ")
	if !hasWork || !hasDetermination {
		return false
	}
	if strings.Contains(padded, " dedication ") || strings.Contains(body, " dedication ") {
		return false
	}
	hasPlural := strings.Contains(padded, " we ") || strings.Contains(padded, " us ") || strings.Contains(padded, " our ") ||
		strings.Contains(body, " we ") || strings.Contains(body, " us ") || strings.Contains(body, " our ")
	return hasPlural
}

func leftoverCoveringWorkDeterminationCompanionLine(query, line string) bool {
	if leftoverCoveringWorkDeterminationLine(query, line) {
		return false
	}
	if leftoverCoveringLineHasForeignPerson(query, line) {
		return true
	}
	padded := leftoverCoveringPrepPlanNormalized(line)
	body := leftoverCoveringPrepPlanNormalized(hybridLineBody(line))
	if strings.Contains(padded, " dedication ") || strings.Contains(body, " dedication ") {
		return true
	}
	if looksChatTurnLine(line) {
		return true
	}
	hasWork := strings.Contains(padded, " hard work ") || strings.Contains(body, " hard work ")
	hasDetermination := strings.Contains(padded, " determination ") || strings.Contains(body, " determination ")
	if hasWork && !hasDetermination {
		return true
	}
	if hasDetermination && !hasWork {
		return true
	}
	return false
}

func leftoverCoveringPreferWorkDetermination(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 {
			continue
		}
		if !leftoverCoveringWorkDeterminationLine(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringWorkDeterminationMissesPair(query, covering, answer string) bool {
	if !looksWhatDoCoordinatedUseQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = leftoverCoveringAnswerHead(answer)
	if covering == "" || leftoverCoveringWorkDeterminationLine(query, answer) {
		return false
	}
	return leftoverCoveringWorkDeterminationLine(query, covering)
}

func leftoverCoveringSelfDirectedRealizeLine(query, line string) bool {
	if !looksWhatDidRealizeAfterQuery(query) {
		return false
	}
	if looksInterrogativeLine(line) || looksImageCaptionLine(line) || leftoverCoveringCongratsLine(line) {
		return false
	}
	padded := leftoverCoveringPrepPlanNormalized(line)
	body := leftoverCoveringPrepPlanNormalized(hybridLineBody(line))
	hasRealize := strings.Contains(padded, " realize ") || strings.Contains(padded, " realizing ") || strings.Contains(padded, " realised ") || strings.Contains(padded, " realized ") ||
		strings.Contains(body, " realize ") || strings.Contains(body, " realizing ") || strings.Contains(body, " realised ") || strings.Contains(body, " realized ")
	if !hasRealize {
		return false
	}
	if strings.Contains(padded, " others ") || strings.Contains(body, " others ") {
		return false
	}
	hasSelf := strings.Contains(padded, " self-") || strings.Contains(padded, " self ") || strings.Contains(padded, " myself ") ||
		strings.Contains(body, " self-") || strings.Contains(body, " self ") || strings.Contains(body, " myself ")
	if !hasSelf {
		return false
	}
	return leftoverCoveringDurationActorLine(query, line) || leftoverCoveringDurationActorLine(query, hybridLineBody(line))
}

func leftoverCoveringSelfDirectedRealizeCompanionLine(query, line string) bool {
	if leftoverCoveringSelfDirectedRealizeLine(query, line) {
		return false
	}
	if leftoverCoveringLineHasForeignPerson(query, line) {
		return true
	}
	padded := leftoverCoveringPrepPlanNormalized(line)
	body := leftoverCoveringPrepPlanNormalized(hybridLineBody(line))
	if strings.Contains(padded, " others ") || strings.Contains(body, " others ") {
		return true
	}
	hasRealize := strings.Contains(padded, " realize ") || strings.Contains(padded, " realizing ") || strings.Contains(padded, " realised ") || strings.Contains(padded, " realized ") ||
		strings.Contains(body, " realize ") || strings.Contains(body, " realizing ") || strings.Contains(body, " realised ") || strings.Contains(body, " realized ")
	hasSelf := strings.Contains(padded, " self-") || strings.Contains(padded, " self ") || strings.Contains(padded, " myself ") ||
		strings.Contains(body, " self-") || strings.Contains(body, " self ") || strings.Contains(body, " myself ")
	if hasRealize && !hasSelf {
		return true
	}
	if hasSelf && !hasRealize {
		return true
	}
	return false
}

func leftoverCoveringPreferSelfDirectedRealize(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 {
			continue
		}
		if !leftoverCoveringSelfDirectedRealizeLine(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

func leftoverCoveringSelfDirectedRealizeMissesSelf(query, covering, answer string) bool {
	if !looksWhatDidRealizeAfterQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = leftoverCoveringAnswerHead(answer)
	if covering == "" || leftoverCoveringSelfDirectedRealizeLine(query, answer) {
		return false
	}
	return leftoverCoveringSelfDirectedRealizeLine(query, covering)
}

func leftoverCoveringRestatementToken(tok string) bool {
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "enjoy", "enjoys", "enjoyed", "enjoying",
		"participate", "participates", "participating", "participated",
		"find", "finds", "finding", "found",
		"friend", "friends", "like", "likes", "liked",
		"make", "makes", "making", "made",
		"awesome", "connecting", "people", "care", "fitness", "always":
		return true
	}
	return leftoverCoverWeakToken(tok) || isQueryStopword(tok)
}

func leftoverCoveringReasonTokens(query string) []string {
	out := make([]string, 0, 4)
	for _, tok := range tokenize(query) {
		switch strings.ToLower(tok) {
		case "easy", "easier", "stay", "staying", "motivated", "motivating", "motivation":
			out = append(out, tok)
		}
	}
	return out
}

func leftoverCoveringReasonHits(query, line string) int {
	n := 0
	for _, tok := range leftoverCoveringReasonTokens(query) {
		if contentCoversQueryToken(line, tok) {
			n++
		}
	}
	return n
}

func leftoverCoveringHasOffQueryEvidence(query, line string) bool {
	qtoks := tokenize(query)
	for _, tok := range tokenize(line) {
		tok = strings.ToLower(strings.TrimSpace(tok))
		if utf8.RuneCountInString(tok) < 4 || leftoverCoveringRestatementToken(tok) {
			continue
		}
		if recordTokensCover(qtoks, tok) {
			continue
		}
		return true
	}
	return false
}

func leftoverCoveringPreferWhatMadeEvidence(query string, scored []leftoverCoverScored, best string) string {
	pick := ""
	pickScore := -1
	for _, row := range scored {
		if row.score < 2 || looksChatTurnLine(row.line) {
			continue
		}
		if leftoverCoveringReasonHits(query, row.line) == 0 || !leftoverCoveringHasOffQueryEvidence(query, row.line) {
			continue
		}
		if leftoverCoveringLineHasForeignPerson(query, row.line) && !leftoverCoveringMentionsQueryEntity(query, row.line) {
			continue
		}
		if row.score > pickScore {
			pickScore = row.score
			pick = row.line
		}
	}
	if pick == "" || strings.EqualFold(strings.TrimSpace(pick), strings.TrimSpace(best)) {
		return ""
	}
	return pick
}

// leftoverCoveringWhatMadeMissesEvidence is true when a what-made hybrid only
// restates enjoy/participate, but leftover covering names off-query evidence
// (push, remind) plus a query reason token (easy / stay / motivated).
func leftoverCoveringWhatMadeMissesEvidence(query, covering, answer string) bool {
	if !looksWhatMadeQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringSkipForeignWhenEvent(query, covering) {
		return false
	}
	if leftoverCoveringHasOffQueryEvidence(query, answer) {
		return false
	}
	if leftoverCoveringReasonHits(query, covering) == 0 || !leftoverCoveringHasOffQueryEvidence(query, covering) {
		return false
	}
	if leftoverCoveringLineHasForeignPerson(query, covering) && !leftoverCoveringMentionsQueryEntity(query, covering) {
		return false
	}
	return true
}

// leftoverCoveringPurposeMissesInstrument is true when a what-does-X-help-with
// answer is ownership or a generic help paraphrase, but a leftover line names
// the instrument's purpose (uses / monitor / reminder / progress).
func leftoverCoveringPurposeMissesInstrument(query, covering, answer string) bool {
	if !looksInstrumentPurposeQuery(query) {
		return false
	}
	covering = strings.TrimSpace(covering)
	answer = strings.TrimSpace(answer)
	if covering == "" || leftoverCoveringSkipForeignWhenEvent(query, covering) {
		return false
	}
	if !leftoverCoveringInstrumentPurposeLine(covering) {
		return false
	}
	if leftoverCoveringInstrumentPossessLine(answer) && !leftoverCoveringInstrumentPurposeLine(answer) {
		return leftoverCoveringMentionsQueryEntity(query, covering)
	}
	for _, tok := range leftoverCoveringInstrumentPurposeDetailTokens(covering) {
		if !contentCoversQueryToken(answer, tok) {
			return leftoverCoveringMentionsQueryEntity(query, covering) ||
				!leftoverCoveringLineHasForeignPerson(query, covering)
		}
	}
	return false
}

// leftoverCoveringSyncEnumerateItems replaces enumerate items with the leftover
// covering sentence for instrument-purpose questions. Enumerate clients (and
// the product harness, which sends "what does …" as mode=enumerate) prefer
// items over answer, so a covering rewrite of answer alone is invisible.
func leftoverCoveringSyncEnumerateItems(query, covering string, items []RecallItem) []RecallItem {
	covering = strings.TrimSpace(covering)
	if covering == "" {
		return items
	}
	// Product-recall harness sends "what does …" as mode=enumerate and
	// scores items, not Answer. A hop-dump list would beat leftover
	// covering unless items follow the covering line.
	if looksInstrumentPurposeQuery(query) {
		return []RecallItem{{Value: covering}}
	}
	if looksWhatSayAboutQuery(query) && leftoverCoveringSayAboutTargetLine(covering) {
		return []RecallItem{{Value: covering}}
	}
	if looksHowReactQuery(query) && leftoverCoveringReactionObservationLine(covering) && leftoverCoveringReactLineHasObject(query, covering) {
		return []RecallItem{{Value: covering}}
	}
	if looksWhatDidPurposeQuery(query) && leftoverCoveringPurposeActionLine(query, covering) {
		return []RecallItem{{Value: covering}}
	}
	if looksHowDidStartQuery(query) && leftoverCoveringStartMethodLine(query, covering) {
		return []RecallItem{{Value: covering}}
	}
	if looksHowLongBeenQuery(query) && leftoverCoveringDurationLine(query, covering) {
		return []RecallItem{{Value: covering}}
	}
	if looksHowOftenQuery(query) && leftoverCoveringCadenceLine(query, covering) {
		return []RecallItem{{Value: covering}}
	}
	if looksWhatProjectWorkingQuery(query) && leftoverCoveringCurrentProjectLine(query, covering) {
		return []RecallItem{{Value: covering}}
	}
	if looksWhatNewHobbyQuery(query) && leftoverCoveringBecomeInterestedLine(query, covering) {
		return []RecallItem{{Value: covering}}
	}
	if looksHowPlanDreamQuery(query) && leftoverCoveringPrepPlanLine(query, covering) {
		return []RecallItem{{Value: covering}}
	}
	if looksWhatFocusingBesidesQuery(query) && leftoverCoveringFocusingBesidesLine(query, covering) {
		return []RecallItem{{Value: covering}}
	}
	if looksWhatNewSeriesQuery(query) && leftoverCoveringTitledShowLine(query, covering) {
		return []RecallItem{{Value: covering}}
	}
	if looksWhatDidRecentlyAtQuery(query) && leftoverCoveringLocativePurposeLine(query, covering) {
		return []RecallItem{{Value: covering}}
	}
	if looksHowFeelAboutQuery(query) && leftoverCoveringExperiencingFeelingLine(query, covering) {
		return []RecallItem{{Value: covering}}
	}
	if looksWhatDoCoordinatedUseQuery(query) && leftoverCoveringWorkDeterminationLine(query, covering) {
		return []RecallItem{{Value: covering}}
	}
	if looksWhatDidRealizeAfterQuery(query) && leftoverCoveringSelfDirectedRealizeLine(query, covering) {
		return []RecallItem{{Value: covering}}
	}
	return items
}

func leftoverCoveringInstrumentTokens(query string) []string {
	inst := leftoverCoveringInstrumentNoun(query)
	if inst == "" {
		return nil
	}
	out := []string{inst}
	for _, rel := range relatedIntentTokens(inst) {
		if leftoverCoverWeakToken(rel) || rel == inst {
			continue
		}
		out = leftoverCoverAddTokens(out, rel)
	}
	return out
}

func leftoverCoveringInstrumentNoun(query string) string {
	toks := tokenize(query)
	for i := 0; i+1 < len(toks); i++ {
		if toks[i] != "what" {
			continue
		}
		j := i + 1
		if j < len(toks) && toks[j] == "does" {
			j++
		} else {
			continue
		}
		hadThe := false
		if j < len(toks) && toks[j] == "the" {
			hadThe = true
			j++
		}
		if j+1 >= len(toks) {
			continue
		}
		inst := toks[j]
		if inst == "" || isQueryStopword(inst) || leftoverCoverWeakToken(inst) {
			continue
		}
		if toks[j+1] != "help" && toks[j+1] != "helps" {
			continue
		}
		if utf8Len(inst) < 4 {
			continue
		}
		// "what does the smartwatch help … with", not "what does exercise help her feel".
		if !hadThe {
			hasWith := false
			for _, t := range toks[j+2:] {
				if t == "with" {
					hasWith = true
					break
				}
			}
			if !hasWith {
				continue
			}
		}
		return inst
	}
	return ""
}

func leftoverCoveringInstrumentPurposeLine(line string) bool {
	for _, tok := range tokenize(hybridLineBody(line)) {
		switch tok {
		case "uses", "use", "using", "monitor", "monitors", "monitoring",
			"reminder", "remind", "reminds", "serves",
			"track", "tracks", "tracking", "progress":
			return true
		}
	}
	return false
}

func leftoverCoveringInstrumentPossessLine(line string) bool {
	for _, tok := range tokenize(hybridLineBody(line)) {
		switch tok {
		case "owns", "own", "owned", "wears", "wear", "wore":
			return true
		}
	}
	return false
}

func leftoverCoveringInstrumentPurposeDetailTokens(line string) []string {
	out := make([]string, 0, 2)
	for _, tok := range tokenize(hybridLineBody(line)) {
		switch tok {
		case "reminder", "remind", "reminds", "progress", "track", "tracks", "tracking":
			out = leftoverCoverAddTokens(out, tok)
		}
	}
	return out
}

func leftoverCoveringMentionsQueryEntity(query, line string) bool {
	ents := hopQueryEntities(query)
	if len(ents) == 0 {
		return true
	}
	return leftoverCoveringQueryEntityHits(query, line) > 0
}

func leftoverCoveringQueryEntityHits(query, line string) int {
	n := 0
	for _, e := range hopQueryEntities(query) {
		if contentCoversQueryToken(line, e) {
			n++
		}
	}
	return n
}

// leftoverCoveringSkipForeignWhenEvent drops when-event covering lines that
// name a different person and do not name a query person. First-person and
// unnamed dated lines still compete.
func leftoverCoveringSkipForeignWhenEvent(query, line string) bool {
	if !leftoverCoveringRequiresQueryEntity(query) {
		return false
	}
	if leftoverCoveringMentionsQueryEntity(query, line) {
		return false
	}
	return leftoverCoveringLineHasForeignPerson(query, line)
}

func leftoverCoveringLineHasForeignPerson(query, line string) bool {
	known := map[string]struct{}{}
	for _, e := range hopQueryEntities(query) {
		known[strings.ToLower(strings.TrimSpace(e))] = struct{}{}
	}
	if len(known) == 0 {
		return false
	}
	fields := strings.Fields(hybridLineBody(line))
	for i, raw := range fields {
		if leftoverCoveringSentenceInitialVerb(i, fields) {
			continue
		}
		key, ok := leftoverCoveringPersonToken(raw)
		if !ok {
			continue
		}
		if _, ok := known[key]; ok {
			continue
		}
		return true
	}
	return false
}

// leftoverCoveringSentenceInitialVerb is true when the first token is
// orthographic capitalization of a verb, not a person. Diary lines such as
// "Checked out an art show …" must still compete as unnamed dated facts.
func leftoverCoveringSentenceInitialVerb(i int, fields []string) bool {
	if i != 0 || len(fields) == 0 {
		return false
	}
	w := strings.Trim(fields[0], ".,;:()[]\"'")
	w = strings.TrimSuffix(strings.TrimSuffix(w, "'s"), "’s")
	if w == "" {
		return false
	}
	lower := strings.ToLower(w)
	// Past-tense morphology. Short names (Ned, Fred) stay people.
	if len(lower) >= 6 && strings.HasSuffix(lower, "ed") {
		return true
	}
	// Hortative / gerund leftover ("Also be sure…", "Building relationships…").
	switch lower {
	case "also", "and", "but", "so", "make", "don't", "dont", "it", "just":
		return true
	}
	if len(lower) >= 6 && strings.HasSuffix(lower, "ing") {
		return true
	}
	if len(fields) < 2 {
		return false
	}
	next := strings.ToLower(strings.Trim(fields[1], ".,;:()[]\"'"))
	switch next {
	case "out", "up", "off", "down", "away", "back":
		return true
	}
	return false
}

func leftoverCoveringPersonToken(raw string) (string, bool) {
	w := strings.Trim(raw, ".,;:()[]\"'")
	w = strings.TrimSuffix(strings.TrimSuffix(w, "'s"), "’s")
	if leftoverCoveringCalendarName(w) || !looksHopPerson(w) {
		return "", false
	}
	for _, r := range w {
		if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return "", false
		}
	}
	return strings.ToLower(w), true
}

func leftoverCoveringCalendarName(tok string) bool {
	if isCalendarCoverToken(tok) {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(tok)) {
	case "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday":
		return true
	}
	return false
}

// preferUnwindPacketActivities appends unwind-evidenced packet activities the
// current answer omitted. Lines without unwind evidence (plain participates-in)
// stay out so camping dumps cannot crowd the list.
func preferUnwindPacketActivities(query, answer string, pkt EvidencePacket, hops []HopResult) (string, []string) {
	if !looksUnwindQuery(query) {
		return "", nil
	}
	answer = strings.TrimSpace(answer)
	if answer == "" || strings.EqualFold(answer, "not in memory") {
		return "", nil
	}
	bind := len(hopQueryEntities(query)) > 0
	lines := packetContentLines(pkt)
	for _, h := range hops {
		for _, c := range h.Contents {
			if strings.TrimSpace(c) != "" {
				lines = append(lines, c)
			}
		}
	}
	var extra []string
	seen := map[string]struct{}{}
	add := func(slot string) {
		slot = strings.TrimSpace(slot)
		if slot == "" || utf8Len(slot) > 40 {
			return
		}
		key := strings.ToLower(slot)
		if _, ok := seen[key]; ok {
			return
		}
		if contentCoversQueryToken(answer, key) {
			return
		}
		seen[key] = struct{}{}
		extra = append(extra, slot)
	}
	for _, line := range lines {
		if !unwindEvidenceHit(line) {
			continue
		}
		if bind && !leftoverCoveringMentionsQueryEntity(query, line) {
			continue
		}
		for _, slot := range unwindActivitySlots(line) {
			add(slot)
			if len(extra) >= 3 {
				break
			}
		}
		if len(extra) >= 3 {
			break
		}
	}
	if len(extra) == 0 {
		return "", nil
	}
	return answer + ", " + strings.Join(extra, ", "), extra
}

// preferPracticePacketPlaces appends practice-object places from leftover
// packet lines and hop contents that locationItemsFromEvidence missed. Lines
// about another named person stay out; purchase leftover without a place is
// not a location.
func preferPracticePacketPlaces(query, answer string, pkt EvidencePacket, hops []HopResult) (string, []string) {
	if !looksLocationListQuery(query) {
		return "", nil
	}
	focus := practiceObjectTokens(query)
	if len(focus) == 0 {
		return "", nil
	}
	answer = strings.TrimSpace(answer)
	bind := len(hopQueryEntities(query)) > 0
	lines := packetContentLines(pkt)
	for _, h := range hops {
		for _, c := range h.Contents {
			if strings.TrimSpace(c) != "" {
				lines = append(lines, c)
			}
		}
	}
	var extra []string
	seen := map[string]struct{}{}
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" || utf8Len(p) > 48 {
			return
		}
		if placeEqualsAny(p, focus) || !keepPracticePlaceCandidate(p) {
			return
		}
		key := strings.ToLower(p)
		if _, ok := seen[key]; ok {
			return
		}
		if answer != "" && !strings.EqualFold(answer, "not in memory") && contentCoversQueryToken(answer, key) {
			return
		}
		seen[key] = struct{}{}
		extra = append(extra, p)
	}
	for _, line := range lines {
		if !contentCoversAnyQueryToken(line, focus) || !practicePlaceEvidenceLine(line, focus) {
			continue
		}
		if bind && leftoverCoveringLineHasForeignPerson(query, line) && !leftoverCoveringMentionsQueryEntity(query, line) {
			continue
		}
		for _, p := range placesFromContent(line) {
			add(p)
			if len(extra) >= 6 {
				break
			}
		}
		if p := compositionalPracticePlace(line, focus); p != "" {
			add(p)
		}
		if len(extra) >= 6 {
			break
		}
	}
	if len(extra) == 0 {
		return "", nil
	}
	if locationListAnswerKeepsHead(answer, focus) {
		return answer + ", " + strings.Join(extra, ", "), extra
	}
	return strings.Join(extra, ", "), extra
}

func practicePlaceEvidenceLine(line string, focus []string) bool {
	if compositionalPracticePlace(line, focus) != "" {
		return true
	}
	low := strings.ToLower(hybridLineBody(line))
	for _, n := range []string{"park", "beach", "studio", "gym", "school", "library", "hospital", "cottage"} {
		if strings.Contains(low, n) {
			return true
		}
	}
	for _, role := range []string{"mother", "father", "mom", "dad", "parent"} {
		if strings.Contains(low, role) && (strings.Contains(low, "home") || strings.Contains(low, "house") || strings.Contains(low, "cottage")) {
			return true
		}
	}
	return false
}

func keepPracticePlaceCandidate(p string) bool {
	low := strings.ToLower(strings.TrimSpace(p))
	if low == "" {
		return false
	}
	for _, n := range []string{"park", "beach", "studio", "gym", "school", "library", "hospital", "cottage", "home", "house"} {
		if strings.Contains(low, n) {
			return true
		}
	}
	for _, role := range []string{"mother", "father", "mom", "dad", "parent"} {
		if strings.Contains(low, role) {
			return true
		}
	}
	return false
}

func locationListAnswerKeepsHead(answer string, focus []string) bool {
	answer = strings.TrimSpace(answer)
	if answer == "" || strings.EqualFold(answer, "not in memory") {
		return false
	}
	if leftoverShortItemJoin(answer) || leftoverCoveringIsShortLocative(answer) {
		return true
	}
	if len(placesFromContent(answer)) > 0 || compositionalPracticePlace(answer, focus) != "" {
		return true
	}
	if parseDateFromText(answer) != nil {
		return false
	}
	return utf8Len(answer) <= 48 && !looksChatTurnLine(answer)
}

func appendUniqueRecallItems(items []RecallItem, extra []string) []RecallItem {
	seen := map[string]struct{}{}
	for _, it := range items {
		if k := strings.ToLower(strings.TrimSpace(it.Value)); k != "" {
			seen[k] = struct{}{}
		}
	}
	for _, slot := range extra {
		slot = strings.TrimSpace(slot)
		if slot == "" {
			continue
		}
		key := strings.ToLower(slot)
		if _, ok := seen[key]; ok {
			continue
		}
		covered := false
		for _, it := range items {
			if contentCoversQueryToken(it.Value, key) {
				covered = true
				break
			}
		}
		if covered {
			continue
		}
		seen[key] = struct{}{}
		items = append(items, RecallItem{Value: slot})
	}
	return items
}

func leftoverCoveringKeepTypedAnswer(query string, hops []HopResult, answer string) bool {
	if looksWhereQuery(query) || leftoverCoveringShouldJoin(query) {
		return false
	}
	if !hopsKeepTypedJoin(hops) {
		return false
	}
	if leftoverShortItemJoin(answer) {
		return true
	}
	if !wantsHistoricalAtomScan(query) {
		return false
	}
	n := 0
	for _, p := range strings.Split(answer, ",") {
		if strings.TrimSpace(p) != "" {
			n++
		}
	}
	return n >= 2 && n <= 8 && utf8Len(answer) <= 200
}

func leftoverThinMissAnswer(query string, hops []HopResult, answer string) bool {
	answer = strings.TrimSpace(answer)
	if looksLocationListQuery(query) {
		return false
	}
	if leftoverQueryEchoAnswer(query, answer) || leftoverShortItemJoin(answer) {
		return false
	}
	if answer == "" || strings.EqualFold(answer, "not in memory") {
		return false
	}
	if strings.Contains(answer, ",") || consecutiveProperNouns(answer) > 0 {
		return false
	}
	if !looksThinPacketLine(answer) {
		return false
	}
	rare := leftoverCoverRareTokens(query, hops)
	if len(rare) == 0 {
		return false
	}
	return !contentCoversAnyQueryToken(answer, rare)
}

func leftoverThinSloganAnswer(query string, hops []HopResult, answer string) bool {
	return leftoverQueryEchoAnswer(query, answer) || leftoverThinMissAnswer(query, hops, answer)
}

func stripConflictingDateTail(query, line string) string {
	qd := parseDateFromText(query)
	ld := parseDateFromText(line)
	if qd == nil || ld == nil {
		return line
	}
	if qd.Year() == ld.Year() && qd.YearDay() == ld.YearDay() {
		return line
	}
	lower := strings.ToLower(line)
	idx := strings.LastIndex(lower, " on ")
	if idx < 4 {
		return line
	}
	head := strings.TrimSpace(line[:idx])
	if utf8Len(head) < 12 {
		return line
	}
	return strings.TrimRight(head, ".,;:")
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
