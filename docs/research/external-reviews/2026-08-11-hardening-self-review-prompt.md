# Self-review prompt for external reviewer — V3 hardening (2026-08-11)

**Copy/paste this entire file to the external reviewer.**  
It is self-contained. Prefer SHA + artifact citations over vibes.

**Canonical pack:** [../external-agent-assessment-pack.md](../external-agent-assessment-pack.md)  
**Intake SOP:** [README.md](./README.md) · **Archive template:** [TEMPLATE.md](./TEMPLATE.md)  
**Live status:** [../program-execution-status.md](../program-execution-status.md)  
**Qualification:** [../../benchmarks/artifacts/recall-contract-v3-hardening-qualification-20260811.md](../../benchmarks/artifacts/recall-contract-v3-hardening-qualification-20260811.md)

**Repo tips at handoff:** `main` = `308d3a1` (production) · `dev` = `1f2f26f` (staging Render live on `1f2f26f`)

---

## Role

You are an independent architecture / SOTA adjudicator for Brainy (Go + Postgres memory service).  
Your job is a **fresh self-review of the KEEP-V3 hardening cycle after it landed on staging and production**. Challenge claims. Spot honesty gaps. Propose the next 3–7 reviewable PRs only.

Do **not** redesign the product from scratch. Do **not** reopen closed architect PR1–PR7 without new contradictory evidence.

---

## Assume true (do not re-litigate)

1. Five-plane target (source → evidence → semantic → projection → recall) remains the course.
2. Architect PR1–PR7 are **closed**.
3. Recall-contract + multi-hop packet depth already on production before this cycle.
4. Prior external guidance was **KEEP V3, harden** (ordered writes → semantic hops → truthful sufficiency). That hardening has now shipped.
5. Default rejects still stand: fusion retune, graph DB, category dictionaries, conversational SOTA claims, “reader-only” as the next default.

---

## What shipped in this cycle (merged)

| PR | Change | Hot files |
| --- | --- | --- |
| #93 | Per-`(tenant,subject)` claim serialization | `internal/store/postgres/store.go` |
| #94 | Provider NONE/DELETE/UPDATE authoritative over baseline (filter-before-merge) | `internal/memory/provider_extractor.go` |
| #95 | LME `--product-recall` + publish `require_pins` | `evals/public/longmemeval/run.py`, `proveability.py`, `judge.py` |
| #96 | Semantic hop executor V2 + `hop_join_proven` | `internal/memory/hop_executor.go`, `planner.go`, `recall.go` |
| #97 | Truthful hybrid `AnswerStatus` + hop-chain reader prompt | `internal/memory/reader_hybrid.go`, `recall.go` |
| #98 | Gate 0 / qualify docs umbrella | artifacts under `docs/benchmarks/artifacts/*20260811*` |

Production cutover: merge onto `main` at `308d3a1`. Staging Render API+worker live on `1f2f26f`.

---

## Measured pins (cite these; do not invent)

### Gate 0 staging baseline (pre-harden deploy `9bad898`)

| Pin | Result | Artifact |
| --- | ---: | --- |
| OpMem | **13/13** | `opmem-staging-gate0-20260811.md` |
| Marketing | **passed** | `marketing-staging-gate0-20260811.md` |
| LoCoMo 1×30 | **18/30 (60%)** · MH **50%** · OD **25%** · temporal **75%** | `locomo-staging-gate0-1x30-pin-20260811.md` |
| LoCoMo 3×90 | **32/90 (35.6%)** · MH **19.4%** · OD **42.9%** | `locomo-staging-gate0-3x90-pin-20260811.md` |

### Post-hardening local (combined #93–#97, before / around merge)

| Pin | Result | Artifact |
| --- | ---: | --- |
| OpMem / marketing | **13/13** / **passed** | `opmem-harden-nonreg-20260811.md`, `marketing-harden-nonreg-20260811.md` |
| LoCoMo 1×30 | **14/30 (46.7%)** · MH **5/10** · OD **2/4** | `locomo-harden-1x30-pin-20260811.md` |
| LME-20 `--publish --product-recall` | Path `/recall` proven on first items; run later aborted on extraction job failure — **not publishable accuracy** | `lme20-product-recall-partial-20260811.md` |

### Post-cutover staging (`1f2f26f`)

