# Paper Topics: Publishing Research from Brainy

**Status:** Active publication roadmap (aligned with [path-to-sota.md](./path-to-sota.md))
**Updated:** 2026-07-24 · originally 2026-07-02

## Publication roadmap (sequenced)

Goal: **surpass Mem0 on multiple axes**, with papers that expand the scoreboard beyond
LOCOMO J-score. Anti-benchmax: papers measure product behavior; they do not justify
dataset special-casing.

| Order | Paper | Axis | Engineering gates | Target shape |
| --- | --- | --- | --- | --- |
| **1** | OpMem operational correctness | A | Multi-system adapters (≥4); supersession fixtures (ENG-86) | Workshop / arXiv benchmark |
| **2** | Vertical packs over primitives | B | Finance pack (ENG-76) on unchanged runtime; ablation | Systems paper |
| **3** | Outcome-grounded conviction / stop-loss | E | Longitudinal outcome loop data | Algorithm paper |

Supporting shorts: TasteSignal; reproducible local-vs-provider eval methodology;
declarative lifecycle as temporal invalidation (section of Paper 2).

**Paper 1 ship checklist**

- [ ] OpMem v0/v1 harness documented + runnable CLI
- [ ] Brainy + Mem0 + ≥2 other systems in one table
- [ ] Pins + reproduce block; public post under `docs/research/posts/`
- [ ] No LOCOMO answer-key coupling

**Paper 2 ship checklist**

- [ ] Marketing + finance packs pass domain evals on same binary
- [ ] Ablation: generic pack vs domain pack
- [ ] Moat table vs Mem0/Zep (capabilities, not LOCOMO)

**Paper 3 ship checklist**

- [ ] Simulated campaigns/trading with outcome→conviction updates
- [ ] Compare observation-only belief baseline (BeliefMem-class)

## Can we write a paper of this nature?

Yes. The agent-memory papers trending right now (GAM, LightMem, HaluMem, LMEB, Mem0)
share a common shape: a named system or benchmark, a clear architectural thesis, an
evaluation against existing memory systems on public benchmarks, and released code.
Brainy already has most of the raw material for that shape:

- A distinctive architecture: 9 cognitive primitives + declarative vertical packs
  (`docs/brainy/architecture/00-cognitive-primitives.md`, `docs/vertical/verticalization-model.md`)
- A belief lifecycle with conviction and stop-loss (`02-belief-lifecycle.md`, `04-conviction-stop-loss.md`)
- A reproducible Tier 1–4 eval harness with a Mem0 side-by-side
  (`docs/benchmarks/METHODOLOGY.md`, `evals/`)
- Competitor analysis across Mem0, Zep, Letta, Cognee, Supermemory, Memobase
  (`docs/brainy/competitors/`)

What's missing for publishability is covered per-topic below, but the common gaps are:
(a) provider-quality embeddings/extraction (we currently use a deterministic local
embedder for CI), (b) evaluation on public benchmarks (LoCoMo, LongMemEval) alongside
our own fixtures, and (c) more baselines than Mem0 alone.

## Where the field is (mid-2026 scan)

