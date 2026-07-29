# Brainy Pre-Release Master Plan — Beating Mem0 (and the Field) For Real

Status: **Active — this is the program of record for the pre-release cycle**
Date: 2026-07-29
Owner: CTO plan (agent-authored, evidence-cited)
Supersedes as guidance: `path-to-sota.md`, `gap-crush-checklist.md`, `public-bench-ladder.md`
(those remain useful history; this doc governs going forward)

**Execution progress (same day):**
- [x] §9.1 P0 de-overfit + CI denylist (`overfit_denylist_test.go`)
- [x] §9.2 publish stack scaffold (`run_full.py`, `publish-stack-pins.md`)
- [x] §9.3 typed atom substrate (`predicates.go`, `memory_atoms` v13, list predicate admit)
- [x] Re-baseline LOCOMO smoke after de-overfit: **16/30** (was 19/30 peak-with-hacks); OpMem 12/12
- [ ] §9.4–9.6 provider golden suite, embedding flip, memory-benchmarks fork

---

## 0. TL;DR — the ten decisions this plan makes

1. **Stop grinding the 30-question LOCOMO pin.** It has done its job (diagnosis). The remaining wins there are not where the market is. Mem0's platform now reports **92.5 on full LoCoMo, 94.4 on LongMemEval, 64.1 on BEAM-1M at ~7K tokens/query** ([Mem0 README](https://github.com/mem0ai/mem0), April 2026 algorithm). Our honest public position today is a 1-conversation smoke pin. Chasing +1 multi-hop question on that pin is invisible progress.
2. **Pay down the benchmark-overfit debt before anything public.** A material slice of our recent multi-hop gains is LOCOMO-shaped regexes (`transWomanRE`, `homeCountryRE`, `dinosaurLikeRE`, prompt examples naming Caroline/Melanie/Sweden) plus answer-side harvest patterns living in the **eval harness**, not the product. This is exactly the "MemPalace" failure mode that got publicly torn apart in April 2026. We de-overfit first (W1) — it is an existential credibility risk, and it also unblocks real generalization.
3. **The real multi-hop fix is architectural, not more boosts:** typed attribute atoms + a structured recall path for enumeration ("what activities does Melanie do?") + a product-side synthesis/answer layer. Lists fail because vector similarity returns *the best match*, not *coverage*. Fix coverage structurally (W2–W4).
4. **Adopt the industry benchmark portfolio: full LoCoMo (1,540 Q), LongMemEval-S (500 Q), BEAM (100K→1M), plus OpMem as our own public benchmark.** Run them on pinned, publishable stacks with multi-seed error bars. Our current same-pin gpt-oss numbers are not comparable to anyone's published numbers.
5. **Counter-run competitors on *their own* harnesses, and get Brainy into *third-party* harnesses.** Mem0 open-sourced their eval framework ([mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks)); Supermemory's [MemoryBench](https://github.com/supermemoryai/memorybench) and Vectorize's [AMB leaderboard](https://agentmemorybenchmark.ai) are pluggable multi-provider harnesses. A Brainy backend in each is worth more than any self-reported table.
6. **Always publish the three baselines that kill fake claims: full-context, naive RAG, filesystem+grep.** Letta scored **74.0 on LoCoMo with a filesystem and `grep`** — any memory feature that doesn't beat that trio isn't a feature. This is our internal ship gate *and* our public credibility weapon.
7. **Attack Mem0 where their architecture is weak, not where it's strong.** Mem0 v3 is **ADD-only — "nothing is overwritten or deleted."** Mem0's own 2026 state-of-memory report lists *memory staleness*, *cross-session structure* ("change as evolution, not replacement"), and *knowledge update* as open problems. Our supersession/correction/lifecycle machinery (OpMem 12/12 vs Mem0 9/12) is the direct answer. Make "**memory that stays true**" the wedge.
8. **Keep and extend the vertical moat.** Marketing pack: Brainy 15/16 vs Mem0 4/16 empirical. Verticals-as-packs is a structural advantage no benchmark-focused competitor has. Ship the second pack (support/CRM — see §3.4) to prove packs generalize.
9. **Make latency/cost a headline metric.** We're a Go binary on Postgres; Mem0/Zep/Hindsight are Python/graph stacks. Zep markets sub-200ms; Mem0 reports p50 0.88–1.09s. A sub-300ms p50 search at publishable accuracy is a winnable, marketable axis (MemScore-style triple: accuracy / latency / tokens).
10. **Two-stack model policy.** Dev iteration stays on the cheap gpt-oss pin; all *publishable* runs use a frozen "publish stack" (frontier judge + answerer, documented prompts, ≥3 seeds, error bars). Never mix the two in one table again.

---

## 1. Ground truth — where we actually are

### 1.1 Honest scoreboard (as of this doc)

| Axis | Brainy | Best competitor signal | Comparable? |
| --- | --- | --- | --- |
| OpMem (ops correctness, 12 tasks) | **12/12** | Mem0 platform 9/12 (measured, same fixtures) | Yes — our harness, both measured |
| Marketing vertical (16 fixtures) | **15/16** empirical | Mem0 4/16 (measured; strict schema = intentional moat) | Yes, with declared policy |
| LOCOMO smoke (1 convo × 30 Q, gpt-oss pin) | 19/30, MH 5/10 | Mem0 same-pin 12/30, MH 6/10 | Only vs our own same-pin Mem0 run |
| LOCOMO full (10 convos, 1,540 Q) | **not run** (partial 3-convo old-stack: 31/90) | Mem0 platform 92.5; Hindsight 92.0; corrected Zep 75.1; Letta-filesystem 74.0; full-context ~73 | No — we have no comparable number |
| LongMemEval-S (500 Q) | **not run** | Mem0 94.4; Hindsight 94.6; Zep ~71 | No |
| BEAM (1M) | **not run** | Mem0 64.1; Hindsight 73.9 | No |
| Search latency | p50 ~730–890ms (staging, best pins) | Mem0 p50 0.88–1.09s; Zep markets <200ms | Roughly |

Sources: `docs/benchmarks/locomo-smoke.md`, `locomo-samepin-brainy-vs-mem0.md`, `opmem-baseline-report.md`, `docs/vertical/marketing-mvp-vs-mem0.md`, `latency-slo.md`; [Mem0 README/research](https://github.com/mem0ai/mem0), [AMB leaderboard](https://agentmemorybenchmark.ai), [Zep corrected result](https://blog.getzep.com/lies-damn-lies-statistics-is-mem0-really-sota-in-agent-memory/), [Letta filesystem result](https://www.letta.com/blog/benchmarking-ai-agent-memory).

### 1.2 The overfit audit (what must not survive to release)

Exhaustive inventory of LOCOMO-shaped code (from code review, 2026-07-29):

| Location | What it is | Disposition |
| --- | --- | --- |
| `internal/memory/attribute_atoms.go` — `transWomanRE`, `dinosaurLikeRE`, `homeCountryRE`, `singleParentRE`, hardcoded activity list incl. `pottery`, `iAmBareRE` special-casing "transgender" | Deterministic regexes emitting exactly the atoms LOCOMO conv-26 needs | **Replace** with the general typed-atom schema (W2). Keep the *mechanism* (deterministic atom synthesis), generalize the *patterns* into a predicate taxonomy |
| `internal/memory/provider_extractor.go` prompt | Few-shot examples literally naming Caroline/transgender, Melanie/pottery, Sweden; comment "multi-hop LOCOMO fails…" | **Rewrite** with synthetic, non-benchmark examples covering the same predicate classes |
| `internal/memory/service.go` — `looksListQuery` cues (`partake`, `destress`, `kids`…), `relatedIntentTokens` (`identity→gender`, `camped`), `queryAttributeIntentBoost`, `topicAlignmentBoost` "transgender/trans" comment | Query-intent machinery tuned to the 30-Q pin's vocabulary | **Generalize** into an intent classifier over the predicate taxonomy (W3); vocabulary must come from the taxonomy, not the benchmark |
| `internal/memory/record.go` LOCOMO session timestamp layouts | Parses `"3:04 pm on 2 January, 2006"` | **Keep** — timestamp format support is legitimate general capability |
| `internal/embedding/local.go` synonym groups | identity/gender, family/adoption, hobby groups | **Retire** with the hash embedder itself (W3.1) |
| `evals/public/judge.py` — `_HARVEST_PATTERNS` incl. `\bis a (transgender woman)\b`, harvest cues for moved/identity/kids/books/destress | **Answer-side** extraction living in the eval harness | **Move** the harvest capability into the product recall API (W4); the eval answerer must become a thin, dumb client. Benchmark-specific patterns die |

Rule going forward (add to CONTRIBUTING and CI review checklist): **no benchmark surface-form (names, places, exact phrases from any eval set) may appear in product code or prompts.** Predicate classes and general linguistic patterns are fine; answer keys are not. This is the anti-MemPalace clause.

### 1.3 Why multi-hop lists still fail — root cause, stated once, precisely

The 5 remaining smoke-pin failures (q15 activities, q18 camped-where, q19 kids-like, q23 books, q24 destress) share one shape: **enumerate all values of an attribute for a subject, scattered across up to 19 sessions.**

Four compounding causes:

1. **Extraction misses atoms.** Quoted media titles ("Nothing is Impossible", "Charlotte's Web") and nested subjects ("Melanie's kids") don't reliably become atoms. Deterministic regexes only cover the patterns we hand-wrote.
2. **Similarity retrieval optimizes best-match, not coverage.** Embedding + lexical search returns the strongest few matches and near-duplicates; a list answer needs *every* distinct supporting atom. MMR diversification is a band-aid; it cannot guarantee attribute coverage.
3. **Budget mismatch.** We evaluate at top_k=30; Mem0 publishes at top_200 (~7K tokens assembled). Several "misses" are actually rank-31-to-80 items.
4. **No synthesis layer in the product.** Union-of-items assembly currently lives in the eval harness. The product returns ranked memories and hopes the answerer aggregates.

The structural fix (this is the heart of the technical program): **atoms carry `(subject, predicate, value, time)` and enumeration queries execute as indexed predicate scans, fused with similarity retrieval.** "What activities does Melanie do?" is `SELECT value WHERE subject=Melanie AND predicate=activity` — a query no top-k vector search can reliably answer but a database answers trivially. Mem0 approximates this with entity-boosted recall at top_200; we can do it exactly, cheaper, and explainably. That is a *better* architecture, not a benchmark hack — and it generalizes to the vertical packs ("list all active campaigns", "all brand rules").

---

## 2. The landscape — competitors, benchmarks, and the credibility crisis

### 2.1 The benchmark theatre (context that shapes everything)

The agent-memory eval field had a public credibility collapse over the last 15 months. Facts we must design around:

- **LoCoMo's answer key is 6.4% wrong** (99/1,540 questions), giving a theoretical ceiling of ~93.6; the standard GPT-4o-mini judge **accepts 62.8% of intentionally wrong but topical answers** ([locomo-audit](https://github.com/dial481/locomo-audit)). The adversarial category (22.5% of the dataset) is excluded by everyone. Per-category CIs are so wide that 56% of adjacent published comparisons are statistically indistinguishable.
- **The Mem0↔Zep war:** Mem0's paper scored Zep at 65.99; Zep countered with 84 ("Lies, Damn Lies & Statistics"); Mem0's CTO found Zep's arithmetic error (Cat-5 in numerator, not denominator); Zep corrected to 75.14 ± 0.17; Mem0's re-run of Zep's own pipeline said 58.44 ± 0.20. Same system, same benchmark: a 26-point range depending on who runs it.
- **Letta's filesystem result:** a plain agent with `grep` over the transcript scored **74.0**, beating Mem0's then-published 68.5. Full-context alone: ~73.
- **MemPalace (April 2026):** a celebrity-boosted "100% on LoCoMo" repo hit 27K GitHub stars, then collapsed within 48 hours when auditors found top-k larger than the candidate pool, retrieval-only scoring passed off as end-to-end QA, and per-question patches the repo itself called "teaching to the test."
- **The correction wave:** [LoCoMo-Refined](https://github.com/mem-eval-suite/LoCoMo_refined) (1,382 cleaned questions, strict Qwen3-14B judge, 86.3% human agreement vs 43.7% original) drops systems 15–22 points (Mem0 ~64 → 48.9 under the strict judge). Dynamic/on-policy benchmarks (MemoryBench-Tsinghua, AMemGym @ ICLR'26, MEMTRACK, CASCADE, ConvoMem) report that **no memory system consistently beats a good RAG baseline** when evaluation is interactive, and GPT-5 tops out at 60% on MEMTRACK.

**Implications for us (non-negotiable policy):**

1. Never publish a single-seed number. ≥3 seeds (10 for headline claims), with std dev.
2. Always publish judge prompts, answerer prompts, pins, and the harness. (We already have `proveable-eval-framework.md` — it becomes mandatory, not aspirational.)
3. Always include full-context, naive-RAG, and filesystem+grep baselines in any table we publish.
4. Report the accuracy/latency/tokens triple, never accuracy alone.
5. Score on both the original LoCoMo judge (industry comparability) *and* LoCoMo-Refined strict judge (credibility) — leading with the strict number is a differentiator precisely because everyone else's number deflates under it.
6. Treat <10-point deltas as noise in public claims; never claim a category win inside CI noise.

### 2.2 Who to benchmark against (the counter-run roster)

| System | Why they're on the roster | What to measure |
| --- | --- | --- |
| **Mem0** (platform + OSS v3) | Market leader, the named target; open harness | Full portfolio, both their harness and ours; OSS-vs-platform gap is itself a finding |
| **Zep** (Graphiti temporal KG) | Temporal/knowledge-update specialist; latency marketer | LongMemEval temporal + knowledge-update slices; latency head-to-head |
| **Letta** | Filesystem baseline owner; agent-runtime framing | Their filesystem agent is a *baseline* in every run we publish |
| **Supermemory** | MemoryBench owner; commercial rival | Run via their MemoryBench harness (provider PR) |
| **Hindsight (Vectorize)** | AMB leaderboard leader (LoCoMo 92.0, LME 94.6, BEAM-1M 73.9) | Compete on AMB itself |
| **Cognee** | Enterprise graph-memory; on AMB (LoCoMo 80.3, PersonaMem 81.8) | Via AMB |
| **Memobase / LangMem / EverMemOS-MemOS** | Long-tail; EverMemOS numbers publicly contested (92.3 claimed vs 38.4 reproduced) | Track only; cite the reproduction gap when relevant |
| **Baselines (always):** full-context, naive RAG over turns, filesystem+grep agent | The Letta test | Every published table |

### 2.3 The benchmark portfolio (what we run, in priority order)

**Tier 1 — headline (must have numbers before release):**

| Benchmark | Size | Why | Target (publish stack) |
| --- | --- | --- | --- |
| **LoCoMo full** (10 convos, 1,540 Q, cats 1–4) | ~26K tokens/convo | The comparison everyone demands; report original + Refined judges | Gate R1: ≥75 original-judge (beat corrected-Zep 75.1 / Letta-grep 74 / full-context 73). Gate R2 (stretch): ≥85 |
| **LongMemEval-S** (500 Q, 6 types) | ~115K tokens/question | Knowledge-update + temporal + multi-session; where our supersession story shows up as *score* | Gate R1: ≥75. Gate R2: ≥88 |
| **BEAM** (100K → 1M buckets first) | 10 abilities incl. **contradiction resolution, knowledge update, abstention** | ICLR'26; scale story; the ADD-only attack surface | Report honestly at 100K + 1M with tokens; target ≥ Mem0's 64.1 on 1M as stretch |
| **OpMem v1** (ours, expanded to ~30 tasks) | Operational memory: supersession, correction, isolation, deletion | Our differentiation benchmark — must become public + multi-vendor | ≥4 systems in one table (Brainy, Mem0, Zep, Letta), adapters open-source |

**Tier 2 — depth (run once Tier 1 is green):** PersonaMem (preference tracking), ConvoMem (Salesforce, 75K QA), LongMemEval-M (~1.5M tokens), LoCoMo-Refined as primary strict rating.

**Tier 3 — direction (watch; one result before GA):** the dynamic/on-policy wave — MemoryBench (Tsinghua), AMemGym, MEMTRACK, LongMemEval-V2 (agentic trajectories, LAFS leaderboard). These will define 2027 credibility; being early with *one honest on-policy result* buys outsized attention.

**Vertical (ours):** marketing MVP suite (already empirical vs Mem0), plus the second pack's suite (§3.4).

### 2.4 How to counter-run marketing benchmarks fairly (direct answer to the standing question)

Three-lane protocol, every public comparison:

- **Lane A — their harness, their defaults, our system plugged in.** Fork [mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks) and add a Brainy backend (it's modular: `benchmarks/{locomo,longmemeval,beam}` with pluggable clients; their README documents the ingest→search→answer→judge loop and CLI pins). Whatever score Brainy gets *there*, at their top_200/top_50 budgets with their judge prompts, is the number nobody can dispute. Do the same via Supermemory's MemoryBench (`-p brainy` provider PR — providers are pluggable TS modules) and submit to AMB (Gemini answerer/judge, open harness).
- **Lane B — our harness, same pins, both systems.** What we already do (same-pin LOCOMO, OpMem, marketing suite). Keep it: it's the only place we control cost. Policy stays: parity fixtures content-only; vertical fixtures strict-schema (the moat is the point — but *label* the policy in every table footer).
- **Lane C — their published numbers, quoted verbatim, with methodology deltas listed.** Never re-plot their numbers next to ours without the caveat table (budget, judge, model stack, seeds).

Fairness rules learned from the Mem0/Zep fiasco: configure competitors with their SDK best practices (session/user roles, native timestamp APIs, parallel search); when in doubt, open an issue asking them how to configure — the paper trail itself is credibility; give every system identical answerer/judge prompts in Lane B; never modify a system prompt for one system only.

---

## 3. Strategy — where Brainy wins

Four moats, in order of defensibility:

### 3.1 Moat 1: Operational memory correctness ("memory that stays true")

Mem0 v3 is **ADD-only by design** — their docs say "memories accumulate; nothing is overwritten or deleted," with temporal metadata ranking dated instances at read time. Their own 2026 report names **memory staleness** and **cross-session evolution** as open problems. Zep invalidates graph edges but doesn't do correction workflows. Nobody has our suppress/correct/supersede/domain-events/lifecycle machinery as first-class API.

Product thesis: **enterprises need memory with an update/forget/correct contract** (compliance, GDPR, support workflows, brand governance). ADD-only + rank-time-filtering is a liability there: stale facts still exist, still retrievable, still leakable.

Actions:
- OpMem v1: expand 12 → ~30 tasks across supersession, correction-stickiness, isolation, deletion/forget-verification, contradiction handling, temporal-update correctness. Write adapters for Mem0, Zep, Letta, Supermemory (spec: `opmem-spec.md` v1 section). Publish dataset + harness + paper (Paper 1, `posts/2026-07-opmem-v0.md` is the seed).
- Map OpMem abilities onto BEAM's *knowledge update / contradiction resolution / abstention* categories so the public benchmark portfolio independently validates the same story.
- Marketing language: "**ADD-only memory never forgets — including the things it must.**"

### 3.2 Moat 2: Structured recall (typed atoms + predicate scans)

§1.3's fix, productized. Nobody in the field offers *exact* attribute enumeration with provenance ("all activities for subject X, each with source turns"). Mem0's entity boost is probabilistic; a typed atom index is deterministic. This is also the honest generalization of everything the LOCOMO grind taught us — the same capability answers "what does Melanie do to destress" and "list all brand no-go words for client Y".

### 3.3 Moat 3: Verticals as packs

Already proven vs Mem0 (15/16 vs 4/16). Extend, don't dilute: primitives + lifecycle + rank policy in YAML, zero code paths per vertical (`verticalization-model.md` holds). The second pack proves generality.

### 3.4 Moat 4: Latency/cost frontier

Go + Postgres single-binary vs Python pipelines. MemoryBench-Tsinghua found leading systems take **>17s per case for memory construction** and some couldn't finish the benchmark at all; Mem0 platform reports p50 ~0.9–1.1s search. Targets: **search p50 ≤ 300ms, p95 ≤ 800ms at top_k=50 on a 10K-memory subject; ingest→queryable p95 ≤ 5s async.** Publish the accuracy/latency/tokens triple on every run (MemScore convention).

### What we do *not* do

- No external graph database (Mem0 removed theirs; Zep built a proprietary one — we win by staying on Postgres).
- No LOCOMO cue lists, GT-derived padding, per-question patches (the anti-benchmax rules in `execution-plan.md` stay, now with the §1.2 CI clause).
- No chasing the full "cognition loop" (belief challenge/retire, stop-loss, reflection engine from `docs/brainy/architecture/`) into the release-critical path. It stays a research track (§7, Phase R) feeding Paper 3 and pack rank policies; the shipped slice remains primitives + outcome→belief synthesis + supersession.
- No calendar-driven ship: gates, not dates (§7).

---

## 4. Technical program — workstreams

Dependency order: W1 → W2 → (W3 ∥ W4) → W5 → W6, with W7 parallel after W2.

### W1 — De-overfit & guardrails (credibility debt)

Scope:
1. Execute the §1.2 disposition table: delete/replace benchmark-shaped regexes and prompt examples; move harvest to product (lands with W4); retire hash-embedder synonym groups (lands with W3.1).
2. Add CI guard: a denylist test asserting no benchmark surface-forms (names/places/titles from LOCOMO, LongMemEval, BEAM fixtures) in `internal/` or prompts (simple grep-based test with an allowlist for docs).
3. Holdout policy, enforced by convention in the harness: **tuning set = LOCOMO convs 1–3; validation = convs 4–10, run at most once per phase gate; LongMemEval touched only at phase gates.** Log every validation run in `docs/benchmarks/runs-log.md` (date, commit, purpose) so we can prove non-overfitting.
4. Re-baseline after removal: expect the smoke pin to drop (likely to ~13–16/30). That drop is the *honest starting line* — record it, don't hide it.

Exit gate: overfit inventory empty; CI guard green; re-baselined numbers recorded.

### W2 — Extraction v2: typed atoms (the general version of attribute atoms)

Replace per-pattern regex atoms with a **predicate taxonomy** emitted by both extractors:

```
atom := (subject, predicate, value, qualifier?, observed_at, valid_from?, valid_to?, provenance)
```

Predicate taxonomy v1 (~20 classes, extensible per pack): `identity`, `relationship_status`, `origin`, `residence`, `occupation`, `education`, `family_member` (with relation qualifier — covers "Melanie's kids"), `activity`, `activity_purpose` (qualifier: destress/fitness/social — covers q24), `event` (kind + location + time — covers q18 camped-at), `media_consumed` (kind: book/film + verbatim title — covers q23), `preference`, `possession`, `health`, `plan`, `belief`, `skill`, `affiliation`, `contact_fact`, `metric`.

Specification details:
- **Fact-instance temporal typing** (adopt Mem0's July 2026 design, openly): each atom classified as `event | state | plan | preference | relationship | absence`, with `observed_at` always set and `valid_from/valid_to` where inferable. This is the substrate for W5.
- **Verbatim span preservation:** quoted titles and proper nouns must be extracted as exact spans (fixes the broken-quote bug class permanently); extractor emits `source_span` offsets into `source_text`.
- **Nested/related subjects:** `family_member` atoms create sub-subject entities (`melanie::kids`) linked in the entity hub, so "what do Melanie's kids like?" resolves via subject graph, not string luck.
- **Speaker attribution:** keep carry-forward (it's general), but re-implement from dialogue structure (turn ownership), not name-specific patterns; property-test with synthetic multi-speaker fixtures.
- **Prompt:** rewrite provider prompt around the taxonomy with fully synthetic few-shots; keep "when in doubt, extract" (ADD-bias is correct — Mem0 proved it); keep deterministic-baseline merge as the no-LLM floor.
- **Storage:** atoms land in `memory_records` with `metadata.predicate`, `metadata.value_norm` (lowercased normalized value), plus a new index table `memory_atoms(tenant, subject, predicate, value_norm, memory_id, observed_at, valid_to)` for W3's predicate scans. Migration + backfill job for existing records.

Tests: golden-file extraction suite on synthetic dialogues per predicate class (no benchmark text); mutation tests for speaker attribution; span-fidelity tests for quoted values.

Exit gate: extraction P/R ≥ 0.8/0.7 per predicate class on the synthetic suite; conv 1–3 tuning-set multi-hop ≥ pre-W1 peak *without any benchmark-specific code*.

### W3 — Retrieval v2: real signals, normalized fusion, budget tiers

**W3.1 Embeddings for real.** The 128-d hash embedder is a dev toy and its synonym tables are hidden overfit. Default the stack (compose + staging + CI eval profile) to a real model via the existing OpenAI-compatible provider path — recommendation: `bge-m3` or `text-embedding-3-small`-class, 768–1024d, served by `evals/tools/local_embeddings_server.py` locally and a hosted endpoint on staging. pgvector HNSW at native dims (drop the 128-d-only restriction). Hash embedder remains for unit tests only.

**W3.2 Lexical for real.** Replace `ILIKE ANY` with Postgres full-text (`tsvector` + `ts_rank_cd`) as BM25 approximation; evaluate `pg_search`/ParadeDB BM25 if ranking quality demands it. Score must be a normalized real value, not a hit flag.

**W3.3 Fusion, once, normalized.** Replace the boost stack (`0.55+0.45*norm` lexical, `+0.28–0.55` attribute boosts, kind boosts, etc.) with one formula:

```
score = (w_sem·sem_norm + w_lex·lex_norm + w_ent·ent_norm + w_pred·pred_match) / (w_sem+w_lex+w_ent+w_pred)
       · lifecycle_multiplier (pack rank policy)
       · temporal_multiplier (W5)
```

All signals min-max calibrated per query. Weights tuned **once** on the conv-1–3 tuning set + parity + vertical + OpMem jointly (a change that wins LOCOMO but breaks OpMem is rejected — this is what killed IDF/entity last time; the fix is joint tuning, not gating). Then defaults ON — retire `BRAINY_ENTITY_RANKING`/`BRAINY_IDF_RANKING` flags in favor of the single fused path (negative-LOC).

**W3.4 Enumeration path.** Query intent classifier (small-LLM or heuristic-over-taxonomy) detects enumeration/attribute queries → executes predicate scan on `memory_atoms` for the resolved subject(s), unions with fused similarity candidates, guarantees one representative per distinct `value_norm`. This is the list fix.

**W3.5 Budget tiers.** `limit` param already exists; make top_k ∈ {10, 30, 50, 200} first-class in the harness and report accuracy-at-k like Mem0 does (their temporal gains show up at top_50). Publish our token count per assembled context.

**W3.6 Graph experiments (evidence-gated).** Entity hub 1-hop expansion (Mem0-style `linked_memory_ids`) and HippoRAG-2-style query-to-triple + Personalized PageRank over the atom/entity graph are *experiments* behind the harness A/B (≥3 convo tuning set + OpMem + vertical joint check), not default until they win. HippoRAG 2's recognition-memory filter (LLM filters retrieved triples) is the most promising piece for multi-hop precision.

Exit gate: joint eval — tuning-set LOCOMO ≥ W2 gate, OpMem 12/12, vertical ≥ 15/16, hybrid 1/1, latency SLOs met at top_k=50.

### W4 — Recall API: synthesis moves into the product

New endpoint (this is the productization of everything the eval harness currently does out-of-band):

```
POST /recall
{ tenant_id, subject_id, q, mode: "context" | "answer" | "enumerate",
  budget_tokens?, top_k?, vertical?, include_historical? }
→ { context_block?, answer?, items?: [{value, evidence:[memory_id], observed_at}],
    abstained?: bool, memories: [...], explain: {...} }
```

- **`context` mode** (default, no LLM): assembled, deduplicated, token-budgeted context block — the Mem0-equivalent product surface (theirs is ~7K tokens/query; make budget explicit and tunable).
- **`enumerate` mode** (no LLM): distinct values union from the W3.4 path with per-item evidence — the list harvest, generalized, deterministic, explainable.
- **`answer` mode** (optional LLM): provider-configured answerer with a fixed public prompt; supports **abstention** ("not in memory") — needed for BEAM's abstention category and LOCOMO's adversarial category, and an honesty feature enterprises want.
- Eval harness (`evals/public/judge.py`) is then gutted to a thin client: ingest → `/recall` → judge. **No answer-shaping logic outside the product.** This single change makes every score we produce a product score.

Exit gate: harness-parity — thin-client scores within noise of current harness scores on the tuning set; `/recall` p50 within latency SLO in context/enumerate modes.

### W5 — Temporal + supersession v2 ("memory that stays true", as score)

- **Read-time temporal reasoning:** using W2's fact-instance types and validity intervals — current-state queries prefer latest valid `state`, past-tense queries retrieve superseded-at-that-time instances, plan queries prefer future-dated. Temporal multiplier in the W3.3 formula (Mem0 gained +6.7 temporal / +9.1 at top_50 from exactly this; Zep's edge-invalidation is the graph flavor).
- **Write-time supersession v2:** same-`(subject, predicate)` `state` atoms with newer `observed_at` auto-supersede prior ones (lineage recorded, prior retrievable via `include_historical`) — "change as evolution": the old fact stays queryable as history, stops polluting current-state recall. Conflict without time order → both kept + `conflict` flag in explain (complement-first principle from `architecture/03`, minimal slice).
- **Forget contract:** verified deletion (suppress → hard-delete pipeline with tombstone audit), per-tenant retention policy hooks. This is the GDPR/enterprise story ADD-only can't tell.
- Measured by: LongMemEval knowledge-update + temporal-reasoning categories, BEAM knowledge-update + contradiction-resolution + event-ordering, OpMem v1 new tasks.

Exit gate: LongMemEval-S knowledge-update ≥ 70 on dev stack; OpMem v1 (expanded) full pass; no regression on Tier-1 tuning sets.

### W6 — Scale & latency engineering

- **BEAM-scale ingestion:** 1M-token conversations mean ~10⁴–10⁵ atoms/subject. Work items: batch ingest endpoint, worker parallelism (poll → LISTEN/NOTIFY or batched leases), embedding batch calls, `memory_atoms` and entity-hub index review at 10⁵ rows/subject, `expandSubjectContentMemories`/`ListMemories` paths must never full-scan a large subject.
- **SLOs (publishable):** search/`recall(context)` p50 ≤ 300ms / p95 ≤ 800ms at top_k=50, 10K-memory subject; ingest→queryable p95 ≤ 5s async at sustained 50 msg/s/tenant. Prometheus histograms already exist — add a `docs/benchmarks/latency-report.md` generator to the harness so every published accuracy table carries the triple.
- **Cost:** tokens per assembled context tracked in `/recall` explain; report per-benchmark averages (Mem0 reports ~6.9K; beating them at equal accuracy with ≤4K would be a headline).

Exit gate: BEAM-100K bucket ingests + evaluates end-to-end within budget; SLOs green on staging under load.

### W7 — Vertical pack #2 + belief slice (parallel track)

- Choose **customer-support/CRM** as pack #2 (rationale: nearest large market where *correction + supersession + isolation* are the buying criteria — ticket state changes, customer facts evolve, cross-customer isolation is mandatory; it exercises exactly the W5 machinery; finance stays M4-deferred per `go-to-market-roadmap.md`).
- Pack contents: vocabulary → primitives mapping, lifecycle rules (resolved-ticket decay), rank policy, ~16-fixture suite + Mem0 counter-run, one design partner from the beta checklist pipeline.
- Belief slice stays scoped: outcome→belief synthesis (shipped) + conviction metadata exposed in explain; full challenge/retire loop remains research (feeds Paper 3).

Exit gate: pack #2 suite ≥ 14/16 empirical; packs-not-code-paths invariant intact (zero new Go branches for the vertical).

---

## 5. Evaluation program — pins, stacks, and gates

### 5.1 Two-stack policy

| | Dev stack | Publish stack |
| --- | --- | --- |
| Purpose | iteration, A/B, CI | any number that leaves the building |
| Answerer/judge | gpt-oss pin (current) | frontier pin, frozen per release (recommendation: GPT-5-class judge to match [mem0ai/memory-benchmarks] defaults; AMB uses Gemini — match per-harness) |
| Embedder | local bge server | same model, hosted |
| Seeds | 1 | ≥3 (10 for headline claims), std dev reported |
| Judges | lexical or LLM | LLM judge, prompt published; LoCoMo additionally scored with Refined strict judge |
| Budget | top_k=30 | accuracy at k ∈ {10, 50, 200} + token counts |

Budget note: publish-stack runs on full LoCoMo + LongMemEval-S at 3 seeds ≈ ~6K judged questions plus ingestion; provision an API budget line and a nightly-batch runner (harness already checkpoints; extend `run_smoke.py` → `run_full.py` with resume).

### 5.2 Eval ladder (supersedes `public-bench-ladder.md` L-numbers)

| Rung | What | Stack | When |
| --- | --- | --- | --- |
| E0 | Go tests + parity/vertical/OpMem/hybrid fixtures | deterministic | every PR (exists) |
| E1 | LOCOMO tuning set (convs 1–3, all cats) | dev | every retrieval/extraction change |
| E2 | LOCOMO validation (convs 4–10) + LongMemEval-S sample (100 Q stratified) | dev | phase gates only (logged) |
| E3 | Full Tier-1 portfolio | publish | release candidates |
| E4 | Counter-runs (Lane A: Mem0 harness, MemoryBench, AMB submission) | per-harness | release candidates |
| E5 | One dynamic-benchmark result (MEMTRACK or AMemGym or MemoryBench-Tsinghua) | publish | before GA |

### 5.3 Regression gates (carry forward from `09-iteration-and-productization.md`, now enforced in CI)

Public-suite score non-decreasing (>0.05 drop blocks), OpMem non-decreasing (any drop blocks), vertical ≥ 15/16, latency +20% warn / +35% block, joint-tuning rule from W3.3 (a change must not trade OpMem/vertical for LOCOMO).

---

## 6. Publication & marketing narrative

Sequence (each gated on the evidence existing first):

1. **Paper/post 1 — OpMem:** "Operational Memory: benchmarking whether memory systems can update, correct, and forget." ≥4 systems, open harness + adapters. Timed with OpMem v1 (W5 exit). This is our authoritative entry into the benchmark-credibility conversation, on ground we define.
2. **Post 2 — the counter-run report:** Brainy on Mem0's own harness + MemoryBench + AMB submission, full methodology, all three baselines, triple-metric tables, both LoCoMo judges. Frame: "we ran ourselves on our competitor's benchmark harness — here's everything." (The field's history — §2.1 — makes radical transparency the highest-leverage marketing move available; we are small, so we can afford honesty the incumbents can't retrofit.)
3. **Post 3 — structured recall:** typed atoms + enumeration with provenance; the "lists are where memory systems lie" essay with LOCOMO-Refined-strict numbers.
4. **Post 4 — vertical memory (Paper 2 seed):** packs model + marketing & support counter-run results.
5. **Launch narrative** (`launch-narrative.md` updated): "Memory that stays true" — correctness (OpMem), structured recall, verticals, latency/cost triple. Explicit non-claims: we do not claim overall LoCoMo SOTA unless E3 actually shows it with error bars.

Cadence rule: no score leaves the building without harness link + pins + seeds + judge prompt. Quarterly competitor re-audit (`quarterly-moat-review-template.md`) resumes, tracking version drift (Zep v3→v4, Mem0 platform changes, Hindsight releases).

---

## 7. Phases & gates (dependency-ordered; no calendar estimates)

| Phase | Contents | Exit gate |
| --- | --- | --- |
| **P0 — Truth** | W1 de-overfit; CI guard; holdout policy; re-baseline; publish-stack setup (models, budget, `run_full.py`) | Overfit inventory empty; honest baseline recorded on E1+E2; publish stack does one full LoCoMo dry run end-to-end |
| **P1 — Substrate** | W2 typed atoms + migration/backfill; W3.1–.2 real embeddings + FTS | W2 gate; E1 ≥ pre-W1 peak without hacks |
| **P2 — Retrieval** | W3.3 fusion, W3.4 enumeration, W3.5 budgets; W4 `/recall` | W3+W4 gates; E2 validation run logged; multi-hop on validation convs ≥ 60% of category |
| **P3 — Truth-over-time** | W5 temporal + supersession v2 + forget contract; OpMem v1 expansion + adapters | W5 gate; OpMem v1 table with ≥4 systems |
| **P4 — Scale & proof** | W6 scale/latency; E3 full portfolio (3 seeds); E4 counter-runs + AMB submission; MemoryBench provider PR | Gate R1 numbers met (LoCoMo ≥75, LME-S ≥75, BEAM reported, triple published); SLOs green |
| **P5 — Release** | Posts 1–3; launch narrative; beta checklist leftovers (ToS, prod deploy, first partner from `commercial-beta-checklist.md`); pack #2 GA | All publications gated on evidence; partner live on staging→prod |
| **R — Research (continuous)** | W3.6 graph experiments; belief/cognition loop; E5 dynamic-benchmark result; Paper 3 | Evidence-gated promotions into W-streams |

Gate R2 (stretch, post-release trajectory): LoCoMo ≥85 original-judge, LME-S ≥88, BEAM-1M ≥64, AMB leaderboard top-3 on ≥2 datasets — i.e., parity-or-better with Mem0/Hindsight on their home turf while holding the OpMem/vertical/latency moats they can't match.

---

## 8. Risks & mitigations

| Risk | Likelihood | Mitigation |
| --- | --- | --- |
| De-overfit (P0) craters scores and morale | High (expected) | Frame as re-baseline; the W2–W4 program is specifically designed to win the points back generally; keep the old numbers in history docs as "peak with hacks" |
| Publish-stack API cost balloons | Medium | Nightly batch + checkpointing; sample-based E2 (100-Q stratified LME); full runs only at RC |
| Mem0 ships again mid-cycle (they ship monthly) | High | We don't compete on their release cadence; moats 1–4 are architecture bets they've publicly declined (ADD-only, Python, no packs). Quarterly re-audit adjusts targets |
| LoCoMo fully discredited before we publish | Medium | Portfolio hedges it: LME-S, BEAM, OpMem, AMB; leading with Refined-strict judge positions us on the right side of the correction |
| Third-party harnesses reject/stall our PRs (MemoryBench provider, AMB) | Medium | Lane A works regardless (fork + publish); the PR attempt itself is content |
| Enumeration path adds latency on big subjects | Medium | It's an indexed scan (cheap); guard with W6 SLO tests; budget caps in `/recall` |
| One-person-cycle scope creep | High | Phases are strictly gated; W3.6/belief/dynamic-bench live in R and cannot block P-gates |

---

## 9. Immediate next actions (first PR-sized bites, in order)

1. P0: delete `attribute_atoms.go` benchmark regexes behind a temporary flag, add the CI denylist guard, record re-baseline on E1/E2. (One PR.)
2. P0: stand up publish stack — `run_full.py` (10-convo, resume, seeds, both judges), model pins doc, one dry run. (One PR + one budget decision.)
3. P1: `memory_atoms` schema + predicate taxonomy constants + deterministic extractor emitting typed atoms for 5 pilot predicates (`activity`, `media_consumed`, `event`, `family_member`, `origin`). (One PR.)
4. P1: provider prompt rewrite on the taxonomy with synthetic few-shots + golden-file suite. (One PR.)
5. P1: embedding default flip (compose/staging/CI to bge via provider path) + pgvector native dims. (One PR.)
6. Fork `mem0ai/memory-benchmarks`, scaffold the Brainy backend against `/ingest`+`/memories/search` (later `/recall`). (One PR in the fork; tracks Lane A from day one so W2–W4 progress is measured on their harness continuously.)

---

## Appendix A — Source index (external)

- Mem0 April 2026 algorithm + numbers: [github.com/mem0ai/mem0](https://github.com/mem0ai/mem0) README; [docs.mem0.ai/core-concepts/memory-evaluation](https://docs.mem0.ai/core-concepts/memory-evaluation); [state-of-ai-agent-memory-2026](https://mem0.ai/blog/state-of-ai-agent-memory-2026); [temporal reasoning post](https://mem0.ai/blog/introducing-temporal-reasoning-in-mem0) (Jul 28, 2026)
- Mem0 open harness: [github.com/mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks)
- LoCoMo audit: [github.com/dial481/locomo-audit](https://github.com/dial481/locomo-audit); LoCoMo-Refined: [github.com/mem-eval-suite/LoCoMo_refined](https://github.com/mem-eval-suite/LoCoMo_refined)
- Field critique synthesis (Mem0/Zep/Letta dispute, MemPalace, dynamic-benchmark wave): "The Benchmark Theatre" ([essays.bloo-mind.ai](https://essays.bloo-mind.ai/posts/2026-05-20-mem-eval/))
- Zep paper: [arXiv:2501.13956](https://arxiv.org/abs/2501.13956); corrected LoCoMo claim thread: getzep/zep-papers#5
- Letta filesystem result: [letta.com/blog/benchmarking-ai-agent-memory](https://www.letta.com/blog/benchmarking-ai-agent-memory)
- LongMemEval: [arXiv:2410.10813](https://arxiv.org/abs/2410.10813); LongMemEval-V2 (agentic, LAFS): [github.com/xiaowu0162/LongMemEval-V2](https://github.com/xiaowu0162/LongMemEval-V2)
- BEAM + LIGHT (ICLR 2026): [arXiv:2510.27246](https://arxiv.org/abs/2510.27246)
- AMB leaderboard (Hindsight/Vectorize): [agentmemorybenchmark.ai](https://agentmemorybenchmark.ai); harness: [github.com/vectorize-io/agent-memory-benchmark](https://github.com/vectorize-io/agent-memory-benchmark)
- Supermemory MemoryBench: [github.com/supermemoryai/memorybench](https://github.com/supermemoryai/memorybench)
- HippoRAG 2 (query-to-triple PPR + recognition memory): [arXiv:2502.14802](https://arxiv.org/abs/2502.14802)
- Dynamic benchmarks: MemoryBench-Tsinghua [arXiv:2510.17281](https://arxiv.org/abs/2510.17281); AMemGym [arXiv:2603.01966](https://arxiv.org/abs/2603.01966); MEMTRACK [arXiv:2510.01353](https://arxiv.org/abs/2510.01353); CASCADE [arXiv:2605.06702](https://arxiv.org/abs/2605.06702)

## Appendix B — Target-number summary (publish stack)

| Benchmark | Floor (Gate R1) | Stretch (Gate R2) | Beat-this context |
| --- | --- | --- | --- |
| LoCoMo full, original judge | ≥75 | ≥85 | full-context ~73, Letta-grep 74.0, Zep-corrected 75.1, Cognee 80.3, Hindsight 92.0, Mem0 platform 92.5 |
| LoCoMo-Refined, strict judge | report | ≥ any published | Mem0 ~48.9 under strict judge |
| LongMemEval-S | ≥75 | ≥88 | Zep ~71, Mem0 94.4, Hindsight 94.6 |
| BEAM 1M | report + tokens | ≥64 | Mem0 64.1, Hindsight 73.9 |
| OpMem v1 (~30 tasks) | Brainy full pass, ≥4-system table | industry citation | nobody else measures this — that's the point |
| Latency (search p50 @ k=50) | ≤300ms | ≤200ms | Mem0 ~0.9s, Zep markets <200ms |
| Tokens per assembled context | report | ≤4K at R1 accuracy | Mem0 ~6.9K |
