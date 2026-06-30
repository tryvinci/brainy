# Marketing Vertical — Use Case Map

**Status:** Approved (ENG-80, 2026-06-19)  
**Decision context:** Marketing chosen as first vertical wedge (ENG-58)  
**Architecture:** Domain terms below are **pack labels**, not DB enums — see [verticalization-model.md](./verticalization-model.md) (ENG-61, approved)  
**Repo baseline:** Brainy Go thin slice — `profile` / `preference` / `fact` kinds, ingest, search, correct, suppress  
**Pack:** `packs/marketing/v1/pack.yaml`  
**Last updated:** 2026-06-19

## Purpose

Map marketing agent jobs to memory requirements and Brainy cognitive primitives. This document is **vertical discovery input** for the marketing pack (`packs/marketing/v1/`), not a schema spec. It feeds ENG-81 (brand voice behavior) and ENG-82 (eval fixtures).

---

## Agent Jobs Overview

| # | Agent job | Primary user | Session pattern |
|---|---|---|---|
| 1 | Brand voice agent | Brand manager, agency strategist | Long-lived; cross-session |
| 2 | Campaign manager | Performance marketer, media buyer | Campaign-scoped; bursts |
| 3 | Content strategist | SEO/content lead | Topic-scoped; recurring |
| 4 | Creative assistant | Copywriter, designer | Asset-scoped; iterative |
| 5 | Audience analyst | Growth marketer, CRM lead | Segment-scoped; data-driven |

---

## 1. Brand Voice Agent

**Job:** Maintain consistent tone, vocabulary, and taboo compliance across all generated content.

### Input sources

| Source | Example | Ingest path |
|---|---|---|
| Brand guidelines PDF / doc | "Never use slang; always lead with benefit" | Manual upload → async extract |
| Style guide conversation | "Our voice is warm but authoritative" | Sync ingest |
| Legal/compliance review | "Do not claim #1 without substantiation" | Correct/suppress on violation |
| Reference content corpus | Top-performing emails, ads | Batch ingest → pattern extraction |

### Memory types needed

| Type | Example | Maps to today | Target primitive |
|---|---|---|---|
| Voice descriptor | "Warm, concise, second-person" | `preference` | **IdentityPrior** |
| Hard taboo | "Never mention competitor X by name" | `fact` (suppressed on violation) | **Principle** (immutable) |
| Soft style preference | "Prefer short sentences" | `preference` | IdentityPrior |
| Vocabulary allowlist | "Use 'clients' not 'customers'" | `preference` | IdentityPrior |

### Retrieval patterns

- **Pre-generation lookup:** "What is the brand voice for {subject_id}?" → rank IdentityPrior + Principle first
- **Taboo check:** Before output, search for Principle-level suppressions matching topic
- **Override hierarchy:** Principle > IdentityPrior > generic `preference` > session default

### Failure modes

| Failure | Impact | Mitigation |
|---|---|---|
| Generic preference overrides brand rule | Off-brand output | Principle immutability + rank boost |
| Stale voice after rebrand | Mixed old/new tone | Supersession on brand refresh event |
| Taboo leak under paraphrase | Compliance risk | Suppression leak tests in eval |
| Multi-brand tenant bleed | Agency cross-contamination | Strict `tenant_id` + `subject_id` isolation |

### Cognitive primitive mapping

```
Brand guidelines ingest → Episode
Repeated style examples → Episode → Pattern ("this phrasing works")
Stable voice rules → IdentityPrior
Non-negotiable taboos → Principle (immutable; governance path for change)
Style ranking in retrieval → TasteSignal
```

---

## 2. Campaign Manager

**Job:** Track campaign context — audience, channels, creative variants, dates — and surface relevant memories during active campaigns.

### Input sources

| Source | Example | Ingest path |
|---|---|---|
| Campaign brief | "Q3 launch, millennials, Instagram + email" | Sync ingest |
| CRM / ad platform sync | Audience segment definitions | Async ingest (future integration) |
| Status updates | "Campaign paused due to budget" | Correct or domain event |
| Post-campaign retro | "Carousel outperformed static 2:1" | Outcome ingest → belief update |

### Memory types needed

| Type | Example | Maps to today | Target kind (proposed) |
|---|---|---|---|
| Campaign descriptor | "Summer Sale 2026, active until Aug 31" | `fact` | `campaign` |
| Audience segment | "High-intent cart abandoners, 25–34" | `profile` | `audience_segment` |
| Channel preference | "Prioritize email over paid social for this campaign" | `preference` | `preference` |
| Performance outcome | "Variant B CTR 3.2% vs 1.8%" | `fact` | `performance_outcome` |

