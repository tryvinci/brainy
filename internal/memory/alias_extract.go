package memory

import (
	"context"
	"regexp"
	"strings"
)

// Dialogue alias capture (S3b). Generic nickname / also-known-as forms only.
var (
	aliasCallMe     = regexp.MustCompile(`(?i)\b(?:call me|called|known as|goes by|nickname is|nicknamed)\s+([A-Z][A-Za-z][A-Za-z'-]{0,30})\b`)
	aliasAlsoKnown  = regexp.MustCompile(`(?i)\b([A-Z][A-Za-z]+(?:\s+[A-Z][A-Za-z]+){0,2})\s+(?:is also known as|aka|a\.k\.a\.)\s+([A-Z][A-Za-z][A-Za-z'-]{0,30})\b`)
	aliasOrNickname = regexp.MustCompile(`\b([A-Z][A-Za-z]+(?:\s+[A-Z][A-Za-z]+)?),\s+or\s+([A-Z][A-Za-z][A-Za-z'-]{0,20}),`)
)

// ExtractDialogueAliases returns (canonical label, alias) pairs from an utterance.
func ExtractDialogueAliases(text, speaker string) [][2]string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	out := make([][2]string, 0, 2)
	seen := map[string]struct{}{}
	add := func(canon, alias string) {
		canon = strings.TrimSpace(canon)
		alias = strings.TrimSpace(alias)
		if canon == "" || alias == "" || strings.EqualFold(canon, alias) {
			return
		}
		if utf8Len(alias) < 2 || utf8Len(alias) > 32 {
			return
		}
		key := NormalizeEntityLabel(canon) + "\x1f" + NormalizeEntityLabel(alias)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, [2]string{canon, alias})
	}
	if speaker != "" {
		for _, m := range aliasCallMe.FindAllStringSubmatch(text, 4) {
			if len(m) > 1 {
				add(speaker, m[1])
			}
		}
	}
	for _, m := range aliasAlsoKnown.FindAllStringSubmatch(text, 4) {
		if len(m) > 2 {
			add(m[1], m[2])
		}
	}
	for _, m := range aliasOrNickname.FindAllStringSubmatch(text, 4) {
		if len(m) > 2 {
			add(m[1], m[2])
		}
	}
	return out
}

// PersistIngestAliases captures nickname/aka forms from the ingest batch (S3b).
func PersistIngestAliases(ctx context.Context, store any, req IngestRequest) {
	for _, msg := range req.Messages {
		speaker := ""
		if i := strings.Index(msg.Content, ":"); i > 0 && i < 40 {
			cand := strings.TrimSpace(msg.Content[:i])
			if cand != "" && !strings.Contains(cand, " ") {
				speaker = cand
			}
		}
		PersistDialogueAliases(ctx, store, req.TenantID, req.SubjectID, speaker, msg.Content)
	}
}

// PersistDialogueAliases merges in-dialogue nicknames onto the canonical entity.
func PersistDialogueAliases(ctx context.Context, store any, tenantID, subjectID, speaker, text string) {
	registry, ok := store.(EntityRegistry)
	if !ok || registry == nil {
		return
	}
	for _, pair := range ExtractDialogueAliases(text, speaker) {
		canon, alias := pair[0], pair[1]
		ent := ResolveCanonicalEntity(ctx, store, tenantID, subjectID, canon)
		if ent.EntityID == "" {
			ent.EntityID = CanonicalEntityID(tenantID, subjectID, canon)
			ent.CanonicalLabel = canon
			ent.TenantID = tenantID
			ent.SubjectID = subjectID
		}
		aliases := append(EntityAliases(ent.CanonicalLabel), NormalizeEntityLabel(alias))
		_ = registry.UpsertMemoryEntity(ctx, MemoryEntity{
			TenantID:       tenantID,
			SubjectID:      subjectID,
			EntityID:       ent.EntityID,
			CanonicalLabel: firstNonEmpty(ent.CanonicalLabel, canon),
			Aliases:        aliases,
		})
	}
}
