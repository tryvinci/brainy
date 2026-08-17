# Brainy competitive parity and implementation-gap review

**Date:** 2026-08-17  
**Status:** archived source (current-SHA deep-research). Live adjudication: [2026-08-17-parity-gap-verdict.md](./2026-08-17-parity-gap-verdict.md).  
**Primary brief:** 2026-08-17 full LoCoMo `/recall` dip self-review prompt  
**Current docs SHA:** `8492ad3`  
**Measured product SHA:** `1b5ab3e`

---

## Executive summary

Brainy should **keep the representation-first course, but adjust the immediate sequence**.

The deep comparison against the current Brainy product code, the August qualification artefacts, Mem0's current OSS implementation and 2026 technical material, Graphiti's source, and the LoCoMo/LongMemEval papers produces a more precise diagnosis than “Brainy needs Mem0 features” or “Brainy needs a graph”.

The central finding is that **Brainy has already implemented more of the competitor checklist than its own gap documents sometimes imply**:

- Conversational/core writes are already append-oriented rather than governed UPDATE/DELETE semantics. `write_policy.go` explicitly separates core conversational behaviour from governed vertical mutation.
- Dense, lexical/BM25-like, entity and temporal retrieval signals already exist; `ScoreAndRankV2Temporal` is not merely dense retrieval.
- The internal candidate pool is not simply 30: the default output is 30, but `CandidatePoolSize` typically overfetches to 120 and caps at 200.
- `EvidencePacket` already separates `ContextEvidence` and `ProofChain`, although the context representation remains largely string-oriented and legacy packet fields still influence answering.
- A Postgres relation store already exists. It is a **v1 string-edge representation**, not yet a canonical entity/relation graph.
- LME publish integrity is substantially hardened: product `/recall`, async completion, exact completed-vs-expected jobs, zero failures and queue preconditions are now enforced.

The real gaps are narrower and more consequential:

1. **Brainy's product answer path is currently much worse than the information surface beneath it.** On the current frozen product SHA, full LoCoMo via product `POST /recall` is 175/1540, or 11.4%, whereas the older search→shared-LLM-answerer path was 49.4% on seed 0 at the same 1540-question scale. The current full run shows especially severe single-hop performance—88/841, or 10.5%—and concrete failures where `/recall` returns a slogan, a random first packet statement, an enumerated word salad, or abstains instead of answering. `recall.go` confirms a deterministic fall-through that can use `firstStatementFromPacket`, enumeration and abstention before or around hybrid composition.

2. **The semantic compiler remains incomplete across diverse conversations.** The current 1×30 head can reach 21/30 and 10/10 MH, but it remains 0/4 open-domain, while full LoCoMo collapses outside that heavily developed conversation; LME-20 is only 4/20 and multi-session is still 0/5. Those measurements strongly imply that the R0–R4 architecture can work where facts/entities/relations are successfully compiled, but coverage does not yet generalise.

3. **Canonical identity is the largest structural representation gap.** Brainy has entity links and a hub, but `entities.go` is fundamentally a string-key extraction/linking system. Graphiti has explicit UUID-addressed `EntityNode`s, while current Mem0 extracts entities at memory and query time and uses them directly as a retrieval feature.

4. **Brainy's relation model is one generation behind Graphiti's.** Brainy's `MemoryRelation` holds string `SrcEntity`, relation, string `DstEntity`, `MemoryID` and `ObservedAt`; Graphiti's `EntityEdge` has stable entity IDs, an explicit fact, embedding, provenance episode IDs, `valid_at`, `invalid_at`, `expired_at` and `reference_time`.

5. **The typed hop executor has the correct shape but weak identity semantics.** `resolveEntityHop` still effectively resolves to mention strings, relation lookup works over those strings, predicate lookup can fall back from entity-scoped to unscoped state, and atom lookup may retain predicate matches even when entity filtering finds no match. This is enough for a favourable small conversation, but not enough for general entity joins.

6. **Brainy should not copy Mem0 blindly.** Mem0's current architecture is particularly instructive because it combines recent conversational context, relevant existing memories, single-call ADD-only extraction, entity extraction, batch embedding and multi-signal retrieval. Its OSS code exposes these mechanisms directly. Mem0's self-reported 2026 managed results—92.5 LoCoMo and 94.4 LongMemEval at roughly 6.8–7.0K retrieved tokens—remain a product-quality bar, not a directly comparable Brainy pin.

7. **Brainy should copy Graphiti's semantics, not its database.** Graphiti's most valuable ideas are source episodes, canonical entities, relation edges, temporal validity, provenance and typed search over nodes/edges; none requires Brainy to abandon Postgres.

The programme I recommend is therefore:

> **Fix structured-first product answering → increase compiler coverage and provenance precision → canonicalise entities → upgrade relations to canonical temporally-valid edges → make hops join canonical IDs → run frozen dual-path qualification.**

The strategic target remains:

> **Mem0-quality high-recall conversational facts + Graphiti-quality entity/relation structure + Brainy-quality provenance, temporal truth, governed operational state and vertical semantics.**

That is a credible differentiation. A graph-database migration, fusion retune, category dictionary or reader-prompt sprint is not.

---

## Current competitive position

| Dimension | Brainy at current product SHA | Mem0 current architecture | Graphiti/Zep | Adjudication |
|---|---|---|---|---|
| Conversational ingestion | Core/conversational append semantics already exist; governed domains can UPDATE/DELETE/suppress. | 2026 algorithm explicitly single-pass ADD-only; agent-generated facts are first-class. | Incremental episodes; changed relations are invalidated rather than history being erased. | **Near parity in policy; gap is extraction coverage, not ADD-vs-UPDATE.** |
| Evidence/provenance | Raw evidence plane, IDs and temporal storage are strong; direct source-span lineage remains incomplete. | Memory-centric; OSS mainly persists memory payload plus metadata/linking. | Explicit episodes; edges retain episode IDs. | **Brainy strength, but add spans and first-class semantic→evidence join rows.** |
| Entity model | `memory_entity_links`/hub and extracted string keys; no durable canonical identity object. | Entity candidates extracted and used in query ranking; current OSS entity extraction is more systematic than Brainy's name regexes. | Native `EntityNode` with UUID, labels, summary, attributes and embedding. | **Major gap.** |
| Relation model | `memory_relations` exists, but endpoints are normalised strings with limited temporal/provenance structure. | 2026 OSS emphasises entity linking rather than requiring a graph subsystem. | Relation edges are first-class with canonical endpoints, validity, reference time, provenance and fact embeddings. | **Moderate-to-major gap.** |
| Temporal representation | Strong bitemporal/current-state foundations; generic facts are not yet uniformly enriched with full event/state/plan/ongoing precision fields. | Each memory can get a time signature: occurrence time, ongoing/completed, precision and memory kind; temporal relevance then reranks. | Relations carry `valid_at`, `invalid_at`, `expired_at`, `reference_time`; episodes carry `valid_at`. | **Brainy is stronger at governed state, weaker at uniform fact-level temporal enrichment.** |
| Retrieval signals | Dense + lexical + entity + temporal already exist. | OSS combines semantic, normalised BM25 and entity boost; managed system adds temporal reasoning. | BM25, cosine, BFS and several rerankers over edges/nodes/episodes/communities. | **Signal parity is not the bottleneck; retrieval-unit quality is.** |
| Candidate budget | Default final `topK=30`, default evidence budget 4K tokens; internal pool normally ~120 and capped at 200. | Published 2026 results use top-200 scoring and ~6.8–7K mean context. | Search config defaults vary by recipe; graph retrieval is multi-surface rather than one flat top-k. | **Measure fixed-token broader recall; do not merely increase top-k.** |
| Packet | `ContextEvidence` and `ProofChain` already exist, plus legacy contents/items. | Memories are assembled for an answerer rather than exposed as Brainy-style proof semantics. | Search returns typed graph units; lineage remains attached to them. | **Brainy architectural advantage, but context must become structured objects rather than strings.** |
| Reader contract | Hybrid JSON contract carries answer, supporting IDs, unresolved targets and abstention; deterministic product path still produces severe answer-quality loss. | Benchmark path is memory retrieval followed by answer generation; published results are not equivalent to Brainy's product `/recall`. | Context retrieval engine, not directly analogous to Brainy's product reader. | **Immediate Brainy gap.** |
| LME harness | Product-recall flag, async barrier, queue precheck and fail-closed publish are implemented. | Published LME 94.4 is Mem0's own current harness/product stack, 500 questions. | Published Zep evidence is a distinct evaluation stack. | **Integrity largely solved; quality is not.** |
| LoCoMo comparability | Current honest product metric is `/recall` 11.4%; historical search+harness is a different path. | Self-reported 92.5 uses 1540 questions/top-200/current Mem0 harness. | Zep numbers use their evaluation methodology. | **Maintain product metric and add a separately labelled industry-format comparator.** |
| Observability | Search trace and a broad failure taxonomy already exist. | OSS exposes score details optionally. | Zep managed product advertises production debug/API tooling; OSS leaves surrounding tooling to users. | **Good foundations; missing stage-quality and cost telemetry.** |