### Retrieval patterns

- **Active campaign boost:** Memories tagged with active campaign get rank multiplier
- **Lifecycle filter:** Completed/archived campaign memories excluded from default search
- **Cross-campaign query:** "What worked for similar audiences?" → Pattern retrieval across Episodes

### Campaign lifecycle → memory state

```
draft     → memories stored, low retrieval weight
active    → retrieval boost, default-visible
paused    → visible but deprioritized
completed → decay rank; not deleted
archived  → suppressed from default search (ENG-83)
```

### Failure modes

| Failure | Impact | Mitigation |
|---|---|---|
| Stale campaign context after end date | Wrong CTA, expired offers | Lifecycle suppression (ENG-83) |
| Segment confusion across campaigns | Wrong audience targeting | Entity linking: campaign → segment |
| Outcome not updating beliefs | Repeats failed creative patterns | Outcome → Belief update path |
| Duplicate brief ingest | Fragmented campaign memory | Dedupe on campaign_id key |

### Cognitive primitive mapping

```
Campaign brief → Episode
Repeated campaign structures → Episode → Pattern
Performance results → Outcome → Belief ("carousel > static for this segment")
Active campaign context → scoped retrieval filter + rank boost
```

---

## 3. Content Strategist

**Job:** Maintain topic authority, SEO context, and content performance learnings across a content calendar.

### Input sources

| Source | Example | Ingest path |
|---|---|---|
| Content calendar | "Publishing 2 posts/week on sustainability" | Sync ingest |
| SEO research | "Target keyword: carbon neutral shipping" | Async extract |
| Analytics feedback | "How-to posts outperform thought leadership 3:1" | Outcome ingest |
| Editorial corrections | "Don't use 'eco-friendly' — use 'sustainably sourced'" | Correct |

### Memory types needed

| Type | Example | Maps to today | Target kind |
|---|---|---|---|
| Topic authority | "Brand is expert in sustainable logistics" | `profile` | `profile` |
| Keyword / SEO fact | "Primary keyword: carbon neutral delivery" | `fact` | `fact` + metadata |
| Content performance belief | "How-to format outperforms for this audience" | `fact` | **Belief** (conviction-weighted) |
| Editorial rule | "Avoid greenwashing terms" | `preference` | **Principle** or IdentityPrior |

### Retrieval patterns

- **Topic-scoped search:** Filter by topic tag / metadata
- **Belief-weighted ranking:** High-conviction performance beliefs rank above low-conviction facts
- **Correction stickiness:** Editorial corrections persist under paraphrase queries (already benchmarked in `evals/correction_stickiness_eval.py`)

### Failure modes

| Failure | Impact | Mitigation |
|---|---|---|
| Outdated SEO facts | Wrong keyword targeting | Temporal supersession on analytics refresh |
| Performance belief not updated after A/B | Suboptimal format choices | Outcome ingestion + conviction update |
| Generic fact overrides editorial rule | Brand/compliance drift | Principle > preference override hierarchy |

### Cognitive primitive mapping

```
Content pieces → Episode
Format performance trends → Pattern → Belief
Topic positioning → profile + IdentityPrior alignment
Editorial corrections → Correct path + lineage (existing)
```

---

## 4. Creative Assistant

**Job:** Help copywriters and designers produce on-brand creative with reference to past assets, revision history, and style preferences.

### Input sources

| Source | Example | Ingest path |
|---|---|---|
| Creative brief | "Hero image + 3 headline variants" | Sync ingest |
| Revision feedback | "Headline B too aggressive, soften tone" | Correct |
| Reference assets | Links to approved past creatives | Async ingest + metadata |
| Designer notes | "Use brand blue #0047AB, not generic blue" | `preference` ingest |

### Memory types needed

| Type | Example | Maps to today | Target kind |
|---|---|---|---|
| Asset reference | "Approved hero from Spring campaign" | `fact` | `creative_asset` |
| Style preference | "Minimal layout, high whitespace" | `preference` | TasteSignal input |
| Revision lineage | v1 → v2 → v3 headline chain | correction history (exists) | lineage via `memory_history` |
| Rejection reason | "Too corporate for Gen Z audience" | `preference` | Episode → Pattern |

### Retrieval patterns

