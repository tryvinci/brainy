# Brainy: State of the System, Gaps, and Path to SOTA Memory

**Audience:** External technical reviewers  
**Date:** 2026-08-01  
**Purpose:** Self-contained briefing on what Brainy is today, what we have measured, which gaps matter, and the proposed roadmap to competitive conversational memory **and** leadership in vertical / operational memory.  
**Related:** [master-plan.md](./master-plan.md) (program of record), [master-plan-execution-status.md](./master-plan-execution-status.md) (latest measured numbers)

---

## 1. Executive summary

Brainy is a **Go-first vertical memory service**: a Postgres-backed API that stores, retrieves, and governs long-lived facts for agents and products. We deliberately compete on **two tracks**:

1. **Operational / truth-over-time memory** — supersession, correction, lifecycle, isolation (our OpMem benchmark).
2. **Vertical / governed memory** — YAML packs that encode domain vocabulary, lifecycle, and ranking (marketing and support today).

We are **ahead** of Mem0 on those two tracks in measured head-to-heads. We are **behind** on industry conversational recall benchmarks (full LoCoMo, LongMemEval, BEAM) after an intentional de-overfit that removed benchmark-shaped hacks from product code.

This document asks reviewers to pressure-test:

- Whether our diagnosis of the conversational gap is correct (coverage + long-haystack + fusion, not more ranking knobs).
- Whether the proposed phases (retrieval fusion → extraction/event memory → vertical canonical layer → proof) are the right order.
- What we are missing relative to 2026 SOTA systems (Mem0, Zep/Graphiti, AtomMem, APEX-MEM, Memanto, and vertical “context graph” patterns).

---

## 2. What Brainy is today

### 2.1 Product thesis

Most memory products optimize for **“retrieve something relevant from chat history.”** Enterprises also need **“memory that stays true”**: facts that can be corrected, superseded, scoped by brand/campaign/ticket, and ranked by domain policy (e.g. brand rules over style preferences).

Brainy’s bet:

- **Cognitive primitives** (principle, belief, episode, outcome, …) rather than opaque embeddings alone.
- **Vertical packs** as first-class config (not hard-coded product logic).
- **Operational contracts** (suppress / supersede / correct / domain events) as API surface.
- **Postgres + Go** for latency, ops simplicity, and explainability — not a mandatory graph DB.

### 2.2 System shape

```
Client ──► API (Go) ──► memory.Service
                │            │
                │            ├─ deterministic extractor (+ optional LLM provider extract)
                │            ├─ attribute atoms (typed predicates)
                │            ├─ hybrid search (FTS + dense + optional entity hub)
                │            └─ POST /recall (context | enumerate | answer)
                │
                └─ Worker (async /ingest/async) ──► same extract + embed path
                           │
                           Postgres (memory_records, embeddings, entity links, atoms, jobs)
```

**Key packages**

| Area | Location |
| --- | --- |
| Core service / search | `internal/memory/service.go` (~2k LOC) |
| Product synthesis | `internal/memory/recall.go` — `POST /recall` |
| Deterministic atoms | `internal/memory/attribute_atoms.go` |
| Predicate taxonomy | `internal/memory/predicates.go` |
| LLM extract | `internal/memory/provider_extractor.go` |
| Entity hub | `internal/memory/entity_hub.go` + `internal/store/postgres/entity_hub.go` |
| Atom index | `internal/store/postgres/atoms.go` (migration v13) |
| FTS | `content_tsv` on `memory_records` (migration v14) |
| Vertical packs | `packs/marketing/v1`, `packs/support/v1` |
| Public evals | `evals/public/{locomo,longmemeval,beam}` |

### 2.3 API surface (product)

