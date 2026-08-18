# Path to a competitive conversational memory system (2026-08-14)

**Status:** accepted course — representation-first; revised after external review  
**Tips:** Fresh remasure (`1b5ab3e`): LoCoMo 1x30 **21/30** (MH **10/10**, OD **0/4**, temporal **11/16**) vs Mem0 Platform **11/30**; full `/recall` **11.4%** is an **answer-path dip** vs hist. 49.8% search+harness (not a vanished compiler; **49.8% is not a current-SHA ceiling**). LME-20 **4/20**; BEAM 100K **8/20**. Archaeology (Wave 1) is historical: keep R0-R4 closed. **R5A structured-first `/recall` landed** (2026-08-17). **R5B–R10 representation stack** is the substrate for a later honest competitive claim (typed packets, named-subject + `she`/`he` coref, canonical entity/relation IDs, hop ID joins, dual-path freeze **wiring**). Not a 70–80% n=1540 claim and not SOTA. Two lanes. [70–80% path](./locomo-full-70-80-path.md) · [R10 freeze](./locomo-dual-path-freeze.md) · [dip why](../benchmarks/artifacts/locomo-full-recall-dip-why-20260817.md) · [live verdict](./external-reviews/2026-08-17-parity-gap-verdict.md). Internal cycle notes: [competitive/cycle-closeout.md](./competitive/cycle-closeout.md).  
**Does not claim:** SOTA, or a LoCoMo/LME target score  
**Review:** [external-reviews/2026-08-14-representation-path-additions.md](./external-reviews/2026-08-14-representation-path-additions.md) (R1c amendment) · [external-reviews/2026-08-17-parity-gap-verdict.md](./external-reviews/2026-08-17-parity-gap-verdict.md) (live: R5A first; do not re-queue R0-R4) · [external-reviews/2026-08-17-competitive-archaeology-verdict.md](./external-reviews/2026-08-17-competitive-archaeology-verdict.md) (historical Wave 1 pin)

## Competitive thesis

The target is no longer:

> Build increasingly sophisticated retrieval over conversational transcripts.

It is:

> **Compile interactions into durable semantic memory, retrieve and reason over that memory, and retain the transcript as immutable provenance.**

Atomic conversational facts plus multi-signal retrieval (semantic, keyword/BM25, entity, temporal) are highly effective. Public graph-memory systems keep episodes as provenance while entities and temporally-valid relations become the semantic substrate. Brainy's opportunity is the combination:

```text
high-recall semantic memory
        +
entity / relation structure
        +
Brainy-style evidence provenance, current-state semantics,
predicate policies, operational correctness, vertical governance,
and truthful answer sufficiency
```

One line:

> **Facts are for recall. Relations are for reasoning. Episodes are for proof. Projections are for truth.**

Wave 1 was useful infrastructure, but it optimized retrieval over the **wrong primary unit**. Competitive systems converge on:

```text
raw interaction
→ compact semantic memories
→ entities / relations / temporal attributes
→ retrieval over those representations
→ source interaction retained as provenance
```

## Direct answers

**Will reading papers get us there?** Papers are necessary **inputs**, not the work.

