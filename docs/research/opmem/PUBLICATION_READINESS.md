# OpMem publication readiness package

**Status:** documentation only · **not a pin** · **no benchmark run in this pass**  
**Inventory SHA:** `bbe55f7` (`origin/dev` / `origin/main` at package authoring)  
**Authoring date:** 2026-09-18  

This package consolidates repo evidence for Paper 1 ([paper-topics.md](../paper-topics.md): *Beyond recall: lifecycle, suppression, and correction correctness*). It does not claim new measurements.

---

## Ready vs missing (executive)

| Area | Ready | Missing |
| --- | --- | --- |
| Task fixtures + runner | 13 JSON tasks, `evals/run_opmem.py`, 3 adapters (Brainy, verbatim, Mem0 Platform) | v1 ~30 tasks; Zep/Letta/LangMem adapters; committed multi-system JSON for 13-task freeze |
| Reproducible protocol | Documented below + [opmem-spec.md](../opmem-spec.md); CI spins embedded API | Pin file (`RunManifest`) for OpMem; SHA `bbe55f7` re-run of Brainy+Mem0 table |
| Brainy 13/13 evidence | Markdown pin + per-task JSON for Brainy-only integrity pass | Full `result.json` with Brainy+verbatim+mem0 on one commit in git |
| Mem0 10/13 evidence | Markdown pin with named failures | Per-task Mem0 JSON in git; Mem0 OSS run; API version drift check |
| Manuscript | LaTeX + PDF at [paper/main.pdf](./paper/main.pdf) | arXiv submission; formal grammar; camera-ready figures |
| Multi-system paper table | Brainy vs Mem0 vs verbatim (partial) | ≥4 systems per paper-topics checklist |
| Threats / limitations | Section below | External review dedicated to OpMem only |