| Endpoint | Role |
| --- | --- |
| `POST /ingest` | Sync extract + upsert |
| `POST /ingest/async` | Queue job; worker runs provider extract + embed |
| `GET /memories/search` | Hybrid retrieval (lexical/FTS + dense + optional entity) |
| `POST /recall` | Product synthesis: `context` / `enumerate` / `answer` (+ abstention) |
| `POST /memories/{id}/supersede`, domain `/events` | Truth-over-time |

`POST /recall` is intentional: **answer shaping lives in the product**, not only in the eval harness (anti-“MemPalace” discipline).

### 2.4 Ingest pipeline (what actually gets stored)

1. **Baseline deterministic extract** (always) — conversational episodes + attribute atoms when speakers are attributed.
2. **Optional provider extract** (async worker) — LLM ADD-only atoms; merged with baseline (not replace).
3. **Embeddings** persisted for dense search.
4. **Entity links** written to `memory_entity_links` (hub-and-spoke).
5. **Typed atoms** written to `memory_atoms` when predicate/value are known.
6. **Auto-supersede** for stateful predicates (identity / relationship / residence / occupation).

Attribute atoms are **general linguistic forms** (identity, origin, activities, quoted titles, family preferences). Benchmark surface-forms were removed under a CI denylist (`overfit_denylist_test.go`) after we found LOCOMO-shaped regexes inflating smoke scores.

### 2.5 Retrieval path (today)

`SearchOpt` approximately:

1. Build lexical patterns from content-bearing tokens.
2. Parallel: FTS/ILIKE search + dense similarity (calibrated min-max).
3. Optionally list full subject corpus for expansion.
4. Admit extra candidates via entity hub, subject-content bridge, and **predicate scan** for list queries.
5. Score with hybrid lexical+dense (+ entity/IDF **gated off by default**).
6. Diversify for list-shaped queries; return top_k (eval default 30–50).

### 2.6 Vertical packs (today)

**Marketing pack** — brand_rule, voice_profile, campaign, audience_segment, creative_asset, performance_outcome, content_belief, … with lifecycle (archived campaigns excluded; active boosted) and primitive rank weights (principle ≫ preference).

**Support pack** — ticket_state, customer_fact, resolution_note, escalation_rule, agent_preference with resolved/closed lifecycle.

Measured: marketing vertical **15/16** Brainy vs **4/16** Mem0 under our strict-schema policy (the policy *is* the moat; we label it in tables).

### 2.7 Design doctrine (non-negotiable)

- **Anti-benchmax:** improve the product; do not tailor to LOCOMO surface forms.
- **CI denylist:** no benchmark names/phrases in product code or prompts.
- **Holdout:** tune on LoCoMo convs 1–3; validate 4–10 sparingly; log runs.
- **Two stacks:** cheap CF gpt-oss for iteration; publish claims only with pinned stack + multi-seed.
- **Baselines required** for any “we beat X” claim: full-context, naive RAG, filesystem+grep (Letta lesson).

---

## 3. Goals (what “SOTA overall, especially vertical” means for us)

We do **not** define success as “one number on LoCoMo.” Success is multi-axis:

| Priority | Goal | Why |
| --- | --- | --- |
| P0 | Stay #1 on **operational correctness** (OpMem) and expand the public suite | Mem0 is ADD-only by design; enterprises need update/forget/correct |
| P0 | Deepen **vertical packs** into canonical domain memory (not just vocabulary) | Structural moat competitors lack |
| P1 | Close conversational gap to **credible** LoCoMo / LME / BEAM (Gate R1 was ≥75 LoCoMo & LME) | Required for market credibility even if not our wedge |
| P1 | Sub-second search p50 under realistic load | Go+Postgres should win latency/cost |
| P2 | Third-party harness presence (Mem0 memory-benchmarks, AMB, MemoryBench) | Self-reported tables are discounted after 2025–26 benchmark theatre |

---

## 4. What we have measured (publish stack)

**Stack for numbers below:** Brainy staging (async provider extract + CF embeddings), answerer/judge = CF gpt-oss via OpenAI-compatible gateway, top_k=30–50, multi-seed where noted.  
**Not** Mem0’s published frontier-GPT / top-200 stack — so absolute comparisons to 92.5 LoCoMo are **directionally** informative, not identical.

