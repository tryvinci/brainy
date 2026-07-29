# Multi-hop fix: learnings from Mem0 OSS v3

**Shipped:** additive extract prompt (provider-v2) · deterministic buried-fact atoms ·
entity hub-and-spoke (`memory_entity_links`) · additive signal fusion.

## Root cause (from same-pin LOCOMO)

Mem0 6/10 vs Brainy ~2/10 multi-hop was primarily **extract quality**: platform stores
atoms (“Caroline is a transgender woman”) while we ranked raw dialogue + acks.
GT phrases often exist in the transcript (Sweden, single parent, dinosaur exhibit).

## Mem0 patterns adopted (anti-benchmax)

1. **ADD-only atomic extraction** — one self-contained fact per attribute; named speakers.
2. **Entity hub** — `entity_key → linked_memory_ids[]`; query entities boost linked memories;
   skip hubs with >40 links (speaker flood).
3. **Additive fusion** — combine lexical + calibrated semantic + entity hub (normalized).

Not adopted: external Neo4j (Mem0 removed it from OSS); LOCOMO cue lists.

## Files

- `internal/memory/provider_extractor.go` — additive prompt + Observation Date
- `internal/memory/attribute_atoms.go` — transgender/single/Sweden/dinosaur atoms
- `internal/memory/entity_hub.go` + `store/postgres/entity_hub.go` — migration v12
- Search admits hub-linked candidates (cap 24)


## Follow-up (2026-07-29)

Speaker carry-forward + query-intent boosts + list harvest → multi-hop **5/10**, overall **19/30**.
Passes single/Sweden atoms. One shy of Mem0 same-pin MH (6/10).