| Source | What to take | What not to take |
| --- | --- | --- |
| [Mem0 paper (arXiv:2504.19413)](https://arxiv.org/abs/2504.19413) + [2026 algorithm](https://mem0.ai/blog/state-of-ai-agent-memory-2026) | Extract **atomic facts** (ADD-only); retrieve those facts; entity linking as a retrieval signal; temporal ranking over **dated facts**; durable assistant-generated facts | Managed-platform LoCoMo as a comparable OSS pin; Neo4j; bench-specific prompts |
| Graphiti / Zep (ENG-69) | **Episode = provenance**. **Entity + edge = retrieval unit**. Validity windows | Graph DB as required substrate (ADR-004); treating Graphiti OSS as Zep Platform |
| LoCoMo / LongMemEval papers | What MH/temporal questions *require* (attribute join, list aggregation, current vs past) | Product rules named after the bench |
| HippoRAG / A-MEM / MemGPT (ENG-71, still backlog) | Later: entity-centric walk, memory evolution, working-vs-archival | Do not start here while the compiler is still thin |

We do **not** need another survey before writing code. ENG-54/ENG-71 annotate this path; they do not block it. The July gap doc already named GAP-C1: ingest does not emit **atomic attribute facts** ([mem0-parity-gaps.md](./mem0-parity-gaps.md)). Wave 1 did not close that.

**Competitor archaeology is required — with OSS vs platform split (POV 13).**

| Track | What it is | How we use it |
| --- | --- | --- |
| **Mem0 OSS** | Inspectable extract / retrieval mechanisms | ADAPT what we can reproduce |
| **Mem0 Platform** | Product quality bar (published benches include proprietary opts) | Match quality; do not assume copying the repo copies the score |
| **Graphiti** | Inspectable entity / relation / episode / validity architecture | ADAPT into Postgres (ADR-004) |
| **Zep Platform** | What that architecture becomes in production | Capability target; not a copy of a proprietary engine |

Two opposite mistakes are both wrong: “they're open source, so cloning the repo reproduces the score,” and “the platform is proprietary, so we cannot learn from it.”

**Is the work “diff our system vs popular systems”?** Yes — the gap is representation, not BM25 weights.

| System | Retrieval unit | Brainy today |
| --- | --- | --- |
| Mem0 | Extracted fact sentences + entity collection | Mix of **conversation_episode transcripts** + some facts |
| Graphiti | Entities + relation edges; episodes stored separately | `memory_entity_links` hub only; **no relation table**; hops = first linked memory ID |
| Brainy (ops/vertical) | Governed records, lifecycle, packs | **Already ahead** (OpMem 13/13, marketing 17/17) — keep |

Wave 1 ledgers saying `READER_MISS` with “coverage supported” were **misleading**. Oracle “supported” meant the gold **substring appears in a retrieved chat turn**. That is `SOURCE_PRESENT`, not “the semantic representation needed to answer is in the reader packet.” Treating it as a reader problem is how we slid into efficiency PRs.

## Architecture: recall-primary vs evidence-primary (POV 11)

```text
                     SOURCE
                       │
                 conversation
                       │
                       ▼
                    EVIDENCE
                       │
                immutable episode
                       │
                       ▼
                    SEMANTIC
           ┌───────────┼───────────┐
           │           │           │
         facts      entities    relations
           │           │           │
           └───────────┼───────────┘
                       │
                       ▼
                     RECALL
```

| Unit | Role |
| --- | --- |
| **episode** | evidence-primary, provenance-primary, fallback-retrievable |
| **fact / entity / relation** | recall-primary |

This is cleaner than “drop episodes.” Episodes are never deleted from the store. They lose recall priority only when the compiler has produced the semantic units a future question needs.

## Why Wave 1 felt like efficiency

| Shipped | What it actually is |
| --- | --- |
| PR4 MaxEvidenceTokens / pool 30–200 / episode −0.15 | Ranking around the same episode corpus |
| PR3 temporal_score + IncludeHistorical | Useful signal on **transcripts**; keep it, move it onto **dated semantic facts** (POV 8) |
| PR5 ContextEvidence vs ProofChain | Packet layout for a reader still fed **dialogue** |
| PR9 skip phatic assistant episodes | Correct filter; do **not** generalize to “assistant turns are not memory” (POV 6) |
| Deferred PR6–PR8 | **Wrong deferral.** MH 3/10 vs Mem0 7/10 is missing **facts/entities/edges** |

Local Wave 1 LoCoMo **14/30** (MH **3/10**) is not an improvement vs Gate 0 18/30. Attribute atoms were even tagged `primitive=episode`, so they took the episode penalty. That is a representation bug, not a reader bug. Temporal 1/16 → 9/16 shows the temporal **signal** works; the unit it scored was still chat.

## Failure model (POV 3) — assign the earliest failing stage

Do not call `READER_MISS` because a gold substring appeared somewhere in a retrieved chat turn.

Presence probes:

```text
SOURCE_PRESENT          immutable evidence contained the required information
REPRESENTATION_PRESENT  ingest created the required fact / entity / relation
RETRIEVAL_PRESENT       that representation was retrieved
PROOF_PRESENT           required joins / coverage could be established
ANSWER_PRESENT          synthesis produced the answer
```

Earliest-stage labels:

```text
1. SOURCE_MISS               required information never entered Brainy
2. WRITE_MISS                source contained it; semantic compiler failed to store it
3. ENTITY_LINK_MISS          fact exists; identity binding missing/wrong
4. RELATION_MISS             facts exist; required relation was not represented
5. TEMPORAL_RESOLUTION_MISS  semantic record exists; temporal interpretation is wrong
6. RETRIEVAL_MISS            correct semantic record exists; candidate retrieval missed it
7. PLANNING_MISS             correct records retrieved; required decomposition was wrong
8. PROOF_MISS                required records exist; join/coverage could not be proven
9. READER_MISS               correct structured answer-bearing packet reached synthesis; answer still wrong
```

Only call `READER_MISS` when the **semantic representation needed to answer is actually in the reader packet**.

This oracle lands **before** interpreting another LoCoMo category regression.

## Execution sequence

R1 and R1b are **one conceptual milestone** (POV 2), even if they stay separate review surfaces:

> A conversational interaction must compile into useful semantic memory before the transcript loses recall priority.

The milestone is **not** `primitive != episode`. The question is: did the source interaction produce the semantic units future questions require? Metric is **semantic coverage**, not fact count. A terrible extractor can emit 40 useless facts from 20 turns; a good one emits 12 that capture everything durable.

**Do not ship R1c as hard episode-suppression before R1b has held-out representation coverage.** That is the sequencing constraint.

### R0 — Representation observability

**Landed in #113** (fact-aware `semantic` / `representation` oracles; gold in episode-only store is `WRITE_MISS`). Evidence-stage `SOURCE_MISS` must mean **zero evidence rows**, not “gold missing from a truncated dump.”

Coverage oracle + stage taxonomy above. Held-out conversation **representation audit** is a merge gate **before** benchmark score (POV 12):

```text
durable source claims identified manually
        ↓
% represented as atomic facts

entity-bearing claims     → % correctly entity-linked
relation-bearing claims   → % represented as edges
dated claims              → % carrying usable temporal attributes
semantic records          → % with valid evidence provenance
```

Example report (illustrative shape, not a pin):

```text
Held-out conversation representation audit
durable claims:             47
atomic facts represented:   43 / 47
entity bindings correct:    36 / 38
relation edges correct:     18 / 20
dated facts usable:         12 / 13
evidence linkage:           43 / 43
```

Those numbers tell us more about R1–R3 than a single LoCoMo result. LoCoMo then measures whether the architecture generalizes to QA. **1×30 remains measurement, not qualification.**

Per held-out conversation also count: `episodes_ingested`, `atomic_facts_created`, `durable_assistant_facts_created`, `entities_created`, `relations_created`, `dated_facts_created`, `facts_with_evidence`, `facts_with_subject/entity binding` — then inspect a small set by hand.

### R1a — Correct primitive semantics

Stop marking real facts/atoms as `primitive=episode`. Episodes remain provenance. Dated provider facts and attribute atoms are recall-primary facts.

**Landed in #113.** It does not by itself complete the representation milestone.

### R1b — Atomic semantic compiler

High-recall durable facts from held-out dialogue. One compiler, not a sentence blob plus a later unrelated graph extractor (POV 4–5).

Atomic unit (conceptual):

```text
fact_id, tenant_id, subject_id
subject_entity_id?  subject_text
predicate
value_text, value_entity_id?
fact_type, speaker_role
event_start, event_end, observed_at, temporal_precision
confidence, evidence_ids, source_span
```

Human-readable sentence remains useful (`"Caroline is originally from Sweden."`). Underneath, Brainy should know `subject=Caroline`, `predicate=origin`, `value=Sweden`.

**Durable assistant facts are first-class** (POV 6). PR9 phatic skip stays. Do not let “assistant turns are noisy” become “assistant turns are not memory.”

```text
PHATIC          "Sure!" / "Happy to help."          → no memory
DURABLE ACTION  "I booked your flight for March 3." → speaker_role=assistant
DURABLE FACT    "The refund was processed yesterday."
COMMITMENT      "I'll send the report Friday."
DECISION        "We selected PostgreSQL for the project."
```

Especially relevant to LongMemEval.

Inspect Mem0 OSS `mem0/configs/prompts.py` (ADD-only fact sentences) as **ADAPT**, not a verbatim copy.

**Exit:** held-out representation audit (POV 12), not a LoCoMo bump. R1+R1b are incomplete until coverage is high enough that dropping transcripts would not hide `WRITE_MISS`.

**Quality gate (landed):** malformed compiler templates are not semantic memory. Light-verb `has done going at …`, failed gerund stems (`participates in runn`), and broken quote shards must not persist, must not complete coverage, and must not outrank provenance. Local remasure **11/30** vs R1c **10/30** (q10 recovered; packet junk 45→6). Remaining LoCoMo misses on that pin are mostly **WRITE_MISS** — the compiler still does not emit the durable claim.

**Coverage slice (landed, not complete):** held-out Jordan/Riley audit is green. Local remasure **15/30 (50.0%)**, MH **6/10 (60.0%)**, WRITE_MISS **15→10**. Episodes stay as fallback while 10 WRITE_MISS remain. See [locomo-r1b-dev-1x30-20260814.md](../benchmarks/artifacts/locomo-r1b-dev-1x30-20260814.md).

### R1c — Fact-primary recall (not unconditional transcript suppression)

**Landed in #113** with coverage-gated fallback (trace: `representation_status`, `episode_fallback`, `episodes_dropped`). Local 1×30 remasure **10/30** (MH 2/10, OD 0/4, temporal 8/16) vs Wave 1 **14/30** — a **dip**, expected while the compiler is thin. See [locomo-r1c-dev-1x30-20260814.md](../benchmarks/artifacts/locomo-r1c-dev-1x30-20260814.md).

```text
facts / edges     = primary evidence
episodes          = provenance + fallback evidence

if structured candidates satisfy query coverage:
    suppress standalone episodes from reader context

if structured representation appears incomplete:
    allow bounded episode_fallback

trace:
    representation_status = complete | partial | empty
    episode_fallback = true | false
```

The rejected rule was: “drop episodes when any non-episode candidate exists.” That converts `WRITE_MISS` into something that looks like `RETRIEVAL_MISS` / `REASONING_MISS`:

```text
episode A: Caroline is from Sweden
episode B: Caroline likes pottery
extractor only emits: Caroline likes pottery
query needs both → a non-episode exists → episode A dropped
```

Safer principle (same as hardening): **represent structurally whenever possible; preserve source text when the compiler demonstrably missed something.** Retrieve broadly enough to answer; prove narrowly enough to trust.

Until R1b coverage is proven, Search **must** keep bounded episode fallback on `partial` / `empty`. Hard suppression is allowed only when `representation_status=complete`.

Enumerate still skips episode *values* (“Yeah, Caroline”) so list answers stay fact-shaped. Episodes may still appear as provenance snippets.

### R2 — Canonical entities (identity, not a name table)

```text
entity_id, tenant_id, subject_scope, entity_type, canonical_label
+ aliases, mentions, evidence, confidence
```

Resolution returns ranked deterministic candidates:

```text
mention → candidate entities → confidence → canonical entity ID
```

Not: mention → matching memories → first memory ID → inferred entity.

Identity remains tenant/subject scoped unless an explicit cross-scope policy exists. **REJECT** Neo4j.

**First slice (landed):** compiler `subject` / `value_norm` are copied onto record metadata and entity links. Not yet canonical IDs + aliases + ranked resolution. Remaining work is **R7**.

### R3 — Relation projection (not a second extractor)

Do not build `fact extractor + separate relation extractor` unless measurement later proves it necessary.

```text
conversation → semantic compiler → AtomicFact
    ├── literal value          (Caroline.age = 32)
    └── entity value           (Caroline.origin = Sweden)
             ↓
         RelationEdge          (Caroline --origin--> Sweden)
```

One semantic truth layer. Facts and graphs must not be able to disagree.

Temporal attributes live on the fact/edge (POV 8):

```text
Caroline moved to London.     fact_type=state_change  event_start=2025-07
Caroline currently lives in London.  fact_type=state  valid_from=2025-07
```

Keep `temporal_score`, `IncludeHistorical`, and current-state resolution — score **dated semantic records**, not conversational prose.

**First slice (landed):** `memory_relations` (mig v20) + `follow_relation` hops. Edges are projected from compiler facts on **both** sync ingest and async extract. R4 remasure MH **9/10**. Remaining quality work is **R8** (canonical-ID endpoints, validity, spans).

### R4 — Relation-aware hops (actual joins) — landed (measurement)

Invariant:

```text
hop[i].output_entity_id == hop[i+1].input_entity_id
```

Local 1x30 (`d48e202`): MH **9/10 (90.0%)** vs Mem0 freeze **7/10**. Remaining MH miss is image-gold, not a hop miss. Canonical entity IDs (R2 full / **R9**) are still not claimed done.

Proof chain holds **entity IDs, relation IDs, fact IDs, evidence IDs** so the join is inspectable. This is not “hop 1 retrieved a relevant memory and hop 2 retrieved another.” Earlier `hop_join_proven` was too permissive when it allowed that.

### R5A — Structured-first `/recall` (first product PR)

```text
StructuredEvidence:
  Caroline.origin = Sweden
  Caroline.relationship_status = single
  Caroline.activity = pottery

SourceEvidence:
  provenance snippets supporting each fact
```

The product answer path consumes structured values first. Source text remains to resolve ambiguity, verify provenance, expose qualification, handle incomplete representation, and explain where memory came from. That is Brainy's evidence-plane advantage over a pure fact-store.

**What to change:** retire `firstStatementFromPacket` as a **normal factual** strategy (`recall.go`: first non-question `pkt.Contents` / `TemporalAnswer`). Scalar answers cite typed fact/atom/relation values; lists enumerate structured values; hops consume proof-chain outputs; abstain only after structured support is assessed. Hybrid reader handles synthesis/composition. Episodes stay provenance + fallback. **Bounded:** not a reader-prompt sprint, not LLM-over-episodes, not a PR named "fix OD".

**Mass vs same-pin trail (2026-08-17):** full LoCoMo single-hop is **88/841 (10.5%)** on product `/recall` because slogan/enumerate/abstain cite chat text instead of structured values. 1x30 OD **0/4 vs Mem0 3/4** is a **diagnostic** inside R5A's test plan (and still WRITE_MISS on some items). Do not restore SH/OD by stuffing episodes into top-k.

**Ceiling honesty:** 11.4% toward ~50% is **directional**. July 49.8% was search+harness on an older stack, not a measured oracle on current storage. After R5A, run a **current-SHA search+harness** on the same stratified 100-200 subset as product `/recall` to size answer-path vs WRITE_MISS. Not a full n=1540 remasure.

**Exit (early checkpoint):** OpMem 13/13 and marketing 17/17 stay green; 1x30 diagnostic; stratified SH/OD/temporal subset; explainable remaining failures.

### R5B — Typed EvidencePacket + spans

**Landed.** `ContextEvidence` is `[]PacketItem` (fact/memory IDs, predicate, subject, value, entity_id, source span). Legacy `Contents` is a compatibility projection of those items. Hops still must not replace context.

### R6 — Compiler Coverage V2

Generalize past conv-26. Full-suite SH 10.5% and LME multi-session 0/5 show R0-R4 work where facts compile, but coverage does not travel. **R6a:** named-subject / addressee binding so reports are not attributed to the reporter; `works as` / `lives in` / `realized that`; held-out audit. **R6 remainder:** `she`/`he` bind to the last named person in the thread (plural `they` stays unbound). `ContextualExtractor` already injects prior memories. Durable assistant facts stay first-class. Held-out representation audit is the merge gate, not a LoCoMo bump. Honest 70–80% map: [locomo-full-70-80-path.md](./locomo-full-70-80-path.md).

### R7 — Canonical Entity V2

**Landed (additive).** Durable `ent:` IDs scoped by tenant/subject. Distinct labels (John Smith vs John Doe) get distinct IDs; a first-name mention resolves only when unique. Dual-write `memory_entities` (mig v22) plus the existing string hub. Ranked resolution is deterministic. Two Johns with the same first-name mention stay ambiguous (no collapse). Tenant/subject scoped. No Neo4j.

### R8 — Relation V2

**Landed (additive).** `memory_relations` keeps v1 string endpoints and dual-writes `src_entity_id` / `dst_entity_id`, validity (`valid_from` = observed_at), and `evidence_span`. Copy Graphiti **semantics**, not Neo4j. `ListRelationsFrom` accepts a mention or an ID.

### R9 — Hop Executor V3 (canonical ID joins)

**Landed.** Invariant remains `hop[i].output_entity_id == hop[i+1].input_entity_id`. `resolve_entity` emits `EntityID`. Fetch/follow copy that ID via DependsOn. Unscoped `GetCurrentState(pred)` and unfiltered atom hits enrich **context** (`proof_kind=context`) and cannot yield `typed_exact` / `hop_join_proven`. Search fallback is context. Do not claim MH-solved from 1x30 10/10 while full MH is 7.4%.

### R10 — Frozen dual-path qualification

**Wiring landed** ([locomo-dual-path-freeze.md](./locomo-dual-path-freeze.md)). `--eval-lane product-recall` vs `--eval-lane industry-search` (industry defaults top-k 200). Do not remasure n=1540 in this pass.

Order when a freeze is requested: LoCoMo 1x30 **diagnostic** -> LoCoMo 3x90 qualification slice -> multi-seed/full **product `/recall`** and a separately labeled **industry-format** search+shared-answerer+shared-judge (n=1540, top-k 200, 3 seeds, report retrieved tokens) -> LME-20 **quality** -> larger LME.

Do not run LME-500 or BEAM 1M as a quality claim while LME-20 is 4/20 and BEAM 100K is 8/20.

Representation health (R0 audit) is a merge gate **before** these scores. 1x30 is never qualification. OpMem/marketing must stay green. No SOTA / beats-Mem0 language.

## Kill list (unchanged)

No fusion-constant fishing (including copying `ENTITY_BOOST_WEIGHT=0.5`), no graph DB default, no category dictionaries, no unbounded top-k, no LoCoMo/LME-named product rules, no SOTA / beats-Mem0 language, no treating 1x30 as the qualification, no hard episode-suppression before coverage, no second unrelated relation extractor, no treating phatic-skip as "assistant is not memory," no spaCy requirement, no v2 schema DDL in R5A, no calendar-week commitments in PoR, no reopening R0-R4 as if missing.

## Linear

- ENG-168 conversational long-memory epic — this path is the engineering response
- ENG-176 multi-hop synthesis — after canonical-ID hops (R9), not instead of R5A
- ENG-69 Graphiti temporal fact model — input to R8
- ENG-71 academic survey — annotate, do not block R5A
- ENG-60 graph layer — **Postgres graph-shaped** (ADR-004), not "build Neo4j"