The benchmark papers support this decomposition. LoCoMo is deliberately long-range and tests temporal and causal dependencies across conversations averaging roughly 300 turns and 9K tokens, while LongMemEval separates extraction, multi-session reasoning, temporal reasoning, knowledge updating and abstention—meaning retrieval quality cannot be diagnosed from an aggregate accuracy alone.

---

# Competitive gap analysis and findings

## The current two-gap thesis is broadly correct, but its numerical ceiling is not proven

The August remasure's proposed decomposition—first recover an answer-path gap from 11.4% towards the old ~50% search+harness level, then solve a representation gap towards competitor-class performance—is **directionally correct but should be stated more carefully**.

The current 11.4% product `/recall` and July 49.4% search+harness seed-0 numbers are not the same product stack, so 49.4% is **not a measured oracle ceiling for current storage**. A current-SHA search+harness run is required to measure that ceiling.

Nevertheless, the answer-path problem is real rather than speculative. Full LoCoMo has 841 single-hop questions and product `/recall` answers only 88 correctly; the handoff documents cases where the gold-bearing material exists but `/recall` emits an unrelated first packet sentence, a malformed enumeration or abstention.

The implementation supports that diagnosis: `Recall` builds an evidence packet, then deterministic answering includes temporal composition, enumeration, multi-hop composition and ultimately `firstStatementFromPacket`; the default top-K is 30 and evidence budget 4,000 tokens.

That makes **structured-first product answering the first PR**, not because representation is solved, but because an answerer that fails on already-retrieved structured facts prevents us from observing the true value of subsequent representation work.

---

## The competitor advantage is semantic-unit quality, not a missing retrieval formula

Current Mem0 is particularly instructive.

Its OSS `main.py` obtains the last ten session messages, retrieves relevant existing memories, then uses a single additive extraction call; extracted memory texts are batch-embedded afterwards.

Its additive prompt explicitly frames extraction as ADD-only and supports existing/recent memories and memory linking.

At search time, its OSS scoring combines vector similarity, normalised BM25 and entity boost.

Mem0's own 2026 documentation attributes large benchmark gains to:

- single-pass ADD-only extraction;
- first-class agent memories;
- entity linking;
- multi-signal retrieval;
- temporal reasoning.

Brainy already possesses versions of those mechanics.

The difference is that Brainy's retrieval corpus still contains heterogeneous `memory_records`, atoms and provenance episodes, with incomplete semantic compilation across conversations.

The accepted Brainy PoR correctly recognises the representation problem and records that R0–R4 have now landed, including:

- fact-primary recall;
- compiler coverage work;
- relation projection;
- relation-aware hopping.

But full-suite performance shows those capabilities are not yet uniformly effective outside the small conversation where MH reached 10/10.

---

## The Graphiti lesson is stable identity plus temporal relations

Graphiti is not mainly valuable because it uses a graph database.

Its code separates `EpisodicNode` from `EntityNode`:

- episodes contain raw source content and validity time;
- entities have stable UUIDs, names, labels, embeddings, summaries and attributes.

Its `EntityEdge` stores:

- source and target UUIDs;
- relation name;
- natural-language fact;
- embedding;
- provenance episode IDs;
- `valid_at`;
- `invalid_at`;
- `expired_at`;
- `reference_time`.

Graphiti's own architecture describes precisely this model: entities, relationship facts with validity windows, and raw episodes as provenance, with semantic, keyword and graph traversal at retrieval.

Brainy's `memory_relations` is already graph-shaped, but only superficially: its endpoints are strings.

That makes the real R2/R3 follow-on work obvious:

> **Upgrade the IDs and lifecycle semantics, do not replace Postgres.**

---

## The entity gap contaminates the hop layer

The strongest code-level gap is between the word “typed” and actual identity semantics.

In `hop_executor.go`:

- resolving an entity still rests on query mentions/entity-hub results rather than a canonical entity object;
- relations are queried by normalised strings;
- predicate lookup can fall back from entity-scoped state to unscoped state;
- the atom path can retain all predicate hits if the entity-filtered subset is empty.

Graphiti avoids that ambiguity because relations point to stable source/target UUIDs.

Mem0's lighter-weight approach is different but still more integrated than Brainy's legacy entity hub: current OSS has a dedicated entity extraction module using NER, proper-noun, quoted, topic and technical-identifier candidates, with filtering and confidence concepts.

Canonical identity therefore has a compounding payoff:

- retrieval accuracy;
- relation correctness;
- proof-chain validity.

All improve from one representation change.

---

## Findings adjudication

| Finding | Verdict | Code/source evidence | Concrete action |
|---|---|---|---|
| “Brainy needs ADD-only conversational ingestion.” | **Modify** | `write_policy.go` already makes core conversational memory append-oriented while preserving governed mutations. | Keep the policy split. Improve extraction coverage rather than reopening UPDATE/DELETE semantics. |
| “Brainy lacks Mem0-style contextual extraction.” | **Modify** | Brainy has `contextual_extractor.go` and provider context handling; Mem0 additionally uses last-session messages plus relevant existing memories in a single additive extraction pipeline. | Benchmark contextual compiler coverage, then adapt missing context inputs/link semantics—not wholesale replacement. |
| “The next major issue is the reader.” | **Accept for the immediate PR; reject as the full programme** | `/recall` is 11.4% full LoCoMo, with 10.5% single-hop and explicit malformed answer examples. `recall.go::firstStatementFromPacket` is a concrete mechanism. | Land structured-first answering first, then return to compiler/entity/relation coverage. |
| “Brainy lacks hybrid retrieval.” | **Reject** | `fusion_v2.go::ScoreAndRankV2Temporal` already combines semantic, BM25, entity and temporal features. | Do not fusion-fish. Measure recall per signal and representation type. |
| “Brainy retrieves only top 30 while Mem0 retrieves 200.” | **Modify** | Brainy's final default limit is 30 but candidate pool is normally 120 and capped at 200. Mem0 reports top-200 benchmark scoring and ~7K context. | Separate `candidate_limit`, final context budget and proof budget; run a fixed-token ablation. |
| “Brainy lacks evidence/proof packet separation.” | **Reject** | `planner.go::EvidencePacket` already contains `ContextEvidence` and `ProofChain`. | Upgrade context entries from strings to typed evidence objects and remove legacy contents from answer logic. |
| “Brainy lacks relation memory.” | **Reject literally; accept quality gap** | `relations.go::MemoryRelation` and migration v20 are present. | Build relation V2 with canonical endpoint IDs, validity and evidence spans. |
| “Brainy has canonical entities.” | **Reject** | `entities.go` centres on extracted string keys and linked memory IDs, not identity records/aliases/merge state. | Add canonical entity, alias and mention tables. |
| “Typed hops are now competitor-grade.” | **Reject** | `hop_executor.go` still resolves and joins on mention/string semantics and has scoped→unscoped fallbacks. | Make canonical entity IDs the hop I/O type; unscoped/fuzzy evidence may inform context but cannot constitute exact proof. |
| “Brainy's temporal model is behind Mem0.” | **Modify** | Brainy has stronger current-state/bitemporal foundations; Mem0 has more uniform conversational-memory time signatures and temporal reranking. | Preserve state architecture; add uniform fact/relation temporal fields and scoring inputs. |
| “LME measurement integrity remains the main blocker.” | **Reject** | Current `run.py` and `require_pins` enforce `/recall`, async ingestion, queue state and jobs-completed equality. | Keep integrity gates. LME's current problem is 4/20 product quality, not the old barrier defect. |
| “We can directly compare Brainy 11.4 to Mem0 92.5 as a bake-off.” | **Reject** | Brainy 11.4 is product `/recall`; Mem0's 92.5 is its own top-200 search/answer evaluation. | Publish separately labelled product and industry-format rows. |
| “Observability is already sufficient to optimise the system.” | **Modify** | Brainy has a broad failure taxonomy and SearchTrace. | Add semantic-coverage precision/recall, per-stage latency/tokens and actual gold-object presence. |
| “A graph database is required to close MH.” | **Reject** | Graphiti's value comes from typed entity/edge semantics; nothing in Brainy's logical model requires abandoning Postgres. | Implement graph-shaped relational tables in Postgres first. |

