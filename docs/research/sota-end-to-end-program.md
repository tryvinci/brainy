---
title: Brainy SOTA End-to-End Program
status: program-of-record
adopted: 2026-08-03
supersedes: sota-assessment-and-action-plan.md (as execution guide; that file remains a short external briefing)
---

# Adoption note (Brainy internal)

This document is the **program of record** for coding-agent execution. It incorporates
external technical review of `sota-assessment-and-action-plan.md` and expands it into
an implementable architecture.

**Realistic sequencing for this repo (Go + Postgres, no graph DB):**

| Phase | Scope in this cycle | Notes |
| --- | --- | --- |
| 0 | Traces, failure taxonomy, oracle modes, baseline freeze | Required before claiming wins |
| 1 | Fusion V2 default-on, over-fetch, evidence-set selection, hot-path scan reduction | Highest ROI for LoCoMo |
| 2 | Immutable evidence shadow + bitemporal fields on atoms/assertions | Transitional: extend existing tables first |
| 3 | Events + procedures + profile projection (minimal viable) | Full schema from §9 lands incrementally |
| 4 | Intent classifier + typed recall tools + answer statuses | Planner starts deterministic |
| 5 | Packs v2 scaffolds (support/marketing) | Deepen before adding a third vertical |
| 6 | Re-measure OpMem, vertical, LoCoMo smoke/full | Gates are measured, not merge-complete |

**Rejected from earlier shorter plan (per §23):** tuning on LoCoMo validation convs 4–10 as product authoring; treating `valid_to=now()` as a complete temporal model; flat top-k as sufficient for multi-hop.

---

# Brainy: End-to-End Program to SOTA Conversational, Operational, and Vertical Memory

**Status:** Program of record and coding-agent execution guide  
**Audience:** Brainy coding agents, technical leads, external reviewers, and evaluators  
**Document date:** 2026-08-03  
**Internal measurement date:** 2026-08-01  
**Objective:** Implement Brainy end to end until it is competitive with the state of the art in conversational memory and demonstrates clear leadership in governed operational and vertical memory.

---

## 0. Directive to the coding agent

This document is not a design memo to summarize and stop at. It is the implementation program.

The coding agent is expected to:

1. inspect the current repository and confirm the actual code paths before changing them;
2. implement the architecture described here through small, reviewable changes;
3. preserve backwards compatibility or provide explicit migrations;
4. run the required tests and evaluations after every material change;
5. record measurements, regressions, failures, model versions, token budgets, and Git SHAs;
6. continue through the measurement–implementation–ablation loop until the SOTA qualification gates are met or a documented external constraint prevents further progress;
7. never improve a benchmark by introducing benchmark-specific product logic.

The agent must not interpret a phase as complete merely because code was merged. A phase is complete only when its exit criteria are measured and recorded.

The required end state is:

> Brainy is an immutable, evidence-grounded, bitemporal memory system with typed semantic projections, domain-governed state, query-specific evidence planning, bounded reasoning operations, and reproducible evaluation across conversational, operational, and vertical workloads.

---

# 1. Executive decision

Brainy should retain its Go-first, Postgres-backed architecture. It should **not** migrate to a dedicated graph database merely to resemble graph-memory systems.

The architecture does, however, require a deeper redefinition.

Brainy currently behaves primarily like a ranked collection of memory records with selected typed atoms and lifecycle operations. The SOTA target requires a complete memory reasoning path:

```text
reliable source capture
    → immutable evidence
    → typed semantic construction
    → entity and temporal resolution
    → materialized state
    → query analysis and planning
    → hybrid and structured retrieval
    → evidence-set construction
    → state/conflict resolution
    → answer or abstention with provenance
```

The central conclusion is:

> Fusion is necessary infrastructure. It is not the full architectural bet. The largest durable gains should come from combining reliable evidence, events, temporal state, query planning, complete evidence assembly, and vertical governance.

Brainy should pursue three separate leadership claims:

1. **Operational memory leadership:** truth-over-time, correction, suppression, isolation, provenance, rollback, and point-in-time state.
2. **Vertical memory leadership:** canonical domain models, state machines, authority rules, workflows, outcomes, and governed derivations.
3. **Conversational memory competitiveness:** credible performance on LoCoMo, LongMemEval, BEAM, and newer memory benchmarks under a pinned and reproducible stack.

---

# 2. Definition of SOTA for Brainy

“SOTA” must not mean one self-reported score on one benchmark.

Brainy qualifies as SOTA, or better, when it satisfies both of the following:

## 2.1 Quality qualification

For each claimed workload, Brainy must either:

- match or exceed the best reproducible system under an equivalent actor model, judge, data version, retrieval budget, and tool-call budget; or
- demonstrate a statistically credible Pareto advantage, such as comparable accuracy with materially lower latency, token usage, cost, or operational complexity.

A fixed score such as 75% is a **credibility floor**, not a SOTA definition.

## 2.2 Product qualification

Brainy must also demonstrate capabilities that generic conversational leaderboards undermeasure:

- current-state correctness;
- historical-state correctness;
- transition reconstruction;
- correction and supersession;
- negative and retracted facts;
- selective suppression and forgetting;
- source authority;
- tenant and object isolation;
- provenance and explainability;
- domain workflows and procedures;
- controlled derivation from outcomes to beliefs;
- low-latency non-generative recall.

## 2.3 Reporting requirements

Every published claim must include:

- Brainy commit SHA;
- dataset and fixture hash;
- model identifiers;
- embedding model;
- extraction model;
- answer model;
- judge model;
- seeds;
- top-k and candidate over-fetch;
- retrieval-token budget;
- maximum retrieval/tool calls;
- latency distribution;
- cost where measurable;
- whether the result uses managed, proprietary, or open components;
- known differences from competitor configurations.

Competitor headline scores are context, not direct evidence of superiority unless reproduced under comparable conditions.

---

# 3. Current position

The following Brainy measurements are internal results supplied as of 2026-08-01 and should be independently reproducible before being used publicly.

| Axis | Current Brainy signal |
|---|---:|
| LoCoMo, 1,540 questions, 3 seeds | approximately 49.8% mean |
| LoCoMo multi-hop | approximately 26% |
| LongMemEval-S, stratified 100 | approximately 4% |
| BEAM, one conversation at 100K, 20 probing questions | approximately 40% |
| OpMem | 13/13 |
| Marketing vertical | 15/16 |
| Support vertical | 3/3 |
| Search at concurrency 8 | p50 approximately 2.4 s; p95 approximately 5.0 s |

The current diagnosis of “coverage + long-haystack + fusion” is directionally correct, but incomplete.

The full failure set is likely:

1. **write failure:** the needed source was not captured correctly;
2. **construction failure:** the source exists but no usable assertion/event/procedure was created;
3. **identity failure:** the memory was attached to the wrong or unresolved entity;
4. **retrieval failure:** useful evidence exists but is not returned;
5. **coverage failure:** one relevant item is returned, but not the complete evidence set;
6. **temporal failure:** evidence is found, but prior/current/transition facts are mixed;
7. **conflict failure:** authoritative and contradictory sources are resolved incorrectly;
8. **planning failure:** the system uses similarity search for a task that requires enumeration, aggregation, or point-in-time lookup;
9. **reader failure:** sufficient evidence is present, but synthesis is incorrect;
10. **abstention failure:** missing or contradictory evidence is converted into an unsupported answer;
11. **judge or harness failure:** the system answer is reasonable but the evaluation path is wrong.

No large architectural decision should be justified solely by the final QA score before these stages are measured.

---

# 4. Non-negotiable product invariants

## 4.1 Anti-benchmax

Product code, extraction prompts, rankers, and domain packs must not contain:

- benchmark character names;
- benchmark-specific places or titles;
- benchmark-specific question templates;
- benchmark answer-key phrases;
- regexes derived from held-out benchmark examples.

General linguistic and domain patterns are valid.

The denylist should apply at minimum to:

```text
internal/memory/
internal/store/
packs/
prompts used by production extraction and recall
```

Benchmark adapters and evaluation code may reference benchmark names.

## 4.2 Evidence is never silently rewritten

Raw evidence is immutable except for explicit legal suppression, retention enforcement, or deletion policy.

Semantic interpretations can evolve. Their source evidence cannot be silently altered to match a later interpretation.

## 4.3 Every semantic object is grounded

Every assertion, event, relation, procedure, profile value, and derived belief must link to:

- one or more evidence records; or
- a deterministic system action whose inputs are themselves grounded.

## 4.4 History is preserved

Correction does not mean deleting history by default.

Brainy must distinguish:

- old state;
- current state;
- transition;
- contradicted claim;
- retracted claim;
- unknown state.

## 4.5 Derived intelligence is not observed truth

Profiles, summaries, and beliefs are projections. They must be rebuildable and provenance-complete.

## 4.6 Tenant and namespace boundaries are enforced below the LLM layer

The planner or model must never be responsible for security filtering. Every store operation must receive and enforce tenant, namespace, subject, and permission scope.

## 4.7 Ordinary recall is bounded

Query-time reasoning must have:

- maximum passes;
- maximum tool calls;
- maximum selected evidence tokens;
- deadlines and cancellation;
- deterministic fallback behavior.

---

# 5. Target architecture

Brainy should use one shared substrate for conversational, operational, and vertical memory.

## 5.1 The five planes

```text
┌────────────────────────────────────────────────────────────┐
│ 1. SOURCE PLANE                                            │
│ Chats, tools, APIs, documents, CRM/ticket/order events      │
└──────────────────────────┬─────────────────────────────────┘
                           ▼
┌────────────────────────────────────────────────────────────┐
│ 2. EVIDENCE PLANE                                          │
│ Immutable source-faithful records, scope, time, provenance  │
└──────────────────────────┬─────────────────────────────────┘
                           ▼
┌────────────────────────────────────────────────────────────┐
│ 3. SEMANTIC PLANE                                          │
│ Entities, assertions, events, relations, procedures,        │
│ contradictions, transitions, and derived beliefs            │
└──────────────────────────┬─────────────────────────────────┘
                           ▼
┌────────────────────────────────────────────────────────────┐
│ 4. PROJECTION PLANE                                        │
│ Current state, profiles, active objects, rollups, domain     │
│ views, event summaries, and search documents                │
└──────────────────────────┬─────────────────────────────────┘
                           ▼
┌────────────────────────────────────────────────────────────┐
│ 5. RECALL PLANE                                            │
│ Query analysis, planning, hybrid search, structured tools,   │
│ evidence packets, temporal resolution, answer/abstention     │
└────────────────────────────────────────────────────────────┘
```

## 5.2 Why this is one architecture

Conversational memory and vertical memory differ mainly in policy and schema, not in their correctness foundation.

Both require:

- faithful evidence;
- entity identity;
- temporal semantics;
- update and conflict rules;
- complete retrieval;
- provenance;
- answer-time resolution.

Vertical packs should extend these primitives with domain objects, mappings, authority, lifecycle, and workflows. They should not create a second memory engine.

---

# 6. Canonical memory model

Brainy should represent at least seven semantic categories.

## 6.1 Evidence

Evidence is the original source material:

- conversation turn;
- tool output;
- API event;
- structured record snapshot;
- imported document section;
- workflow trajectory observation.

Evidence properties:

- immutable content;
- source identity;
- tenant and namespace;
- subject/session;
- actor and role;
- occurred time;
- recorded time;
- source version;
- content hash;
- suppression state;
- searchable text and optional embedding.

## 6.2 Assertions

Assertions are claims about entities.

Examples:

- a user lives in Paris;
- a customer prefers email;
- a ticket has high priority;
- a campaign targets a particular audience.

Required fields:

```text
subject
predicate
object or typed value
assertion kind
confidence
world-valid interval
system-valid interval
observed time
source authority
status
provenance
```

Assertion kinds should include:

```text
explicit
observed
imported
inferred
derived
corrective
negative
```

## 6.3 Events

Events represent occurrences or transitions.

Examples:

- the user moved;
- a campaign launched;
- a ticket reopened;
- payment was received;
- the agent attempted a workflow and encountered an error.

Required fields:

```text
event type
participants and roles
start/end range
time precision
related entities
resulting state changes
confidence
provenance
```

Do not reduce event memory to session buckets. Session and time buckets may be indexing aids, not the event definition.

## 6.4 Relations

Relations connect canonical entities.

Examples:

- ticket belongs to customer;
- order is linked to ticket;
- asset belongs to campaign;
- outcome supports belief;
- procedure applies to product.

Relations may be temporal and must support provenance.

## 6.5 Procedures and experience

This is mandatory for vertical SOTA.

Procedure memory should include:

```text
workflow
prerequisite
policy
failure mode
environment gotcha
workaround
strategy
```

Procedures must retain where they came from and when they were last validated.

## 6.6 Derived beliefs

Beliefs are conclusions supported by evidence, such as:

- a creative approach appears to outperform another for a specific audience;
- a customer may be at escalation risk;
- a workflow tends to fail before a prerequisite is completed.

A belief must contain:

```text
scope
supporting evidence
contradicting evidence
confidence
derivation version
first-supported time
last-supported time
review or expiry time
```

A belief must never supersede an observed fact merely because it is newer.

## 6.7 Projections

Examples:

- current profile;
- current ticket state;
- active campaigns;
- user preference inventory;
- event rollup;
- outcome summary.

Projections are caches. They are rebuildable from semantic objects and evidence.

---

# 7. Temporal and state architecture

## 7.1 Full bitemporal model

A single `valid_to` column is insufficient.

For stateful assertions and temporal relations, Brainy should support:

```text
valid_from      # when the claim became true in the represented world
valid_to        # when it stopped being true in the represented world
recorded_at     # when Brainy learned or stored it
retired_at      # when Brainy stopped treating this version as live
observed_at     # when the source observation occurred
superseded_by
```

This permits two different queries:

- **world-time:** What was true on a particular date?
- **system-time:** What did Brainy believe on a particular date?

If Brainy learns today that a move happened three months ago, `valid_to = now()` is not correct.

## 7.2 State roles

The resolver should explicitly label evidence as:

```text
CURRENT_STATE
HISTORICAL_STATE
TRANSITION_EVENT
CONTRADICTED
RETRACTED
UNKNOWN
```

This follows the useful direction of state-aware memory research: bank maintenance, retrieval, and answer-time state resolution should be independently measurable.

## 7.3 Predicate-specific policy

Do not implement one universal “latest wins” rule.

Each predicate or relation should declare one of the following behaviors:

```text
SINGLE_CURRENT_STATE
TEMPORAL_STATE
APPEND_ONLY_EVENT
ACCUMULATING_SET
SCOPED_PREFERENCE
AUTHORITY_RESOLVED_RULE
DERIVED_BELIEF
```

Examples:

- residence: `TEMPORAL_STATE`;
- ticket status: `TEMPORAL_STATE`;
- attended conference: `APPEND_ONLY_EVENT`;
- languages spoken: usually `ACCUMULATING_SET`;
- communication preference: `SCOPED_PREFERENCE`;
- brand policy: `AUTHORITY_RESOLVED_RULE`.

## 7.4 Conflict policy

Conflict resolution should consider:

1. explicit correction;
2. source authority;
3. matching scope;
4. world-valid interval;
5. observation time;
6. system record time;
7. confidence;
8. corroboration;
9. negative or retraction status.

