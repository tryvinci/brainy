package memory

import (
	"context"
	"strings"
)

// ContextualExtractor wraps an Extractor and injects recent conversation /
// related memory context into the ingest request for provider compile.
type ContextualExtractor struct {
	inner Extractor
	store Store
	limit int
}

func NewContextualExtractor(inner Extractor, store Store) *ContextualExtractor {
	if inner == nil {
		inner = NewDeterministicExtractor()
	}
	return &ContextualExtractor{inner: inner, store: store, limit: 12}
}

func (c *ContextualExtractor) Identity() ExtractorIdentity {
	if ident, ok := c.inner.(interface{ Identity() ExtractorIdentity }); ok {
		return ident.Identity()
	}
	return ExtractorIdentity{Name: "contextual", Provider: "contextual"}
}

func (c *ContextualExtractor) Stats() ExtractorStats {
	if s, ok := c.inner.(interface{ Stats() ExtractorStats }); ok {
		return s.Stats()
	}
	return ExtractorStats{}
}

func ExtractorIdentityOf(e Extractor) ExtractorIdentity {
	if e == nil {
		return ExtractorIdentity{Name: "none", Provider: "none"}
	}
	if ident, ok := e.(interface{ Identity() ExtractorIdentity }); ok {
		return ident.Identity()
	}
	return ExtractorIdentity{Name: "unknown", Provider: "unknown"}
}

func ExtractorStatsOf(e Extractor) ExtractorStats {
	if e == nil {
		return ExtractorStats{}
	}
	if s, ok := e.(interface{ Stats() ExtractorStats }); ok {
		return s.Stats()
	}
	return ExtractorStats{}
}

func (c *ContextualExtractor) Extract(ctx context.Context, req IngestRequest) ([]ExtractedMemory, error) {
	enriched := req
	if c.store != nil {
		ctxBlock := c.buildContextBlock(ctx, req)
		if ctxBlock != "" {
			if enriched.Metadata == nil {
				enriched.Metadata = map[string]any{}
			}
			// Copy metadata map so we do not mutate the queued payload permanently
			// beyond the extract call (idempotent re-runs stay stable).
			meta := make(map[string]any, len(enriched.Metadata)+1)
			for k, v := range enriched.Metadata {
				meta[k] = v
			}
			meta["extract_context"] = ctxBlock
			enriched.Metadata = meta
		}
	}
	return c.inner.Extract(ctx, enriched)
}

func (c *ContextualExtractor) buildContextBlock(ctx context.Context, req IngestRequest) string {
	limit := c.limit
	if limit <= 0 {
		limit = 12
	}
	var b strings.Builder
	// Recent active memories (write-time continuity).
	if recent, err := c.store.ListActiveMemories(ctx, req.TenantID, req.SubjectID); err == nil && len(recent) > 0 {
		// ListActiveMemories returns newest-first (ORDER BY updated_at DESC).
		// Keep that order so "Recent memories" is actually recent.
		b.WriteString("Recent memories:\n")
		n := 0
		for _, mem := range recent {
			if n >= limit {
				break
			}
			content := strings.TrimSpace(mem.Content)
			if content == "" || strings.HasSuffix(content, "?") {
				continue
			}
			b.WriteString("- ")
			if id := strings.TrimSpace(mem.MemoryID); id != "" {
				b.WriteString("[")
				b.WriteString(id)
				b.WriteString("] ")
			}
			b.WriteString(truncateRunes(content, 180))
			b.WriteString("\n")
			n++
		}
	}
	// Related memories via light lexical search on new message tokens.
	probe := strings.TrimSpace(strings.Join(func() []string {
		parts := make([]string, 0, len(req.Messages))
		for _, m := range req.Messages {
			parts = append(parts, m.Content)
		}
		return contentBearingTokens(tokenize(strings.Join(parts, " ")))
	}(), " "))
	if probe != "" {
		if hits, err := c.store.SearchActiveMemories(ctx, req.TenantID, req.SubjectID, []string{probe}, 8); err == nil && len(hits) > 0 {
			b.WriteString("Related existing memories:\n")
			for _, h := range hits {
				content := strings.TrimSpace(h.Content)
				if content == "" {
					continue
				}
				b.WriteString("- ")
				b.WriteString(truncateRunes(content, 180))
				b.WriteString("\n")
			}
		}
	}
	return strings.TrimSpace(b.String())
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
