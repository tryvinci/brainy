# Competitive positioning — Brainy vs industry memory systems

**Date:** 2026-08-06  
**Purpose:** Realistic answer to “where would we stand on a fair industry benchmark?”  
**Rule:** Same-pin evidence only for lead/trail claims. Competitor *headlines* are context, not scoreboard rows.

**Related:** [external-agent-assessment-pack.md](../research/external-agent-assessment-pack.md) · [METHODOLOGY.md](./METHODOLOGY.md) · [locomo-samepin-brainy-vs-mem0.md](./locomo-samepin-brainy-vs-mem0.md) · [staging-competitive-report.md](./staging-competitive-report.md)

---

## Executive answer

On a **fair, same-harness** comparison today, Brainy is a **leader on operational correctness and governed vertical memory**, and a **mid-pack / not-yet-credible challenger on conversational SOTA benches** (LoCoMo / LongMemEval / BEAM).

| Axis | Realistic stand vs industry | Evidence class |
| --- | --- | --- |
| Operational memory (suppress / correct / supersede / isolation) | **Leads Mem0** on same fixtures | Same-pin |
| Marketing / brand vertical governance | **Leads Mem0** on same fixtures | Same-pin |
| Support vertical | Green Brainy-only; **no competitor pin** | Incomplete |
| Conversational LoCoMo (fair) | **Beats Mem0 on one old same-pin smoke**; **far behind** published Mem0/Zep/Hindsight headlines | Mixed |
| LongMemEval / BEAM | **Not competitive** on absolute Brainy pins; no fair competitor run | Absolute only |
| Latency / cost | Smoke ~0.7–1.5s search; load p50 ~2.4s — **not** industry-leading vs Zep/Mem0 marketing claims | Incomplete fair pin |

**One-line publishable wedge:** *Brainy wins where agents need durable, governed, correctable memory (ops + vertical packs). It does not yet win the headline conversational memory race.*

### Constructed bake-off scorecard (if run tomorrow, equal harness)

| Track | Expected finish | Confidence |
| --- | --- | --- |
| OpMem / correction & suppression | **1st vs Mem0** (already proven) | High |
| Marketing governed memory | **1st vs Mem0** (already proven) | High |
| Support vertical | 1st until a competitor adapter exists | Medium (Brainy-only) |
| LoCoMo same-pin smoke | Likely **ahead of Mem0 overall**, still **behind Mem0 on multi-hop** until reader lands; far behind Mem0/Zep *headlines* | Medium (old pin stale) |
| LoCoMo / LME full suite vs SOTA gates | **Not competitive** at ≤50% / 4% absolute | High |
| Latency + cost bake-off | Unknown / likely mid-pack under load | Low (no fair pin) |

---

## Scoreboard A — Fair same-pin (use this for claims)

These are runs where Brainy and the competitor used the **same fixtures / judge / budget** (or the same Brainy harness adapter).

### A1. Operational memory (OpMem)

| System | Score | Notes |
| --- | ---: | --- |
| **Brainy** | **12/12** (later **13/13** expanded) | Staging |
| **Mem0 Platform** | **9/12** | Fails correction / suppression / update cases |
| Verbatim baseline | 9/12 | Control |

Sources: [staging-competitive-report.md](./staging-competitive-report.md), [opmem-staging-vs-mem0.json](./opmem-staging-vs-mem0.json), [opmem-latest.json](./opmem-latest.json).

**Stand:** Clear lead vs Mem0 on operational correctness. This is Brainy’s strongest fair claim.

### A2. Marketing vertical

| System | Score | Notes |
| --- | ---: | --- |
| **Brainy** | **15/16** empirical (later **17/17** declared suite) | Lifecycle, brand rules, outcomes |
| **Mem0 Platform** | **4/16** empirical | Same fixture JSON |

Source: [docs/vertical/marketing-mvp-vs-mem0.md](../vertical/marketing-mvp-vs-mem0.md).