### 4.1 Scoreboard

| Axis | Brainy | Best competitor signal |
| --- | --- | --- |
| Full LoCoMo (1540 Q, 3 seeds) | **Mean ≈49.8%** (49.4 / 49.4 / 50.6); multi-hop ≈**26%** | Mem0 92.5; Hindsight 92.0; Zep corr. ~75; Letta-fs 74 |
| LongMemEval-S | **4%** on stratified **100** Q (full 500 deferred) | Mem0 94.4; Zep 90.2 |
| BEAM | **40%** on **1 conversation** @ 100K (20 probing Q) | Mem0 64.1 @ BEAM-1M |
| OpMem | **13/13** | Mem0 platform 9/12 (same fixtures) |
| Marketing vertical | **15/16** | Mem0 4/16 |
| Support vertical | **3/3** fixtures | — |
| Search latency @ concurrency 8 | p50 **2403** / p95 **4997** ms | Mem0 ~0.9–1.1s; Zep markets <200ms |

Artifacts live under `docs/benchmarks/artifacts/`.

### 4.2 Interpretation

- **Stack mismatch does not excuse ~50% LoCoMo or ~4% LME.** Absolute performance shows missing **coverage** (atoms never stored) and **long-haystack retrieval** (signal drowned in 40–500 session histories).
- **Multi-hop is the conversational hole** (~26%). Lists of attributes across sessions fail because similarity optimizes best-match, not attribute coverage — even with `/recall` enumerate and atom index.
- **Vertical / OpMem remains the differentiator** and should stay the public wedge while conversational scores climb.

### 4.3 Master-plan completion status

| Phase | Status |
| --- | --- |
| P0 Truth (de-overfit) | Done |
| P1 Substrate (atoms, FTS) | Done |
| P2 Retrieval + `/recall` | Done (fusion not default-on) |
| P3 Truth-over-time | Done (OpMem 13/13) |
| P5 Vertical pack #2 | Done (support v1) |
| P4 Scale/proof | **Partial** — baselines measured; Gate R1 **missed**; AMB / Lane A / full LME-500 / BEAM-1M open |

---

## 5. Gaps we have identified (internal)

### 5.1 Conversational recall gaps

| ID | Gap | Evidence | Likely fix |
| --- | --- | --- | --- |
| **G1** | Fused retrieval not default-on | `entityRankingEnabled` / `idfRankingEnabled` default **false**; Mem0 runs semantic+BM25+entity always | Default-on fusion with Mem0-shaped additive score, over-fetch, semantic threshold gate |
| **G2** | Extraction coverage incomplete | Multi-hop 26%; missed compound/agent statements; LME 4% | Broader atoms + event atoms + profile rollup |
| **G3** | No bi-temporal read-time “latest wins” | Supersession exists; Graphiti-style `valid_at/invalid_at` query filters do not | `memory_atoms.valid_to` + query-time current-only for state predicates |
| **G4** | Long-haystack collapse | LME ~115K tokens/question; single subject corpus + top_k≈30–50 | Event memory, rolling profile, atom-level indexing, dynamic budget |
| **G5** | Latency under load | c=8 p50 2.4s; full subject list on many searches | Avoid full scans; batch embeds; build FTS GIN off boot |
| **G6** | Synthesis still thin | `/recall` enumerate exists but depends on atoms being present | Make enumerate the primary path for list intents; improve atom density first |
| **G7** | Eval vs product budget | We often evaluate at top_k=30; Mem0 publishes ~top_200 | Raise budget with latency SLOs; report tokens/query |

### 5.2 Vertical memory gaps

