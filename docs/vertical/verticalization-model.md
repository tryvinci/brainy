# Verticalization Model

**Status:** Approved (ENG-61, 2026-06-19)  
**Decision:** General verticalization framework — not per-vertical DB kinds  
**Last updated:** 2026-06-19

## Problem

Two bad options:

| Approach | Failure mode |
|---|---|
| **Per-vertical schema** | `brand_rule`, `campaign`, `thesis`, `market_event` as DB enums → combinatorial sprawl; finance and marketing share nothing |
| **Generic memory clone** | `profile` / `preference` / `fact` only → no vertical moat; ranking and lifecycle stay dumb |

We need a **general platform** that verticalizes through configuration and primitives, not through forked schemas.

---

## Core thesis

> **Verticals are packs, not code paths.**

The runtime is domain-agnostic. A vertical pack declares how domain language maps onto cognitive primitives, what metadata is valid, how retrieval ranks, and which eval scenarios prove correctness.

Marketing is the **first pack**, not a special schema.

---

## Three layers

```
┌─────────────────────────────────────────────────────────┐
│  Layer 3: Domain vocabulary (vertical pack)             │
│  marketing: "brand_rule"  finance: "policy"             │
│  → labels only; not storage enums                       │
├─────────────────────────────────────────────────────────┤
│  Layer 2: Vertical pack (declarative config)            │
│  metadata schemas, rank policy, lifecycle, extraction,  │
│  eval fixtures, governance defaults                     │
├─────────────────────────────────────────────────────────┤
│  Layer 1: Cognitive primitives (stable runtime)         │
│  Principle, IdentityPrior, Episode, Pattern, Belief,    │
│  Outcome, Experiment, TasteSignal, Reflection           │
└─────────────────────────────────────────────────────────┘
```

---

## Layer 1 — Runtime (build once)

### Primitive field on every memory record

Extend `memory_records` with:

| Field | Purpose |
|---|---|
| `primitive` | Cognitive primitive type (enum, stable) |
| `vertical` | Pack ID, e.g. `marketing`, `finance`, `core` |
| `label` | Domain term from pack vocabulary, e.g. `brand_rule` |
| `scope` | Optional entity scope, e.g. `campaign:summer-2026` |
| `metadata` | JSONB validated against pack schema |
| `lifecycle_state` | Generic state machine (see below) |
| `conviction` | For Belief primitive (nullable) |

Retire widening the `kind` enum per vertical. Keep `kind` temporarily for thin-slice migration (`profile` → IdentityPrior or profile primitive, etc.).

### Generic lifecycle state machine

One machine, pack configures transitions and retrieval effects:

```
active → deprioritized → suppressed
       → superseded (links supersedes_id)
       → archived (excluded from default search)
```

Marketing "campaign ended" and finance "earnings revised" are **pack-defined triggers** on the same machine — not separate status enums.

### Generic rank pipeline

```
1. Filter: tenant, subject, vertical, lifecycle (exclude suppressed/archived)
2. Scope boost: active scope entities (campaign, filing period, etc.)
3. Primitive precedence: Principle > IdentityPrior > Belief > Pattern > Episode
4. Conviction / recency / semantic score blend (pack-weighted)
5. TasteSignal tie-break (pack optional)
```

Finance and marketing differ in **weights and scope rules**, not in ranker implementation.

### Generic extraction pipeline

```
ingest → deterministic baseline (always) → optional provider enrich
       → classify primitive (pack rules + prompts)
       → validate metadata (pack JSON Schema)
       → persist
```

---

## Layer 2 — Vertical pack (configure per domain)

A pack is a versioned bundle (YAML/JSON in repo: `packs/marketing/v1/`):

