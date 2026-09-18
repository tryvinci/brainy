# Inventory and reproduction plan — 2026-09-18

**Status:** inventory only. **Not a pin. Not a remasure. Not a paper claim.**  
**Inventory SHA:** `origin/dev` = `origin/main` = **`bbe55f7`** (PR #184 merged).  
**This pass:** no benchmark run, no provider spend, no merge.

Anti-benchmax (binding): papers and evals measure product behavior. They do not justify LoCoMo/LME-named product rules, leftover-covering detectors queued from a remaining-item ledger, category dictionaries, or mixing smoke n with full n. Honesty stop: [competitive/benchmax-audit-2026-08-25.md](./competitive/benchmax-audit-2026-08-25.md).

Independently verifiable product goal (not claimed here):

| Gate | Bar | Last honest public pin |
| --- | --- | --- |
| LoCoMo product `POST /recall` | ≥ **1386/1540 (90%)**, MH / OD / SH / temporal reported | **175/1540 = 11.4%** on SHA `1b5ab3e` |
| LongMemEval-S | ≥ **400/500 (80%)**, including knowledge-update and abstention | **4/20** product `/recall`; LME-500 **not run** |
| OpMem / marketing | keep **13/13** and **17/17** | `1b5ab3e` same-cycle remasure |
| Personal-assistant / marketing qualification cases | ≥ **90/100** each | **suites do not exist in this repo** |

90% on the skip-ingest diagnostic 180 is **162/180**. That is **not** public LoCoMo. Publishing 152/180, 137/180, or 1×30 21/30 as 90% / n=1540 / Mem0 same-pin is forbidden.

---

## 1. Three research-paper artifacts

Roadmap source: [paper-topics.md](./paper-topics.md) (updated 2026-07-24). There is **no LaTeX / arXiv / Overleaf manuscript** in the tree (`git ls-tree origin/dev` has zero `.tex` / `.bib`). Publication state is markdown specs + evals + blog-shaped drafts.

### 1.1 Paper 1 — OpMem (operational memory correctness)

| | |
| --- | --- |
| **Working title** | Beyond recall: lifecycle, suppression, and correction correctness |
| **Intended claim** | Public memory benchmarks score retrieval; none score whether forget / correct / isolate / supersede *hold as a system*. OpMem is a small binary suite that discriminates production failure modes. |
| **Target venue** | Workshop / arXiv benchmark |
| **Publication readiness** | **Not ready.** Closest of the three, but the public draft is stale vs the live 13-task pin, adapters stop at Mem0+verbatim, and there is no camera-ready manuscript. |

**Exact paths**

| Role | Path |
| --- | --- |
| Spec | `docs/research/opmem-spec.md` |
| Public draft (blog-shaped) | `docs/research/posts/2026-07-opmem-v0.md` |
| Fixtures (v0+`upd03`) | `fixtures/opmem/*.json` (**13** tasks) |
| Runner | `evals/run_opmem.py` |
| Adapters | `evals/opmem_adapters.py` — `verbatim`, `brainy`, `mem0` (Platform) |
| CI | `internal/api/eval_e2e_test.go` → `TestOpMemBenchmarkAgainstHTTPServer` |
| Live Brainy pin | `docs/benchmarks/artifacts/opmem-fresh-local-20260815.md` |
| Live Mem0 pin | `docs/benchmarks/artifacts/opmem-mem0-fresh-20260815.md` |
| Older staging JSON | `docs/benchmarks/opmem-staging-vs-mem0.json` |

**Manuscript / data / code state**

- Spec still says **“v0: 12 tasks”**. The live suite is **13** (`upd03_state_supersession.json` landed after the spec). The July public post still headlines **12/12 vs Mem0 9/12**.
- Code is a real, CI-gated HTTP harness. Scoring is pass/fail substring on `/memories/search` (not product `/recall`).
- Adapters: **3** (Brainy, verbatim, Mem0 Platform). Paper checklist wants **≥4** (Zep, Letta, LangMem named; none exist).
- ENG-86 supersession API exists (`docs/research/supersession-v1.md`, `POST /memories/{id}/supersede`); OpMem `revise` still maps to in-place `/correct`.
- No released dataset DOI. Fixtures are in-repo JSON.

**Missing sections (paper)**

- Related-work (HaluMem / LMEB / MemoryAgentBench contrast, written only as bullets in `paper-topics.md`)
- Formal task grammar + inter-annotator / fixture-authoring protocol
- Multi-system results table with **pinned dates** (Mem0 10/13 is 2026-08-15 Platform, not OSS)
- Failure taxonomy beyond the five categories
- Limitations: 13 tasks, domain-neutral English, substring matching, no concurrency/expiry v1 tasks

**Missing experiments / data**

- Zep / Letta / LangMem / Graphiti OSS adapters and one frozen table
- v1 expansion (~30 tasks: expiry, async visibility, concurrent write) listed in the spec, not built
- Re-run Mem0 **OSS** (Platform 10/13 is explicitly not OSS-reproducible)
- Camera-ready numbers must cite **13/13 vs 10/13**, not the July 12/12 vs 9/12 post

**Reproduce (no LLM, cheap)**

```bash
python3 evals/run_opmem.py --systems verbatim,brainy --base-url http://127.0.0.1:8080
# MEM0_API_KEY=... python3 evals/run_opmem.py --systems verbatim,brainy,mem0 --base-url http://127.0.0.1:8080
```

---

### 1.2 Paper 2 — Vertical packs over primitives

| | |
| --- | --- |
| **Working title** | Verticals are packs, not code paths |
| **Intended claim** | One domain-agnostic runtime (cognitive primitives + generic lifecycle/rank) specializes per domain through versioned YAML packs, not forked schemas. Evidence: two packs pass domain evals on an **unchanged** binary, plus generic-pack vs domain-pack ablation. |
| **Target venue** | Systems paper |
| **Publication readiness** | **Not ready.** Marketing pack + 17 fixtures are a product moat, not a paper. The second pack (finance) is still Linear discovery. Support pack is a stub. No manuscript. |

**Exact paths**

| Role | Path |
| --- | --- |
| Thesis / model | `docs/vertical/verticalization-model.md` |
| Blog handoff (not a paper) | `docs/research/blog-handoff-vertical-memory.md` |
| Marketing pack | `packs/marketing/v1/pack.yaml`, `packs/marketing/v2/{pack,entities,state-machines}.yaml` |
| Support pack | `packs/support/v1/pack.yaml`, `packs/support/v2/{pack,entities,state-machines}.yaml` |
| Pack runtime | `internal/pack/pack.go`, `internal/memory/vertical.go`, `internal/memory/outcome.go` |
| Marketing fixtures | `fixtures/vertical/marketing/` (**17** JSON) |
| Support fixtures | `fixtures/vertical/support/` (**4** JSON) |
| Runners | `evals/run_vertical_eval.py`, `evals/run_marketing_mvp_benchmark.py` |
| Capability matrix | `evals/marketing_mvp_matrix.json` |
| Moat report | `docs/benchmarks/marketing-moat-report.md` |
| Live Brainy pin | `docs/benchmarks/artifacts/marketing-fresh-local-20260815.md` |
| Live Mem0 pin | `docs/benchmarks/artifacts/marketing-mem0-fresh-20260815.md` |
| Vetting gate | `docs/vertical/marketing-vetting-gate.md` |
| Finance pack | **absent** (`packs/finance/` does not exist) |

**Linear (not in repo)**

- [ENG-76](https://linear.app/engramhq/issue/ENG-76) finance taxonomy — **Backlog**
- [ENG-56](https://linear.app/engramhq/issue/ENG-56) finance epic — **Todo**, pack implementation gated M4
- [ENG-78](https://linear.app/engramhq/issue/ENG-78) ≥10 finance goldens — **Backlog**

**Manuscript / data / code state**

- Marketing v2 YAML is real and loaded by the Go runtime. Rank weights, lifecycle rules, and metadata schemas are executable.
- CI runs marketing 17 via `TestVerticalEvalHarnessAgainstHTTPServer` / MVP benchmark. Support fixtures are **not** a CI merge bar.
- Mem0 “4/17 empirical” is a **schema/capability** miss under `strict_schema=True` as much as a quality miss; the paper must say that, or it is a straw man.
- No ablation harness (generic pack vs marketing pack on the same 17 tasks).

**Missing sections**

- Systems architecture (planes, pack load, validation) at paper length
- Ablation: same binary, pack on vs pack off / generic vs domain
- Second-domain experiment (finance **or** a completed support matrix ≥10)
- Threats: fixture authorship, Mem0 not implementing pack semantics by design

**Missing experiments / data**

- Finance pack on unchanged runtime (the paper’s key experiment)
- Generic-vs-domain ablation table
- Support pack expanded and scored as a second in-repo domain (4 fixtures ≠ a domain eval)
- 90/100 “fresh qualification cases” for marketing / personal-assistant: **not in repo**

**Reproduce**

```bash
python3 evals/run_vertical_eval.py --base-url http://127.0.0.1:8080
python3 evals/run_marketing_mvp_benchmark.py --base-url http://127.0.0.1:8080
```

---

### 1.3 Paper 3 — Outcome-grounded conviction / stop-loss

| | |
| --- | --- |
| **Working title** | Conviction and stop-loss: closing the loop between task outcomes and agent beliefs |
| **Intended claim** | Belief-memory systems (BeliefMem, Hindsight, Kumiho) update from *observations*. Brainy updates from *outcomes*: conviction, expected-vs-observed delta, stop-loss, experiment, retire. |
| **Target venue** | Algorithm paper |
| **Publication readiness** | **Not ready.** Spec is a one-page policy note. Shipped code is a rank bonus + outcome→belief synthesis. No stop-loss loop, no longitudinal dataset, no BeliefMem baseline, no manuscript. |

**Exact paths**

| Role | Path |
| --- | --- |
| Policy note | `docs/brainy/architecture/04-conviction-stop-loss.md` (~15 lines) |
| Belief states | `docs/brainy/architecture/02-belief-lifecycle.md` |
| Hypothesis ledger schema | `docs/brainy/architecture/06-hypothesis-ledger-schema.md` |
| Ledger seed | `docs/brainy/architecture/hypothesis-ledger.seed.json` |
| Master-plan scope | `docs/research/master-plan.md` (full challenge/retire loop is **research**, not release-critical) |
| Shipped synthesis | `internal/memory/outcome.go` (`synthesizeBeliefFromOutcome`, `applyConvictionBoost`) |
| One golden | `fixtures/vertical/marketing/ob05_outcome_updates_belief_rank.json` |
| Python prototype tests cited by HYP-001 | `archive/brainy-python-prototype/tests/test_reflection.py` (not the Go service) |

**Manuscript / data / code state**

- `04-conviction-stop-loss.md` names parameters (`min_observations`, `delta_threshold`, `persistence_window`, `max_conviction_drop_per_cycle`) and actions (soft / hard / terminal). **None of those parameters are implemented** as a control loop in `internal/memory/`.
- What shipped: ingest of `performance_outcome` can synthesize a `content_belief` with a conviction float; retrieval adds `conviction * 2` to score. That is ranking, not stop-loss.
- HYP-001 falsification test `tests/test_reflection.py::test_reflection_downgrades_conviction_on_outcome_failure` is **not** in the Go tree.
- No simulated campaign/trading corpus. No BeliefMem reimplementation.

**Missing sections**

- Update rule (formula, identifiability, calibration)
- Experiment design (bandit / simulated campaigns)
- Comparison to observation-only belief revision
- Failure modes (over-drop, confirmation stickiness)

**Missing experiments / data**

- Longitudinal outcome traces (the paper-topics gate)
- BeliefMem-class baseline
- Any metric that is not `ob05` first-hit ranking

Do **not** treat `ob05` 1/1 as Paper 3 evidence.

---

## 2. Current benchmark protocol (pinned)

Authoritative contracts: [proveable-eval-framework.md](./proveable-eval-framework.md), [public-bench-ladder.md](./public-bench-ladder.md), [locomo-dual-path-freeze.md](./locomo-dual-path-freeze.md), [publish-stack-pins.md](./publish-stack-pins.md), [benchmarks/METHODOLOGY.md](../benchmarks/METHODOLOGY.md), [benchmarks/README.md](../benchmarks/README.md).

`eval-protocol-v2` (always `mode:"answer"`, transport stays transport, malformed judge JSON = `UNRESOLVED`) lives on **PR #186**, not on `bbe55f7`. Until that merges, `origin/dev` public scoring is **protocol v1** (`evals/public/locomo/run_smoke.py`).

### 2.1 Datasets

| Suite | Upstream | Bytes pin | Local path (gitignored) |
| --- | --- | --- | --- |
| LoCoMo | [snap-research/locomo](https://github.com/snap-research/locomo) `data/locomo10.json` | **SHA256 `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`** | `datasets/locomo/locomo10.json` |
| LongMemEval-S | HuggingFace `xiaowu0162/longmemeval-cleaned` `longmemeval_s_cleaned.json` | **SHA256 `d6f21ea9d60a0d56f34a05b609c79c88a451d2ae03597821ea3d5a9678c3a442`** (LME-20 pin) | `datasets/longmemeval/longmemeval_s_cleaned.json` |
| BEAM | HuggingFace `Mohammadta/BEAM` (10M: `Mohammadta/BEAM-10M`) | **no content SHA in-repo** (download cache) | `datasets/beam/beam_{size}.json` |
| OpMem | in-repo | git tree of `fixtures/opmem/` | — |
| Marketing / support / parity | in-repo | git tree of `fixtures/` | — |

### 2.2 Full vs smoke sample sizes

| Run | n | What it is | Publishable as the suite? |
| --- | ---: | --- | --- |
| LoCoMo 1×30 | 30 (conv-26 head, cats 1–4; MH 10 / OD 4 / temporal 16) | smoke / diagnostic | **No** (measurement only) |
| LoCoMo S0 stratified | **180**, `--seed 1`, 10 convos, proportional SH/MH/OD/temporal | iteration currency | **No** (not n=1540) |
| LoCoMo 3×90 | 90 × 3 seeds | integrity / holdout | **No** |
| LoCoMo full | **1540** scored cats 1–4 (10 convos; adversarial cat 5 traced, excluded from overall; 1986 judged on the 11.4% pin) | public LoCoMo | **Yes**, with lane label |
| LME-20 | 20, `--seed 1` | smoke | **No** |
| LME-S 100 | 100 (historical) | smoke | **No** |
| LME-500 | 500 | LongMemEval-S | **Yes** if jobs complete fail-closed; **not run** |
| BEAM 100K/20q | 20 probes, conv-0 | smoke | **No** |
| BEAM 1M / 10M | 700 / 200 in Mem0 docs | full BEAM tiers | **not run** |
| OpMem | 13 | full own suite | **Yes** (ops, not LoCoMo) |
| Marketing vertical | 17 | full own suite | **Yes** (vertical, not LoCoMo) |
| Parity | 4 | full own suite | **Yes** (thin slice) |
| Support | 4 | stub | **No** |
| Personal-assistant 100 / marketing 100 qualification | — | **missing** | **No** |

Holdout rule ([publish-stack-pins.md](./publish-stack-pins.md)): tuning = LoCoMo convs 1–3; validation = 4–10 at most once per phase gate. `runs-log.md` is the log.

### 2.3 Lanes (never mix in one percent)

| Lane | Flag | Path | Default top-k |
| --- | --- | --- | ---: |
| **Product** | `--eval-lane product-recall` or `BRAINY_USE_RECALL=1` | `POST /recall` | 30 |
| **Industry** | `--eval-lane industry-search` | search → shared answerer → shared judge | **200** |

July **49.8%** is industry / old stack. Current **11.4%** is product `/recall` on `1b5ab3e`. Mem0 published **92.5%** is their harness, top-k 200, n=1540 — **context**, not same-pin.

### 2.4 Answerer / judge

| Knob | Protocol |
| --- | --- |
| Judge temperature | **0.0** (`require_pins()` fails otherwise) |
| Default model | `LLM_MODEL`, else local `llama3.1`, else `gpt-4o-mini` (`evals/public/llm.py`) |
| Dev / CI typical | Cloudflare AI Gateway + Workers AI (`publish-stack-pins.md`; model id often redacted in artifacts) |
| Publish stack | GPT-class OpenAI-compatible, model id written in the manifest |
| Lexical fallback | `lexical-overlap-v0` — **not** a publishable J-score |
| v1 judge miss | unparseable JSON may substring-fallback to CORRECT (`evals/public/judge.py`) |
| v2 (PR #186) | malformed JSON = `UNRESOLVED`; blocks `result.json` |
| Tokens | recorded when an LLM answerer is used; lexical mode has none |

Pinned 1×30 / full 11.4% artifacts name **temp=0** and the CF/Workers family; they do **not** print a stable public model id (redacted). Reproducing the exact J-score requires the same `LLM_BASE_URL` + `LLM_MODEL` that wrote the UnifiedResult — those UnifiedResult files are **not in git** (`docs/benchmarks/runs/` is gitignored).

### 2.5 Embedding path

| Profile | Embedder | Dim | ANN | When |
| --- | --- | --- | --- | --- |
| CI / unit | `local-hash-v1` | 128 | off | `go test`; hash dump remasure |
| Integrity / qualify | hosted, pinned | **768** | pgvector HNSW | fail-closed / qualification |
| A/B (2026-08-20) | BGE-768 vs OpenAI `text-embedding-3-small`@768 vs `text-embedding-3-large`@768 vs hash-128 | 768 / 128 | retrieval@k, **not** QA | [embedding-ab-20260820.md](../benchmarks/artifacts/embedding-ab-20260820.md) |

Qualification profile (PR #186 `--qualification-profile`): API/worker signatures match, **fallbacks=0**, ANN active. Hash fallback is a **remeasure**, not qualification.

### 2.6 Metrics

- LoCoMo / LME: LLM-judge **CORRECT/WRONG** on cats 1–4; accuracy = correct/n; category breakdown required.
- Search p50 / p95 on the pin (harness observation, **not** a production SLO).
- Tokens/query when claiming deployability.
- Stage-oracle earliest-fail (WRITE / RETRIEVAL / PROOF / READER / HARNESS) on ledgers under `docs/benchmarks/artifacts/failure-ledger/`.
- OpMem / vertical: binary fixture pass/fail, **not** J-score.
- SuperMemory Recall@k is **not** an LLM-judge percent.

### 2.7 Latency / cost capture

- `UnifiedResult.metrics` includes search p50/p95; per-item retrieval latency is in `evaluations[]`.
- Cost is **not** a first-class field. Estimate from model × call counts, or from provider bills.
- Local vs Platform latency must not be published as an SLO.

### 2.8 Seeds

| Run | Seed |
| --- | --- |
| LoCoMo smoke / 1×30 | `--seed 1` (stratified only when `--stratified` > 0; 1×30 is conv-26 head, not a random 30) |
| S0 180 | `--seed 1` |
| Integrity 3×90 | documented in [locomo-integrity-3x90-20260820.md](../benchmarks/artifacts/locomo-integrity-3x90-20260820.md) |
| Full 11.4% | **1 seed** (`locomo-fresh-full-20260815-s0-33161a`) |
| Historical 49.8% | **3 seeds** (`2a6a04`, `e7ba5b`, `9b61f5`) |
| LME-20 | `--seed 1`, `--limit 20` |

### 2.9 Competitor-comparison rules

1. **Same-pin** = same dataset SHA, same judge temp 0, same answerer, same question set, same harness. Only same-pin may be used for lead/trail.
2. **Published %** = sourced vendor headlines with n and metric labeled ([published-claims.md](../benchmarks/published-claims.md)). Never a scoreboard row for lead/trail.
3. **Mem0 Platform ≠ Mem0 OSS.** Graphiti OSS ≠ Zep Platform.
4. Fair Mem0 LoCoMo (audit 2026-08-22): v3 add/search, chunk **1**, unix `timestamp`, top_k **200**, event wait. The frozen 11/30 used v2, chunk 8, no timestamps, top_k 30 — **handicapped**. Do not refresh lead/trail from 21/30 vs 11/30 until a fair stratified 180 exists. Fair 180 was quota-blocked until **2026-09-01**.
5. No pin for a vendor we did not run.
6. No SOTA / beats-Mem0 in product copy without a frozen same-pin win **and** explicit approval.
7. Cycle closeout order: Landed → Own pins (by category) → **detailed competitor compare in** [cycle-closeout.md](./competitive/cycle-closeout.md) → Why → Next. README gets published-% + same-pin summary only.
8. Do not merge leftover-covering PRs **#133, #131, #143, #145**. Do not queue new `looks*Query` detectors from the skip-ingest 180 remaining ledger.

---

## 3. Claim map (code · data · artifact · SHA)

Status: **supported** (reproducible from git + documented runner) · **summary-only** (markdown/JSON summary in git, UnifiedResult not in git) · **smoke-only** · **stale** · **unsupported** · **in-flight (not a pin)**.

### 3.1 Own suites (merge bar)

| Claim | Code | Data | Artifact | SHA | Status |
| --- | --- | --- | --- | --- | --- |
| OpMem **13/13** | `evals/run_opmem.py`, `TestOpMemBenchmarkAgainstHTTPServer` | `fixtures/opmem/` (13) | [opmem-fresh-local-20260815.md](../benchmarks/artifacts/opmem-fresh-local-20260815.md) | **`1b5ab3e`** | **supported** on that SHA; **not re-run** on `bbe55f7` |
| OpMem Mem0 **10/13** | `evals/opmem_adapters.py` `Mem0OpAdapter` | same 13 fixtures | [opmem-mem0-fresh-20260815.md](../benchmarks/artifacts/opmem-mem0-fresh-20260815.md) | Platform, 2026-08-15 | **supported** as Platform; **not OSS** |
| Marketing **17/17** | `evals/run_vertical_eval.py` | `fixtures/vertical/marketing/` | [marketing-fresh-local-20260815.md](../benchmarks/artifacts/marketing-fresh-local-20260815.md) | **`1b5ab3e`** | **supported** on that SHA; **not re-run** on `bbe55f7` |
| Marketing Mem0 **4/17** | `--systems brainy,mem0` | same 17 | [marketing-mem0-fresh-20260815.md](../benchmarks/artifacts/marketing-mem0-fresh-20260815.md) | Platform, 2026-08-15 | **supported** empirical; schema-moat |
| Parity **4/4** | `evals/run_eval.py` | `fixtures/parity/` | [parity-fresh-local-20260815.md](../benchmarks/artifacts/parity-fresh-local-20260815.md) | **`1b5ab3e`** | **supported** |
| OpMem **12/12** vs Mem0 **9/12** | same runner, 12-task set | old fixture set | [posts/2026-07-opmem-v0.md](./posts/2026-07-opmem-v0.md), staging JSON | ~2026-07-14 | **stale** — do not cite as current |
| Marketing **15/16** / **16/16** | older fixture set | — | ladder / July post | July 2026 | **stale** (`bv06` now in 17/17) |

### 3.2 LoCoMo

| Claim | Code | Data | Artifact | SHA | Status |
| --- | --- | --- | --- | --- | --- |
| Full product `/recall` **175/1540 = 11.4%** | `evals/public/locomo/run_full.py` + `/recall` | LoCoMo SHA `79fa87e…` | [locomo-fresh-full-20260815.md](../benchmarks/artifacts/locomo-fresh-full-20260815.md), [summary JSON](../benchmarks/artifacts/locomo-fresh-full-20260815-summary.json) | product **`1b5ab3e`**, run `locomo-fresh-full-20260815-s0-33161a` | **summary-only** (no UnifiedResult in git) |
| Full search+harness **49.8% mean** | older `run_full` path | same dataset SHA | [locomo-full-publish-summary.json](../benchmarks/artifacts/locomo-full-publish-summary.json) | old stack **2026-07-31**, 3 seeds | **summary-only**; **not current SHA** |
| 1×30 **21/30 (70%)** MH 10/10 OD **0/4** temporal 11/16 | `run_smoke.py --conversations 1 --questions 30` | conv-26, cats 1–4 | [locomo-fresh-1x30-20260815.md](../benchmarks/artifacts/locomo-fresh-1x30-20260815.md), ledger JSONL | **`1b5ab3e`** | **smoke-only** (allowed as measurement) |
| Mem0 1×30 **11/30** | `run_smoke.py --system mem0` | same 30 | [locomo-mem0-fresh-1x30-20260815.md](../benchmarks/artifacts/locomo-mem0-fresh-1x30-20260815.md) | Platform, handicapped knobs | **smoke-only**; **not fair Mem0 protocol** |
| Integrity S0 product **32/180** / industry **62/180** | `run_s0.py` | seed 1, 180 | [locomo-integrity-s0-20260819.md](../benchmarks/artifacts/locomo-integrity-s0-20260819.md) | integrity stack 2026-08-19/20 | **smoke-only**; later invalidated numbers called out in README |
| This-VM `diag-mh-135` industry **62/180** | skip-ingest industry lane | frozen tenant | [locomo-s0-diag-mh-135-20260822.md](../benchmarks/artifacts/locomo-s0-diag-mh-135-20260822.md) | reader-off pin | **smoke-only**; industry **unchanged** by leftover covering |
| P84 product **152/180** | skip-ingest hybrid `/recall` | tenant `diag-mh-135` + conv-30, **no re-extract** | [locomo-s0-diag-mh-135-p84-20260828.md](../benchmarks/artifacts/locomo-s0-diag-mh-135-p84-20260828.md) | product **`d7d55ab`** (ancestor of `bbe55f7`) | **smoke-only diagnostic**; **not** n=1540 / 90% / Mem0 same-pin |
| P53 covering **137/180** | leftover covering reader | same tenant | [benchmax-audit](./competitive/benchmax-audit-2026-08-25.md) | `ae15e40` | **smoke-only**; saturating; do not extend |
| Hash n=1540 protocol-v2 in-flight | PR #186 `run_staged` | restored dump (VM-local) | `/opt/cursor/artifacts/eval-runs/locomo-full-n1540-bbe55f7-protocol-v2/` | harness `b1dd655`, product `bbe55f7` | **in-flight, not a pin** (no `result.json` at inventory time) |
| `bbe55f7` full n=1540 | — | — | — | `bbe55f7` | **unsupported** — never scored at 1540 |
| Fair Mem0 180 / n=1540 | audit recipe | — | [mem0-harness-audit-2026-08-22.md](./competitive/mem0-harness-audit-2026-08-22.md) | — | **unsupported** (no pin) |

Category counts on the 11.4% pin (supported as labeled n): MH **21/282 (7.4%)**, OD **5/96 (5.2%)**, SH **88/841 (10.5%)**, temporal **61/321 (19.0%)**.

### 3.3 LongMemEval / BEAM

| Claim | Code | Data | Artifact | SHA | Status |
| --- | --- | --- | --- | --- | --- |
| LME-20 **4/20** product `/recall` | `evals/public/longmemeval/run.py --product-recall --limit 20 --seed 1` | SHA `d6f21ea9…` | [lme20-fresh-20260815.md](../benchmarks/artifacts/lme20-fresh-20260815.md) | **`1b5ab3e`**, jobs 4829=4829 | **smoke-only**; KU 0/3, multi-session 0/5 |
| LME-20 integrity **0/20** | same | same SHA/seed | [lme20-product-recall-pr1-20260812-pin.md](../benchmarks/artifacts/lme20-product-recall-pr1-20260812-pin.md) | 2026-08-12 | **smoke-only** historical |
| LME-S 100 **4/100** | search+harness | LME-S | [lme-s-100.md](../benchmarks/artifacts/lme-s-100.md) | 2026-08-01 | **smoke-only**; different path/n from 4/20 |
| LME-500 | same runner, `--limit 500` | LME-S | — | — | **not run** |
| BEAM 100K **8/20** | `evals/public/beam/run.py --chat-size 100K --conversations 0-0` | HF download | [beam-100k-fresh-20260815.md](../benchmarks/artifacts/beam-100k-fresh-20260815.md) | **`1b5ab3e`** | **smoke-only**; search+harness, **not** `/recall` |
| BEAM 1M / 10M | same | — | — | — | **not run** |

### 3.4 Representation / retrieval (not QA)

| Claim | Code | Artifact | SHA | Status |
| --- | --- | --- | --- | --- |
| Provider extract coverage **161/180** vs det **139/180** | `evals/public/extraction_ceiling.py` | [extraction-ceiling-20260819.md](../benchmarks/artifacts/extraction-ceiling-20260819.md) | integrity 2026-08-19 | **supported** as coverage, **not** J-score |
| OpenAI small@768 strongest r@100/200 on rebuild | `evals/public/embedding_ab.py` | [embedding-ab-20260820.md](../benchmarks/artifacts/embedding-ab-20260820.md) | integrity tenant rebuild | **supported** retrieval@k; **not** QA |

### 3.5 Paper claims vs evidence

| Paper claim | Evidence | Status |
| --- | --- | --- |
| OpMem discriminates production ops vs Mem0 | 13/13 vs 10/13 Platform, 2026-08-15 | **supported** as a small own-suite pin; **not** a published paper |
| “Verticals are packs, not code paths” | one pack (marketing 17/17); support stub; no finance; no ablation | **unsupported** as a systems-paper result |
| Outcome-grounded stop-loss | `ob05` rank bonus | **unsupported** |

### 3.6 Reproduction holes (apply to several rows)

- `docs/benchmarks/runs/` and `datasets/` are **gitignored**. Third-party reproduce of 11.4% / 21/30 / 4/20 needs those UnifiedResult files or a re-run.
- VM dumps (`/opt/cursor/artifacts/n1540-live/…`, qualify DB) are **not** git artifacts.
- Current product SHA **`bbe55f7` includes leftover covering + P54–P84 S2 + PR #184 think-about**. The public 11.4% / 21/30 pins are **`1b5ab3e`**. Do not treat README percents as current-SHA scores.

---

## 4. Why drafts #185, #186, #187 stay drafts

All three target **`dev` @ `bbe55f7`**, are **open drafts**, and were **not** merged (this pass does not merge them). GitHub combined status on each head was empty/`pending` at inventory time (no recorded check runs via the status API).

| PR | Branch | Head | Role |
| --- | --- | --- | --- |
| [#185](https://github.com/tryvinci/brainy/pull/185) | `pr/n1540-handoff-1e9e` | `cea4630` | docs + enqueue/score helpers |
| [#186](https://github.com/tryvinci/brainy/pull/186) | `pr/qualified-memory-measurement-1e9e` | `b1dd655` | eval-protocol-v2 staged runner + worker keep-alive |
| [#187](https://github.com/tryvinci/brainy/pull/187) | `pr/s5-rrf-retrieval-1e9e` | `80fce86` | versioned RRF + S2c abstain keep + cherry-picked keep-alive |

### 4.1 #185 — handoff only

**Why draft:** no product change; it exists so a later agent can resume n=1540 extract/score. The live n=1540 JSON still does not exist, so there is nothing to close out.

**Superseded by #186:** `git log origin/dev..origin/pr/qualified-memory-measurement-1e9e` starts with `cea4630`. #186 **contains** #185. Merging #185 after #186 is redundant.

### 4.2 #186 — measurement, incomplete pin

**Why draft:**

1. Protocol-v2 skip-ingest n=1540 on the restored **hash** dump has **no `result.json`**. A mid-run CORRECT rate is not a pin.
2. Identity lock: `identity.json` records harness SHA `b1dd655`. Further commits on that branch break resume.
3. Hash + leftover-covering fusion cannot meet the 90% qualification bar. This PR does not claim to.
4. Worker keep-alive is a real product bugfix and is independently testable (`processor_test.go`), but the PR is bundled with an unfinished remasure.

**Does not supersede #187.** Complementary: harness vs retrieval.

### 4.3 #187 — qualify substrate, incomplete extract/score

**Why draft:**

1. RRF default **off**. Merge gate still requires OpMem 13/13 and marketing 17/17 with the flag off — **unchecked on this PR**.
2. Live-provider extract on `brainy_qualify` (te3-small 768) was **not drained to 0 failed / 0 fallbacks** at last operational note. `--qualification-profile` must not run mid-extract (`store_identity` includes embedding counts).
3. S2c education-keep is a small reader fix with unit tests; it is not a LoCoMo pin.
4. Do not point the hash n=1540 API (`:18200` / `brainy_n1540`) at this binary.

### 4.4 Conflict vs overlap

| Files both #186 and #187 change vs `dev` | Relationship |
| --- | --- |
| `internal/jobs/processor.go` + `_test.go` | **Same keep-alive patch.** #187 commit `66997e0` is a cherry-pick of #186 `f48c4ff`. Merge one first; the second rebases cleanly if the patch is identical. |
| `internal/store/postgres/runtime.go` | **Same `current_database()` field.** Same cherry-pick story (`9a49f0d` vs #186). |

No other file overlap. **Neither supersedes the other.** Recommended land order if/when merging: **#186 (measurement + keep-alive) then rebase #187 (RRF)**. Drop #185. Do **not** merge leftover-covering `pr/locomo-180-p29-1e9e`.

#185/#186/#187 do **not** replace README 11.4%. They do **not** conflict with Paper 1–3 except that honest LoCoMo numbers are a prerequisite for any conversational section of those papers.

---

## 5. Smallest next experiments (costed, not run)

Estimates use public list prices as **order-of-magnitude** (OpenAI `text-embedding-3-small` ≈ $0.02 / 1M tokens; `gpt-4o-mini` ≈ $0.15 / 1M in + $0.60 / 1M out; CF gpt-oss is typically similar or cheaper). Runtime from documented pins and prior VM notes, not a new run. **Do not execute in this pass.**

### Do not run

| Experiment | Why forbidden / premature |
| --- | --- |
| Another leftover-covering 180 increment | Honesty stop; saturating; not n=1540 |
| LME-500 as a quality claim | LME-20 is 4/20; plan forbids LME-500 until LME-20 is non-embarrassing |
| BEAM 1M / 10M | After LoCoMo/LME; huge download + judge cost |
| Full n=1540 as a **claim** on `bbe55f7` | No current-SHA stratified ledger yet; SOTA plan: 180 before 1540 |
| Re-extract frozen `diag-mh-135` | Pin is skip-ingest; re-extract breaks attribution |
| Paper-3 longitudinal sim | No algorithm to test |
| Finance pack eval | Pack does not exist |

### Run next (highest evidence per dollar)

| ID | Experiment | Closes | Calls (order) | $ (order) | Runtime (order) | Gate before running |
| --- | --- | --- | ---: | ---: | ---: | --- |
| **E0** | OpMem 13 + marketing 17 + parity 4 on **local API built from `bbe55f7`**, hash embedder, no `LLM_*` | Merge-bar non-reg after PR #184 | ~30 HTTP ingest/search | **~$0** | **<10 min** | `go test ./...` already the CI bar; this is the labeled ops/vertical remasure |
| **E1** | `go test ./internal/jobs ./evals` equivalent on #186 only (no LoCoMo) | Whether protocol-v2 / keep-alive can merge without a 1540 JSON | unit tests | **~$0** | **minutes** | none |
| **E2** | **New-tenant** S0 dual-lane 180, seed 1, async live extract, **product + industry**, protocol v2 if #186 landed, **no** leftover-covering fishing | What `bbe55f7` actually scores (WRITE vs READER vs RETRIEVAL) | ~1571 extract + ~180 `/recall` + ~180 industry answer + ~360 judge + embeddings | **~$5–40** extract LLM + **< $1** te3-small + **~$1–8** answer/judge | extract **1–3 h** (10 workers); score **1–3 h** sequential | E0 green; new tenant (not `diag-mh-135`); fail-closed ANN+768 if calling it qualification |
| **E3** | Fair Mem0 Platform **180** (v3, chunk 1, timestamps, top_k 200) + same judge | Same-pin lead/trail (replaces handicapped 11/30) | ~thousands Mem0 add/search + 180 judge | **Mem0 bill unknown** + **~$1–4** judge | **hours**; 429 risk | E2 Brainy industry row exists; `MEM0_API_KEY`; org/project IDs if 180 looks weak |
| **E4** | Conv-26 only, te3-small 768 + `BRAINY_RETRIEVAL_RRF=1`, protocol-v2 1×30 | Whether qualify stack moves OD/SH vs hash 1×30 | ~111 extract jobs + 30 `/recall` + 30 judge | **~$1–8** | **30–90 min** after extract | Extract complete, fallbacks=0, histogram frozen; RRF **on** this tenant only |
| **E5** | Finish hash n=1540 protocol-v2 **if** dump + identity SHA still match | Replace or keep 11.4% with a labeled current-SHA **hash** remasure | remaining ~1100 answer+judge if mid-run | **~$2–15** judge/reader (no embed) | **4–20 h** sequential | Do not enable RRF or 768 on `brainy_n1540`; do not `--qualification-profile` |

**E0 then E2** are the highest-value pair: they are small, labeled, and they tell the truth about `bbe55f7` without benchmaxxing the 180 leftover ledger.

After E2’s stage-oracle histogram, allocate S1–S5 per [sota-execution-plan.md](./sota-execution-plan.md). Only then: 3×90 → n=1540 both lanes → LME-20 ≥16/20 → LME-500.

### Paper-shaped smallest extras (still not this pass)

| Paper | Smallest evidence increment | $ / time | Note |
| --- | --- | --- | --- |
| 1 OpMem | Refresh 13/13 on `bbe55f7` (E0) + one **verbatim vs Brainy vs Mem0 Platform** JSON | E0 + Mem0 13 tasks **< $1 / 15 min** | Still 3 systems; Zep adapter is a code task, not an eval spend |
| 2 Vertical | Support 4 fixtures in CI; finance pack is engineering, not a cheap eval | ~$0 | Ablation = product work |
| 3 Conviction | Unit-test the documented stop-loss parameters **after** they exist | ~$0 | Do not simulate campaigns first |

---

## 6. Reproduction plan (commands; do not run here)

Work from **`bbe55f7`** (or a branch based on it). Do not check out leftover-covering `pr/locomo-180-p29-1e9e` as the product SHA.

```text
# 0. Identity
git rev-parse HEAD   # expect bbe55f7 if on origin/dev
unset BRAINY_API_KEYS BRAINY_REQUIRE_API_KEY
export BRAINY_ENV=local

# 1. Merge bar (no public suite)
gofmt -l .
go vet ./...
go test ./...

# 2. E0 ops/vertical (local API, no provider required)
#    start API per AGENTS.md; pg_trgm required; pgvector optional
python3 evals/run_opmem.py --systems verbatim,brainy --base-url http://127.0.0.1:8080 \
  --json-out docs/benchmarks/opmem-bbe55f7.json
python3 evals/run_vertical_eval.py --base-url http://127.0.0.1:8080
python3 evals/run_eval.py --base-url http://127.0.0.1:8080

# 3. Dataset pin (download once; do not commit)
cd evals && python3 -c "from public.locomo.dataset import ensure_dataset; p,s=ensure_dataset(); print(p,s)"
# expect 79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4

# 4. E2 current-SHA S0 (AFTER extract drain; new tenant prefix)
#    product then industry; write UnifiedResult under a durable artifact root (not /tmp)
python -m public.locomo.run_s0 --base-url http://127.0.0.1:8080 \
  --stratified 180 --seed 1 --conversations 10

# 5. Proveability
#    require_pins() empty; jobs_failed=0; fallbacks=0 if fail-closed;
#    file cycle-closeout.md in Landed → Own pins → competitor compare → Why → Next
#    do not overwrite README 11.4% until a completed n=1540 JSON exists
```

If using PR #186 staged runner:

```text
export BRAINY_EVAL_PROTOCOL=eval-protocol-v2
python -m public.locomo.run_staged --base-url http://127.0.0.1:18200 \
  --run-id locomo-full-n1540-bbe55f7-protocol-v2 \
  --tenant-prefix locomo-n1540-bbe55f7 \
  --conversations 10 --skip-ingest --top-k 30 \
  --artifact-root /opt/cursor/artifacts/eval-runs
```

Refuse non-loopback `BRAINY_BASE_URL` (cloud injects staging). Refuse `/tmp` artifacts.

---

## 7. What this pass did not do

- No LoCoMo / LME / BEAM / Mem0 / OpMem / marketing run
- No provider embeddings or LLM calls
- No merge of #185 / #186 / #187 / leftover-covering PRs
- No README percent change
- No new paper manuscript
- No cycle-closeout scores section (inventory is not a remasure)

Next human/agent review: accept E0→E2 as the measurement sequence, or reject and keep 11.4% as the only full-product pin until a UnifiedResult for `bbe55f7` exists.