**Stand:** Clear lead on governed marketing memory. Mem0 is not a peer on this track.

### A3. API parity slice

| System | Score |
| --- | ---: |
| Brainy | 4/4 |
| Mem0 | 4/4 |

Source: [competitor-parity-staging.json](./competitor-parity-staging.json). Tie — baseline API competence, not differentiation.

### A4. LoCoMo smoke — last Mem0 same-pin (2026-07-29)

| System | Overall | Temporal | Multi-hop | Open |
| --- | ---: | ---: | ---: | ---: |
| Brainy (diversify peak)* | 19/30 | 13/16 | ~2–5/10 | strong |
| Mem0 Platform | 12/30 | 2/16 | **6/10** | 4/4 |

\*Peak Brainy pin includes pre-deoverfit surface forms; honest post-deoverfit Brainy smoke was **16/30**. Still beats Mem0 overall on that stack; Mem0 **won multi-hop**.

Source: [locomo-samepin-brainy-vs-mem0.md](./locomo-samepin-brainy-vs-mem0.md).

**Stand:** Overall lead on that old same-pin; **not** a multi-hop lead. **Stale** relative to current Fusion V2 / planner / `/recall` stack — a fresh Mem0 same-pin is still open.

---

## Scoreboard B — Brainy absolute pins (not competitor compares)

Use these to describe Brainy’s own maturity. Do **not** put Mem0 92% / Zep 75% in the same table cell.

| Bench | Brainy result | Pin | Gate |
| --- | ---: | --- | --- |
| LoCoMo full 3-seed | **≈49.8%** (MH ≈26%) | 2026-08-01 publish stack, gpt-oss | Gate R1 ≥75 **missed** |
| LoCoMo smoke LLM-over-search | **60% (18/30)** | `c223da3d`, 2026-08-04 | Directional |
| LoCoMo smoke product `/recall` | **43.3% (13/30)**, MH 50% | `f722342a`, 2026-08-05 | Reader baseline |
| LongMemEval-S stratified | **4% (4/100)** | 2026-08-01 | Far from credible |
| BEAM-100K sample | **40% (8/20)** | 2026-08-01 | Tiny sample |
| Support vertical | **4/4** | 2026-08-04 | No competitor |
| Search latency (recall smoke) | p50/p95 ≈ **683 / 1108 ms** | 2026-08-05 | OK for smoke |
| Search latency (load c=8) | p50/p95 ≈ **2.4 / 5.0 s** | 2026-07-31 | SLO miss |

Sources: assessment pack §5; [locomo-full-publish-summary.json](./artifacts/locomo-full-publish-summary.json); [lme-s-100.md](./artifacts/lme-s-100.md); [beam-100k-c0-async.md](./artifacts/beam-100k-c0-async.md); [locomo-smoke-recall-reader-20260805.md](./artifacts/locomo-smoke-recall-reader-20260805.md).

---

## Scoreboard C — Industry headlines (context only)

These are **published / blog / paper claims**. Different judges, budgets, top-k, and often managed platforms. **Do not claim Brainy is “X points behind” these as if same-pin.**

| System | Claimed LoCoMo | Claimed LME | Claimed BEAM | Caveat |
| --- | ---: | ---: | ---: | --- |
| Mem0 | ~92.5 | ~94.4 | ~64.1 (1M) | GPT / large top-k / full suite |
| Hindsight (AMB) | ~92.0 | ~94.6 | ~73.9 (1M) | External citation |
| Zep | ~65–84 (disputed; often ~75) | ~71–90 (docs differ) | — | Contested methodology; **0 in-repo runs** |
| Letta (filesystem+grep) | ~74 | — | — | Citation only |

Rough visual (absolute Brainy vs headlines — **illustrative, not fair**):

```text
LoCoMo overall (illustrative)
Mem0 / Hindsight headlines ████████████████████████████████████  ~92%
Zep / Letta headlines       ██████████████████████████████       ~74–75%
Brainy full 3-seed          ████████████████████                 ~50%
Brainy /recall smoke        █████████████████                    ~43%
Gate R1                     ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  75%
```

