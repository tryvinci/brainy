package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// MemoryEntity is a tenant/subject-scoped canonical identity (R7).
// Names are not unique identities: "John Smith" and "John Doe" coexist.
type MemoryEntity struct {
	TenantID       string
	SubjectID      string
	EntityID       string
	CanonicalLabel string
	Aliases        []string
}

// EntityRegistry persists canonical identities and ranked mention resolution.
type EntityRegistry interface {
	UpsertMemoryEntity(ctx context.Context, ent MemoryEntity) error
	ResolveMemoryEntity(ctx context.Context, tenantID, subjectID, mention string) (MemoryEntity, bool, error)
}

// CanonicalEntityID is a durable ID for a normalized mention in one
// tenant/subject scope. The same surface string in another subject is a
// different person. Distinct labels (John Smith vs John Doe) get distinct IDs.
func CanonicalEntityID(tenantID, subjectID, mention string) string {
	mention = strings.TrimSpace(mention)
	if mention == "" {
		return ""
	}
	if strings.HasPrefix(mention, "ent:") {
		return mention
	}
	label := NormalizeEntityLabel(mention)
	if label == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(
		strings.ToLower(strings.TrimSpace(tenantID)) + "\x1f" +
			strings.ToLower(strings.TrimSpace(subjectID)) + "\x1f" +
			label,
	))
	return "ent:" + hex.EncodeToString(sum[:12])
}

// NormalizeEntityLabel lowercases and collapses a mention for identity keys.
func NormalizeEntityLabel(mention string) string {
	s := strings.ToLower(strings.TrimSpace(mention))
	s = strings.Trim(s, ".,;:!?\"'`")
	s = strings.ReplaceAll(s, "’s", "")
	s = strings.TrimSuffix(s, "'s")
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// EntityAliases returns lookup keys for a label: the full form, then the
// first token when the label is multi-word (first name / given name).
func EntityAliases(label string) []string {
	n := NormalizeEntityLabel(label)
	if n == "" {
		return nil
	}
	out := []string{n}
	fields := strings.Fields(n)
	if len(fields) >= 2 && fields[0] != n {
		out = append(out, fields[0])
	}
	return out
}

func attachEntityIdentity(record *MemoryRecord) {
	if record == nil {
		return
	}
	subj := entitySubjectOf(*record)
	if subj == "" {
		return
	}
	eid := entityIDOf(*record)
	if eid == "" {
		eid = CanonicalEntityID(record.TenantID, record.SubjectID, subj)
	}
	if eid == "" {
		return
	}
	if record.Metadata == nil {
		record.Metadata = map[string]any{}
	}
	if record.Explain == nil {
		record.Explain = map[string]any{}
	}
	record.Metadata["entity_id"] = eid
	record.Explain["entity_id"] = eid
	if strings.TrimSpace(entitySubjectOf(*record)) != "" {
		record.Metadata["subject"] = strings.TrimSpace(subj)
		record.Explain["subject"] = strings.TrimSpace(subj)
	}
}

func entityIDOf(record MemoryRecord) string {
	if record.Metadata != nil {
		if v, ok := record.Metadata["entity_id"].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	if record.Explain != nil {
		if v, ok := record.Explain["entity_id"].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// ResolveCanonicalEntity ranked-resolves a mention to a stored entity, or
// synthesizes a stable ID from the mention when the registry misses.
func ResolveCanonicalEntity(ctx context.Context, store any, tenantID, subjectID, mention string) MemoryEntity {
	mention = strings.TrimSpace(mention)
	if mention == "" {
		return MemoryEntity{}
	}
	if registry, ok := store.(EntityRegistry); ok && registry != nil {
		if ent, found, err := registry.ResolveMemoryEntity(ctx, tenantID, subjectID, mention); err == nil && found {
			return ent
		}
	}
	return MemoryEntity{
		TenantID:       tenantID,
		SubjectID:      subjectID,
		EntityID:       CanonicalEntityID(tenantID, subjectID, mention),
		CanonicalLabel: strings.TrimSpace(mention),
		Aliases:        EntityAliases(mention),
	}
}

// RankEntityResolution picks one candidate for a mention.
// Exact ID / full-label wins; a first-name alias wins only when unique.
func RankEntityResolution(candidates []MemoryEntity, mention string) (MemoryEntity, bool) {
	n := NormalizeEntityLabel(mention)
	raw := strings.TrimSpace(mention)
	if n == "" && raw == "" {
		return MemoryEntity{}, false
	}
	var exact, aliasHits []MemoryEntity
	seen := map[string]struct{}{}
	add := func(dst *[]MemoryEntity, e MemoryEntity) {
		if e.EntityID == "" {
			return
		}
		if _, ok := seen[e.EntityID]; ok {
			return
		}
		seen[e.EntityID] = struct{}{}
		*dst = append(*dst, e)
	}
	for _, e := range candidates {
		if raw != "" && (e.EntityID == raw || strings.EqualFold(e.EntityID, raw)) {
			add(&exact, e)
			continue
		}
		if n != "" && NormalizeEntityLabel(e.CanonicalLabel) == n {
			add(&exact, e)
			continue
		}
		for _, a := range e.Aliases {
			if n != "" && a == n {
				add(&aliasHits, e)
				break
			}
		}
	}
	if len(exact) == 1 {
		return exact[0], true
	}
	if len(exact) > 1 {
		best := exact[0]
		for _, e := range exact[1:] {
			if utf8Len(e.CanonicalLabel) > utf8Len(best.CanonicalLabel) {
				best = e
			}
		}
		return best, true
	}
	if len(aliasHits) == 1 {
		return aliasHits[0], true
	}
	return MemoryEntity{}, false
}

func PersistCanonicalEntity(ctx context.Context, store any, record MemoryRecord) {
	if store == nil {
		return
	}
	registry, ok := store.(EntityRegistry)
	if !ok {
		return
	}
	subj := entitySubjectOf(record)
	if subj == "" {
		return
	}
	eid := entityIDOf(record)
	if eid == "" {
		eid = CanonicalEntityID(record.TenantID, record.SubjectID, subj)
	}
	if eid == "" {
		return
	}
	_ = registry.UpsertMemoryEntity(ctx, MemoryEntity{
		TenantID:       record.TenantID,
		SubjectID:      record.SubjectID,
		EntityID:       eid,
		CanonicalLabel: subj,
		Aliases:        EntityAliases(subj),
	})
}
