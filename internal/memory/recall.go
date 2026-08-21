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
		hopResults, byKey = s.executeTypedHops(ctx, req.TenantID, req.SubjectID, req.Vertical, hist, plan, topK)
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
			merged := mergeSearchResults(search.Results, extra, topK)
			search.Results = merged
			out.Memories = merged
		}
		// Second pass only when join not yet proven — single unresolved hop probe.
		if plan.BudgetPasses >= 2 && !hopJoinProven(hopResults) {
			if unc := uncoveredTargets(pkt); len(unc) > 0 {
				probe := nextHopProbe(plan, pkt)
				if probe != "" {
					if second, err := s.SearchOpt(ctx, req.TenantID, req.SubjectID, req.Vertical, "", probe, SearchOptions{
						IncludeHistorical: hist,
						Limit:             topK,
					}); err == nil && len(second.Results) > 0 {
						merged := mergeSearchResults(search.Results, second.Results, topK)
						search.Results = merged
						out.Memories = merged
						hopResults2, byKey2 := s.executeTypedHops(ctx, req.TenantID, req.SubjectID, req.Vertical, hist, plan, topK)
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
			if probe != "" {
				if second, err := s.SearchOpt(ctx, req.TenantID, req.SubjectID, req.Vertical, "", probe, SearchOptions{
					IncludeHistorical: hist,
					Limit:             topK,
				}); err == nil && len(second.Results) > 0 {
					merged := mergeSearchResults(search.Results, second.Results, topK)
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
		out.Items = items
		out.ContextBlock = formatEnumerateContext(items)
		out.Explain["item_count"] = len(items)
		out.AnswerStatus = AnswerSupported
		if len(items) == 0 {
			out.AnswerStatus = AnswerNotFound
			out.Abstained = true
		}
		out.Coverage = map[string]any{
			"targets":   len(plan.CoverageTargets),
			"satisfied": len(items) > 0,
			"source":    "evidence_packet",
		}
	case "answer":
		out.ContextBlock = assembleContextFromPacket(pkt, budget)
		if (plan.NeedsEnumeration || looksListQuery(tokenize(req.Query)) || looksUnwindQuery(req.Query) || looksSuperlativeQuery(req.Query) || transferRecipient(req.Query) != "") && !looksConsequenceQuery(req.Query) && !looksWhereQuery(req.Query) {
			enumerated = true
			items := s.enumerateFromSearch(ctx, req, search.Results, hopResults)
			items = filterBesides(req.Query, items)
			items = s.filterHopEvidence(ctx, req, items, hopResults)
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
			if ans := whereAnswerFromHops(hopResults); ans != "" {
				out.Answer = ans
				out.AnswerStatus = AnswerSupported
				out.Explain["where_answer"] = true
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
		if strings.TrimSpace(out.Answer) == "" && plan.NeedsMultiHop && hopComposeAllowed(req.Query) {
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
		if strings.TrimSpace(out.Answer) == "" {
			if ans := pickStructuredAnswer(req.Query, search.Results); ans != "" {
				out.Answer = ans
				out.AnswerStatus = AnswerSupported
				out.Explain["structured_answer"] = true
			}
		}
		if strings.TrimSpace(out.Answer) == "" {
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
	if mode == "answer" {
		hybrid := s.synthesizeHybridAnswer(ctx, req.Query, plan, pkt)
		if hybrid.Reason != "" {
			out.Explain["hybrid_reader_reason"] = hybrid.Reason
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
		lockedDate := looksWhenEventQuery(req.Query) && out.Explain["date_answer"] == true
		lockedWhere := looksWhereQuery(req.Query) && out.Explain["where_answer"] == true
		if hybrid.OK && !lockedDate && !lockedWhere {
			out.Answer = strings.TrimSpace(hybrid.Answer)
			if hopComposeAllowed(req.Query) {
				grounded := groundToHopValues(hybrid.Answer, hopResults)
				out.Answer = grounded
				if composed := composeFromHopValues(hopResults); composed != "" && grounded == composed && grounded != strings.TrimSpace(hybrid.Answer) {
					out.Explain["hybrid_grounded_to_hops"] = true
				}
			}
			out.Abstained = false
			out.AnswerStatus = hybridAnswerStatus(hybrid, plan, pkt, packetOK)
			out.Explain["reader_source"] = "hybrid_llm_packet"
			plan.Tools = append(plan.Tools, "hybrid_reader")
			out.Explain["query_plan"] = plan
			out.Explain["tools_executed"] = plan.Tools
		} else if hybrid.Abstain && !lockedDate && !lockedWhere {
			if hopComposeAllowed(req.Query) {
				if composed := composeFromHopValues(hopResults); composed != "" {
					out.Answer = composed
					out.Abstained = false
					out.AnswerStatus = AnswerSupported
					out.Explain["reader_source"] = "multihop_bridge_chain"
				} else {
					out.Abstained = true
					out.Answer = "not in memory"
					out.AnswerStatus = AnswerInsufficient
					out.Explain["reader_source"] = "hybrid_llm_packet"
				}
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
	}
	pkt.Plan = plan
	out.Explain["evidence_packet"] = pkt
	return out, nil
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

	// Prefer indexed atoms when hops missed or to fill additional values.
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
	}
	if len(preds) > 0 {
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

	if len(order) == 0 {
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
	return items
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
		return hopPred == PredicateEvent || hopPred == PredicatePreference || hopPred == PredicateSkill
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
	if strings.HasSuffix(t, "s") && len(t) > 4 {
		t = t[:len(t)-1]
	}
	return t
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
	add := func(n string) {
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
	for _, h := range hops {
		if h.Kind == "resolve_entity" || h.Source == "unresolved" || h.Source == "search_fallback" {
			continue
		}
		for _, v := range h.Values {
			add(v)
		}
		add(h.Value)
		for _, c := range h.Contents {
			for _, w := range strings.Fields(c) {
				add(strings.Trim(w, "?,.!'\"()"))
			}
		}
	}
	if len(names) == 0 {
		return ""
	}
	return strings.Join(names, ", ")
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
	if toks := groupCompanionTokens(req.Query); len(toks) > 0 {
		items = s.filterItemsByTokens(ctx, req, items, hops, toks)
	}
	if toks := forClauseTokens(req.Query); len(toks) > 0 {
		items = s.filterItemsByTokens(ctx, req, items, hops, toks)
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

func whereAnswerFromHops(hops []HopResult) string {
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
	for _, h := range hops {
		if h.Kind == "resolve_entity" || h.Source == "unresolved" || h.Source == "search_fallback" {
			continue
		}
		for _, c := range h.Contents {
			add(placeFromContent(c))
		}
		if h.Value != "" {
			add(placeFromContent(h.Value))
		}
	}
	if len(places) == 0 {
		return ""
	}
	return strings.Join(places, ", ")
}

func placeFromContent(content string) string {
	content = strings.TrimSpace(stripTrailingStamp(content))
	if content == "" {
		return ""
	}
	lower := strings.ToLower(content)
	best := ""
	bestIdx := -1
	for _, prep := range []string{" in ", " at ", " near ", " around "} {
		i := strings.LastIndex(lower, prep)
		if i < 0 || i < bestIdx {
			continue
		}
		if cand := cleanPlaceCandidate(content[i+len(prep):]); cand != "" {
			best = cand
			bestIdx = i
		}
	}
	return best
}

func cleanPlaceCandidate(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if j := strings.IndexAny(s, ".([;!?"); j >= 0 {
		s = strings.TrimSpace(s[:j])
	}
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	first := strings.ToLower(strings.Trim(fields[0], ",'\"'"))
	switch first {
	case "his", "her", "their", "its", "my", "our", "this", "that", "these", "those",
		"order", "common", "fact", "addition":
		return ""
	}
	stopAt := map[string]struct{}{
		"last": {}, "during": {}, "when": {}, "after": {}, "before": {},
		"because": {}, "with": {}, "while": {}, "where": {}, "and": {},
	}
	keep := make([]string, 0, 4)
	for i, f := range fields {
		w := strings.ToLower(strings.Trim(f, ",'\"'"))
		if _, ok := stopAt[w]; ok && i > 0 {
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
	out := strings.Join(keep, " ")
	if anaphoricSlotValue(out) {
		return ""
	}
	return out
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
