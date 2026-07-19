# Public benchmark ladder

Path from **today’s OpMem/moat** to **LOCOMO / LongMemEval / BEAM**-style public research, without claiming SOTA early.

Reference industry surfaces:

- [mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks) — open runners
- [SuperMemory research](https://supermemory.ai/research/) · [LongMemBench](https://supermemory.ai/research/longmembench/) — packaging / narrative
- [snap-research/locomo](https://github.com/snap-research/locomo) — LOCOMO ground truth

---

## Layers (do in order)

```
L0  Own CI suites          ← DONE (parity, vertical, OpMem vs Mem0+verbatim)
L1  Research portal + cites ← DONE (this folder + benchmarks/README)
L2  Brainy adapter for public harness ← DONE (evals/public/)
L3  LOCOMO smoke (top-k / subset) — **remeasured 6/30** post async+event-time (2026-07-19); L4 gate unmet
L4  Full LOCOMO + latency/tokens + blog
L5  LongMemEval + BEAM subset
L6  Vertical “MarketingMem” public track (Brainy’s differentiation)
```

---

## L0 — shipped

| Metric | Brainy | Mem0 | Artifact |
| --- | ---: | ---: | --- |
| Parity | 4/4 | 4/4 | [staging report](../benchmarks/staging-competitive-report.md) |
| OpMem | 12/12 | 9/12 | same |
| Marketing vertical | 16/16 | N/A | [moat](../benchmarks/marketing-moat-report.md) |

---

## L2 — Brainy adapter (DONE)

Shipped **option B**: in-repo [`evals/public/`](../../evals/public/) with Mem0-compatible `UnifiedResult`, `RunManifest` pins, Brainy HTTP backend, lexical + OpenAI judges.

Contract: [proveable-eval-framework.md](./proveable-eval-framework.md)

```bash
cd evals && python -m public.locomo.run_smoke --conversations 1 --questions 30
```

---

## L3 — LOCOMO smoke (publish mid scores OK)

- Dataset: [locomo10.json](https://github.com/snap-research/locomo/blob/main/data/locomo10.json)
- Start with **1 conversation** or **category subset** (~30–50 Qs) to bound cost
- Output: `docs/benchmarks/locomo-smoke.md` with:

```markdown
| System | Overall | Single-hop | Multi-hop | Temporal | Adversarial | p95 latency | tokens/q |
| Brainy | TBD | TBD | TBD | TBD | TBD | TBD | TBD |
| Mem0   | TBD | ... |  |  |  |  |  |
```

**Do not** compare to Mem0’s published 66.9% until we match their judge + settings; either re-run Mem0 ourselves or mark cells “from Mem0 blog (cite)” with date.

---

## L4 — Full public post

Template (copy into `docs/research/posts/YYYY-MM-locomo.md`):

1. Headline metric
2. Systems + versions
3. Result tables (accuracy × category, latency, tokens)
4. Failure analysis (where Brainy loses — intentional)
5. Methodology + outlinks to LOCOMO + harness
6. Reproduce block

---

## L5 — LongMemEval / BEAM

Same adapter; larger cost. Schedule after L3 confidence.

---

## L6 — MarketingMem (differentiation)

Public track that *Mem0 LOCOMO won’t cover*: brand rules, campaign lifecycle, durable suppress, principle > preference. Ship as sibling to LOCOMO — not a substitute.

Fixtures already exist under `fixtures/vertical/marketing/` (16) + OpMem (12). Expand to a named public matrix + blog once LOCOMO smoke exists.

---

## Cost / ops notes

- LOCOMO full + judge LLM ≈ material OpenAI spend; use staging Brainy (already up)
- Pin: judge model, embedding if any, Brainy commit SHA, Mem0 API date
- Never commit API keys; results JSON only

---

## Decision log

| Date | Decision |
| --- | --- |
| 2026-07-14 | Keep OpMem as CI gate; treat LOCOMO as Track B research, not merge gate |
| 2026-07-14 | Prefer in-repo `evals/public/` adapter over hard-forking Mem0 harness |
| 2026-07-14 | Publish incomplete tables rather than estimated scores |
| 2026-07-14 | Proveability = dataset SHA + brainy commit + judge pins; lexical ≠ publishable J-score |
| 2026-07-14 | Linear Track D: milestone Public Proveable Eval Framework (ENG-162…167) |
| 2026-07-15 | Anti-benchmax: LOCOMO fails = product bugs; no dataset special-casing; re-measure after product |
| 2026-07-15 | L3 smoke 2/30 — 28/30 retrieval miss; driver = extract+rank+no event time |
| 2026-07-19 | Post ENG-172 async path + relative event-time: L3 smoke **6/30 (20%)**; temporal 5/16; multi-hop still 0/10. L4 gate (≥12/30) not met — next: real embeddings + multi-hop synthesis |