| ID | Gap | Evidence | Likely fix |
| --- | --- | --- | --- |
| **V1** | Packs are vocabulary + lifecycle, not **canonical domain models** | Marketing/support YAML; no CRM/ticket/order entity resolution | Canonical schemas + source mappings (GTM context-graph pattern) |
| **V2** | Weak cross-object linking | Ticket ↔ customer ↔ escalation rule not first-class | Support pack v2 with typed edges / atom links |
| **V3** | Only two packs | Marketing mature; support early | Second depth pack + one new vertical (e.g. sales/CRM) |
| **V4** | Outcome→belief underused in public story | Code exists (`outcome.go`) | Productize in marketing pack + fixtures |

### 5.3 Ops / platform gaps

| ID | Gap | Evidence |
| --- | --- | --- |
| **O1** | Staging worker SIGTERM ~60s after boot | Observed on Render; async extract drained via local workers |
| **O2** | FTS GIN creation OOMs/timeouts on boot | Deferred; needs controlled `BRAINY_ENSURE_FTS_INDEX=1` |
| **O3** | Provider extract throughput | ~1–6 jobs/min on CF gpt-oss; serializes large evals |

### 5.4 What we deliberately removed (context for reviewers)

Prior multi-hop gains included LOCOMO-shaped regexes and prompt examples (names, places, titles). We **removed** them (W1) and accepted a smoke drop (peak-with-hacks 19/30 → honest ~16/30). Reintroducing those patterns is **out of scope** and blocked by CI. Any proposed fix must generalize.

---

## 6. External research & competitor takeaways

Sources reviewed for this plan: Mem0 OSS (`mem0/memory/main.py`, `utils/scoring.py`, `utils/entity_extraction.py`, prompts), Zep Graphiti (`search_filters.py` bi-temporal), AtomMem (arXiv 2606.19847), APEX-MEM (ACL 2026), Memanto (typed semantic memory), Mem0 “State of AI Agent Memory 2026”, FunnelStory GTM context graph, Atlan agent-memory patterns.

| Idea | Who | Implication for Brainy |
| --- | --- | --- |
| Additive fusion: semantic + sigmoid-BM25 + entity boost (0.5), over-fetch `limit×4`, semantic threshold **before** hybrid | Mem0 OSS | Closest high-ROI copy; we already have FTS + entity hub + partial fusion |
| spaCy entities: PROPER / QUOTED / TOPIC / IDENTIFIER | Mem0 | Upgrade entity extract beyond our simple heuristics |
| Bi-temporal edges + point-in-time filters | Graphiti / APEX-MEM | Map onto atoms + supersession (`valid_to`) without Neo4j |
| Atomic facts → event memory → temporal profile | AtomMem / Memanto | Natural extension of our atom index |
| Multi-tool retrieval agent at query time | APEX-MEM | Emulate with `/recall` modes + predicate scan (“GRAPHSQL-lite”) |
| Canonical typed models + source mappings (~95% stable ontology) | FunnelStory / Atlan | **Vertical packs v2** — this is how we win “especially vertical memory” |
| Always beat full-context / RAG / grep baselines | Letta / field consensus | Mandatory in any public table |

---

## 7. Proposed action plan (for review)

### Phase A — Retrieval truth (1–2 weeks)

**Intent:** Make search behave like a modern multi-signal system by default, without bench-specific logic.

1. Default-on fusion (feature-flagged): calibrated semantic + `ts_rank_cd` BM25 (sigmoid params by query length) + entity hub ≤0.5; over-fetch; semantic gate first.
2. Migration 15: `memory_atoms.valid_to`; supersede sets `valid_to`; query-time latest-wins for state predicates.
3. Tune on LoCoMo **validation** convs 4–10 only; OpMem 13/13 + support 3/3 non-regression.
4. Rebuild FTS GIN in a quiet window; re-measure latency c=8.

**Exit:** Full LoCoMo ≥ **60%** (from ~50); OpMem/support green; p95 trending down.

### Phase B — Extraction + event/profile memory (2–3 weeks)