- **Similar creative lookup:** "Show past headlines for product launch" → semantic + kind filter
- **Revision trace:** Query memory history for asset evolution
- **TasteSignal ranking:** Style-matched references rank above generic preferences

### Failure modes

| Failure | Impact | Mitigation |
|---|---|---|
| Rejected creative resurfaces | Wasted iteration | Suppress rejected variants |
| Style drift across revisions | Inconsistent output | IdentityPrior anchor on each retrieval |
| Asset reference without provenance | Cannot verify approval status | Source metadata + status field |

### Cognitive primitive mapping

```
Creative brief + feedback → Episode
Repeated style choices → Pattern
Style ranking in search → TasteSignal
Revision chain → correction lineage (existing)
Approved references → creative_asset with provenance
```

---

## 5. Audience Analyst

**Job:** Encode and retrieve audience segment preferences, behavioral patterns, and A/B test outcomes to personalize agent output.

### Input sources

| Source | Example | Ingest path |
|---|---|---|
| Segment definition | "VIP customers: LTV > $500, 3+ purchases" | Sync ingest |
| Survey / interview | "This segment prefers email over SMS" | Sync ingest |
| A/B test results | "Subject line A won by 12% open rate" | Outcome ingest |
| Behavioral signal | "Mobile users bounce on long-form" | Async ingest |

### Memory types needed

| Type | Example | Maps to today | Target kind |
|---|---|---|---|
| Segment profile | "VIP: high LTV, price-insensitive" | `profile` | `audience_segment` |
| Channel preference | "Prefers SMS for flash sales" | `preference` | `preference` + segment metadata |
| A/B outcome | "Emoji subject lines +12% open rate for segment X" | `fact` | `performance_outcome` → Belief |
| Behavioral heuristic | "Short copy wins on mobile for this segment" | `fact` | Pattern |

### Retrieval patterns

- **Segment-scoped query:** Filter by `audience_segment` metadata
- **Cross-segment conflict:** Same brand, different segment prefs → complement-first reconciliation (ENG architecture)
- **Outcome-driven re-ranking:** Recent A/B outcomes boost related beliefs

### Failure modes

| Failure | Impact | Mitigation |
|---|---|---|
| Segment bleed | Wrong personalization | subject_id / segment metadata isolation |
| Stale A/B belief | Suboptimal messaging | Conviction decay + outcome refresh |
| Conflicting segment prefs unresolved | Inconsistent recommendations | Conflict reconciliation protocol |

### Cognitive primitive mapping

```
Segment data → profile (audience_segment)
A/B results → Outcome → Belief (conviction-weighted)
Behavioral patterns → Episode → Pattern
Segment-scoped retrieval → metadata filter + rank
```

---

## Marketing pack vocabulary (labels → primitives)

Domain terms for the marketing vertical pack — **not** proposed DB `kind` enums. Storage uses `primitive` + `label` + `metadata` per [verticalization-model.md](./verticalization-model.md).

| Pack label | Primitive | Legacy kind (compat) | Example |
|---|---|---|---|
| `brand_rule` | **Principle** | `fact` | "Never claim market leadership without proof" |
| `voice_profile` | **IdentityPrior** | `preference` | "Warm, authoritative, second-person" |
| `campaign` | **Episode** (scoped) | `fact` | "Summer Sale 2026, active" |
| `audience_segment` | **Episode** (scoped) | `profile` | "VIP: LTV > $500" |
| `creative_asset` | **Episode** | `fact` | "Approved hero image, Spring 2026" |
| `performance_outcome` | **Outcome** | `fact` | "Variant B CTR 3.2%" |
| `content_belief` | **Belief** | `fact` | "How-to outperforms thought leadership (0.8 conviction)" |
| `preference` | **IdentityPrior** (soft) | `preference` | "Prefer short sentences" |
| `profile` | **Episode** / profile | `profile` | "Brand expert in sustainable logistics" |
| `fact` | **Episode** | `fact` | "Primary keyword: carbon neutral delivery" |

---

## Override Hierarchy (retrieval rank)

When memories conflict at retrieval time, apply this precedence:

```
1. Principle (brand_rule) — immutable, highest rank
2. IdentityPrior (voice_profile) — brand voice anchor
3. Active campaign context — lifecycle boost
4. Belief (content_belief) — conviction-weighted
5. preference / profile / fact — default rank
6. Suppressed / archived — excluded
```

---

## Generic Mem0 Gap Analysis

What generic memory (Mem0-style) misses for these jobs:

| Gap | Marketing impact | Brainy differentiation |
|---|---|---|
| No Principle immutability | Brand taboos overridden by later preferences | Principle > preference hierarchy |
| No campaign lifecycle | Stale campaign context surfaces | Lifecycle state + suppression |
| No TasteSignal / IdentityPrior | Generic recall, not brand-aware ranking | Cognitive primitive ranking |
| No Outcome → Belief loop | A/B learnings don't update retrieval | Conviction-weighted beliefs |
| No correction stickiness under semantic search | Editorial fixes lost on paraphrase | Already started in repo |
| No segment isolation | Audience bleed in multi-brand/agency setups | tenant + subject + segment metadata |

---

## Eval Scenario Seeds (for ENG-82 / Gate M3)

Each agent job yields ≥2 golden eval scenarios. **Gate M3** requires a passing fixture (or tracked `skip`) for every row.

| # | Scenario | Agent job | Validates | Fixture status |
|---|---|---|---|---|
| 1 | Brand rule overrides generic preference | Brand voice | Principle hierarchy | ✅ `bv01` |
| 2 | Taboo term never surfaces (paraphrase query) | Brand voice | Suppression leak | ✅ `bv02` |
| 5 | A/B outcome updates retrieval rank | Audience analyst | Outcome → Belief | ✅ `ob05` |
| 6 | Editorial correction sticks under paraphrase | Content strategist | Correction stickiness | ✅ `bv04` |
| 7 | Multi-brand isolation (tenant A ≠ tenant B) | All | tenant_id isolation | ✅ `bv06` |
| 8 | Cross-campaign pattern retrieval | Campaign manager | Episode → Pattern | ✅ `pt08` |
| 9 | Style-matched creative reference ranks first | Creative assistant | TasteSignal | ✅ `ts09` |
| 10 | Conflicting segment prefs coexist scoped | Audience analyst | Complement-first reconciliation | ✅ `sg10` |

Vetting policy: [`marketing-vetting-gate.md`](./marketing-vetting-gate.md)

---

## Implementation Phases (marketing pack v1 on general runtime)

| Phase | Scope | Repo touchpoints | Gate |
|---|---|---|---|
| **MVP-1** | Verticalization skeleton + static marketing pack | `model.go`, `packs/marketing/v1/`, rank policy | M1 ✅ |
| **MVP-2** | Generic lifecycle state machine + campaign rules in pack | `lifecycle_rules`, ENG-83 | M1 ✅ |
| **MVP-5** | Marketing eval suite (ENG-82) + CI (ENG-90) + benchmark (ENG-93) | `fixtures/vertical/marketing/`, `evals/` | M1 ✅ |
| **MVP-3** | Outcome ingest + Belief conviction in rank pipeline | new endpoint, primitive precedence | **M3** |
| **MVP-4** | Pack JSON Schema validation on ingest | pack loader, extractor | **M3** |
| **MVP-1.1** | Pack-driven `classification_rules` | `pack.yaml`, remove Go heuristics | **M3** |
| **ENG-87** | Semantic / hybrid retrieval without deterministic regression | pgvector, new fixtures | **M3** |
| **Finance pack** | Second vertical | `packs/finance/` — **blocked until Gate M4** | M4 |

Do not start finance until **Gate M3** clears. See [`marketing-vetting-gate.md`](./marketing-vetting-gate.md).

---

## Open Questions (for ENG-81, ENG-83)

1. **Principle changes:** Governance path for brand refresh — who can retire a Principle?
2. **Agency model:** One `tenant_id`, many brand `subject_id`s — confirm isolation model (ENG-65)
3. **Campaign events:** Push-based lifecycle (webhook) vs poll-based expiry?
4. **TasteSignal implementation:** Rule-based vs embedding-based style matching for MVP-1?
5. **Belief scope:** Full lifecycle (candidate → active → challenged → retired) or conviction float only for MVP?

~~ENG-61 taxonomy approach~~ — resolved: primitives + vertical packs (see verticalization-model.md).

---

## References

- `docs/brainy/architecture/00-cognitive-primitives.md`
- `docs/brainy/architecture/01-taste-evolution-model.md`
- `docs/brainy/architecture/02-belief-lifecycle.md`
- `docs/brainy/architecture/03-conflict-reconciliation.md`
- `.omx/plans/v1-contracts-mem0-go-rebuild.md`
- Linear: ENG-58 (marketing wedge), ENG-81 (brand voice), ENG-82 (eval fixtures), ENG-83 (campaign lifecycle)