```yaml
id: marketing
version: "1.0.0"

vocabulary:
  brand_rule:    { primitive: principle,      kind: fact }
  voice_profile: { primitive: identity_prior, kind: preference }
  campaign:      { primitive: episode,        kind: fact, scoped: true }
  audience_segment: { primitive: episode,     kind: profile, scoped: true }
  performance_outcome: { primitive: outcome,   kind: fact }

metadata_schemas:
  campaign:
    type: object
    required: [name, status, start_date, end_date]
    properties:
      name: { type: string }
      status: { enum: [draft, active, paused, completed, archived] }
      start_date: { type: string, format: date }
      end_date: { type: string, format: date }

lifecycle_rules:
  - when: "metadata.status == 'archived'"
    then: { lifecycle_state: archived, exclude_from_search: true }
  - when: "metadata.status == 'completed'"
    then: { lifecycle_state: deprioritized, rank_multiplier: 0.5 }

rank_policy:
  primitive_weights:
    principle: 100
    identity_prior: 80
    belief: 60
    pattern: 40
    episode: 20
  scope_boost:
    active_campaign: 1.5

extraction:
  deterministic_rules: [...]   # regex/heuristics for CI
  provider_prompt: "..."       # optional LLM prompt template

eval_fixtures: fixtures/vertical/marketing/

governance:
  default_tags: []
  retention_days: null
```

Finance pack (`packs/finance/v1/`) swaps vocabulary and schemas — same runtime:

```yaml
vocabulary:
  policy:        { primitive: principle, kind: fact }
  thesis:        { primitive: belief,    kind: fact }
  market_event:  { primitive: episode,   kind: fact, triggers_invalidation: true }
  entity:        { primitive: episode,   kind: profile, scoped: true }
```

---

## Layer 3 — Domain vocabulary (labels only)

Vertical docs and UIs use domain terms. Storage uses `primitive` + `label` + `metadata`.

| Marketing label | Finance label | Primitive |
|---|---|---|
| `brand_rule` | `policy` | Principle |
| `voice_profile` | `risk_profile` | IdentityPrior |
| `campaign` | `filing_period` | Episode (scoped) |
| `performance_outcome` | `market_outcome` | Outcome |
| `content_belief` | `thesis` | Belief |
| `audience_segment` | `entity` | Episode (scoped) |

Same retrieval semantics. Different words in eval fixtures and customer-facing API.

---

## What we do NOT verticalize

Keep these universal — do not fork per vertical:

| Concern | Reason |
|---|---|
| Ingest / async / worker / DLQ | Already built; vertical-agnostic |
| Dedupe / idempotency | Universal |
| Correct / suppress / lineage | Universal |
| Postgres + pgvector | Universal |
| Observability / audit event shape | Universal |
| Tenant + subject isolation | Universal |

---

## What each vertical pack owns

| Concern | Pack responsibility |
|---|---|
| Domain → primitive mapping | `vocabulary` |
| Typed entity metadata | `metadata_schemas` |
| Lifecycle triggers | `lifecycle_rules` |
| Rank weights and scope boosts | `rank_policy` |
| Extraction prompts/rules | `extraction` |
| Golden eval scenarios | `eval_fixtures` |
| Compliance defaults | `governance` |

---

## API shape

### Ingest (unchanged envelope, pack-aware validation)

```json
{
  "tenant_id": "t1",
  "subject_id": "brand-acme",
  "vertical": "marketing",
  "messages": [...],
  "source_type": "conversation"
}
```

Pack selected by `vertical` field (default `core` for thin-slice compat).

### Search (pack-aware ranking)

```
GET /memories/search?tenant_id=t1&subject_id=brand-acme&vertical=marketing&q=...
```

Optional: `scope=campaign:summer-2026` to boost/filter scoped memories.

### Pack registry (operator)

```
GET /packs                    → list registered packs
GET /packs/marketing/v1       → pack manifest (schemas, vocabulary)
```

No CRUD for packs at runtime in MVP — packs are versioned repo artifacts, loaded at startup.

---

## Migration from current thin slice

Current Go model has `kind: profile|preference|fact` only.