**Intent:** Fix multi-hop and start fixing LME by storing the right units.

1. Broader deterministic + provider atoms (compounds, assistant statements, nested subjects, events with `observed_at`).
2. `memory_events` (session/time buckets) + async rollup; entity-overlap propagation at recall (AtomMem-style).
3. Rolling **profile memory** per subject (identity / activities / prefs inventory) + `/recall` profile mode.
4. Re-run LoCoMo (3 seeds) + LME-S 500 + BEAM-100K sample.

**Exit:** LoCoMo ≥ **65%**; LME-S 500 ≥ **40%**; BEAM-100K sample ≥ **50%**.

### Phase C — Vertical moat (2–4 weeks, parallelizable)

**Intent:** Make “especially vertical memory” undeniable.

1. Canonical schemas + source mappings for **support** (ticket / customer / order) and deepen **marketing**.
2. Support pack v2: typed links, profile rollup, 5+ fixtures (escalation, time-bounded ticket state, customer isolation).
3. Publish OpMem + vertical multi-vendor table with baselines.

**Exit:** Support 5/5; marketing ≥15/16; ≥3 systems on expanded OpMem.

### Phase D — Proof (1–2 weeks after A/B)

1. Publish-stack LoCoMo ×3 seeds, LME-S 500, BEAM-100K multi-convo.
2. Lane A: Brainy backend in mem0ai/memory-benchmarks; baseline triple table.
3. AMB / MemoryBench if numbers are defensible.

**Exit:** Honest public portfolio; no Gate R1 claim until numbers clear ≥75 on LoCoMo **and** LME (or we revise gates with rationale).

---

## 8. Questions for external reviewers

We would especially value feedback on:

1. **Diagnosis:** Is “coverage + long-haystack + fusion” the right root-cause set for ~50% LoCoMo / ~4% LME, or are we under-weighting synthesis / judge / embedding model choice?
2. **Architecture:** Should we stay Postgres+atoms+entity hubs (Memanto-style), or invest in a real temporal KG (Graphiti-style) for conversational SOTA?
3. **Vertical:** Is “canonical schema + mappings” the right next step for packs, or should packs stay thin and push ontology to the application?
4. **Ordering:** Fusion-first (Phase A) vs extraction/event-memory-first (Phase B) — which moves LoCoMo/LME more?
5. **Eval strategy:** Given LoCoMo’s known answer-key/judge issues, how much weight should Gate R1 still carry vs OpMem + vertical + on-policy benches?
6. **What did we miss?** Techniques, papers, or production patterns (2025–26) not reflected above that you would prioritize for a Go/Postgres memory service.

---

## 9. Appendix — quick file map for reviewers

```
internal/memory/service.go          # search, ranking, supersession
internal/memory/recall.go           # POST /recall
internal/memory/attribute_atoms.go  # deterministic typed atoms
internal/memory/provider_extractor.go
internal/memory/entity_hub.go       # fusion helpers + entity links
internal/memory/predicates.go
internal/memory/vertical.go
internal/store/postgres/{store,atoms,entity_hub,migrations}.go
packs/{marketing,support}/v1/pack.yaml
evals/public/{locomo,longmemeval,beam}/
docs/research/master-plan.md
docs/benchmarks/artifacts/          # measured run reports
```

### Reproduce (high level)

```bash
# OpMem
python3 evals/run_opmem.py --systems brainy --base-url "$BRAINY_BASE_URL"

# Marketing vertical
python3 evals/run_marketing_mvp_benchmark.py --base-url "$BRAINY_BASE_URL" --systems brainy,mem0

# Full LoCoMo (publish stack)
cd evals && python -m public.locomo.run_full --conversations 10 --seeds 3 --top-k 50
```

---

*Document prepared for external technical review. Numbers are from Brainy’s publish-stack runs as of 2026-08-01; competitor headline scores are cited from public materials and may use different judges, budgets, and model stacks.*