**Publication verdict:** **not ready** for submission. Closest of the three paper tracks in [inventory-reproduction-plan-2026-09-18.md](../inventory-reproduction-plan-2026-09-18.md) (merge via PR #188), but requires manuscript, expanded systems, and frozen artifact bundle.

---

## Reproducible protocol

### Scope

OpMem measures **operational correctness** via deterministic operation scripts over `remember` / `recall` / `revise` / `forget`. Scoring is **binary pass/fail** per task on `GET /memories/search` results (Brainy adapter), not product `POST /recall` and not LLM-judge accuracy.

### Task set (v0 + `upd03`)

| Category | Task IDs | Count |
| --- | --- | ---: |
| suppression | `sup01_basic_forget`, `sup02_targeted_forget`, `sup03_durable_forget` | 3 |
| correction | `cor01_basic_revision`, `cor02_correction_stickiness`, `cor03_revised_retrievable` | 3 |
| isolation | `iso01_subject_isolation`, `iso02_tenant_isolation`, `iso03_forget_isolated` | 3 |
| staleness | `upd01_stale_fact`, `upd02_preference_change`, `upd03_state_supersession` | 3 |
| idempotency | `dup01_idempotent_remember` | 1 |
| **Total** | | **13** |

`upd03_state_supersession.json` added after the July 12-task public draft; it is the sole cardinality change for the 2026-08-15 pin.

### Systems and mappings

| Adapter | `remember` | `recall` | `revise` | `forget` |
| --- | --- | --- | --- | --- |
| Brainy | `POST /ingest` | `GET /memories/search` | `POST /memories/{id}/correct` | `POST /memories/{id}/suppress` |
| Verbatim | in-process append | token overlap rank | in-place content replace | delete by id |
| Mem0 Platform | `POST /v1/memories/` | `POST /v2/memories/search/` | `PUT /v1/memories/{id}/` | `DELETE /v1/memories/{id}/` |

Hermeticity: Brainy and Mem0 adapters prefix tenants per task (`opmem-{nonce}-{task}`) so shared backends do not collide.

### Assertion semantics

Case-insensitive **substring** match on ranked search `content` fields. See [opmem-spec.md](../opmem-spec.md) for `min_results`, `max_results`, `top_contains`, `must_exclude`, etc.

### Runner exit code

`evals/run_opmem.py` exits **non-zero only on infrastructure errors** (HTTP/adapter exceptions), not on task `fail`. CI `TestOpMemBenchmarkAgainstHTTPServer` therefore proves **harness health**, not that Brainy passes 13/13, unless the test is extended to parse JSON (it does not today).

### Commands (no external LLM required for Brainy+verbatim)

```bash
# CI-equivalent (embedded Postgres + in-process API; no Mem0 key)
go test ./internal/api/ -run TestOpMemBenchmarkAgainstHTTPServer -count=1

# Local API (start API+worker per CONTRIBUTING.md / AGENTS.md)
unset BRAINY_API_KEYS BRAINY_REQUIRE_API_KEY
export BRAINY_ENV=local
python3 evals/run_opmem.py --systems verbatim,brainy --base-url http://127.0.0.1:8080 \
  --json-out /tmp/opmem-report.json

# Mem0 Platform counter-run (requires MEM0_API_KEY; costs API calls)
MEM0_API_KEY=... python3 evals/run_opmem.py --systems mem0 \
  --json-out /tmp/opmem-mem0-report.json
```

### Pin fields to record on any published table

| Field | 2026-08-15 freeze value |
| --- | --- |
| Brainy git SHA | `1b5ab3e` (not re-validated on `bbe55f7` in this inventory) |
| Fixture set | 13 files under `fixtures/opmem/` (git tree at SHA) |
| Mem0 surface | **Platform** API via `Mem0OpAdapter` — **not** Mem0 OSS |
| Mem0 run date | 2026-08-15 |
| Brainy env (fresh local pin) | dedicated DB `brainy_bench`, `BRAINY_RECALL_LLM=1`, `BRAINY_USE_RECALL=1`, worker concurrency 8 |
| Scoring endpoint | `/memories/search` (not `/recall`) |

Artifacts: [opmem-fresh-local-20260815.md](../../benchmarks/artifacts/opmem-fresh-local-20260815.md), [opmem-mem0-fresh-20260815.md](../../benchmarks/artifacts/opmem-mem0-fresh-20260815.md).

---

## Exact claims and evidence table

| Claim | Allowed wording | Evidence in repo | Gaps / caveats |
| --- | --- | --- | --- |
| Brainy passes all **13** OpMem tasks on search path | “13/13 on OpMem v0+`upd03` at SHA `1b5ab3e`” | [opmem-fresh-local-20260815.md](../../benchmarks/artifacts/opmem-fresh-local-20260815.md); [opmem-integrity-20260819.json](../../benchmarks/artifacts/opmem-integrity-20260819.json) (per-task `passed: true`) | Not re-run on `bbe55f7`; no committed Brainy+Mem0 combined JSON |
| Mem0 Platform **10/13** on same fixtures | “10/13 Platform, 2026-08-15; fails `cor02`, `sup03`, `upd02`” | [opmem-mem0-fresh-20260815.md](../../benchmarks/artifacts/opmem-mem0-fresh-20260815.md) | No per-task Mem0 JSON in git; not OSS; API may drift |
| Brainy leads Mem0 on ops | Same-pin **only** with above pins | Category table in mem0-fresh artifact | Do not imply LoCoMo lead; do not use July **9/12** as current Mem0 |
| Verbatim baseline **10/13** | From Brainy fresh run same day | [opmem-fresh-local-20260815.md](../../benchmarks/artifacts/opmem-fresh-local-20260815.md) | Not Mem0-comparable headline |
| July **12/12 vs 9/12** | Historical staging only | [posts/2026-07-opmem-v0.md](../posts/2026-07-opmem-v0.md), [opmem-staging-vs-mem0.json](../../benchmarks/opmem-staging-vs-mem0.json) | **Stale as current** — 12-task set, pre-`upd03` |
| CI guarantees Brainy 13/13 | **Do not claim** | `TestOpMemBenchmarkAgainstHTTPServer` | Exit code ignores task failures |
| “Beats Mem0” on conversational recall | **Forbidden** without same-pin LoCoMo | README / cycle-closeout | OpMem lead ≠ LoCoMo lead |

---

## Independent evidence review (13/13 vs 10/13)

Performed from static repo artifacts only (no re-run).

### Fixture cardinality

`fixtures/opmem/*.json` → **13** files. Sorted runner order matches integrity JSON task list (13 entries).

### Brainy 13/13

- [opmem-fresh-local-20260815.md](../../benchmarks/artifacts/opmem-fresh-local-20260815.md): aggregate **13/13**, category splits correction/isolation/suppression **3/3**, staleness **3/3**, idempotency **1/1**.
- [opmem-integrity-20260819.json](../../benchmarks/artifacts/opmem-integrity-20260819.json): 13 results, each `brainy.passed: true`, `failures: []`.

**Not inflated:** we do not extrapolate to `bbe55f7` without a new run; integrity JSON is Brainy-only, not the full three-system counter-run.

### Mem0 10/13

- [opmem-mem0-fresh-20260815.md](../../benchmarks/artifacts/opmem-mem0-fresh-20260815.md): **10/13 (76.9%)**; failures documented as `cor02`, `sup03`, `upd02` with mechanism notes (ruby stickiness, forget leak, sms preference).
- Arithmetic: 13 − 3 failures = 10 passes. Category table in artifact sums to 10 passed tasks.

**Not inflated:** Platform-only; no reproducible OSS pin; no machine-readable Mem0 per-task log in git to independently re-score without re-running.

### Verbatim cross-check (same Brainy run doc)

Brainy fresh doc reports verbatim **10/13** — shows suite discriminates; does not change Mem0 pin.

### Older 12-task JSON

[opmem-staging-vs-mem0.json](../../benchmarks/opmem-staging-vs-mem0.json) lists **12** tasks (staleness **2/2**, no `upd03`). Brainy **12/12**, Mem0 **9/12**. This is consistent with the July set; it must not be merged into the 13-task pin.

---

## Missing experiments and ablations

From [opmem-spec.md](../opmem-spec.md) roadmap and [inventory-reproduction-plan-2026-09-18.md](../inventory-reproduction-plan-2026-09-18.md):

| Experiment | Why it matters | Repo state |
| --- | --- | --- |
| Mem0 **OSS** same 13 tasks | Fair open comparison | Not run; Platform 10/13 explicit |
| Zep / Letta / LangMem adapters | Paper-topics ≥4 systems | Not implemented |
| v1 task expansion (~30) | expiry, async visibility, races | Not built |
| `revise` vs `POST .../supersede` | ENG-86 supersession API exists | OpMem still maps revise → `/correct` |
| Brainy `/recall` vs `/search` on OpMem | Product path may differ | Harness fixes search lane only |
| Ablation: correction vs suppression vs rank-only | Explain mechanism | No harness |
| Re-run on current `dev` SHA | Current merge bar | **Done** — `20533f2` pin (`opmem-pin-20260918.json`) |
| Commit frozen 3-system JSON (per-task) | Artifact review | **Done** — `docs/benchmarks/artifacts/opmem-pin-20260918.json` |
| Statistical stability | 13 tasks, single run | **Partial** — 5× repeat in `opmem-stability.json` (infra flakiness on correction tasks; see manuscript) |

---

## Threats to validity

1. **Small n:** 13 binary tasks; wide confidence intervals; one-task swings change percentages (~7.7% per task).
2. **Substring scoring:** Passing does not prove semantic correctness or safe production answers.
3. **Search vs recall:** Pins use `/memories/search`; product agents may use `/recall` with different ranking/abstain behavior.
4. **CI gap:** Harness success does not assert pass rate (see protocol).
5. **Mem0 Platform drift:** Vendor API/version changes can move 10/13 without Brainy changing.
6. **Mem0 ≠ OSS:** Headline competitor number may not reproduce for readers on open-source Mem0.
7. **English, domain-neutral fixtures:** May not represent multilingual or vertical-pack content.
8. **No concurrency / TTL:** v0 scripts are sequential; race and expiry bugs undetected.
9. **Adapter fidelity:** Mapping neutral ops to vendor endpoints may not match recommended vendor patterns.
10. **SHA staleness:** Public README cites 13/13 on `1b5ab3e` cycle; `bbe55f7` merge bar not re-measured for OpMem in inventory.
11. **Authoring bias:** Fixtures authored by Brainy team; independent fixture audit not documented.
12. **Semantic contracts:** `sup03` durable forget and staleness without explicit supersede are **normative choices**; failing systems may reflect different valid policies.

---

## Artifact instructions (reviewers and authors)

### Minimal reproduction (Brainy + verbatim)

1. Checkout git SHA to pin (document in report; inventory used `1b5ab3e` for last full pin).
2. `go test ./internal/api/ -run TestOpMemBenchmarkAgainstHTTPServer -count=1`
3. Optional: write JSON via `python3 evals/run_opmem.py --systems verbatim,brainy --base-url <api> --json-out opmem-report.json`

### Full competitive table (authors)

1. Run Brainy and Mem0 in separate invocations or one `--systems brainy,verbatim,mem0` with live API + `MEM0_API_KEY`.
2. Commit **one** JSON with `benchmark`, `systems`, `results[]` per task, `summary`, and manifest: git SHA, date, Mem0 surface (Platform/OSS), env vars affecting retrieval.
3. Update [docs/benchmarks/artifacts/](../../benchmarks/artifacts/) markdown summary; do not replace LoCoMo pins.

### What not to ship

- July **12/12 vs 9/12** as current (use 13-task pin or label historical).
- Platform Mem0 as “Mem0” without qualifier.
- OpMem lead as LoCoMo or LME lead.

### Doc consistency check (maintainers)

```bash
# Spec and ladder should not say "12 tasks" for current v0
rg -n 'v0: 12 tasks|12 tasks\)' docs/research/opmem-spec.md docs/research/public-bench-ladder.md

# Fixture count should match spec
find fixtures/opmem -name '*.json' | wc -l
```

---

## Manuscript outline

**Working title:** *Beyond recall: operational correctness for agent memory systems*

1. **Abstract** — Retrieval benchmarks miss forget/correct/isolate/supersede; OpMem binary suite; Brainy 13/13 vs Mem0 Platform 10/13 on frozen 2026-08-15 pin (qualified).
2. **Introduction** — Production failure modes; gap vs LoCoMo/LME/HaluMem.
3. **Related work** — Retrieval QA benchmarks; memory product surveys; Mem0 staleness/update gaps (cite their reports, not as same-pin scores).
4. **Neutral operation model** — remember/recall/revise/forget; actors; adapter contract.
5. **Task suite** — 13 tasks, five categories; semantic contracts (`sup03`, staleness).
6. **Methodology** — Substring assertions; hermetic tenants; Platform vs OSS boundary; search endpoint.
7. **Results** — Table: Brainy, Mem0 Platform, verbatim; per-task failure taxonomy for Mem0.
8. **Discussion** — What failures imply; policy disagreements on durable forget.
9. **Limitations** — n=13, English, no concurrency, CI vs full pin, SHA age.
10. **Artifact appendix** — Fixture URLs, runner commands, JSON schema.
11. **Future work** — v1 tasks, more adapters, supersede mapping, `/recall` lane.

Target venue: workshop or arXiv benchmark track per [paper-topics.md](../paper-topics.md).

---

## Overlap with draft PRs #185–#187

| PR | Topic | OpMem overlap | Consolidation |
| --- | --- | --- | --- |
| [#185](https://github.com/tryvinci/brainy/pull/185) | n=1540 LoCoMo handoff | Mentions keeping **13/13** green; no OpMem code | No merge conflict; publication package is independent |
| [#186](https://github.com/tryvinci/brainy/pull/186) | eval-protocol-v2 staged LoCoMo | OpMem unchanged; LoCoMo qualification only | Do not conflate protocol-v2 with OpMem runner |
| [#187](https://github.com/tryvinci/brainy/pull/187) | RRF retrieval flag | Test plan: OpMem **13/13** with `BRAINY_RETRIEVAL_RRF=0` before merge | Product change could affect search ranking; **re-run OpMem** after merge if RRF path touches shared rank |

Related: PR [#188](https://github.com/tryvinci/brainy/pull/188) (`inventory-reproduction-plan-2026-09-18.md`) — §1.1 OpMem; this package **implements** Paper 1 deliverables from that inventory without duplicating the full three-paper inventory.

---

## Doc checks (this PR)

- [x] `fixtures/opmem/` count = **13**
- [x] [opmem-spec.md](../opmem-spec.md) updated to **13 tasks** and `upd03`
- [x] [public-bench-ladder.md](../public-bench-ladder.md) L0 OpMem row → **13/13 vs 10/13** with artifact pointer
- [x] [mem0-parity-gaps.md](../mem0-parity-gaps.md) executive OpMem row updated
- [x] [master-plan.md](../master-plan.md) §1.1 OpMem row updated (13 tasks, current pin)
- [x] [posts/2026-07-opmem-v0.md](../posts/2026-07-opmem-v0.md) supersession banner (July numbers preserved)
- [ ] Historical reports (`opmem-baseline-report.md`, `launch-narrative.md`, staging JSON) — intentionally **dated**; not rewritten
- [x] Re-run OpMem on current SHA — `opmem-manuscript-20260918.md`
- [x] LaTeX manuscript — `docs/research/opmem/paper/`
- [ ] Mem0 OSS / Zep / Letta / LangMem — **blocked** (adapters or spend approval)