---

# Target technical design

The intended architecture should now make semantic units—not transcript text—the default answer surface while preserving episodes as the audit substrate.

```text
Source interaction
        ↓
Immutable evidence / episode
        ↓
Semantic compiler
   ├── Atomic facts
   ├── Canonical entities + aliases
   ├── Temporal relations
   └── Events / temporal features
        ↓
Governed projections
        ↓
Candidate retrieval
        ↓
┌──────────────────────────────┐
│ Evidence context             │
│ broad bounded evidence       │
└──────────────────────────────┘
        +
┌──────────────────────────────┐
│ Proof chain                  │
│ canonical joins              │
└──────────────────────────────┘
        ↓
Structured-first answer
        ↓
Supported / partial / conflicted / insufficient
```

This remains consistent with Brainy's five-plane architecture while taking:

- Mem0's atomic-memory retrieval;
- Graphiti's entity/relation/provenance split.

---

## Postgres canonical entity design

The following DDL is a proposed Brainy-native design, not a claim about an existing schema.

It deliberately uses application-generated `TEXT` IDs to minimise disruption to current Brainy identifiers.

```sql
CREATE TABLE memory_entities_v2 (
    tenant_id              TEXT        NOT NULL,
    namespace              TEXT        NOT NULL DEFAULT 'default',
    entity_id              TEXT        NOT NULL,
    subject_scope          TEXT,
    entity_type            TEXT        NOT NULL DEFAULT 'unknown',
    canonical_name         TEXT        NOT NULL,
    canonical_norm         TEXT        NOT NULL,
    identity_key           TEXT,
    confidence             DOUBLE PRECISION NOT NULL DEFAULT 1.0
                           CHECK (confidence >= 0 AND confidence <= 1),
    status                 TEXT        NOT NULL DEFAULT 'active'
                           CHECK (status IN ('active', 'merged', 'retired')),
    merged_into_entity_id  TEXT,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (tenant_id, namespace, entity_id)
);

CREATE INDEX memory_entities_v2_name_idx
    ON memory_entities_v2 (tenant_id, namespace, canonical_norm)
    WHERE status = 'active';

CREATE UNIQUE INDEX memory_entities_v2_identity_idx
    ON memory_entities_v2 (tenant_id, namespace, identity_key)
    WHERE identity_key IS NOT NULL AND status = 'active';


CREATE TABLE memory_entity_aliases_v2 (
    tenant_id      TEXT        NOT NULL,
    namespace      TEXT        NOT NULL DEFAULT 'default',
    alias_id       TEXT        NOT NULL,
    entity_id      TEXT        NOT NULL,
    alias_text     TEXT        NOT NULL,
    alias_norm     TEXT        NOT NULL,
    alias_kind     TEXT        NOT NULL DEFAULT 'observed',
    confidence     DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    first_seen_at  TIMESTAMPTZ,
    last_seen_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (tenant_id, namespace, alias_id),
    FOREIGN KEY (tenant_id, namespace, entity_id)
        REFERENCES memory_entities_v2 (tenant_id, namespace, entity_id)
);

CREATE INDEX memory_entity_aliases_v2_lookup_idx
    ON memory_entity_aliases_v2
    (tenant_id, namespace, alias_norm);


CREATE TABLE memory_entity_mentions_v2 (
    tenant_id       TEXT        NOT NULL,
    namespace       TEXT        NOT NULL DEFAULT 'default',
    mention_id      TEXT        NOT NULL,
    entity_id       TEXT        NOT NULL,
    evidence_id     TEXT        NOT NULL,
    mention_text    TEXT        NOT NULL,
    start_byte      INTEGER     NOT NULL CHECK (start_byte >= 0),
    end_byte        INTEGER     NOT NULL CHECK (end_byte > start_byte),
    confidence      DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    resolver        TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (tenant_id, namespace, mention_id),
    FOREIGN KEY (tenant_id, namespace, entity_id)
        REFERENCES memory_entities_v2 (tenant_id, namespace, entity_id)
);

CREATE INDEX memory_entity_mentions_v2_entity_idx
    ON memory_entity_mentions_v2
    (tenant_id, namespace, entity_id);
```

Names are deliberately **not unique identities**.

Two Johns can coexist.

A strong external ID or domain key may populate `identity_key`; otherwise alias/context resolution remains conservative.

---

## Atomic facts and temporal features

Brainy's current atoms already have predicate/value and bitemporal additions.

Rather than forcing every new capability into metadata JSON, the next representation should make the temporal semantics queryable.

```sql
CREATE TABLE memory_facts_v2 (
    tenant_id            TEXT        NOT NULL,
    namespace            TEXT        NOT NULL DEFAULT 'default',
    fact_id              TEXT        NOT NULL,
    subject_id           TEXT        NOT NULL,

    subject_entity_id    TEXT,
    subject_text         TEXT,

    predicate            TEXT        NOT NULL,

    value_text           TEXT,
    value_norm           TEXT,
    value_entity_id      TEXT,

    fact_type            TEXT        NOT NULL DEFAULT 'fact',
    speaker_role         TEXT,
    confidence           DOUBLE PRECISION NOT NULL DEFAULT 1.0,

    -- Event/world time
    event_start           TIMESTAMPTZ,
    event_end             TIMESTAMPTZ,
    valid_from            TIMESTAMPTZ,
    valid_to              TIMESTAMPTZ,

    -- Knowledge/system time
    observed_at           TIMESTAMPTZ,
    recorded_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    retired_at            TIMESTAMPTZ,

    temporal_precision    TEXT NOT NULL DEFAULT 'unknown',
    temporal_status       TEXT NOT NULL DEFAULT 'unknown',
    temporal_expression   TEXT,
    reference_time        TIMESTAMPTZ,

    status                TEXT NOT NULL DEFAULT 'active',
    extraction_version    TEXT,
    memory_id             TEXT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (tenant_id, namespace, fact_id),

    CHECK (
        value_text IS NOT NULL
        OR value_entity_id IS NOT NULL
    )
);

CREATE INDEX memory_facts_v2_sp_idx
    ON memory_facts_v2
    (tenant_id, namespace, subject_entity_id, predicate);

CREATE INDEX memory_facts_v2_predicate_idx
    ON memory_facts_v2
    (tenant_id, namespace, predicate, value_norm);

CREATE INDEX memory_facts_v2_validity_idx
    ON memory_facts_v2
    (tenant_id, namespace, valid_from, valid_to);
```

Recommended controlled temporal values:

| Field | Suggested values/purpose |
|---|---|
| `fact_type` | `state`, `event`, `plan`, `preference`, `relationship`, `absence`, `attribute`, `decision`, `action`, `procedure` |
| `temporal_precision` | `exact`, `minute`, `hour`, `day`, `week`, `month`, `year`, `range`, `relative`, `unknown`, `timeless` |
| `temporal_status` | `past`, `current`, `ongoing`, `future`, `timeless`, `unknown` |
| `event_start/end` | When an event actually occurs |
| `valid_from/to` | Interval in which a state/relation is true |
| `observed_at` | Source observation/conversation time |
| `recorded_at/retired_at` | Brainy's knowledge-system time |
| `reference_time` | Anchor used to resolve phrases such as “last week” |
| `temporal_expression` | Source-faithful phrase, e.g. `"last summer"` |

This combines Brainy's bitemporal strength with:

- the generic memory-time signature Mem0 now applies during retrieval;
- the validity model Graphiti stores on edges.

---

## Relation V2