When uncertainty remains, preserve the conflict and return `conflicted` rather than inventing a winner.

---

# 8. Storage architecture

## 8.1 Remain on Postgres

Brainy should stay on Postgres for the primary roadmap.

Graph semantics do not imply a graph database. Postgres can support:

- typed entities and edges;
- temporal SQL;
- recursive CTEs;
- aggregations;
- JSONB pack extensions;
- full-text search;
- vector search;
- materialized projections;
- transactional provenance;
- tenant isolation.

## 8.2 Reconsider graph storage only when measured

A dedicated graph backend should be evaluated only if:

1. common production queries require deep variable-length traversal;
2. recursive SQL is a dominant measured bottleneck;
3. graph-specific indexing produces a material quality or latency gain;
4. consistency across stores can be guaranteed;
5. the additional operational burden is justified by customer or benchmark results.

This is a gated experiment, not a Phase 1 dependency.

---

# 9. Target logical schema

The exact migration should fit the existing repository. The following is the required logical shape.

## 9.1 `memory_evidence`

```sql
CREATE TABLE memory_evidence (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    namespace TEXT NOT NULL,
    subject_id UUID,

    source_type TEXT NOT NULL,
    source_ref TEXT,
    source_version TEXT,

    session_id TEXT,
    actor_id TEXT,
    actor_role TEXT,

    content TEXT NOT NULL,
    content_hash BYTEA NOT NULL,

    occurred_at TIMESTAMPTZ,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    language TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',

    suppression_status TEXT NOT NULL DEFAULT 'active',
    suppressed_at TIMESTAMPTZ,
    suppression_reason TEXT
);
```

Required indexes:

- tenant/namespace;
- subject/session;
- occurred time;
- source reference;
- content hash;
- FTS GIN;
- vector index where used.

## 9.2 `memory_entities`