| Pin | Result | Artifact |
| --- | ---: | --- |
| OpMem / marketing | **13/13** / **passed** | `opmem-staging-postcutover-20260811.md`, `marketing-staging-postcutover-20260811.md` |
| LoCoMo 1×30 | **15/30 (50%)** · MH **50%** · OD **25%** · temporal **56.2%** | `locomo-staging-postcutover-1x30-pin-20260811.md` |
| LoCoMo 3×90 | in progress at handoff — do not invent | `locomo-staging-postcutover-3x90-*-20260811.md` when present |

### Context pins (pre this harden cycle)

| Pin | Result |
| --- | ---: |
| LoCoMo V3 early 1×30 | 16/30 · MH 50% · hybrid 17/30 |
| Mem0 same-pin 1×30 | 12/30 · MH 70% |
| LoCoMo V3 3×90 | 31/90 |

**Honesty rule for the dip:** harden local 14/30 and post-cutover staging 15/30 are **not** improvements vs Gate 0 18/30. Expected risk from stricter `hop_join_proven` (lexical bridge no longer counts as proven). Treat as a measured tradeoff, not a win. Do not blend Gate 0 / harden-local / post-cutover pins.

---

## Claims discipline (enforce)

**Allowed now**

- Ops + marketing non-reg hold through Gate 0 and harden.
- Gate 0 staging conversational pins above.
- Harden local 14/30 with explicit dip honesty.
- Product `/recall` path proven under LME `--product-recall`.
- Hardening PRs merged to `dev` and `main`.

**Forbidden**

- Unqualified “beats Mem0”
- SOTA / “MH solved”
- Publishable LME accuracy before a full `--publish --product-recall` run completes without abort
- Calling harden 1×30 an improvement vs Gate 0
- Calling Gate 0 3×90 MH **50%** (it is **19.4%**)
- Treating local pins as staging pins when SHAs differ

---

## Required reading (in order)

1. This prompt  
2. [README.md](./README.md) (revised intake / priority)  
3. [recall-contract-v3-hardening-qualification-20260811.md](../../benchmarks/artifacts/recall-contract-v3-hardening-qualification-20260811.md)  
4. Gate 0 + harden pins listed above  
5. [program-execution-status.md](../program-execution-status.md)  
6. Hot path skim: `hop_executor.go`, `reader_hybrid.go`, `provider_extractor.go`, claim serialization in `store.go`, LME `--product-recall` harness  
7. [external-agent-assessment-pack.md](../external-agent-assessment-pack.md) for architecture context only  
8. Prior briefs only if needed: [2026-08-10-v3-rereview-brief.md](./2026-08-10-v3-rereview-brief.md), [2026-08-07-recall-contract-verdict.md](./2026-08-07-recall-contract-verdict.md)

---

## Return format (mandatory)

Use [TEMPLATE.md](./TEMPLATE.md). Fill every section. In the verdict paragraph, answer **Keep hardening course / adjust / replace**.

Also answer these five questions explicitly:

1. **Course** — Given Gate 0 18/30 → harden local 14/30 with truthful hop-join, was KEEP-V3-harden the right call, or did sufficiency/hop-join overshoot?
2. **MH gap** — Mem0 same-pin MH was 70% vs Brainy ~50% on 1×30. Is the residual gap retrieval binding, extract freshness/ops, hop executor coverage, or answer composition?
3. **Next 3–7 PRs** — Ordered, reviewable, failure-class tagged. Prefer: finish isolated LME-20 publish; staging re-pin on `1f2f26f` if missing; Mem0 same-pin; multi-seed LoCoMo; pack authority/procedures/conflicts. Avoid architecture reopen.
4. **Claims** — What may we say publicly after this cutover vs what remains forbidden?
5. **Kill list** — Confirm what not to do next (fusion retune / graph DB / category dicts / SOTA language / etc.).

### Findings table requirement

For each finding: **Accept / Modify / Reject**, code evidence (file + symbol), and a concrete action. Reject vibes-only findings.

---

## Explicit non-goals

- Do not invent LME scores from partial/aborted runs.
- Do not treat Gate 0 and harden pins as interchangeable.
- Do not propose LOCOMO-named regexes or held-out prompt tuning.
- Do not default to “add a graph database.”
- Do not reopen architect PR1–PR7 without new measured contradiction.