```sql
CREATE TABLE memory_relations_v2 (
    tenant_id           TEXT        NOT NULL,
    namespace           TEXT        NOT NULL DEFAULT 'default',
    relation_id         TEXT        NOT NULL,
    subject_id          TEXT        NOT NULL,

    src_entity_id       TEXT        NOT NULL,
    predicate           TEXT        NOT NULL,

    dst_entity_id       TEXT,
    dst_value           TEXT,
    dst_value_norm      TEXT,

    fact_text           TEXT,

    valid_from          TIMESTAMPTZ,
    valid_to            TIMESTAMPTZ,
    reference_time      TIMESTAMPTZ,

    observed_at         TIMESTAMPTZ,
    recorded_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    retired_at          TIMESTAMPTZ,

    temporal_precision  TEXT NOT NULL DEFAULT 'unknown',
    confidence          DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    status              TEXT NOT NULL DEFAULT 'active',
    fact_id             TEXT,
    memory_id           TEXT,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (tenant_id, namespace, relation_id),

    FOREIGN KEY (tenant_id, namespace, src_entity_id)
        REFERENCES memory_entities_v2 (tenant_id, namespace, entity_id),

    CHECK (dst_entity_id IS NOT NULL OR dst_value IS NOT NULL)
);

CREATE INDEX memory_relations_v2_out_idx
    ON memory_relations_v2
    (tenant_id, namespace, src_entity_id, predicate)
    WHERE status = 'active';

CREATE INDEX memory_relations_v2_in_idx
    ON memory_relations_v2
    (tenant_id, namespace, dst_entity_id, predicate)
    WHERE dst_entity_id IS NOT NULL AND status = 'active';

CREATE INDEX memory_relations_v2_validity_idx
    ON memory_relations_v2
    (tenant_id, namespace, valid_from, valid_to);
```

The critical difference from Brainy's current `memory_relations` is not “SQL vs graph”.

It is:

```text
current:
"caroline" --origin--> "sweden"

target:
entity:7d1... --origin--> entity:9ab...
               |
               + validity
               + confidence
               + fact/evidence provenance
```

That matches the useful semantic properties of Graphiti's `EntityEdge` without importing its storage backend.

---

## Evidence and source spans

Derived facts and relations should support more than one evidentiary source.

```sql
CREATE TABLE memory_fact_evidence_v2 (
    tenant_id       TEXT NOT NULL,
    namespace       TEXT NOT NULL DEFAULT 'default',
    fact_id         TEXT NOT NULL,
    evidence_id     TEXT NOT NULL,
    support_kind    TEXT NOT NULL DEFAULT 'direct',
    start_byte      INTEGER,
    end_byte        INTEGER,
    span_hash       TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (tenant_id, namespace, fact_id, evidence_id, start_byte),
    FOREIGN KEY (tenant_id, namespace, fact_id)
        REFERENCES memory_facts_v2 (tenant_id, namespace, fact_id),

    CHECK (
        (start_byte IS NULL AND end_byte IS NULL)
        OR
        (start_byte >= 0 AND end_byte > start_byte)
    )
);

CREATE TABLE memory_relation_evidence_v2 (
    tenant_id       TEXT NOT NULL,
    namespace       TEXT NOT NULL DEFAULT 'default',
    relation_id     TEXT NOT NULL,
    evidence_id     TEXT NOT NULL,
    support_kind    TEXT NOT NULL DEFAULT 'direct',
    start_byte      INTEGER,
    end_byte        INTEGER,
    span_hash       TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (tenant_id, namespace, relation_id, evidence_id, start_byte),
    FOREIGN KEY (tenant_id, namespace, relation_id)
        REFERENCES memory_relations_v2
        (tenant_id, namespace, relation_id)
);
```

Byte offsets are preferable for a Go service because they can map consistently to the stored UTF-8 evidence payload.

The original text remains canonical; `span_hash` makes accidental text drift detectable.

---

## Packet V2

Brainy's current packet already has the conceptual split.

The next change should make it **typed rather than string-first**.

```json
{
  "query": "Where is Melanie's friend Caroline originally from?",
  "plan": {
    "intents": ["multi_hop"],
    "answer_shape": "scalar",
    "required_slots": ["origin"],
    "hops": [
      {
        "hop_id": "h1",
        "operation": "follow_relation",
        "input_entity_id": "ent_melanie",
        "predicate": "friend"
      },
      {
        "hop_id": "h2",
        "operation": "fetch_predicate",
        "depends_on": "h1",
        "predicate": "origin"
      }
    ]
  },
  "budgets": {
    "candidate_limit": 120,
    "context_token_limit": 6000,
    "proof_limit": 8,
    "episode_fallback_token_limit": 1200
  },
  "context_evidence": [
    {
      "kind": "fact",
      "id": "fact_caroline_origin",
      "memory_id": "mem_42",
      "content": "Caroline is originally from Sweden.",
      "subject": {
        "entity_id": "ent_caroline",
        "label": "Caroline"
      },
      "predicate": "origin",
      "value": {
        "text": "Sweden",
        "entity_id": "ent_sweden"
      },
      "temporal": {
        "status": "timeless",
        "precision": "unknown",
        "observed_at": "2025-04-10T00:00:00Z"
      },
      "evidence": [
        {
          "evidence_id": "ev_991",
          "span": {
            "start_byte": 84,
            "end_byte": 119
          }
        }
      ],
      "scores": {
        "semantic": 0.81,
        "bm25": 0.44,
        "entity": 0.50,
        "temporal": 0.00,
        "combined": 0.73
      }
    }
  ],
  "proof_chain": [
    {
      "hop_id": "h1",
      "operation": "follow_relation",
      "input_entity_id": "ent_melanie",
      "predicate": "friend",
      "output_entity_id": "ent_caroline",
      "relation_id": "rel_friend_12",
      "evidence_ids": ["ev_850"],
      "proof_level": "typed_exact"
    },
    {
      "hop_id": "h2",
      "operation": "fetch_predicate",
      "input_entity_id": "ent_caroline",
      "predicate": "origin",
      "output_entity_id": "ent_sweden",
      "fact_id": "fact_caroline_origin",
      "evidence_ids": ["ev_991"],
      "proof_level": "typed_exact"
    }
  ],
  "coverage": {
    "required_slots": ["origin"],
    "resolved_slots": ["origin"],
    "unresolved": []
  },
  "answer_contract": {
    "status": "supported",
    "claims": [
      {
        "text": "Caroline is originally from Sweden.",
        "support_ids": [
          "rel_friend_12",
          "fact_caroline_origin"
        ]
      }
    ]
  }
}
```

The reader should be allowed to see `context_evidence`, while only `proof_chain` may establish strict multi-hop support.

That preserves Brainy's truthfulness advantage without forcing the planner to perfectly formalise every useful context item.

---

# Ordered implementation programme

The sequence below deliberately starts with the read-path defect exposed by full LoCoMo, then resumes the representation programme.

Engineering-effort estimates are directional assumptions for one engineer familiar with the codebase; they are not measured delivery commitments.

## Budget assumptions

These defaults should be treated as **starting hypotheses to ablate, not hard-coded truths**.

| Budget | Current relevant behaviour | Proposed production default | Qualification variants | Rationale |
|---|---:|---:|---:|---|
| `candidate_limit` | ~120 normal internal overfetch; cap 200 | **120** | 60 / 120 / 200 | Do not destabilise ranking merely to imitate Mem0. |
| final memory/result limit | 30 | **30 initially** | 30 / 50 | Increase only if fixed-token evidence recall improves. |
| `context_token_limit` | default 4,000 | **6,000** after structured facts | 4K / 6K / 7K | Mem0 reports ~6.8–7.0K tokens/query; structured facts should make 6K substantially denser than transcripts. |
| `proof_limit` | implicit packet/hop size | **8 items/hops** | 4 / 8 / 12 | Proof should remain compact even when context recall is broad. |
| `episode_fallback_token_limit` | part of common evidence budget | **1,200** | 0 / 600 / 1,200 | Provenance fallback should not crowd out semantic facts. |
| fair industry comparator `top_k` | — | **200** | fixed | Match the competitor evaluation depth, not the production default. |

---

## Recommended PR sequence

