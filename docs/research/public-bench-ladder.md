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
L2  Brainy adapter for public harness
L3  LOCOMO smoke (top-k / subset) — publish warts and all
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

## L2 — Brainy adapter (next eng)

Goal: plug Brainy into the same harness Mem0 uses so numbers are comparable.

Options (pick one):

| Option | Pros | Cons |
| --- | --- | --- |
| **A.** PR / fork of `mem0ai/memory-benchmarks` + Brainy backend | Same CLI as Mem0 blog | Upstream dependency, Python SDK mismatch |
| **B.** Thin `evals/public/` that downloads LOCOMO JSON + implements Mem0 harness schema | Stays in-repo | More glue code |
| **C.** Call Mem0 harness with a Brainy “Mem0-compatible” shim API | Fastest path if we only expose add/search | Couples to their client |

**Recommend B** for control; mirror their `schema.py` metrics and cite LOCOMO upstream.

Minimum Brainy surface for LOCOMO:

1. Sessionized ingest of dialogue turns → `/ingest` (or batch API)
2. Per-question retrieve → `/memories/search`
3. Optional answerer LLM (OpenAI) — *judge is not Brainy*; pin model + temp=0
4. Record: accuracy by category, p50/p95 search+answer latency, tokens

Staging URL: `BRAINY_BASE_URL` + `BRAINY_API_KEY` as already used.

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