---

## Where we would stand in a constructed industry bake-off

If we ran a **credible public bake-off** tomorrow with equal judge, top-k, ingest path, and token budget:

### Track 1 — Operational / agent reliability
**Brainy: favorite.** Same-pin OpMem already shows Mem0 missing correction/suppression/update semantics Brainy implements. Most “memory for agents” products are weak here; this is the natural launch axis.

### Track 2 — Vertical / enterprise governed memory
**Brainy: favorite vs Mem0** on marketing fixtures; support pack green but needs competitor adapters. Zep/Graphiti may compete on temporal KG narrative, but Brainy’s pack FSM + lifecycle is a different product claim (Postgres + packs, not graph DB).

### Track 3 — Conversational long-horizon QA (LoCoMo / LME / BEAM)
**Brainy: underdog / mid-pack at best.** Absolute pins (~50% full LoCoMo, 4% LME-100) are not SOTA-credible. Even the best recent smoke (60% LLM-over-search) is a 30-Q directional pin. Product `/recall` deterministic reader (43%) shows the synthesis gap clearly. Industry leaders would win this track until reader quality + multi-seed proof land.

### Track 4 — Latency / cost
**Brainy: unproven vs marketing claims.** Smoke latency is decent; load latency is not. No fair Mem0/Zep token+latency pin on the current stack.

---

## Positioning matrix (honest)

| If the buyer cares about… | Pitch | Risk if overclaimed |
| --- | --- | --- |
| Agent memory that can be corrected, suppressed, superseded | **Lead with OpMem 12/12 vs Mem0 9/12** | Expanding OpMem without re-measuring Mem0 on 13 tasks |
| Brand / campaign governed memory | **Lead with marketing 15/16 vs Mem0 4/16** | Treating declared 17/17 as empirical Mem0 loss |
| “We beat Mem0 on LoCoMo” | Only with **fresh same-pin**; old pin is stale | Quoting Mem0 92% next to Brainy 50% |
| “SOTA conversational memory” | **Do not claim** | Immediate credibility loss vs Mem0/Zep/Hindsight |
| Postgres / self-host / five-plane architecture | Architecture story is strong; **proof is mid-migration** | Selling planes as finished |

---

## What a fair industry benchmark package would require next

To turn this brief into a publishable bake-off (not headlines):

1. **Mem0 same-pin LoCoMo smoke** on current staging (`BRAINY_USE_RECALL` + LLM-over-search dual report), identical judge/top-k.
2. **Mem0 same-pin OpMem 13** + marketing 17 (refresh empirical, not declared).
3. **LME-100 stratified** completion + failure ledger (Brainy absolute first; Mem0 adapter second).
4. Optional Lane-B: Zep/Letta only if adapters exist under equal budget — otherwise keep them in Scoreboard C.

Until (1)–(3), the honest public line remains:

> Brainy leads Mem0 on operational and marketing vertical fixtures under equal conditions. Conversational benches are still below SOTA gates and are not yet re-compared same-pin on the post–architect stack.

---

## Appendix — pins to cite

| Claim | Cite |
| --- | --- |
| OpMem lead | `docs/benchmarks/staging-competitive-report.md` |
| Marketing lead | `docs/vertical/marketing-mvp-vs-mem0.md` |
| Old LoCoMo same-pin | `docs/benchmarks/locomo-samepin-brainy-vs-mem0.md` |
| Full LoCoMo ~50% | `docs/benchmarks/artifacts/locomo-full-publish-summary.json` |
| LME 4% | `docs/benchmarks/artifacts/lme-s-100.md` |
| Latest `/recall` smoke | `docs/benchmarks/artifacts/locomo-smoke-recall-reader-20260805.md` |
| Methodology / anti-mismatch | `docs/benchmarks/METHODOLOGY.md`, assessment pack §5 |