| Step | Change |
|---|---|
| 1 | Add `primitive`, `vertical`, `label`, `metadata`, `lifecycle_state` columns (nullable) |
| 2 | Default `vertical=core`, infer primitive from kind for existing records |
| 3 | Load marketing pack v1; new ingests with `vertical=marketing` use pack |
| 4 | Rank pipeline reads primitive precedence when `primitive` is set; falls back to kind |
| 5 | Deprecate widening `kind` enum; document mapping in parity matrix |

No breaking change to existing parity fixtures (`vertical=core` or omitted).

---

## MVP scope (marketing pack v1)

Ship the **framework minimally** with one pack — do not over-build the registry.

| Ship | Defer |
|---|---|
| `primitive` + `vertical` + `label` + `metadata` on records | Dynamic pack CRUD API |
| Static marketing pack v1 in `packs/marketing/v1/` | Finance pack (Phase 2) |
| Primitive precedence in ranker | Full Belief lifecycle |
| Generic lifecycle states | Event webhook invalidation |
| Marketing eval fixtures | Multi-pack CI matrix |
| Finance pack | **Gate M4** — see marketing vetting gate |

### Vetting policy

Marketing must pass **Gate M3** (technical proof) before finance or any second vertical. ENG-93 (MVP benchmark) is **Gate M1** only — not sufficient to start finance.

Canonical policy: [`docs/vertical/marketing-vetting-gate.md`](./marketing-vetting-gate.md)

---

## How marketing use case map fits

`docs/vertical/marketing-use-case-map.md` describes **agent jobs and failure modes** for the marketing pack. Domain terms there (`brand_rule`, `campaign`, etc.) are **pack labels**, not proposed DB enums.

ENG-81 (brand voice) becomes: IdentityPrior + Principle behavior in the rank pipeline, validated by marketing pack rules — not a `voice_profile` table.

Spec: `docs/vertical/marketing-brand-voice-spec.md` (ENG-81, approved 2026-06-19).

ENG-82 (eval fixtures) become: marketing pack golden scenarios against the general runtime.

---

## Decision record (ENG-61) — approved 2026-06-19

| Option | Verdict |
|---|---|
| A. Extend `kind` enum per vertical | **Rejected** — schema sprawl |
| B. Metadata overlay on 3 kinds only | **Rejected** — loses primitive semantics and rank policy |
| C. Hybrid kinds + subtype | **Rejected** — still verticalizes the wrong layer |
| D. Cognitive primitives + vertical packs | **Approved** — one runtime, many domains |

First pack: `packs/marketing/v1/pack.yaml`

---

## Open questions

1. ~~**Pack format:** YAML in repo vs JSON vs embedded Go~~ — **Resolved:** YAML in repo (`packs/{vertical}/v1/pack.yaml`)
2. **Primitive enum stability:** All 9 primitives in v1, or ship 4 (Principle, IdentityPrior, Belief, Episode) first?
3. **Scope syntax:** `scope` as opaque string vs structured `{type, id}`?
4. **Cross-vertical memories:** Can one subject span packs (agency with finance + marketing clients)?

---

## Roadmap & Linear issue updates (2026-06-19)

Decisions locked since project creation. Update Linear manually if MCP sync failed.

### Approved decisions

| Issue | Status | Decision |
|---|---|---|
| ENG-58 | Done | Marketing first vertical wedge |
| ENG-61 | Done | Primitives + YAML packs (not per-vertical DB kinds) |
| ENG-80 | Done | Use case map → `docs/vertical/marketing-use-case-map.md` |
| Marketing pack v1 | Done | `packs/marketing/v1/pack.yaml` |

### Issues to reframe (title / scope change)

