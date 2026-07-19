# Blog handoff: Why vertical memory?

**Audience of this doc:** Writer / research agent drafting a blog about the need for *vertical* (domain-specialized) memory systems for AI agents.

**Product / repo:** [tryvinci/brainy](https://github.com/tryvinci/brainy) — Go memory service; marketing is the first vertical wedge. Do **not** invent metrics. Use only the numbers and claims below.

**Release pin:** `v0.1.0` developer preview · OpMem + Mem0 staging comparison dated ~2026-07-14

---

## One-line thesis for the post

Generic agent memory optimizes for *remembering stuff*. Vertical memory optimizes for *governed domain behavior* — what must never be said, what outranks a soft preference, what expires when a campaign ends, and whether a correction sticks after the wrong content is re-ingested.

---

## The problem (editorial framing)

Most trending “agent memory” work (Mem0, GAM, LightMem, HaluMem, LMEB, LOCOMO-style evals) treats memory as:

1. Extract facts from dialogue  
2. Store / consolidate  
3. Retrieve by similarity for the next turn  

That is necessary but not sufficient for **production vertical agents** (marketing brand voice, finance theses, clinical policy, etc.). Those agents fail when:

| Failure | Why generic RAG/memory misses it |
| --- | --- |
| Soft preference overrides a hard brand rule | No precedence among memory types |
| Archived / ended campaign still surfaces in copy | No lifecycle machine |
| “Never mention competitor X” leaks after paraphrase or re-ingest | Suppression isn’t durable |
| Correction loses after the old sentence is re-seen | No correction stickiness |
| Agency tenant A’s brand bleeds into tenant B | Isolation is assumed, rarely stress-tested |
| Two conflicting preferences coexist unsorted | No recency / supersede policy |

**Useful contrast with public literature (allude, don’t claim SOTA):**

- Surveys (“Memory in the Age of AI Agents”, Dec 2025) still frame memory by form/function/dynamics — not by *domain configuration*.
- Domain *agents* (FinMem, AD-Bench, etc.) often verticalize by forking the system; few publish **one runtime + declarative packs**.
- Public suites (LOCOMO / LongMemEval) score recall & multi-hop QA — they do **not** certify operational correctness (suppression, correction, isolation). Brainy is explicit that OpMem ≠ LOCOMO.

---

## Brainy’s architectural claim (safe to explain)

> **“Verticals are packs, not code paths.”**

Three layers ([`docs/vertical/verticalization-model.md`](../vertical/verticalization-model.md)):

1. **Cognitive primitives** (runtime, domain-agnostic): Principle, IdentityPrior, Episode, Pattern, Belief, Outcome, Experiment, TasteSignal, Reflection — [`docs/brainy/architecture/00-cognitive-primitives.md`](../brainy/architecture/00-cognitive-primitives.md)
2. **Vertical pack** (YAML): vocabulary → primitive mapping, metadata schemas, lifecycle rules, rank weights, eval fixtures — e.g. `packs/marketing/v1/pack.yaml`
3. **Domain labels only** in customer language: `brand_rule`, `voice_profile`, `campaign` — **not** per-vertical DB enums

**Rank sketch (marketing):**  
Principle > IdentityPrior > Belief > Pattern > Episode, with scope boosts (active campaign) and TasteSignal as tie-break.

**Belief loop (differentiate from observation-only belief mem systems):**  
Beliefs carry conviction; outcomes challenge them past a stop-loss — see architecture docs `02-belief-lifecycle.md`, `04-conviction-stop-loss.md`. Good blog paragraph: other systems revise from new *observations*; vertical ops often need revise from *task outcomes*.

---

## Concrete marketing stories (use as vignettes)

From [`docs/vertical/marketing-use-case-map.md`](../vertical/marketing-use-case-map.md) and golden fixtures:

1. **Brand voice agent** — hard taboos as Principles; soft style as IdentityPrior; Principle always wins (`bv01`, `bv02`, `bv08`)  
2. **Campaign manager** — archived campaign hidden; active outranks completed (`lc01`, `lc02`)  
3. **Multi-brand agency** — subject/tenant isolation (`bv06`, OpMem `iso01`–`iso03`)  
4. **Correction stickiness** — user corrected “warm tone” → “crisp professional”; re-ingest of the old line must not revive it (`bv04`, OpMem `cor02`)  
5. **Outcome → belief** — campaign result updates belief ranking (`ob05`)  
6. **Scoped segments** — VIP prefers email, Mobile prefers SMS; both coexist under `scope=` (`sg10`)

These are **reproducible fixtures** under `fixtures/vertical/marketing/` and `fixtures/opmem/` — strong “show your work” material for the blog.

---

## Numbers you may cite (verified in-repo)

### OpMem v0 — operational correctness (domain-neutral)

| | Brainy | Verbatim (raw-RAG stand-in) | Mem0 (staging, platform API) |
| --- | ---: | ---: | ---: |
| OpMem | **12/12** | 9/12 | **9/12** |
| Thin-slice parity | 4/4 | — | 4/4 |

**Sources:** [`docs/benchmarks/opmem-baseline-report.md`](../benchmarks/opmem-baseline-report.md), [`docs/benchmarks/staging-competitive-report.md`](../benchmarks/staging-competitive-report.md)

Mem0’s OpMem misses (allude as “why generic stores fail ops tests”):

- `cor02` correction stickiness  
- `sup03` durable forget (suppressed memory resurfaces on re-remember)  
- `upd01` stale fact not outranked by newer restatement  

### Marketing vertical moat

- **16/16** marketing fixtures; Gate M3 Tier 1–4 green  
- Differentiation vs Mem0 (from moat report): principle-over-preference, lifecycle suppression, outcome→belief, taste ranking, scoped segments, never-sentence → brand_rule  

**Source:** [`docs/benchmarks/marketing-moat-report.md`](../benchmarks/marketing-moat-report.md)

### Launch framing already drafted

[`docs/benchmarks/launch-narrative.md`](../benchmarks/launch-narrative.md) — can be adapted into blog voice; don’t treat as published prose until legal/marketing review.

---

## Honest limits (must include for credibility)

From launch narrative + research portal:

1. **CI uses a deterministic local embedder** for reproducibility; Mem0 may still win on provider-quality semantic embeddings at scale until Brainy ships provider extraction.  
2. **OpMem / marketing fixtures are not LOCOMO or LongMemEval.** We have not published LOCOMO accuracy/latency/token numbers yet — ladder is tracked in [`docs/research/public-bench-ladder.md`](./public-bench-ladder.md).  
3. **Finance pack is Gate M4 research** — marketing is the only shipped vertical pack. The *architecture* claims multi-vertical; the *evidence* is marketing-first.  
4. Prefer empty cells to invented latency/token numbers.

---

## Suggested blog outline

1. **Hook** — Agent wrote on-brand… then named the competitor we banned last quarter. Memory “worked”; governance didn’t.  
2. **Generic memory’s job** — durable context, RAG over history (cite Mem0 / surveys lightly).  
3. **Vertical agent’s job** — precedence, lifecycle, isolation, correction, outcome feedback.  
4. **Thesis** — verticals as config over cognitive primitives, not forked schemas.  
5. **Evidence** — OpMem 12/12 vs Mem0/verbatim 9/12; marketing moat fixtures; one vignette deep-dive (taboo or campaign lifecycle).  
6. **Reproduce** — clone + `docker compose` + `python3 evals/run_opmem.py` / `run_vertical_eval.py` (from launch narrative; use `messages[]` ingest shape, not `"text"`).  
7. **What’s next** — public-bench ladder (LOCOMO adapter); second vertical when marketing proof holds.  
8. **Close** — Memory for agents is becoming infrastructure; verticals make it trustworthy.

---

## Copy-paste quotes / phrases from the repo

Use sparingly; attribute as product language:

- “Verticals are packs, not code paths.”  
- “Generic memory APIs optimize for semantic search. Brainy optimizes for governed marketing memory.”  
- “Structured memory must get updates, suppressions, and isolation right — not just embedding similarity.”  
- OpMem semantic contracts: **durable forget** (forgotten ≠ resurrect on re-ingest) and **staleness** (restated fact/preference outranks stale).  

---

## Key file index for the writing agent

| Path | Why open it |
| --- | --- |
| `docs/vertical/verticalization-model.md` | Core architecture thesis |
| `docs/brainy/architecture/00-cognitive-primitives.md` | Primitive definitions |
| `docs/vertical/marketing-use-case-map.md` | Jobs + failure modes |
| `docs/benchmarks/launch-narrative.md` | Existing public framing |
| `docs/benchmarks/opmem-baseline-report.md` | OpMem scores |
| `docs/benchmarks/staging-competitive-report.md` | Brainy vs Mem0 staging |
| `docs/benchmarks/marketing-moat-report.md` | Marketing differentiation |
| `docs/research/opmem-spec.md` | OpMem design / why ops ≠ recall |
| `docs/research/public-bench-ladder.md` | What we have *not* claimed yet |
| `docs/research/README.md` | Research portal tone & format rules |
| `packs/marketing/v1/pack.yaml` | Concrete pack example |
| `fixtures/opmem/*.json` | Readable op scripts for examples |
| `fixtures/vertical/marketing/*.json` | Marketing golden scenarios |

---

## Do / don’t for the writing agent

**Do**

- Argue from **failure modes of production agents**, then show evidence.  
- Contrast **operational correctness** vs **retrieval accuracy**.  
- Link GitHub / fixture paths for reproducibility.  
- Keep “marketing first, packs general” honest.

**Don’t**

- Claim LOCOMO / LongMemEval / SOTA.  
- Invent latency, token cost, or user-count metrics.  
- Imply finance (or other verticals) are shipped.  
- Call Brainy a Mem0 fork — Mem0 is a **pinned behavioral reference**, not the codebase.  
- Exaggerate: Mem0 matches **4/4** on thin-slice parity; Brainy’s win is **ops + vertical**, not “remembers more.”