```sql
CREATE TABLE memory_entities (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    namespace TEXT NOT NULL,

    canonical_name TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    vertical_type TEXT,

    external_ids JSONB NOT NULL DEFAULT '{}',
    attributes JSONB NOT NULL DEFAULT '{}',

    merge_status TEXT NOT NULL DEFAULT 'canonical',
    merged_into UUID REFERENCES memory_entities(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## 9.3 `memory_entity_aliases`

```sql
CREATE TABLE memory_entity_aliases (
    id UUID PRIMARY KEY,
    entity_id UUID NOT NULL REFERENCES memory_entities(id),

    alias TEXT NOT NULL,
    normalized_alias TEXT NOT NULL,
    alias_type TEXT,

    source_evidence_id UUID REFERENCES memory_evidence(id),
    confidence DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## 9.4 `memory_assertions`

```sql
CREATE TABLE memory_assertions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    namespace TEXT NOT NULL,

    subject_entity_id UUID REFERENCES memory_entities(id),
    predicate TEXT NOT NULL,

    object_entity_id UUID REFERENCES memory_entities(id),
    value_json JSONB,
    value_text TEXT,
    value_type TEXT NOT NULL,

    assertion_kind TEXT NOT NULL,
    confidence DOUBLE PRECISION NOT NULL,
    source_authority DOUBLE PRECISION,

    valid_from TIMESTAMPTZ,
    valid_to TIMESTAMPTZ,

    observed_at TIMESTAMPTZ,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    retired_at TIMESTAMPTZ,

    superseded_by UUID REFERENCES memory_assertions(id),
    status TEXT NOT NULL DEFAULT 'active',

    pack_id TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'
);
```

## 9.5 `memory_events`

```sql
CREATE TABLE memory_events (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    namespace TEXT NOT NULL,

    event_type TEXT NOT NULL,
    title TEXT,
    description TEXT,

    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    time_precision TEXT,

    location_entity_id UUID REFERENCES memory_entities(id),

    confidence DOUBLE PRECISION NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    pack_id TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'
);
```

## 9.6 `memory_event_participants`

```sql
CREATE TABLE memory_event_participants (
    event_id UUID NOT NULL REFERENCES memory_events(id),
    entity_id UUID NOT NULL REFERENCES memory_entities(id),
    participant_role TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    PRIMARY KEY (event_id, entity_id, participant_role)
);
```

## 9.7 `memory_relations`

```sql
CREATE TABLE memory_relations (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    namespace TEXT NOT NULL,

    source_entity_id UUID NOT NULL REFERENCES memory_entities(id),
    relation_type TEXT NOT NULL,
    target_entity_id UUID NOT NULL REFERENCES memory_entities(id),

    valid_from TIMESTAMPTZ,
    valid_to TIMESTAMPTZ,

    confidence DOUBLE PRECISION NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    retired_at TIMESTAMPTZ,

    pack_id TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'
);
```

## 9.8 `memory_procedures`

```sql
CREATE TABLE memory_procedures (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    namespace TEXT NOT NULL,

    procedure_type TEXT NOT NULL,
    title TEXT NOT NULL,
    instructions JSONB NOT NULL,

    applies_to_entity_id UUID REFERENCES memory_entities(id),
    applies_to_type TEXT,

    confidence DOUBLE PRECISION NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',

    first_observed_at TIMESTAMPTZ,
    last_validated_at TIMESTAMPTZ,
    review_after TIMESTAMPTZ,

    pack_id TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'
);
```

## 9.9 `memory_provenance`

```sql
CREATE TABLE memory_provenance (
    id UUID PRIMARY KEY,

    evidence_id UUID NOT NULL REFERENCES memory_evidence(id),

    assertion_id UUID REFERENCES memory_assertions(id),
    event_id UUID REFERENCES memory_events(id),
    relation_id UUID REFERENCES memory_relations(id),
    procedure_id UUID REFERENCES memory_procedures(id),

    extraction_method TEXT NOT NULL,
    extractor_version TEXT NOT NULL,

    source_span_start INTEGER,
    source_span_end INTEGER,
    confidence DOUBLE PRECISION,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

## 9.10 `memory_current_state`

```sql
CREATE TABLE memory_current_state (
    tenant_id UUID NOT NULL,
    namespace TEXT NOT NULL,
    subject_entity_id UUID NOT NULL,
    predicate TEXT NOT NULL,

    winning_assertion_id UUID NOT NULL,
    resolved_value JSONB NOT NULL,

    resolution_policy TEXT NOT NULL,
    resolved_at TIMESTAMPTZ NOT NULL,

    PRIMARY KEY (
        tenant_id,
        namespace,
        subject_entity_id,
        predicate
    )
);
```

This table is a rebuildable projection, never canonical truth.

---

# 10. End-to-end ingestion

## 10.1 Dual-speed write path

### Synchronous evidence path

The synchronous path must:

1. validate tenant, namespace, source, and actor;
2. normalize the source envelope;
3. persist immutable evidence;
4. calculate content hash;
5. make raw evidence searchable;
6. create a transactional outbox job;
7. return success.

It must not depend on an LLM completing.

### Asynchronous semantic path

The enrichment worker must:

1. claim a job using a lease;
2. segment the source when needed;
3. run deterministic extraction;
4. run provider extraction where configured;
5. resolve/create entity candidates;
6. extract assertions, events, relations, and procedures;
7. normalize temporal expressions;
8. validate semantic output;
9. write semantic records and provenance;
10. update search indexes and embeddings;
11. recompute affected projections;
12. mark the job complete.

If enrichment fails, the raw evidence remains queryable.

## 10.2 Exactly-once effects

Use:

- idempotency keys;
- content hashes;
- source identity;
- extractor versions;
- unique constraints;
- transactional outbox;
- leased jobs;
- retry counters;
- dead-letter status.

Reprocessing the same evidence with the same extractor version must not duplicate effects.

Reprocessing with a new extractor version should create a versioned projection or deliberately replace only derived projections.

## 10.3 Extraction contract

Provider extraction must produce strict structured output.

Illustrative shape:

```json
{
  "entities": [
    {
      "temporary_id": "e1",
      "name": "Acme",
      "type": "organization",
      "aliases": ["Acme Corp"],
      "external_ids": {},
      "confidence": 0.97
    }
  ],
  "assertions": [
    {
      "subject": "e1",
      "predicate": "support_tier",
      "value": "enterprise",
      "value_type": "string",
      "assertion_kind": "explicit",
      "valid_from": null,
      "valid_to": null,
      "evidence_spans": [{"start": 20, "end": 54}],
      "confidence": 0.94
    }
  ],
  "events": [],
  "relations": [],
  "procedures": []
}
```

The provider proposes candidates. Brainy validates and resolves them.

## 10.4 Deterministic versus model extraction

Use deterministic extraction where precision is high:

- dates and ranges;
- explicit identifiers;
- ticket/order/campaign IDs;
- quoted titles;
- currency and quantities;
- explicit pack enums;
- direct structured-source fields.

Use model extraction for:

- implicit events;
- nested subjects;
- coreference;
- temporal anchoring;
- semantic predicates;
- procedure/gotcha extraction;
- candidate entity links.

Both paths must merge through one typed contract.

## 10.5 Learned extraction is a later phase

Build a reviewed corpus first from:

- write misses;
- temporal mistakes;
- entity-resolution mistakes;
- real customer failures;
- difficult procedures;
- contradictory sources.

Only then train or fine-tune an extractor.

---

# 11. Entity resolution

Entity resolution is a subsystem, not a rank boost.

## 11.1 Resolution sequence

1. exact external identifier;
2. exact normalized alias within scope;
3. canonical name plus entity type;
4. alias plus contextual entities;
5. lexical/semantic candidate retrieval;
6. conservative merge classifier;
7. create a new entity when uncertain.

False merges are generally more damaging than temporary duplicates.

## 11.2 Reversible merges

Every merge must record:

- considered entities;
- matching signals;
- resolver version;
- confidence;
- evidence;
- human override.

Merges must be reversible.

## 11.3 Pack-level identifiers

Packs should declare authoritative IDs.

Example:

```yaml
support:
  customer:
    primary_id: crm_customer_id
  ticket:
    primary_id: ticket_id
  order:
    primary_id: order_id
```

Stable external IDs outrank name similarity.

---

# 12. Retrieval and recall architecture

## 12.1 Target path

```text
query
  → query analysis
  → retrieval plan
  → parallel candidate generation
  → score/rank fusion
  → evidence-set selection
  → temporal and conflict resolution
  → sufficiency check
  → optional bounded second pass
  → evidence packet
  → answer or abstention
```

Do not collapse these stages into one score.

## 12.2 Query intents

The analyzer should identify one or more intents:

```text
point_fact
current_state
historical_state
temporal_sequence
enumeration
aggregation
multi_hop
preference
procedure
outcome
belief
provenance
abstention_sensitive
```

It should also identify:

- entities;
- identifiers;
- predicates;
- date windows;
- requested output type;
- current versus historical view;
- expected evidence categories.

## 12.3 Retrieval plan

Illustrative plan:

```json
{
  "intents": ["enumeration", "multi_hop"],
  "entities": ["user"],
  "predicates": ["activity"],
  "operations": [
    {
      "tool": "list_assertions",
      "arguments": {
        "predicate": "activity",
        "current_only": false
      }
    },
    {
      "tool": "search_evidence",
      "arguments": {
        "query": "activities hobbies sports"
      }
    }
  ],
  "coverage_targets": [
    "all distinct activity values"
  ],
  "max_passes": 2
}
```

Start with deterministic planning and optional model assistance. Do not begin with an unconstrained agent loop.

## 12.4 Parallel candidate channels

### Exact channel

- entity IDs;
- ticket/order/campaign IDs;
- quoted phrases;
- external identifiers.

### Lexical channel

- Postgres FTS;
- phrase search;
- language-aware tokenization;
- BM25-like ranking.

### Dense channel

- semantic retrieval over evidence;
- assertions;
- events;
- procedures.

### Entity channel

- entity match;
- linked assertions;
- linked events;
- adjacent relations.

### Structured channel

- predicate lookup;
- current state;
- point-in-time state;
- temporal range query;
- SQL aggregation;
- constrained relation traversal.

### Domain channel

- pack-specific indexes;
- authoritative source mappings;
- lifecycle and isolation filters.

### Raw-corpus fallback

Exact grep/file-style retrieval should remain a baseline and a fallback for rare names, exact phrases, or evaluation diagnosis. Research shows that harness and tool-output design can matter as much as the nominal retriever.

## 12.5 Fusion

Implement an observable baseline:

1. over-fetch independently from channels;
2. apply channel quality gates;
3. normalize within each channel;
4. fuse with reciprocal-rank fusion or calibrated additive scoring;
5. add explicit bounded boosts for exact identifiers and authoritative sources;
6. log every score component.

Illustrative:

```text
final_score =
    semantic_weight  * semantic_score
  + lexical_weight   * lexical_score
  + entity_weight    * entity_score
  + temporal_weight  * temporal_fit
  + authority_weight * source_authority
  + exact_match_bonus
```

Entity or lexical boosts must not rescue completely irrelevant semantic candidates without explicit evidence.

## 12.6 Evidence-set selection

Top-k similarity is insufficient for lists and multi-hop questions.

The selector should maximize marginal coverage of:

- distinct predicates;
- distinct entities;
- distinct time windows;
- distinct event types;
- uncovered subquestions;
- source authority.

It should penalize:

- duplicate text;
- same-session repetition;
- semantically redundant evidence;
- repeated proof of an already-covered fact.

A greedy set-cover implementation is appropriate initially:

```text
while budget remains:
    choose candidate maximizing
        marginal coverage
      + relevance
      + authority
      + temporal fit
      - redundancy
```

## 12.7 Temporal and conflict resolver

The resolver should:

1. separate state assertions from events;
2. order evidence by world time;
3. distinguish record time from valid time;
4. identify transition events;
5. apply predicate and source-authority rules;
6. preserve unresolved conflicts;
7. emit current, historical, and transition labels.

Illustrative output:

```json
{
  "status": "resolved",
  "current_value": "Paris",
  "previous_values": [
    {
      "value": "London",
      "valid_from": "2024-01-01",
      "valid_to": "2025-04-10"
    }
  ],
  "supporting_assertions": ["a2"],
  "superseded_assertions": ["a1"],
  "transition_events": ["e9"],
  "conflicts": []
}
```

## 12.8 Sufficiency

Before generation, verify:

- all subquestions have evidence;
- list scope was exhaustively scanned;
- result limits did not truncate the set;
- temporal ordering is available;
- conflicts are resolved or exposed;
- the answer is permitted by suppression policy.

Allow one additional retrieval pass only when a declared coverage target is unmet.

## 12.9 Internal tools

Provide constrained typed operations:

```text
search_evidence
search_assertions
search_events
search_procedures
lookup_entity
get_entity_history
get_current_state
get_state_as_of
list_by_predicate
list_related_entities
aggregate_events
trace_provenance
resolve_conflict
```

Do not expose unrestricted model-generated SQL in the first implementation.

## 12.10 Evidence packet

The answerer should receive a structured packet, not a flat top-k list:

```text
question interpretation
current-state evidence
historical-state evidence
transition events
procedures/gotchas
unresolved conflicts
source authority
coverage status
provenance
```

## 12.11 Answer status

Support at least:

```text
supported
partially_supported
conflicted
not_found
insufficient_evidence
suppressed
```

Abstention is a correct product behavior, not a failure by default.

---

# 13. Vertical packs v2

Packs must become executable domain models.

## 13.1 Pack responsibilities

A pack should define:

- entity types;
- canonical identifiers;
- aliases;
- predicates;
- events;
- relations;
- state machines;
- temporal behavior;
- source mappings;
- source authority;
- retrieval policy;
- derivation rules;
- retention and governance;
- fixtures and evaluation.

## 13.2 Suggested structure

```text
packs/support/v2/
  pack.yaml
  entities.yaml
  predicates.yaml
  events.yaml
  relations.yaml
  state-machines.yaml
  source-mappings.yaml
  resolution.yaml
  retrieval.yaml
  derivations.yaml
  governance.yaml
  fixtures/
```

## 13.3 Three schema levels

### Core

Universal:

- entity;
- assertion;
- event;
- relation;
- procedure;
- time;
- provenance.

### Vertical pack

Domain-wide:

- support ticket;
- campaign;
- order;
- escalation;
- creative asset.

### Customer extension

Customer-specific:

- custom states;
- internal codes;
- proprietary categories;
- source fields.

Customer extensions must not require forking core code.

## 13.4 Support pack v2

Canonical objects:

```text
customer
account
ticket
order
product
incident
agent
resolution
escalation
```

Events:

```text
ticket_created
ticket_assigned
status_changed
customer_replied
agent_replied
order_shipped
order_cancelled
incident_reported
ticket_reopened
ticket_resolved
```

Stateful predicates:

```text
ticket_status
priority
assigned_agent
escalation_level
customer_tier
order_status
```

Experience:

```text
resolution_workflow
escalation_rule
system_gotcha
required_prerequisite
known_workaround
```

Required fixtures:

- same-name customer isolation;
- ticket reopening;
- late-arriving state update;
- order–ticket linking;
- CRM versus agent-note conflict;
- effective-date escalation rule;
- provenance of resolution;
- sensitive-field suppression;
- account isolation;
- procedure recall.

## 13.5 Marketing pack v2

Canonical objects:

```text
brand
campaign
audience
asset
channel
experiment
creative_variant
metric
outcome
rule
belief
```

Events:

```text
campaign_launched
campaign_paused
asset_published
experiment_started
experiment_ended
budget_changed
creative_approved
performance_observed
```

Stateful predicates:

```text
campaign_status
budget
target_audience
active_asset
brand_rule
approval_status
```

Outcome-to-belief derivation must preserve:

- supporting outcomes;
- contradicting outcomes;
- confidence;
- audience/channel scope;
- first/last support;
- expiry/review date;
- derivation version.

Required fixtures:

- conflicting brand rules;
- active versus archived policy;
- channel-scoped preference;
- contradictory experiments;
- asset–campaign linking;
- brand isolation;
- time-bounded guidance;
- unsupported belief abstention.

## 13.6 Procedures are a primary vertical moat

Before creating a third shallow pack, deepen support and marketing with:

- workflows;
- prerequisites;
- failure modes;
- environment gotchas;
- workarounds;
- strategy notes.

This aligns better with enterprise agents than only adding another domain noun taxonomy.

---

# 14. API evolution

## 14.1 Preserve public simplicity

Brainy may retain:

```text
POST /ingest
POST /ingest/async
GET /memories/search
POST /recall
```

The internal contracts should become richer.

## 14.2 Proposed recall response

```json
{
  "answer": "...",
  "answer_status": "supported",
  "intent": ["historical_state"],
  "evidence": [
    {
      "evidence_id": "...",
      "source_ref": "...",
      "occurred_at": "...",
      "excerpt": "..."
    }
  ],
  "resolved_facts": [],
  "transitions": [],
  "conflicts": [],
  "coverage": {
    "targets": 2,
    "satisfied": 2
  },
  "retrieval": {
    "passes": 1,
    "candidates_considered": 84,
    "items_selected": 7,
    "tokens": 2310
  }
}
```

## 14.3 Temporal request options

Add explicit options such as:

```json
{
  "query": "...",
  "view": "current",
  "as_of": null,
  "include_history": false,
  "include_provenance": true,
  "max_evidence_tokens": 4000
}
```

Supported views:

```text
current
historical
system_as_known
trajectory
all
```

---

# 15. Operations and reliability

## 15.1 Worker lifecycle

Implement:

- durable jobs;
- lease expiration;
- worker heartbeat;
- graceful shutdown;
- retry policy;
- dead-letter state;
- provider rate limits;
- extraction timeout;
- circuit breakers.

On SIGTERM:

1. stop claiming jobs;
2. complete or release leases;
3. flush metrics;
4. exit within the platform grace period.

## 15.2 Index management

Never create expensive indexes during ordinary application boot.

Use controlled migrations for:

- FTS GIN;
- vector indexes;
- temporal indexes;
- foreign-key indexes.

For large indexes:

- create concurrently where available;
- measure disk headroom;
- expose progress;
- run during a controlled window;
- verify plans using `EXPLAIN (ANALYZE, BUFFERS)`.

## 15.3 Latency SLOs

Separate retrieval from generation.

| Stage | Initial gate | Mature target |
|---|---:|---:|
| Indexed candidate retrieval p50 | <500 ms | <250 ms |
| Indexed candidate retrieval p95 | <1 s | <500 ms |
| Non-generative context recall p95 | <1.5 s | <750 ms |
| Generative recall | report separately | model-dependent |
| Async enrichment lag | measured SLO | workload-specific |

A 2.4-second p50 for search indicates a hot-path or query-plan problem and must not be normalized as an acceptable target.

## 15.4 Observability

Every recall trace should include:

```text
query analysis
retrieval plan
channel timings
candidate counts
score components
selected evidence
coverage calculation
state resolution
answer model
token usage
final latency
```

Required metrics:

- ingest p50/p95/p99;
- enrichment lag;
- extraction failures;
- semantic objects per evidence record;
- entity merge rate;
- false-merge review rate;
- search latency by channel;
- evidence recall;
- selected context tokens;
- answer latency;
- abstention rate;
- stale-state rate;
- conflict rate;
- projection rebuild lag.

---

# 16. Evaluation architecture

## 16.1 Stage-level truth

For every benchmark or replay question, determine:

1. Does the raw source exist?
2. Is the correct semantic object present?
3. Is the entity correct?
4. Is the object retrievable?
5. Is the complete evidence set assembled?
6. Is state resolution correct?
7. Is synthesis correct?
8. Is abstention correct?
9. Is the judge correct?

## 16.2 Failure taxonomy

```text
SOURCE_MISS
WRITE_MISS
REPRESENTATION_MISS
ENTITY_LINK_MISS
RETRIEVAL_MISS
EVIDENCE_COVERAGE_MISS
TEMPORAL_RESOLUTION_MISS
CONFLICT_RESOLUTION_MISS
PLANNING_MISS
READER_MISS
ABSTENTION_MISS
JUDGE_MISS
HARNESS_ERROR
```

Record one primary and optional secondary causes.

## 16.3 Oracle modes

### Oracle reader

Give exact gold evidence to the answerer.

Measures reader and judge ceiling.

### Oracle raw evidence

Confirm that the source text exists after ingestion.

Measures source/write correctness.

### Oracle semantic representation

Confirm that the correct assertion/event/procedure exists.

Measures extraction coverage.

### Oracle retrieval

Supply the correct semantic objects to state resolution and synthesis.

Measures resolver and reader quality.

## 16.4 Metrics

### Retrieval

- evidence recall@5/10/30/50/100;
- complete-evidence-set recall;
- selected token count;
- redundant-evidence ratio;
- per-channel contribution.

### Enumeration

- item precision;
- item recall;
- exact-set match;
- truncation rate.

### Temporal

- current-state accuracy;
- historical-state accuracy;
- trajectory accuracy;
- transition detection;
- stale-state rate.

### Safety and governance

- abstention precision/recall;
- false-premise correction;
- suppression leakage;
- tenant-isolation failures;
- provenance completeness.

### Operations

- write latency;
- enrichment lag;
- retrieval latency;
- full recall latency;
- LLM calls;
- tokens;
- cost.

## 16.5 Benchmark portfolio

### Conversational

- LoCoMo;
- LongMemEval V1;
- BEAM;
- LoCoMo-Plus where stable.

### State and failure analysis

- MemTrace-style knowledge-point probes;
- A-TMA-style current/historical/trajectory probes;
- Brainy state-transition mutations.

### Agent experience

- LongMemEval-V2;
- workflow recall;
- environment gotchas;
- dynamic state;
- premise awareness.

### General memory capabilities

- MemoryAgentBench;
- EvoMemBench or equivalent;
- selective forgetting;
- contradiction resolution;
- test-time learning where relevant.

### Brainy-specific

- expanded OpMem;
- SupportBench;
- MarketingBench;
- real anonymized customer replays;
- tenant isolation;
- provenance;
- suppression;
- rollback/version consistency.

### Required baselines

- full context;
- BM25/FTS;
- dense RAG;
- hybrid RAG;
- filesystem/grep or coding-agent retrieval;
- Brainy ablations;
- at least two external memory systems.

## 16.6 Scale-conditioned testing

Do not evaluate only one corpus size.

For a fixed relevant evidence set, progressively add irrelevant history and measure:

- evidence recall;
- answer accuracy;
- calls;
- tokens;
- latency.

Report Brainy’s **usable-scale boundary** under fixed budgets.

## 16.7 Holdout policy

The current policy should be corrected and made explicit:

- development/tuning: LoCoMo conversations 1–3 only, plus generalized non-benchmark fixtures;
- validation: conversations 4–10, used sparingly and logged;
- publication: full set, multi-seed;
- product prompts and rules may never be authored from held-out examples.

Prefer growing a separate development corpus so that even conversations 1–3 do not become the product specification.

---

# 17. Implementation roadmap

Durations are planning ranges, not permission to stop early. The coding agent should proceed based on dependencies and measured evidence.

## Phase 0 — Baseline freeze and diagnostic truth

**Typical duration:** 1–2 weeks

### Implement

- reproduce all current Brainy numbers;
- tag baseline commit;
- archive model/config/data hashes;
- add recall trace envelope;
- implement failure taxonomy;
- implement oracle-reader mode;
- implement oracle-evidence mode;
- implement oracle-semantic mode;
- verify full-context, FTS, dense, hybrid, and grep baselines;
- manually adjudicate a stratified failure sample;
- determine why LongMemEval is approximately 4%.

### Exit

- at least 95% of sampled failures have defensible primary labels;
- oracle reader establishes a meaningful ceiling;
- harness errors are separated from product failures;
- all baseline commands are reproducible;
- no architecture claim depends only on aggregate QA accuracy.

### Important

Fusion work may begin in parallel once traces exist. Event/extraction work must not be blocked by an arbitrary LoCoMo score if diagnosis shows representation failure dominates.

---

## Phase 1 — Reliability, indexing, and modern retrieval

**Typical duration:** 1–2 weeks

### Implement

- controlled FTS GIN migration;
- vector-index verification;
- query-plan analysis;
- remove ordinary full-subject scans;
- exact, lexical, dense, entity, and structured channels in parallel;
- over-fetch;
- transparent rank fusion;
- exact-ID bonuses;
- optional reranking;
- dynamic retrieval budgets;
- evidence-set selector;
- feature-flagged rollout;
- latency and token traces.

### Exit

- indexed retrieval p50 <500 ms and p95 <1 s at target load;
- mature target trajectory documented;
- evidence recall improves at equal token budget;
- no full subject scan in normal paths;
- OpMem and vertical suites do not regress;
- ablations identify which channels add value.

### Score expectation

A LoCoMo lift is expected, but Phase 1 is not considered failed solely because an arbitrary target such as 60% is missed. Failure attribution determines the next move.

---

## Phase 2 — Immutable evidence and bitemporal substrate

**Typical duration:** 3–4 weeks

### Implement

- `memory_evidence`;
- source hashes and idempotency;
- full provenance;
- bitemporal assertions;
- bitemporal relations where relevant;
- predicate policy registry;
- temporal supersession/retirement;
- current-state projection;
- current and point-in-time reads;
- transactional outbox;
- leased workers;
- additive backfill;
- compatibility adapters for current `memory_records` and `memory_atoms`.

### Exit

- raw evidence is searchable immediately;
- semantic enrichment may fail without losing source access;
- every semantic record has provenance;
- current and historical state are deterministic;
- late-arriving corrections pass;
- reprocessing is idempotent;
- existing API contracts are compatible or versioned.

---

## Phase 3 — Typed event, relation, profile, and procedure memory

**Typical duration:** 3–4 weeks

### Implement

- first-class events;
- event participants and roles;
- time-range normalization;
- transition generation;
- temporal event search;
- entity relations;
- procedure/gotcha storage;
- current profile projection;
- profile provenance;
- backfill for evaluation histories;
- extraction coverage metrics;
- provider and deterministic extraction through one contract.

### Exit

- event extraction precision/recall measured;
- profiles are fully traceable and rebuildable;
- procedures are distinct from facts;
- temporal and multi-session questions improve;
- LME failures shift measurably away from representation misses;
- no profile value exists without source support.

---

## Phase 4 — Query planner, evidence packets, and state resolver

**Typical duration:** 3–4 weeks

### Implement

- deterministic intent classifier;
- optional model-assisted plan generation;
- typed internal tools;
- coverage targets;
- evidence-set construction;
- temporal resolver;
- conflict resolver;
- one bounded second pass;
- answer statuses;
- provenance in recall response;
- abstention behavior;
- query budgets and deadlines.

### Exit

- enumeration optimizes complete sets;
- multi-hop tracks subquestion coverage;
- temporal questions use structured state resolution;
- aggregate questions use constrained SQL;
- ordinary point facts remain low latency;
- tool, token, and pass budgets are enforced;
- unsupported answers abstain reliably.

---

## Phase 5 — Vertical packs v2

**Typical duration:** 4–6 weeks, parallel with Phase 4 where safe

### Implement

- pack v2 schema;
- support canonical model;
- support state machine;
- marketing canonical model;
- source mappings;
- source authority;
- domain retrieval policy;
- procedures and gotchas;
- outcome-to-belief derivation;
- customer-extension mechanism;
- expanded fixtures.

### Exit

- support passes at least 10 serious fixtures;
- marketing passes at least 16 existing plus expanded outcome/conflict fixtures;
- customer extensions require configuration, not a core fork;
- vertical answers expose provenance;
- multi-vendor comparison runs under equal fixtures;
- OpMem expands beyond the original 13 cases.

---

## Phase 6 — Neutral proof, scale, and SOTA qualification

**Typical duration:** 2–3 weeks, then continuous

### Implement

- full LoCoMo, multi-seed;
- LongMemEval full set;
- BEAM at multiple scales;
- LongMemEval-V2;
- state-aware and knowledge-point probes;
- current external baselines under the same stack;
- Brainy adapter for neutral harnesses;
- latency/cost reporting;
- statistical comparisons;
- public reproducibility package.

### Credibility floor

- LoCoMo and LongMemEval at or above 75 under the pinned stack;
- no major collapse in multi-hop or temporal categories;
- operational and vertical suites fully green.

### SOTA gate

Brainy must satisfy at least one:

1. equal or higher accuracy than the strongest comparable reproducible system at equal or lower resource budget;
2. accuracy within a small statistically defensible band while materially improving latency, tokens, cost, governance, or operational correctness;
3. clear leadership on operational and vertical suites while remaining within the credible conversational frontier.

Do not claim SOTA based on mismatched proprietary headline scores.

---

## Phase 7 — Associative reachability research

This is a gated experiment after the core is stable.

### Hypothesis

Similarity-based retrieval misses memories connected to a future query only by a latent semantic arc.

### Experiment

Add write-time retrieval cues at two levels:

- atomic item;
- scene/episode.

Cue families may include:

- entity cue;
- semantic bridge;
- likely future-use context;
- horizon/commitment cue.

### Gate

Ship only if the method improves held-out associative recall without:

- unacceptable false positives;
- large write cost;
- material latency increase;
- benchmark-specific cues.

---

## Phase 8 — Learned memory policy research

After deterministic actions and traces are stable, evaluate learning:

- extraction policy;
- memory CRUD decisions;
- retrieval planning;
- budget allocation;
- retention and consolidation;
- procedure induction.

Prerequisites:

- explicit tools;
- attributable rewards;
- leak-audited datasets;
- deterministic fallbacks;
- safe rollout.

Do not make learned policy a dependency for the first SOTA-capable release.

---

# 18. Transitional migration plan

The coding agent should not replace the current system in one change.

## 18.1 Stage 1 — Shadow evidence

- retain existing `memory_records`;
- write new evidence records in parallel;
- verify count, hash, and source fidelity;
- no read-path change.

## 18.2 Stage 2 — Shadow semantic projections

- write assertions/events/provenance alongside atoms;
- compare semantic coverage;
- run dual-read diagnostics;
- no user-visible behavior change.

## 18.3 Stage 3 — Hybrid reads

- enable new retrieval channels under tenant/benchmark flags;
- compare old and new candidate sets;
- log disagreement.

## 18.4 Stage 4 — State resolver

- use new temporal resolution for selected predicates;
- retain old supersession as fallback;
- expand predicate by predicate.

## 18.5 Stage 5 — Projection and pack migration

- build current-state and profile projections;
- migrate support and marketing to pack v2;
- preserve v1 compatibility during rollout.

## 18.6 Stage 6 — Default and cleanup

Only remove old paths when:

- backfill is complete;
- rollback is tested;
- old/new disagreement is understood;
- benchmark and production non-regression gates pass.

---

# 19. Initial engineering work breakdown

## Measurement

- **MEM-001:** Recall trace envelope
- **MEM-002:** Question-level failure taxonomy
- **MEM-003:** Oracle reader
- **MEM-004:** Oracle evidence and semantic modes
- **MEM-005:** Baseline and artifact manifest
- **MEM-006:** Scale-conditioned evaluation runner

## Reliability and retrieval

- **MEM-010:** Controlled FTS GIN migration
- **MEM-011:** Parallel candidate channels
- **MEM-012:** Score attribution
- **MEM-013:** Evidence-set selector
- **MEM-014:** Dynamic retrieval budgets
- **MEM-015:** Remove hot-path subject scans
- **MEM-016:** Exact/grep fallback

## Evidence and time

- **MEM-020:** Immutable evidence store
- **MEM-021:** Transactional outbox
- **MEM-022:** Leased jobs and graceful shutdown
- **MEM-023:** Bitemporal assertion fields
- **MEM-024:** Predicate temporal-policy registry
- **MEM-025:** Current-state resolver
- **MEM-026:** Point-in-time reads
- **MEM-027:** Provenance constraints

## Semantic memory

- **MEM-030:** Entity aliases and external IDs
- **MEM-031:** Reversible entity merge
- **MEM-032:** Event schema and participants
- **MEM-033:** Event extraction and temporal anchoring
- **MEM-034:** Relation storage
- **MEM-035:** Procedure/gotcha memory
- **MEM-036:** Rolling profile projection
- **MEM-037:** Extraction-version backfill

## Query controller

- **MEM-040:** Intent classifier
- **MEM-041:** Typed recall operations
- **MEM-042:** Structured plan
- **MEM-043:** Coverage targets
- **MEM-044:** Temporal/conflict resolver
- **MEM-045:** Bounded second pass
- **MEM-046:** Evidence packet
- **MEM-047:** Answer statuses and abstention

## Vertical packs

- **MEM-050:** Pack v2 contract
- **MEM-051:** Support canonical entities
- **MEM-052:** Support state machine
- **MEM-053:** Support procedure memory
- **MEM-054:** Marketing canonical entities
- **MEM-055:** Outcome-to-belief derivation
- **MEM-056:** Source mapping and authority
- **MEM-057:** Customer extension layer
- **MEM-058:** Expanded vertical benchmark

## Proof

- **MEM-060:** Neutral harness adapter
- **MEM-061:** Multi-scale BEAM
- **MEM-062:** LongMemEval-V2 integration
- **MEM-063:** State-aware probe suite
- **MEM-064:** Public artifact generator
- **MEM-065:** Comparable external baseline runs

---

# 20. Architecture decision records

Create and maintain:

```text
ADR-001 Canonical evidence and memory planes
ADR-002 Bitemporal semantics
ADR-003 Append-only evidence and semantic retirement
ADR-004 Postgres graph-shaped model
ADR-005 Entity resolution and reversible merge
ADR-006 Query planner and bounded tools
ADR-007 Evidence-set selection
ADR-008 Vertical pack v2 contract
ADR-009 Provenance and derived beliefs
ADR-010 Dual-speed ingestion and outbox
ADR-011 Benchmark reproducibility
ADR-012 Learned memory-policy boundary
```

---

# 21. Feature flags

Suggested flags:

```text
BRAINY_TRACE_V2
BRAINY_FUSION_V2
BRAINY_EVIDENCE_STORE
BRAINY_BITEMPORAL_ASSERTIONS
BRAINY_EVENT_MEMORY
BRAINY_PROCEDURE_MEMORY
BRAINY_PROFILE_PROJECTION
BRAINY_QUERY_PLANNER
BRAINY_SECOND_PASS_RETRIEVAL
BRAINY_PACKS_V2
BRAINY_ASSOCIATIVE_TRIGGERS
```

Flags should support:

- tenant rollout;
- evaluation rollout;
- comparative tracing;
- fast rollback.

---

# 22. Pull-request discipline

Good PR boundaries:

- add evidence schema;
- add one bitemporal migration;
- implement parallel candidate generation;
- add event participants;
- implement current-state resolver;
- implement support state machine.

Bad boundaries:

- rewrite ingest, schema, ranker, and answer prompt together;
- change benchmark data while changing product logic;
- tune multiple unexplained weights in one commit;
- merge a profile system without provenance;
- introduce a graph service before a measured need.

For each substantive PR:

1. state the hypothesis;
2. identify the affected failure class;
3. include tests;
4. include latency/storage impact;
5. run operational non-regression;
6. run a small evaluation;
7. attach an ablation when relevant.

---

# 23. Adjudication of external suggestions

## Accepted for implementation now

- default-on hybrid retrieval;
- over-fetch and budget control;
- exact/lexical/dense/entity complementarity;
- immutable source evidence;
- typed assertions and events;
- complete bitemporal state;
- event and raw-evidence dual retrieval;
- canonical vertical schemas;
- source mappings and authority;
- procedures and gotchas;
- bounded multi-tool recall;
- neutral baselines and third-party harnesses.

## Accepted with modification

### “Retrieval first”

Implement retrieval early, but do not gate representation work on an arbitrary score. Stage-level diagnosis controls ordering.

### “ADD-only extraction”

Apply append-only semantics to evidence and historical events. Assertions may be retired, superseded, contradicted, or retracted while history remains preserved.

### “Rolling profile”

Use only as a versioned, provenance-complete materialized projection.

### “GraphSQL-lite”

Use constrained typed operations before unrestricted SQL.

### “Latest wins”

Use predicate-specific temporal and authority policy, not universal recency.

## Deferred research

- associative write-time triggers;
- learned CRUD/memory policies;
- dedicated graph database;
- unrestricted model-generated SQL;
- autonomous unbounded retrieval agents.

## Rejected

- benchmark-specific extractor patterns;
- tuning on validation conversations 4–10;
- considering `valid_to = now()` a complete temporal model;
- accepting a flat top-k list as sufficient for multi-hop/list questions;
- declaring SOTA from unmatched headline scores.

---

# 24. Research reading map

Use primary sources. For each source, record what to borrow, what to test, and what not to assume.

## 24.1 Modern retrieval baseline

### Mem0 repository and benchmark notes

Use for:

- managed-system benchmark context;
- single-pass retrieval budgets;
- hybrid retrieval pragmatics;
- external harness integration.

Do not assume managed-platform scores reproduce in OSS.

- https://github.com/mem0ai/mem0
- https://github.com/mem0ai/memory-benchmarks

## 24.2 Temporal event retrieval

### Chronos: Temporal-Aware Conversational Agents with Structured Event Retrieval for Long-Term Memory

Use for:

- raw-turn plus event-calendar architecture;
- normalized event tuples;
- date-range filtering;
- query-specific retrieval instructions;
- tool-based multi-hop retrieval;
- ablation design.

- https://arxiv.org/abs/2603.16862

## 24.3 Semi-structured and SQL-assisted reasoning

### APEX-MEM

Use for:

- append-only temporal history;
- entity-centric events;
- search plus structured operations;
- retrieval-time conflict resolution.

Do not copy high call budgets without cost/latency ablation.

- https://arxiv.org/abs/2604.14362

## 24.4 State-aware failure decomposition

### A-TMA

Use for:

- bank/retrieval/answer decomposition;
- current/historical/transition labels;
- evidence packets;
- ghost-memory analysis.

- https://arxiv.org/abs/2607.01935

## 24.5 Knowledge-point evaluation

### MemTrace: Probing What Final Accuracy Misses

Use for:

- evaluating facts across memory age;
- current/earlier/trajectory questions;
- missing and false-premise conditions;
- separating retrieval from evidence use.

- https://arxiv.org/abs/2606.17328

## 24.6 Atomic facts, events, and temporal profiles

### AtomMem: Building Simple and Effective Memory System for Personalized LLM Agents

Use for:

- value-dense facts;
- event hierarchy;
- temporal profiles;
- associative recall;
- extraction-model research.

- https://arxiv.org/abs/2606.19847
- https://github.com/MINE-USTC/AtomMem

### TriMem: Beyond Atomic Facts

Use for:

- preserving raw evidence;
- atomic facts;
- synthesized profiles;
- multi-granularity representation.

- https://arxiv.org/abs/2605.19952

## 24.7 Associative reachability

### T-Mem

Use for:

- descriptive versus associative retrieval;
- item and scene granularity;
- write-time trigger hypotheses.

Treat as a gated experiment because generated cues may add noise and cost.

- https://arxiv.org/abs/2606.15405

## 24.8 Learned memory operations

### AtomMem: Learnable Dynamic Agentic Memory with Atomic Memory Operation

Use for:

- CRUD as explicit memory actions;
- learned policy research;
- future task-conditioned memory behavior.

- https://arxiv.org/abs/2601.08323
- https://github.com/RUCBM/AtomMem

## 24.9 Experienced-agent memory

### LongMemEval-V2

Use for:

- static and dynamic state;
- workflow knowledge;
- environment gotchas;
- premise awareness;
- context-gathering evaluation;
- accuracy/latency trade-offs.

- https://arxiv.org/abs/2605.12493

## 24.10 Agent harness and corpus interaction

### Is Grep All You Need?

Use for:

- grep/exact-search baseline;
- harness sensitivity;
- inline versus file-based tool output;
- scale-conditioned distraction tests.

- https://arxiv.org/abs/2605.15184

### ByteRover

Use for:

- human-readable hierarchical memory;
- provenance and lifecycle;
- progressive retrieval;
- agent-curated memory as a competing architecture.

Do not replace Brainy’s governed service architecture with files merely because the benchmark result is strong.

- https://arxiv.org/abs/2604.01599

## 24.11 Semantic versioning and rollback

### ChronoMem

Use as a research reference for:

- whole-memory versions;
- rollback intent;
- post-exposure rollback evaluation.

Brainy should first implement bitemporal semantic history; snapshot rollback is a separate product capability.

- https://arxiv.org/abs/2607.27773

---

# 25. Coding-agent execution loop

The coding agent must repeat this loop until the SOTA gates are met:

```text
1. select the dominant measured failure class
2. state a general product hypothesis
3. implement the smallest architecture-consistent change
4. run unit and integration tests
5. run OpMem and vertical non-regression
6. run oracle and stage metrics
7. run a held-out evaluation slice
8. compare accuracy, evidence coverage, tokens, and latency
9. keep, revise, or revert
10. record result and update the program status
```

Do not move directly from “score did not improve” to rank-weight tuning.

When a result fails:

- if source is absent, fix ingestion;
- if source exists but semantics are absent, fix construction;
- if semantics exist but retrieval fails, fix channels/indexing;
- if items are found but the set is incomplete, fix evidence selection/planning;
- if state is wrong, fix temporal/conflict resolution;
- if evidence is sufficient, fix reader or judge;
- if everything is correct but cost is high, optimize the plan.

---

# 26. Definition of done

The program is not done when the new tables or planner exist.

It is done when all of the following are true.

## Architecture

- immutable evidence is the canonical source layer;
- semantic objects are typed and provenance-complete;
- stateful facts use full bitemporal semantics;
- events, relations, and procedures are first-class;
- projections are rebuildable;
- entity merges are reversible;
- Postgres reads are indexed and bounded;
- query planning is typed and budgeted;
- recall produces evidence packets and answer statuses.

## Operational correctness

- expanded OpMem is fully green;
- current, historical, and transition queries pass;
- corrections and late-arriving facts pass;
- suppression and tenant isolation have zero known leaks;
- provenance is returned and auditable.

## Vertical leadership

- support and marketing packs implement canonical models, state, source authority, workflows, and derivations;
- customer extensions do not fork core;
- multi-vendor vertical evaluation demonstrates a material advantage.

## Conversational frontier

- credibility floor is cleared under a pinned stack;
- multi-hop, temporal, and enumeration categories no longer collapse;
- Brainy is equal to or better than the strongest comparable system, or is Pareto superior on accuracy/resources;
- results are reproducible outside the authors’ ad hoc notebook.

## Performance

- indexed retrieval meets stated SLOs;
- non-generative recall meets stated SLOs;
- generation latency is reported separately;
- no routine full-corpus scan exists;
- ingestion remains available when enrichment is degraded.

## Research integrity

- no benchmark-specific product logic;
- holdout use is logged;
- all public runs include model and budget disclosure;
- negative ablations and regressions are preserved;
- the final SOTA claim is narrowly scoped to what the evidence supports.

---

# 27. Final mandate

The coding agent should implement this program as a continuous engineering and evaluation effort, not as a one-time rewrite.

The intended end state is not merely “Brainy with better retrieval.”

It is:

> A governed memory system that preserves source truth, constructs typed and temporal knowledge, understands domain state and procedures, retrieves complete evidence according to the question, resolves historical and current truth correctly, and returns an answer—or an abstention—with inspectable provenance.

That architecture is the shared foundation required to pursue:

- SOTA conversational memory;
- better-than-SOTA operational correctness;
- category leadership in vertical memory;
- production-grade latency, governance, and explainability.

The agent should continue iterating against the measured frontier until those outcomes are demonstrated.