| PR | Work | Primary failure classes | Effort / cost | Exit criteria |
|---|---|---|---|---|
| **R5A — Structured-first `/recall`** | Make scalar/list answers consume typed fact/atom/relation values first. Retire `firstStatementFromPacket` as a normal factual answer strategy. Hybrid reader handles synthesis/composition; episode text is fallback/context only. | `READER_MISS`, `ABSTENTION_MISS`, answer-path failures | **M: 4–7 eng-days**. Read tokens neutral or lower. Medium OpMem risk because output logic changes. | Unit/golden scalar, list, temporal, preference, provenance, insufficient-evidence tests; OpMem 13/13 and marketing 17/17 remain green; current-SHA 1×30 diagnostic and a stratified single-hop sample improve or produce explainable failures. |
| **R5B — Typed EvidencePacket + spans** | Convert `ContextEvidence []string` into structured objects; retain ProofChain separately; attach fact/relation/evidence/source-span IDs. Legacy `Contents/Items` become compatibility-only. | `EVIDENCE_COVERAGE_MISS`, `PROOF_MISS`, provenance defects | **M: 4–7 days**. Minimal read latency; modest DB/write amplification. Low-to-medium OpMem risk. | Every factual claim emitted by structured reader references a semantic ID and evidence ID; 100% provenance coverage in held-out audit. |
| **R6 — Compiler Coverage V2** | Adapt Mem0's strongest ingestion mechanics: recent-session context, relevant existing facts, one rich extraction call, additive conversational memories, durable assistant actions/facts, memory linking; emit temporal fields and source spans. Preserve governed vertical policies. | `WRITE_MISS`, `REPRESENTATION_MISS`, `TEMPORAL_RESOLUTION_MISS` | **L: 7–12 days**. Likely +10–25% write tokens depending on structured schema; no material read cost. Highest OpMem risk; isolate core mode carefully. | Held-out durable-claim recall ≥90%; direct evidence linkage 100%; malformed semantic outputs <1%; assistant durable-fact recall ≥90%; no OpMem/marketing regression. |
| **R7 — Canonical Entity V2** | Add entity/alias/mention tables; deterministic resolution candidates; type/scope; explicit merge lifecycle. Dual-write old entity hub during migration. | `ENTITY_LINK_MISS`, `RETRIEVAL_MISS`, `PROOF_MISS` | **L: 7–10 days**. Additional DB lookups ~5–20 ms target, write/index overhead moderate. Medium OpMem risk. | Held-out entity-bearing claims ≥95% correct resolution; duplicate-person and same-name tests; all typed hops expose canonical IDs. |
| **R8 — Relation V2** | Project entity-valued facts to canonical ID edges with validity, status, confidence, reference time and evidence spans. Dual-write/backfill v1 relation strings. | `RELATION_MISS`, `TEMPORAL_RESOLUTION_MISS`, `PROOF_MISS` | **L: 7–10 days**. Storage/index growth moderate; read cost low with correct indexes. Medium OpMem risk. | ≥90% held-out relation recall and ≥95% precision on audited relation claims; historical validity tests; every edge traceable to source evidence. |
| **R9 — Hop Executor V3** | Make canonical IDs the hop input/output. Enforce `hop[i].output == hop[i+1].input`. Unscoped/fuzzy matches may enrich context but cannot yield `typed_exact` proof. Remove lexical target binding as strong proof. | `PLANNING_MISS`, `ENTITY_LINK_MISS`, `PROOF_MISS`, MH failures | **M/L: 5–8 days**. ~10–50 ms expected extra DB traversal depending on query; no additional LLM needed for deterministic paths. Low-to-medium OpMem risk. | Synthetic two/three-hop chains, ambiguous-name tests, historical relation tests; no false proof when entity mismatch; 3×90/multi-conversation MH must improve before declaring success. |
| **R10 — Frozen dual-path qualification** | Freeze code; measure product `/recall` and industry-format search→shared answerer separately; fresh Mem0 pin under shared models/budgets where API permits; LME-20; OpMem; marketing. | `HARNESS_ERROR`, `JUDGE_MISS`, programme measurement | **M: 3–6 eng-days plus model/API cost**. No production code risk. | Exact SHA/dataset/model manifests; no tuning between systems; publish product and comparator rows separately. |

---

## Why structured-first answering comes before another compiler wave

This is the one place to deliberately adjust a pure “representation first at all times” reading of the PoR.

R0–R4 have already created semantic material.

Full `/recall` demonstrates that product synthesis can discard its value.

Improving extraction before repairing `/recall` risks storing better facts only to have:

- `firstStatementFromPacket`;
- generic enumeration;
- abstention;

still emit the wrong answer.

But R5A must remain **bounded**.

It should not turn into reader-prompt optimisation.

Once structured facts are actually used correctly, the next quality limit is compiler coverage and entity/relation semantics.

---

## Tests that should accompany every representation PR

The unit test corpus should contain ordinary product-language cases, not benchmark labels:

```text
same person, changed residence
same name, two different people
nickname and formal name
user fact and assistant action
multi-valued hobbies
exclusive current employer
historical employer
relative date anchored to observation time
relation A → B → C
correction vs historical change
explicit deletion in governed vertical
fact absent from compiler but present in evidence fallback
```

The existing `overfit_denylist_test.go` should remain part of merge gates.

---

# Competitor reproduction programme

The right approach is not “clone Mem0” or “port Graphiti”.

It is a maintained **BORROW / ADAPT / REJECT ledger** against their current source.

---

## Mem0 OSS archaeology

| OSS path | What it demonstrates | Brainy decision |
|---|---|---|
| `mem0/memory/main.py` | V3 phased pipeline: recent session messages, relevant existing-memory retrieval, one additive LLM extraction call, batch embedding, entity extraction and hybrid search. | **ADAPT pipeline sequencing/context.** Brainy already has contextual extraction; measure and close the delta. |
| `mem0/configs/prompts.py::ADDITIVE_EXTRACTION_PROMPT` and `generate_additive_extraction_prompt` | Rich self-contained ADD-only memory extraction, existing/recent memory context and optional linking. | **ADAPT semantics, not verbatim wording.** Use Brainy's structured subject/predicate/value/time/evidence contract. |
| `mem0/utils/entity_extraction.py` | NER/proper/quoted/topic/identifier candidate extraction with filtering and candidate confidence concepts. | **ADAPT candidate taxonomy and precision tests.** Do not require spaCy if Go-native/LLM methods perform better. |
| `mem0/utils/scoring.py` | Semantic + normalised BM25 + entity boost; explains signal combination. | **INSPECT, mostly REJECT porting.** Brainy already has equivalent-plus-temporal fusion; do not copy `ENTITY_BOOST_WEIGHT=0.5` merely because Mem0 uses it. |
| Mem0 benchmark framework | Current competitor budget/model methodology. Mem0 reports top-200 and roughly 7K retrieved tokens. | **ADAPT evaluation preset**, not production top-k. |
| Mem0 2026 temporal algorithm | Separate time signature and additive temporal reranking. | **ADAPT fact-level temporal features.** Preserve Brainy's stronger bitemporal projections. |

A subtle but important point:

Mem0's repository still contains legacy UPDATE/DELETE prompts as well as the newer additive V3 path.

The existence of those older symbols should not be confused with the current default algorithm; current V3 uses the additive extraction machinery.

---

## Graphiti archaeology

| OSS path | What it demonstrates | Brainy decision |
|---|---|---|
| `graphiti_core/graphiti.py` | Episode ingestion and graph-construction orchestration. | **ADAPT conceptual pipeline**, not driver/storage code. |
| `graphiti_core/nodes.py::EpisodicNode` | Source content, source type, `valid_at`, referenced entities. | **BORROW concept.** Brainy's evidence plane already supplies most of it. |
| `graphiti_core/nodes.py::EntityNode` | Stable UUID identity, labels, embeddings, summaries and attributes. | **ADAPT strongly** into Postgres canonical entities. |
| `graphiti_core/edges.py::EntityEdge` | Source/target IDs, fact, embedding, provenance episodes, validity interval and reference time. | **ADAPT strongly** into Relation V2. |
| `graphiti_core/prompts/extract_nodes.py`, `extract_edges.py`, `extract_nodes_and_edges.py` | Typed node/edge extraction surfaces. | **Inspect schema/contracts; adapt concepts.** Do not build a second independent truth extractor if entity-valued Brainy facts can project relations. |
| `dedupe_nodes.py`, `dedupe_edges.py` | Explicit identity/relation consolidation. | **Adapt conservative dedupe/merge semantics**, including reversible decisions. |
| `search/search_config_recipes.py` | Edge/node/episode/community BM25+cosine, BFS, RRF/MMR/cross-encoder and graph-distance patterns. | **Borrow later, selectively.** Relation retrieval first; cross-encoder or graph-distance only after a measured need. |
| Graph backends | Neo4j, FalkorDB, Kuzu, Neptune and pluggable graph abstractions. | **REJECT as immediate dependency.** Postgres is adequate for Brainy's current scale and logical needs. |

---

## Licence and legal handling

Both relevant OSS repositories are Apache-2.0 licensed.

Practically, Brainy should distinguish **ideas/adaptations** from **copied source**:

- Reimplementing an architecture or algorithmic idea does not require pasting competitor code.
- If substantial source code is copied or modified, preserve applicable copyright/licence notices and comply with Apache-2.0's NOTICE/licensing requirements.
- Check third-party dependency licences independently; the repository's Apache licence does not automatically relicense every external dependency.
- Mem0 Platform and Zep managed-service behaviour must not be treated as Apache-licensed merely because their OSS cores are.

This is an engineering/licence-management recommendation, not legal advice.

---

# Measurement, observability and qualification

## Representation-coverage oracle

The most important measurement improvement is to stop asking only whether the answer string exists somewhere in memory.

For a small, frozen, manually annotated corpus of held-out conversations, annotate durable source claims at ingestion time.

Example:

```json
{
  "claim_id": "hc_041",
  "evidence_id": "ev_991",
  "source_span": {
    "start_byte": 84,
    "end_byte": 119
  },
  "speaker_role": "user",
  "subject": "Caroline",
  "predicate": "origin",
  "value": "Sweden",
  "entity_required": true,
  "relation_required": true,
  "temporal": {
    "required": false
  }
}
```

Then compute stage-level precision and recall.

| Oracle | Question |
|---|---|
| `SOURCE_PRESENT` | Did raw evidence contain the required claim? |
| `FACT_PRESENT` | Did the compiler create a usable atomic fact? |
| `ENTITY_PRESENT` | Did the claim bind to the right canonical entity? |
| `RELATION_PRESENT` | Was an entity-valued relationship projected correctly? |
| `TEMPORAL_PRESENT` | Were date/status/validity fields represented correctly? |
| `RETRIEVED` | Did candidate retrieval contain the gold semantic object? |
| `CONTEXT_PRESENT` | Did it survive evidence-budget selection? |
| `PROOF_PRESENT` | Were required canonical joins established? |
| `ANSWER_USED` | Did the reader cite/use the correct semantic object? |
| `ANSWER_CORRECT` | Did the final answer match the required result? |

LongMemEval's own decomposition supports this type of analysis: the benchmark is explicitly about extraction, multi-session reasoning, temporal reasoning, knowledge updates and abstention.

---

## Suggested representation-quality gates

| Measure | Initial gate |
|---|---:|
| Durable-claim fact recall | **≥90%** |
| Fact precision on audited outputs | **≥95%** |
| Entity resolution accuracy | **≥95%** on entity-bearing claims |
| Relation recall | **≥90%** |
| Relation precision | **≥95%** |
| Temporal-field correctness | **≥90%** on temporally-bearing claims |
| Direct semantic→evidence linkage | **100%** |
| Invalid/malformed semantic objects | **<1%** |
| Durable assistant-fact recall | **≥90%** |

These are proposed engineering thresholds, not industry standards.

---

## Diagnostic taxonomy mapped to code

Brainy already has most required names in `trace.go`.

The important change is to make every label depend on actual oracle evidence.

| Failure | Meaning | Primary code surface |
|---|---|---|
| `SOURCE_MISS` | Required information was never ingested into immutable evidence | `evidence_plane.go`, ingest path |
| `WRITE_MISS` | Evidence contains it but no usable fact/event/relation exists | `provider_extractor.go`, `attribute_atoms.go`, semantic compiler |
| `REPRESENTATION_MISS` | A row exists but is malformed, over-coarse or semantically unusable | compiler validation / atoms |
| `ENTITY_LINK_MISS` | Correct fact exists but canonical identity is absent/wrong | `entities.go`, future Entity V2 |
| `RELATION_MISS` | Required entity-valued relation was not projected | `relations.go` / Relation V2 |
| `TEMPORAL_RESOLUTION_MISS` | Correct semantic object exists but time/status/validity is wrong | `temporal.go`, temporal enrichment |
| `RETRIEVAL_MISS` | Gold semantic object exists but candidate retrieval omitted it | `service.go::SearchOpt`, `fusion_v2.go` |
| `EVIDENCE_COVERAGE_MISS` | Candidate existed but budget/packet construction omitted it | `BuildEvidencePacket`, packet selection |
| `PLANNING_MISS` | Correct semantic evidence exists but query decomposition asks for wrong operations | `planner.go` |
| `PROOF_MISS` | Necessary objects exist but canonical joins were not established | `hop_executor.go` |
| `READER_MISS` | Correct answer-bearing structured packet reached the reader; answer still wrong | `recall.go`, `reader_hybrid.go` |
| `ABSTENTION_MISS` | Reader answers unsupported query or abstains despite sufficient structured evidence | reader/status contract |
| `JUDGE_MISS` | Evaluation judge failed or was inconsistent | eval judge |
| `HARNESS_ERROR` | Ingest/path/job/model pin invalidates evaluation | LME/LoCoMo harness |

In particular:

> A transcript substring is **not sufficient evidence for `READER_MISS`**.

If no fact/entity/relation was compiled, the earliest failure is `WRITE_MISS` or `REPRESENTATION_MISS`.

---

## Per-stage operational metrics

Brainy's `SearchTrace` already records:

- lexical hits;
- dense admissions;
- entity-hub admissions;
- atom admissions;
- candidate overfetch;
- episode fallback;
- representation status;
- channel scores.

Extend that rather than creating a parallel telemetry framework.

### Write path

```text
evidence_records_written
facts_proposed / accepted / rejected
facts_per_episode
assistant_facts
entities_proposed / resolved / created / ambiguous
relations_projected
temporal_features_present
semantic_objects_without_evidence
source_span_coverage
compiler_prompt_tokens
compiler_completion_tokens
compiler_latency_ms
job_queue_age_ms
```

### Search path

```text
candidate_count_by_channel
candidate_limit
candidate_unique_count
semantic / lexical / entity / temporal recall@candidate
structured_fact_share
episode_fallback_share
context_token_count
context_fact_count
context_relation_count
context_episode_count
retrieval_latency_by_channel_ms
```

### Proof and answer

```text
planned_hops
typed_exact_hops
fuzzy_hops
unresolved_hops
proof_items
reader_source
answer_mode
supporting_ids_valid
unresolved_targets
answer_status
reader_prompt_tokens
reader_completion_tokens
reader_latency_ms
```

### Evaluation

```text
source_present
representation_present
retrieval_present
context_present
proof_present
answer_used_gold
failure_class
judge_retry_count
judge_failure
```

---

## LME publish-integrity checklist

The current harness already implements many of these, so this is primarily a release checklist rather than another architecture PR.

A publishable LME-20 must record:

- exact Brainy product SHA and evaluation-framework SHA;
- dataset URL/hash and exact selected question IDs;
- answerer, extractor/embedder where applicable, judge model and temperatures;
- product `POST /recall` on every answer-bearing item;
- asynchronous ingestion;
- empty-queue precondition;
- exact `jobs_expected == jobs_completed`;
- `jobs_failed == 0`;
- candidate/context/proof budgets;
- reader-source distribution;
- no silently discarded question;
- judge failures distinctly labelled, not scored as ordinary wrong answers;
- per-question answer, category and failure-stage trace.

Brainy's current 4/20 LME result should be treated as a genuine product-quality signal now that jobs completed 4829/4829, not dismissed as the old harness problem.

Running LME-500 immediately would be poor use of spend because LME-20 already shows:

- multi-session 0/5;
- actionable architectural deficits.

The larger run becomes valuable when the 20-question diagnostic can no longer distinguish competing engineering hypotheses.

---

# Fair LoCoMo measurement

Brainy should maintain **two different headline measurements**.

| Row | Purpose | Brainy path | Competitor path |
|---|---|---|---|
| **Product recall** | What Brainy users actually get | `POST /recall` | No forced competitor equivalent unless directly available |
| **Industry-format memory bake-off** | Memory/index comparison | search → shared answerer → shared judge | Mem0 search → same shared answerer → same shared judge |

For the fair comparator:

```text
n = 1540 cats 1–4
same dataset hash
candidate/top_k = 200
same answerer model
same answer prompt
same judge model
judge temperature = 0
same seeds / question ordering
same maximum final context-token budget where technically enforceable
report actual retrieved tokens, not only requested top_k
```

The current Mem0 92.5 result remains useful as a **vendor-published quality bar**, but not as the same-run denominator for Brainy's 11.4 product `/recall`.

Likewise, Brainy's 21/30 versus Mem0 11/30 current conv-26 pin is a useful same-cycle diagnostic but cannot stand in for full LoCoMo.

It is especially unrepresentative because Brainy is 10/10 MH there while full-suite MH is only 21/282.

---

# Risks, economics and claims