| Issue | Old framing | New framing |
|---|---|---|
| **ENG-85** | Vertical memory schema extension (kinds, metadata, entities) | **Verticalization runtime skeleton** — add `primitive`, `vertical`, `label`, `metadata`, `lifecycle_state`; pack loader; no new `kind` enums |
| **ENG-81** | Brand voice memory model | **Principle + IdentityPrior rank behavior** — pack rules + rank pipeline; labels `brand_rule`, `voice_profile` stay in YAML only |
| **ENG-83** | Campaign lifecycle semantics | **Pack lifecycle_rules** — generic lifecycle engine applies rules at ingest + search — **done** |
| **ENG-76** | Finance memory taxonomy proposal | **Research only** — no `packs/finance/` until marketing **Gate M3** (`marketing-vetting-gate.md`) |
| **ENG-56** epic | Finance discovery (equal priority) | **Blocked at Gate M4** — research notes OK; implementation after marketing technical proof |
| **ENG-82** | Marketing golden eval fixtures | Fixtures validate **pack + runtime**, not marketing-specific schema — **done** (BV-01–BV-10) |
| **ENG-90** | Vertical eval harness | Run pack evals from `eval_fixtures` path in pack YAML — **done** via `TestVerticalEvalHarnessAgainstHTTPServer` in CI |
| **ENG-93** | Vertical MVP | **Marketing pack MVP on general runtime** — benchmark in `evals/run_marketing_mvp_benchmark.py` |

### Issues unchanged

| Issue | Notes |
|---|---|
| ENG-91 | Push dev + CI — still first engineering task |
| ENG-87 | pgvector + hybrid retrieval — unchanged |
| ENG-92 | Provider extraction — unchanged |
| ENG-86 | Temporal invalidation — generic runtime; finance-heavy but useful for campaign end dates |
| ENG-59 | Temporal model PD — still open; pack lifecycle_rules are interim |
| ENG-62 | Belief scope PD — still open |
| ENG-63–67 | Embedding, extraction, tenancy, compliance, API PDs — still open |

### Revised delivery phases

```
Phase 0 — Publish thin slice (now)
  ENG-91: push dev, CI, PR
  Update parity matrix ✅

Phase 1 — Verticalization skeleton (MVP-1) ✅ skeleton landed
  ENG-85: primitive/vertical/label/metadata columns + pack loader
  Wire rank_policy.primitive_weights from pack
  Ingest accepts vertical=marketing

Phase 2 — Brand voice behavior (MVP-1 cont.)
  ENG-81: Principle > IdentityPrior precedence in ranker
  Suppression leak tests for brand_rule label

Phase 3 — Lifecycle engine (MVP-2) ✅
  ENG-83: generic lifecycle_state machine driven by pack lifecycle_rules
  Campaign active/archived retrieval behavior

Phase 4 — Evals (MVP-5) ✅
  ENG-82: fixtures/vertical/marketing/
  ENG-90: CI integration (go test ./... runs vertical eval e2e)

Phase 5 — Semantic retrieval (feeds Gate M3)
  ENG-87: pgvector (after PD ENG-63)

Phase 6 — Marketing MVP benchmark ✅ (Gate M1)
  ENG-93: benchmark report (`docs/vertical/marketing-mvp-benchmark.md`)

Phase 6b — Marketing technical proof (Gate M3) — **before finance**
  Tier 4 eval seeds, MVP-3/4, semantic non-regression
  Policy: `docs/vertical/marketing-vetting-gate.md`

Phase 7 — Finance (Gate M4 only)
  ENG-56/76/78: finance pack + evals — blocked until M3 clears

Go-to-market (OSS, benchmarks, commercial): `docs/vertical/go-to-market-roadmap.md`
```

### Cancelled / rejected approaches

- Per-vertical `kind` enum (`brand_rule`, `thesis`, etc. in Postgres)
- Marketing-specific Go code paths
- Finance as first wedge
- Parallel finance + marketing build for MVP

---

## References

- `docs/brainy/architecture/00-cognitive-primitives.md`
- `docs/vertical/marketing-use-case-map.md` (domain discovery)
- `.omx/plans/` historical rebuild contracts
- Linear: ENG-58 (marketing first pack), ENG-81 (`docs/vertical/marketing-brand-voice-spec.md`)
