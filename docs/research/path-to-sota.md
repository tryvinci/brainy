# Surpass Mem0: multi-axis research + product plan

**Doctrine: improve the product, never tailor the benchmark.** LOCOMO / LongMemEval /
BEAM are diagnostics. OpMem and vertical fixtures measure production failure modes.
North star is **surpassing** Mem0 and peers on multiple axes — not LOCOMO parity alone —
with publishable research that expands what “SOTA memory” means.

See also: [paper-topics.md](./paper-topics.md) · [public-bench-ladder.md](./public-bench-ladder.md) · [README](./README.md)

---

## Multi-axis scoreboard

| Axis | Why it matters | Today (honest) | Surpass bar |
| --- | --- | --- | --- |
| **A. Operational correctness** | Production / regulated | OpMem **12/12** vs Mem0 **9/12** | Hold lead; publish multi-system OpMem matrix |
| **B. Vertical / governed memory** | Domain agents | Marketing **16/16**; Mem0 N/A | Second pack (finance) on unchanged runtime; systems paper |
| **C. Conversational recall** | Industry headline | LOCOMO smoke **16/30** (gpt-oss, 1×30); multi-hop **3/10** | Fair full LOCOMO under same pins ≥ Mem0; then beat multi-hop + temporal |
| **D. Latency / cost** | Deployability | Search p50 **~730–890 ms** | Beat Mem0 p50/p95 and tokens/query on same pin |
| **E. Outcome-grounded belief** | Research novelty | Spec + `ob05` | Mechanism paper vs observation-only belief systems |

**External “surpassed Mem0” claim** requires: win **A + D** publicly, win **B** by category,
and on **C** either beat under identical pins or match within CI while clearly leading A/B/D.
Never a LOCOMO-only victory lap from a 30-Q smoke.

---

## Where Brainy is (2026-07-24, `dev`)

- Conversational path: async provider extract, atomic episodes, event-time
  (`observed_at` incl. relative dates), content-bearing + hybrid ranking,
  low-info downrank, subject-content bridge, parallel lexical+dense search.
- Entity linking: generic extraction + persistence; graph propagation available,
  gated off by default until boost re-tune proves non-regressing.
- Embeddings: CF AI Gateway Workers AI `bge-base-en-v1.5` on staging (768-d);
  local hash fallback for CI.
- Own suites: OpMem 12/12, marketing vertical 16/16.
- LOCOMO smoke (1 conv / 30 Q, staging dense, entity OFF, gpt-oss): **16/30**;
  multi-hop **3/10**; search p50 ~730–890 ms. See [locomo-smoke.md](../benchmarks/locomo-smoke.md).

---

## What the leaders do (and Brainy's status)

| Technique | Source | Brainy |
| --- | --- | --- |
| Single-pass ADD-only extraction | Mem0 v3 | Done (provider extract + episodes) |
| Entity linking + entity-boosted retrieval | Mem0, Zep, A-MEM | Infra done; ranking gated |
| Multi-signal fusion (semantic + BM25 + entity) | Mem0 | Partial; IDF gated pending re-tune |
| Temporal metadata + query temporal-intent | Mem0 | Partial: `observed_at` + when-query boosts |
| Bi-temporal fact invalidation / supersession | Zep/Graphiti | **ENG-86 v1 shipped** — see [supersession-v1.md](./supersession-v1.md); pack auto-rules later |
| Entity graph + Personalized PageRank multi-hop | HippoRAG | Prototype (rerank-only, gated) |
| Note construction + memory evolution | A-MEM | Not started |
| Declarative vertical packs | (gap in literature) | **Differentiator** — Paper 2 thesis |
| Outcome-grounded belief / stop-loss | (gap vs BeliefMem) | Spec + fixtures — Paper 3 |

---

## Doctrine (non-negotiable)

1. No LOCOMO speaker/answer special-cases in product or eval answerers; no GT padding.
2. Same pins across iterations; attribute deltas to product commits.
3. Fair comparison or no claim — do not cite Mem0’s ~92 against gpt-oss 30-Q smoke.
4. Experimental boosts stay opt-in until OpMem + vertical + same-pin LOCOMO non-regress.
5. Speed (p50/p95, tokens/query) is first-class in every public table.

---

## Research ↔ engineering map

| Phase | Engineering | Paper |
| --- | --- | --- |
| P0 | Dense emb, content-dense ranking, latency | — (baseline) |
| **P1** | **Temporal supersession (ENG-86)** | Strengthens Paper 1 + 2 |
| P2 | Multi-hop product path (graph re-tune, multi-span) | L4 LOCOMO post |
| P3 | Multi-signal fusion default-on when proven | Conversational claim |
| P4 | Full LOCOMO + fair Mem0 re-run; LongMemEval | Public L4/L5 |
| P5 | Finance pack on unchanged runtime | **Paper 2** |
| P6 | Longitudinal outcome / conviction loops | **Paper 3** |

**Paper 1 (OpMem)** ships first — fixtures already beat Mem0; no LOCOMO SOTA required.
Details: [paper-topics.md](./paper-topics.md).

---

## Ordered product path (highest leverage)

1. **Temporal supersession (ENG-86).** `supersedes_id` + `superseded_at`; search excludes
   superseded by default; supersede API + domain-event batch invalidation; opt-in historical recall.
2. **Multi-hop product path.** Entity-graph re-tune off the 30-Q smoke; multi-span aggregation
   in product retrieval (not eval hacks).
3. **BM25/IDF + entity fusion default-on** after staging A/B non-regression.
4. **Full LOCOMO + LongMemEval** under pinned judges; re-run Mem0 under identical pins when possible.
5. **Finance pack + Paper 2**; then outcome-loop **Paper 3**.

---

## Measurement discipline

- Pin: dataset URL + SHA256, Brainy URL/commit, embed model, judge/answerer, temperature=0, top_k, ingest mode.
- Publish honest mid scores; empty latency cells beat invented ones.
- Reproduce dense runs with `evals/tools/local_embeddings_server.py` when gateway blocked.

---

## Why not claim “SOTA” yet

Mem0’s headline LOCOMO uses a managed platform, top-200 budget, and a GPT-class judge.
We already **lead on operational correctness** and **vertical governance**. Conversational
recall is mid and improving (14→16/30) without benchmax. A fair surpass claim waits for
identical pins on axis C plus published wins on A/B/D.