## Risk and benefit by major change

| Change | Main benefit | Engineering effort | Runtime/token effect | OpMem / vertical risk | Principal risk |
|---|---|---:|---|---|---|
| Structured-first `/recall` | Exposes value of semantic memory immediately; attacks 841-question SH mass | Medium | Usually neutral/lower tokens; minimal storage cost | **Medium** | Over-constraining answers or changing proven vertical response behaviour |
| Typed packet + source spans | Auditable claims and cleaner reader input | Medium | Slight packet size; small write/storage increase | **Low–medium** | Migration/compatibility complexity |
| Compiler Coverage V2 | Biggest representation-recall upside | High | +write LLM/schema tokens; read unchanged | **Medium–high** | Extraction precision regression or governed-mode contamination |
| Canonical entities | Improves retrieval, relation quality and MH together | High | Small DB latency/index overhead | **Medium** | False merges are worse than duplicate entities |
| Relation V2 | Makes graph-style reasoning reliable without new DB | High | Moderate storage; low indexed read overhead | **Medium** | Over-projecting relations / temporal conflicts |
| Hop V3 | Converts “typed” chain into actual identity join | Medium–high | Several indexed reads; normally no extra LLM | **Low–medium** | Planner fails to express relations even though graph is correct |
| Frozen qualification | Stops anecdotal optimisation and establishes parity truth | Medium + API spend | Evaluation-only | **None** | Temptation to tune between competitor runs |

The largest cost increase is likely on **write-time semantic compilation**, not retrieval.

That is acceptable strategically.

Mem0's present approach also invests in semantic extraction so reads can stay compact, and its 2026 system reports under ~7K retrieved tokens/query on its major benchmarks.

Graphiti is also a warning that extraction richness has provider cost and reliability consequences.

Brainy's existing subject-ordered worker and bounded concurrency should therefore remain.

Increasing concurrency is not a substitute for compiler correctness.

---

# Public claims now

The August 15/17 freeze supports the following precise statements:

| Claim | Status |
|---|---|
| Brainy OpMem **13/13 vs Mem0 10/13** on the current empirical freeze | **Allowed** |
| Brainy marketing suite **17/17 vs Mem0 4/17** empirical on the current freeze | **Allowed**, with “empirical suite” qualification |
| Brainy LoCoMo conv-26 **21/30 vs Mem0 11/30**, with Brainy MH 10/10 but OD 0/4 | **Allowed as a 30-question measurement**, never as qualification |
| Brainy full product `/recall` **175/1540 = 11.4%** | **Must be disclosed as the current full product pin**, not hidden behind the 21/30 result |
| Brainy LME-20 **4/20** after successful job completion | **Allowed as Brainy's own diagnostic pin**, not competitor parity |
| Mem0 reports **92.5 LoCoMo / 94.4 LongMemEval** at roughly 6.8–7.0K tokens/query | **May be cited as vendor-published figures**, clearly labelled non-same-pin |

Still forbidden:

- “Brainy beats Mem0” without a narrowly specified same-pin scope;
- “Brainy is SOTA”;
- “multi-hop solved” on the basis of one conversation's 10/10;
- presenting 21/30 as a 70% full-LoCoMo score;
- presenting historical 49.8% search+harness as the current Brainy product score;
- placing Mem0 92.5 and Brainy 11.4 in a table that implies identical answer paths/budgets;
- comparing Brainy's 4/20 LME sample directly with Mem0's 500-question 94.4 as though it were a same-run benchmark;
- restoring open-domain/single-hop quality by indiscriminately stuffing transcript episodes back into the reader context.

---

# Execution roadmap and qualification checkpoints

A practical programme is roughly nine to eleven weeks with overlap between database/schema work and evaluation tooling.

These are planning assumptions, not commitments.

## Early checkpoint

After structured-first answering and before waiting for the entire representation migration:

**Freeze the SHA** and run:

- OpMem 13;
- marketing 17;
- LoCoMo 1×30 only as regression/diagnostic;
- a frozen, stratified 100–200 question single-hop/open-domain/temporal subset;
- a **current-SHA search+harness oracle run** over the same subset.

The main question is not:

> “Did the headline increase?”

It is:

> **How much of the product `/recall` gap disappears when the reader consumes the structured semantic values already present?**

A material `/recall` improvement without representation changes validates the answer-path diagnosis.

A small improvement would mean WRITE/representation misses dominate more strongly than the current smoking-gun examples suggest.

---

## Representation checkpoint

After Compiler Coverage V2 and canonical entities:

Run the manual representation audit first.

Merge gates should focus on:

- semantic coverage;
- entity accuracy;
- provenance;

not a specific LoCoMo number.

Then run:

- current 1×30 diagnostic;
- 3×90 or another frozen multi-conversation slice;
- LME-20 product `/recall`;
- OpMem/marketing non-regression.

Do not run another full 1540-question LoCoMo after every PR.

---

## Relation checkpoint

After Relation V2 and Hop V3:

Run:

- synthetic entity/relation/historical proof suite;
- multi-conversation MH slice;
- LME-20 multi-session breakdown;
- structured evidence/proof inspection.

The multi-hop success criterion should be:

> **Generalisation across conversations**

not preservation of conv-26's 10/10.

---

## Frozen competitive qualification

Only once answer-path and representation checkpoints both move.

### Product track

```text
Brainy POST /recall
full LoCoMo
LME-20
then LME-100/500 if the 20-Q sample no longer determines the next decision
OpMem
marketing
BEAM 100K / larger BEAM when scale is the active uncertainty
```

### Industry-format track

```text
Brainy search
Mem0 search
top_k 200
shared answerer
shared prompt
shared judge
shared dataset/questions
identical seed protocol
actual context tokens reported
```

At that point, the meaningful parity question becomes:

> **At an equal answerer and roughly equal evidence-token budget, does Brainy's compiled fact/entity/relation index retrieve enough information to match Mem0-class conversational QA while preserving Brainy's operational and governance advantages?**

That is a much more useful test than comparing two differently constructed headline percentages.

---

# Explicit answers to the 2026-08-17 self-review questions

## 1. Course

**Verdict: KEEP representation-first course, ADJUST sequence.**

The two-gap split is directionally correct:

1. a real product answer-path gap;
2. a deeper representation/compiler/entity/relation gap.

However:

> **11.4% → ~50% is not yet a proven numeric decomposition.**

The historical ~49.8% search+harness result came from an older path.

Run a current-SHA search+harness diagnostic to measure the present answer-path ceiling.

The slogan/enumeration/abstention failures are **causal on a meaningful subset**, not merely cosmetic symptoms.

But WRITE/representation misses remain the larger long-term barrier to Mem0-class performance.

---

## 2. Published metric

Keep **both metrics**, clearly separated.

### Product metric

Use:

```text
POST /recall
n=1540
cats 1–4
```

This is the honest measure of what Brainy users receive.

Current freeze:

```text
175/1540 = 11.4%
```

### Industry-format metric

Also pin:

```text
search
→ shared answerer
→ shared judge
```

for Brainy and Mem0 under the same evaluation contract.

### README recommendation

The primary Brainy product row should remain `/recall`.

A separate clearly labelled row should show:

> **Industry-format memory bake-off**

Do not collapse the two.

---

## 3. First product PR

The first PR should be:

> **R5A — Structured-first `/recall`**

Primary failure class:

```text
READER_MISS / ANSWER_PATH_MISS
```

Specifically:

- scalar factual questions should consume structured fact values;
- list questions should enumerate structured values, not arbitrary strings;
- multi-hop should consume proof-chain outputs;
- `firstStatementFromPacket` should stop being a normal factual-answer mechanism;
- abstention should happen only after structured support is assessed.

This should land **before R5-on-OD as a benchmark-specific focus**.

R5-on-OD remains useful as a diagnostic because OD is 0/4, but the product PR should be general structured answering, not “fix OD”.

---

## 4. Fair Mem0 stack

A fair eventual n=1540 comparison should use:

```text
LoCoMo cats 1–4
n = 1540
same dataset hash
top_k / candidate output = 200
same answerer model
same answer prompt
same judge model
judge temperature = 0
same ordering
same seed protocol
actual retrieved context tokens reported
same max evidence-token budget where enforceable
```

Recommended seeds:

```text
3 seeds minimum
```

for any claimed comparative result.

Also report:

```text
mean / p50 / p95 retrieved tokens
answerer tokens
latency
cost
category accuracy
```

Do not call Mem0's published 92.5 a current same-pin result.