- **Surveys:** "Memory in the Age of AI Agents" (Dec 2025) organizes the field by
  forms/functions/dynamics and names frontiers: memory automation, RL integration,
  multimodal, multi-agent, trustworthiness. "Memory for Autonomous LLM Agents"
  (2026) explicitly flags **uncertainty-aware memory** ("confidence levels updated as
  new data arrives") as "something most current memory systems handle poorly or not at all."
- **Belief-oriented memory is emerging fast:** BeliefMem (probabilistic candidate
  conclusions, noisy-OR updates), Hindsight (opinion networks + disposition profiles),
  Kumiho (AGM belief revision on a versioned graph), MnemeBrain (Belnap truth states,
  evidence polarity). None of these ground belief updates in **task outcomes**
  (expected-vs-observed results); they update from new *observations* only.
- **Benchmarks:** HaluMem (hallucination per memory operation), LMEB (memory
  embeddings), MemoryAgentBench, MemBench. None evaluate **operational correctness**:
  lifecycle suppression, correction stickiness, tenant isolation, governance.
- **Domain agents:** FinMem/FinAgent (finance, layered memory), AD-Bench (marketing
  agents). Domain *benchmarks* exist; a **domain-configurable memory runtime** does not
  appear in the literature — verticalization is done by forking systems, not by config.

## Candidate topics, ranked

### 1. Vertical packs over cognitive primitives (systems paper) — best fit

**Working title:** *"Verticals are packs, not code paths: domain-specialized agent
memory through declarative configuration over cognitive primitives."*

**Thesis:** a domain-agnostic runtime of typed cognitive primitives (Principle,
IdentityPrior, Belief, Episode, Outcome, ...) with a generic lifecycle machine and
rank pipeline can be specialized per domain purely through versioned YAML packs —
vocabulary mapping, metadata schemas, lifecycle triggers, rank weights, eval fixtures.

**Why it's novel:** the published alternatives are either generic stores (Mem0, Zep)
or domain-forked systems (FinMem). Nobody has published the config-not-code
verticalization claim with evidence.

**Evidence we have:** marketing pack passes Gate M3 (Tier 1–4, 10/10 use-case seeds);
the moat report documents 10 capabilities where Mem0 has no equivalent behavior.

**What it needs:** a second pack (finance is already drafted as a research direction,
ENG-76) to prove the "one runtime, many domains" claim — the paper's key experiment is
showing both packs pass domain evals on an *unchanged* runtime, plus an ablation
(generic pack vs. domain pack on the same tasks). Provider embeddings so the semantic
retrieval comparison is fair.

### 2. Operational memory correctness benchmark — cheapest to ship

**Working title:** *"Beyond recall: benchmarking lifecycle, suppression, and
correction correctness in agent memory systems."*

**Thesis:** existing memory benchmarks measure retrieval accuracy and hallucination;
none measure whether a memory system *behaves correctly as a system*: does a
suppressed memory ever leak into retrieval? Does a correction stick across
re-ingestion? Does an archived campaign stop ranking? Is tenant/subject isolation
airtight? These are the failure modes that matter in production and in regulated
domains, and HaluMem/LMEB/MemoryAgentBench don't cover them.

**Evidence we have:** the Tier 2/4 fixture suite is exactly this — taboo suppression
(`bv02`), correction stickiness (`bv04`), multi-brand isolation (`bv06`), lifecycle
suppression (`lc01`/`lc02`), scoped coexistence (`sg10`).

**What it needs:** generalize fixtures beyond marketing into a domain-neutral task
format, run 4–6 systems through it (Mem0, Zep, Letta, LangMem, a raw-RAG baseline —
adapters partially exist), and publish the harness. This fits the "trustworthiness"
frontier the Dec 2025 survey calls out, and a benchmark paper doesn't require our
system to win — only the benchmark to discriminate.

### 3. Outcome-grounded belief lifecycle (algorithm paper) — most novel, most work

**Working title:** *"Conviction and stop-loss: closing the loop between task outcomes
and agent beliefs."*

**Thesis:** belief-memory systems (BeliefMem, Hindsight, Kumiho) revise beliefs from
new *observations*. We revise them from *outcomes*: every belief carries conviction,
an expected-vs-observed outcome delta challenges it past a stop-loss threshold, and
competing beliefs are resolved through explicit experiments. This is the
"uncertainty-aware memory" gap the 2026 survey names, attacked from the calibration
side rather than the probabilistic-storage side.

**Evidence we have:** the belief lifecycle and conviction/stop-loss are specified
(`02-belief-lifecycle.md`, `04-conviction-stop-loss.md`, `06-hypothesis-ledger-schema.md`)
and the outcome→belief rank loop passes `ob05`.

**What it needs:** the most new work of the three — a longitudinal eval where an agent
acts on beliefs, observes outcomes, and measurably improves (e.g., simulated marketing
campaigns or trading episodes); comparison against BeliefMem-style probabilistic
updates; likely an RL or bandit framing for the conviction update rule. High upside:
this is a mechanism paper, not a systems paper, so it survives even if Brainy doesn't.

### Smaller / supporting topics

- **TasteSignal:** modeling non-functional preference ("taste") as a distinct memory
  primitive that influences ranking and synthesis (`01-taste-evolution-model.md`).
  Under-explored; could be a workshop paper or a section of topic 1.
- **Deterministic reproducible memory evaluation:** our CI-reproducible local embedder
  vs. provider embeddings is a methodology point (memory benchmarks are notoriously
  non-reproducible across provider versions). Workshop/short paper at best, but
  strengthens the methodology section of topics 1–2.
- **Declarative lifecycle rules as temporal invalidation:** pack-defined triggers on a
  generic state machine vs. Zep/Graphiti-style temporal knowledge graphs. Probably a
  section of topic 1 rather than standalone.

## Recommendation

Sequence them: **topic 2 first** (benchmark paper — fixtures exist, no second vertical
needed, establishes credibility and a citable artifact), **topic 1 second** (systems
paper, gated on the finance pack which is already the post-M3 roadmap), **topic 3**
as the long-term differentiator once outcome loops have real longitudinal data.

Topic 2 also de-risks the others: its harness becomes the evaluation section of
topic 1, and its correction/suppression tasks are prerequisites for trusting the
belief loop in topic 3.

## Reference papers (from the trending list + scan)

| Paper | Relevance |
| --- | --- |
| Mem0 (arXiv:2504.19413) | Primary baseline; we already track parity |
| GAM: General Agentic Memory (BAAI, Nov 2025) | JIT-context framing; contrast for our precomputed-primitive approach |
| Memory in the Age of AI Agents (Dec 2025 survey) | Taxonomy to position against; names trustworthiness frontier |
| LightMem (arXiv, Oct 2025) | Efficiency baseline and eval methodology (LongMemEval) |
| HaluMem (Nov 2025) | Operation-level eval precedent; topic 2 extends this to lifecycle/governance ops |
| LMEB | Long-horizon memory retrieval eval; relevant when we add provider embeddings |
| BeliefMem (arXiv:2605.05583) | Closest prior work for topic 3; observation-driven, not outcome-driven |
| Hindsight (arXiv:2512.12818) | Opinion networks + disposition profiles; contrast for TasteSignal/IdentityPrior |
| Kumiho (arXiv:2603.17244) | AGM belief-revision formalism; possible theoretical grounding for topic 3 |
| Memory for Autonomous LLM Agents (arXiv:2603.07670) | 2026 survey naming the uncertainty-aware-memory gap |
| FinMem / FinAgent / AD-Bench | Domain-agent prior work; motivates topic 1's config-not-fork claim |
