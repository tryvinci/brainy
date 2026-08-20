package memory

import (
	"context"
	"math"
	"strings"
	"time"
)

// EntityLinker persists Mem0-style hub-and-spoke entity → memory links.
type EntityLinker interface {
	LinkMemoryEntities(ctx context.Context, tenantID, subjectID, memoryID string, entities []string) error
	// EntityHubBoosts returns per-memory boosts for memories linked to any of
	// the query entities. Boost is attenuated for ubiquitous entities.
	EntityHubBoosts(ctx context.Context, tenantID, subjectID string, queryEntities []string) (map[string]float64, error)
}

const entityHubBoostWeight = 0.5

// combineRetrievalSignals fuses lexical, calibrated semantic, and entity-hub
// scores the Mem0 v3 way: (sum active signals) / max_possible, capped at 1,
// then rescaled into Brainy's existing score range via base lexical.
//
// lexical is the pre-hybrid lexical score (>0). semantic and entityHub are in [0,1]
// and [0, entityHubBoostWeight] respectively.
func combineRetrievalSignals(lexical, semantic, entityHub float64) (combined float64, explain map[string]float64) {
	explain = map[string]float64{
		"lexical":    lexical,
		"semantic":   semantic,
		"entity_hub": entityHub,
	}
	if lexical <= 0 && semantic <= 0 && entityHub > 0 {
		combined = entityHub * 1.25
		explain["combined"] = combined
		return combined, explain
	}
	maxPossible := 1.0
	raw := math.Max(lexical, 0)
	if semantic > 0 {
		maxPossible += 1.0
		raw += semantic
	}
	if entityHub > 0 {
		maxPossible += entityHubBoostWeight
		raw += entityHub
	}
	norm := raw / maxPossible
	if norm > 1 {
		norm = 1
	}
	combined = lexical*(0.55+0.45*norm) + semantic*0.55 + entityHub
	explain["combined"] = combined
	return combined, explain
}

func (s *Service) persistEntityLinks(ctx context.Context, record MemoryRecord) {
	linker, ok := s.store.(EntityLinker)
	if !ok {
		return
	}
	ents := recordEntities(record)
	ents = appendIdentityEntities(ents, record)
	if eid := entityIDOf(record); eid != "" {
		ents = append(ents, eid)
	}
	if len(ents) == 0 {
		return
	}
	_ = linker.LinkMemoryEntities(ctx, record.TenantID, record.SubjectID, record.MemoryID, ents)

	PersistCanonicalEntity(ctx, s.store, record)
	PersistDialogueAliases(ctx, s.store, record.TenantID, record.SubjectID, entitySubjectOf(record), record.Content)
	if record.SourceText != "" && record.SourceText != record.Content {
		PersistDialogueAliases(ctx, s.store, record.TenantID, record.SubjectID, entitySubjectOf(record), record.SourceText)
	}

	if indexer, ok := s.store.(AtomIndexer); ok {
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
		if pred != "" && val != "" {
			_ = indexer.UpsertMemoryAtom(ctx, record.TenantID, record.SubjectID, pred, val, record.MemoryID, record.ObservedAt)
		}
	}
	if rel, ok := ProjectMemoryRelation(record); ok {
		if indexer, ok := s.store.(RelationIndexer); ok {
			_ = indexer.UpsertMemoryRelation(ctx, rel)
		}
	}
}

// AtomIndexer persists typed atoms for predicate enumeration (W2/W3).
type AtomIndexer interface {
	UpsertMemoryAtom(ctx context.Context, tenantID, subjectID, predicate, value, memoryID string, observedAt *time.Time) error
	ListAtomMemoryIDs(ctx context.Context, tenantID, subjectID, predicate, valueNorm string, limit int) ([]string, error)
}

// AtomRetirer retires superseded state atoms (bitemporal valid_to / retired_at).
type AtomRetirer interface {
	RetireMemoryAtom(ctx context.Context, tenantID, subjectID, predicate, value, memoryID string, endedAt *time.Time) error
}

func appendIdentityEntities(ents []string, record MemoryRecord) []string {
	add := func(raw string) {
		e := strings.ToLower(strings.TrimSpace(raw))
		if e == "" || utf8Len(e) < 2 {
			return
		}
		if i := strings.IndexAny(e, ",;("); i > 0 {
			e = strings.TrimSpace(e[:i])
		}
		if utf8Len(e) > 48 {
			return
		}
		for _, existing := range ents {
			if existing == e {
				return
			}
		}
		ents = append(ents, e)
	}
	if record.Explain != nil {
		if s, ok := record.Explain["subject"].(string); ok {
			add(s)
		}
		if v, ok := record.Explain["value_norm"].(string); ok {
			add(v)
		}
	}
	if record.Metadata != nil {
		if s, ok := record.Metadata["subject"].(string); ok {
			add(s)
		}
		if v, ok := record.Metadata["value_norm"].(string); ok {
			add(v)
		}
	}
	return ents
}

func (s *Service) entityHubBoostMap(ctx context.Context, tenantID, subjectID, query string) map[string]float64 {
	linker, ok := s.store.(EntityLinker)
	if !ok {
		return nil
	}
	qEnts := ExtractEntities(query)
	if len(qEnts) == 0 {
		// Fall back to name-like content tokens as soft entities.
		for _, t := range nameLikeTokens(contentBearingTokens(tokenize(query))) {
			qEnts = append(qEnts, t)
		}
	}
	if len(qEnts) == 0 {
		return nil
	}
	boosts, err := linker.EntityHubBoosts(ctx, tenantID, subjectID, qEnts)
	if err != nil || len(boosts) == 0 {
		return nil
	}
	return boosts
}