It remains a vendor-published quality bar.

---

## 5. Skipped suites

Skipping:

- LME-500;
- BEAM 1M;
- BEAM 10M;

was correct.

### LME-500 becomes useful when:

- LME-20 is no longer obviously failing due to the same architectural holes;
- multi-session is no longer 0/5;
- a larger sample could change which engineering hypothesis is selected.

### BEAM 1M / 10M becomes useful when:

- scale itself becomes the active uncertainty;
- 100K performance is materially stronger;
- the main question is memory robustness under much larger histories rather than current representation/read-path correctness.

Do not use expensive scale benchmarks as substitutes for architecture diagnosis.

---

## 6. Next 3–7 PRs

Recommended order:

1. **R5A — Structured-first `/recall`**  
   Failure class: `READER_MISS / ANSWER_PATH_MISS`

2. **R5B — Typed EvidencePacket + source spans**  
   Failure class: `EVIDENCE_COVERAGE_MISS / PROOF_MISS`

3. **R6 — Compiler Coverage V2**  
   Failure class: `WRITE_MISS / REPRESENTATION_MISS`

4. **R7 — Canonical Entity V2**  
   Failure class: `ENTITY_LINK_MISS`

5. **R8 — Relation V2**  
   Failure class: `RELATION_MISS / TEMPORAL_RESOLUTION_MISS`

6. **R9 — Hop Executor V3**  
   Failure class: `PLANNING_MISS / PROOF_MISS`

7. **R10 — Frozen dual-path qualification**  
   Failure class: `MEASUREMENT / HARNESS / COMPARABILITY`

Do not add:

- fusion retune;
- graph DB migration;
- another full benchmark sweep;
- benchmark-specific heuristics;

between these unless new contradictory evidence appears.

---

## 7. Claims

### Allowed

- Brainy OpMem 13/13 vs Mem0 10/13 on this empirical freeze.
- Brainy marketing 17/17 vs Mem0 4/17 empirical on this freeze.
- Brainy LoCoMo 1×30 21/30 vs Mem0 11/30 on the same current conv-26 diagnostic.
- Brainy full LoCoMo product `/recall` 175/1540 = 11.4%.
- Brainy LME-20 4/20 after 4829/4829 job completion.
- BEAM 100K 8/20 as the current non-regression diagnostic.
- Mem0's published 92.5 LoCoMo / 94.4 LME as vendor-published, non-same-pin figures.

### Forbidden

- unqualified “Brainy beats Mem0”;
- SOTA;
- “multi-hop solved”;
- 70% full LoCoMo;
- restoring historical 49.8% as current;
- calling 11.4% a harness glitch;
- comparing 11.4 product `/recall` directly to 92.5 as if paths/budgets were identical;
- comparing LME 4/20 directly to Mem0 94.4;
- comparing BEAM 8/20 100K to Mem0 64.1 BEAM 1M;
- treating OpMem/marketing leadership as conversational-memory leadership.

---

## 8. Kill list

Confirm unchanged:

- no fusion-constant fishing;
- no graph DB default;
- no category dictionaries;
- no LoCoMo/LME-named product rules;
- no unbounded top-k;
- no episode stuffing to restore SH/OD;
- no reader-prompt sprint as the programme;
- no reopening architect PR1–PR7 without contradictory measured evidence;
- no reopening R0–R4 as if they never landed;
- no SOTA / beats-Mem0 language;
- no another full LoCoMo remasure as the next task;
- no LME-500-as-quality;
- no BEAM 1M/10M until scale is the active uncertainty;
- no benchmark-target score promises.

---

# Final programme judgement

Brainy's competitive problem is **not that Mem0 and Graphiti possess a secret architectural primitive that Brainy lacks entirely**.

Much of the competitive substrate is now present:

```text
append-oriented conversational history       ✓
immutable evidence                            ✓
atomic facts / atoms                          ✓ but incomplete coverage
dense + lexical retrieval                     ✓
entity retrieval                              ✓ but string-based
temporal scoring                              ✓
context vs proof packet                       ✓ but under-typed
relation table                                ✓ but v1
typed hop executor                            ✓ but weak identity semantics
hybrid reader                                 ✓ but poor product answer routing
measurement integrity                         ✓ substantially hardened
```

The remaining difference is the **quality and consistency of the contracts between those components**.

Mem0's advantage is that its extracted memories are consistently treated as the primary retrieval unit.

Its current OSS pipeline supplies:

- recent history;
- related existing memories;
- one additive extraction pass;
- semantic retrieval;
- lexical retrieval;
- entity signals.

Its managed system adds rich temporal reasoning.

Graphiti's advantage is that identity and relationships are not inferred anew from arbitrary memory strings during every hop:

- episodes;
- entities;
- temporally valid edges;

are explicit persistent objects.

Brainy's advantage is that it already has the foundations neither architecture should cause it to throw away:

- dedicated evidence plane;
- bitemporal/current-state semantics;
- governed update/suppression behaviour;
- vertical state models;
- explicit answer sufficiency;
- operational evaluation.

The current freeze's OpMem and marketing results are evidence that those capabilities create real differentiated behaviour even while conversational recall remains weak.

The engineering mandate should therefore be:

> **Do not rebuild Brainy as Mem0. Do not rebuild Brainy as Graphiti. Complete the semantic contracts Brainy has already chosen.**

Specifically:

> **Structured facts should answer simple questions. Canonical entities should establish identity. Temporal relations should express changing relationships. Proof chains should join stable IDs. Episodes should prove where every fact came from. Governed projections should decide operational truth.**

Or, retaining the strongest formulation from the existing programme:

> **Facts are for recall. Relations are for reasoning. Episodes are for proof. Projections are for truth.**

If Brainy executes R5A through R9 and then survives a frozen equal-budget qualification, it will no longer merely have a promising architecture.

It will have the specific:

- retrieval units;
- identity semantics;
- temporal relations;
- answer contract;

required to make:

> **Mem0-quality conversational recall + Graphiti-quality relational reasoning + Brainy-quality governed truth**

a testable competitive proposition.

---

# Source references used by the deep-research review

## Brainy internal basis

Primary review brief:

- `2026-08-17 full LoCoMo /recall dip self-review prompt`
- current docs SHA: `8492ad3`
- measured product SHA: `1b5ab3e`

Key internal surfaces referenced by the review:

```text
internal/memory/recall.go
internal/memory/reader_hybrid.go
internal/memory/write_policy.go
internal/memory/contextual_extractor.go
internal/memory/entities.go
internal/memory/relations.go
internal/memory/hop_executor.go
internal/memory/fusion_v2.go
internal/memory/planner.go
internal/memory/trace.go
evals/public/longmemeval/run.py
```

Measured pins supplied in the review brief:

```text
OpMem:                 13/13
Marketing:             17/17
LoCoMo 1×30:           21/30
  MH:                  10/10
  OD:                   0/4
  Temporal:            11/16

LoCoMo full /recall:   175/1540 = 11.4%
  MH:                  21/282 = 7.4%
  OD:                   5/96 = 5.2%
  SH:                  88/841 = 10.5%
  Temporal:            61/321 = 19.0%

Historical July:
LoCoMo search+harness: 49.8% mean
seed-0:                49.4%

LME-20:                4/20
jobs:                  4829/4829

BEAM 100K conv-0:      8/20
```

## Public competitor / benchmark sources

Mem0:

- https://github.com/mem0ai/mem0
- https://docs.mem0.ai/core-concepts/memory-evaluation
- https://docs.mem0.ai/migration/platform-v2-to-v3
- https://mem0.ai/blog/state-of-ai-agent-memory-2026
- https://mem0.ai/blog/ai-memory-benchmarks-in-2026
- https://mem0.ai/blog/introducing-temporal-reasoning-in-mem0
- https://mem0.ai/blog/the-token-efficient-memory-algorithm-now-has-temporal-reasoning
- https://mem0.ai/research

Graphiti / Zep:

- https://github.com/getzep/graphiti
- https://github.com/getzep/graphiti/blob/main/graphiti_core/graphiti.py
- https://arxiv.org/abs/2501.13956

Benchmarks:

- LoCoMo — https://arxiv.org/abs/2402.17753
- LongMemEval — https://arxiv.org/abs/2410.10813

---

# One-line handoff

> **KEEP representation-first, but fix structured `/recall` consumption first; then close compiler coverage, canonical identity, temporal relation and exact-hop contracts before the next frozen competitive qualification.**
