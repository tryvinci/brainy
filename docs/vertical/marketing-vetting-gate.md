# Marketing Vetting Gate

**Status:** Approved product policy  
**Purpose:** Define how Brainy proves vertical memory **on marketing** before any second vertical (finance, etc.) or major runtime expansion.  
**Last updated:** 2026-06-23

---

## Principle

The general vertical runtime (`primitive`, packs, lifecycle, rank policy) is only considered **proven** when marketing agent jobs pass reproducible, automated vetting at each tier. Finance and other packs are **blocked** until **Gate M3** clears.

No second vertical until marketing technical proof — not discovery docs, not pack YAML drafts alone.

---

## Vetting ladder

```
Tier 0  Unit + package tests (go test ./...)
Tier 1  Mem0 parity fixtures (core ingest/search/dedupe/correct)
Tier 2  Marketing golden fixtures (BV + LC per pack eval_fixtures)
Tier 3  Marketing MVP benchmark (parity + vertical + Mem0 gap matrix)
Tier 4  Marketing capability depth (all use-case eval seeds + semantic non-regression)
Tier 5  Second vertical unlock (finance research → pack only after Gate M3)
```

| Gate | Tiers required | Meaning | Status |
| --- | --- | --- | --- |
| **M1 — Deterministic MVP** | 0–3, all green in CI | Marketing pack runs on general runtime; Mem0 parity held; documented differentiation | **Passed** (local `dev`) |
| **M2 — Publish** | M1 + push `dev`, PR, CI on origin | Reproducible off one laptop | **Open** (ENG-91) |
| **M3 — Marketing technical proof** | M2 + Tier 4 | All marketing use-case eval seeds covered; pgvector does not regress deterministic suite | **Open** |
| **M4 — Second vertical** | M3 + explicit architecture sign-off | Finance pack work may begin (vocabulary + fixtures, not schema fork) | **Blocked** |

---

## How each tier runs

### Tier 0 — Unit and integration tests

```bash
go test ./...
```

Includes embedded Postgres API tests that run parity, vertical, and MVP benchmark harnesses against a real HTTP server.

### Tier 1 — Parity (Mem0 thin-slice baseline)

| Item | Location |
| --- | --- |
| Fixtures | `fixtures/parity/` |
| Runner | `evals/run_eval.py` |
| CI | `TestEvalHarnessAgainstHTTPServer` |

**Pass criteria:** Every fixture `passed: true`. Regressions block merge.

Mem0 reference pin: `docs/mem0-parity-matrix.md` (commit `a670333d…`).

### Tier 2 — Marketing golden scenarios

| Item | Location |
| --- | --- |
| Fixtures | `fixtures/vertical/marketing/` (`packs/marketing/v1/pack.yaml` → `eval_fixtures`) |
| Runner | `evals/run_vertical_eval.py` |
| CI | `TestVerticalEvalHarnessAgainstHTTPServer` |

**Pass criteria:** All non-`skip` fixtures pass. New marketing behavior requires a new fixture before merge.

### Tier 3 — Marketing MVP benchmark

| Item | Location |
| --- | --- |
| Capability matrix | `evals/marketing_mvp_matrix.json` |
| Runner | `evals/run_marketing_mvp_benchmark.py` |
| Report | `docs/vertical/marketing-mvp-benchmark.md` |
| CI | `TestMarketingMVPBenchmarkAgainstHTTPServer` |

**Pass criteria:**

- `mvp_ready: true` (parity + vertical suites green)
- Differentiation score: every `mem0_has: false` capability in the matrix must have `brainy_pass: true`
- Report committed or regenerated in CI on release branches

This tier answers: *“Does Brainy beat generic Mem0 on marketing-specific capabilities we claim?”*

### Tier 4 — Marketing capability depth (Gate M3)

Tier 4 is **not** satisfied by M1 alone. It requires coverage of all eval scenario seeds in `docs/vertical/marketing-use-case-map.md` and semantic retrieval without deterministic regression.

| Seed | Scenario | Fixture / test | M3 status |
| --- | --- | --- | --- |
| 1 | Principle over preference | `bv01_principle_over_preference` | Done |
| 2 | Taboo suppression leak | `bv02_suppression_leak` | Done |
| 3 | Active campaign ranks above completed | `lc02_active_campaign_ranks_above_completed` | Done |
| 4 | Campaign end suppresses stale context | `lc01_archived_campaign_hidden` | Done |
| 5 | A/B outcome updates retrieval rank | — | **Missing** (needs MVP-3 belief/outcome) |
| 6 | Correction stickiness (paraphrase) | `bv04_correction_stickiness` | Done |
| 7 | Multi-brand isolation | `bv06_multi_brand_isolation` | Done |
| 8 | Cross-campaign pattern retrieval | — | **Missing** |
| 9 | Style-matched creative ranks first | — | **Missing** (needs TasteSignal or semantic) |
| 10 | Scoped segment prefs coexist | — | **Missing** |

Additional M3 requirements:

| Capability | ENG / MVP | M3 status |
| --- | --- | --- |
| Pack JSON Schema validation on ingest | MVP-4 | Not started |
| Outcome → Belief rank loop | MVP-3 | Not started |
| Semantic / hybrid retrieval | ENG-87 (after ENG-63 PD) | Not started |
| Paraphrase robustness under embeddings | ENG-87 + new fixtures | Not started |
| Pack-driven `classification_rules` | MVP-1.1 | Not started |

**Pass criteria for M3:**

1. All ten use-case eval seeds have passing fixtures (or documented `skip` with issue link and expiry).
2. `go test ./...` green including Tiers 1–3.
3. After ENG-87 lands: hybrid search suite passes; deterministic fixtures still pass unchanged.
4. Architecture review recorded (hypothesis + falsification test per new primitive behavior).

### Tier 5 — Second vertical (Gate M4)

**Blocked until M3.**

Allowed before M4:

- Research notes, competitor scans, vocabulary sketches (no `packs/finance/` merge without M3)
- Generic runtime work that benefits all verticals **if** marketing fixtures still pass

Allowed after M4:

- `packs/finance/v1/pack.yaml` + `fixtures/vertical/finance/`
- Finance eval runner mirroring marketing pattern
- No new Postgres `kind` enums; finance uses same primitives + pack labels

---

## CI and release policy

| Event | Required vetting |
| --- | --- |
| Every PR to `dev` / `main` | Tier 0–3 (via `go test ./...`) |
| Release candidate | Tier 0–3 + manual Tier 4 checklist until Tier 4 is fully automated |
| Finance pack PR | Gate M4 + finance fixture suite (future) |
| pgvector / provider extraction PR | Tier 1–3 must not regress; add Tier 4 fixtures for new behavior |

Merge policy: **no fixture deletion** to green CI without replacing coverage. **no `skip` without** issue ID and target date.

---

## What “marketing MVP” meant vs “marketing proven”

| Term | Gate | Honest summary |
| --- | --- | --- |
| **Marketing MVP (ENG-93)** | M1 | Deterministic runtime + pack + golden evals + benchmark report. **Shipped on `dev`.** |
| **Marketing technically proven** | M3 | Full agent-job coverage + semantic layer + pack validation. **Not yet.** |
| **Ready for finance** | M4 | M3 + sign-off. **Blocked.** |

Do not treat ENG-93 completion as permission to start finance implementation.

---

## Testing & exposure

Vetting (Tiers 0–4) is not the same as product exposure or competitor benchmarking. Three modes run in parallel after **M2**.

| Mode | What | When | Merge gate? |
| --- | --- | --- | --- |
| **Fixture / CI** | Golden scenarios via `go test ./...` | **Now** | Yes — every PR |
| **Staging API** | Shared HTTP endpoint for dogfood + partners | **M2** | Deploy gate |
| **Live competitor benchmarks** | Same fixtures against Mem0/Zep APIs | **Start M2**, publish **M3** | No — evidence only |

### Fixture testing (now)

Primary quality bar. Does not require staging, API keys, or public repo.

```bash
go test ./...
# or against local API
go run ./cmd/api
python3 evals/run_marketing_mvp_benchmark.py --base-url http://127.0.0.1:8080
```

### API exposure

The HTTP API already exists locally (`POST /ingest`, `GET /memories/search`, correct/suppress, async worker). Exposure is an **ops milestone**, not a greenfield build.

| Audience | Gate | Notes |
| --- | --- | --- |
| Local / team | Now | Runbook: `docs/external-postgres-runbook.md` |
| Internal dogfood | **M2** | Staging Postgres + stable URL |
| Design partners | **M3** | Marketing vertical credible; API keys |
| Production customers | **M3 + commercial** | Auth, billing, SLA — see GTM roadmap |

### Competitor benchmarks

| Stage | What | Gate |
| --- | --- | --- |
| Documented Mem0 gap (Tier 3) | `marketing_mvp_matrix.json` — expected behavior, not live API | **M1 ✅** |
| Live Mem0 parity side-by-side | `fixtures/parity/` on both APIs | **M2** |
| Published marketing moat report | Vertical fixtures + methodology | **M3** |
| Multi-vendor (Zep, etc.) | Public scenarios only | **M3+** |

Live competitor runs are **not** a substitute for fixture CI. They support positioning and sales after staging exists.

Full open source, benchmark publication, and commercial timelines: [`go-to-market-roadmap.md`](./go-to-market-roadmap.md).

---

## Product plan alignment

| Document | Role |
| --- | --- |
| This file | Canonical gate definitions |
| `docs/vertical/verticalization-model.md` | Architecture; Phase 7 finance deferred to M4 |
| `docs/vertical/marketing-use-case-map.md` | Eval seeds → fixture backlog for M3 |
| `docs/mem0-parity-matrix.md` | Tier 1 reference behavior |
| `docs/brainy/09-iteration-and-productization.md` | Release regression policy |
| `evals/README.md` | Operator commands |
| `docs/vertical/go-to-market-roadmap.md` | Open source, benchmark publication, commercial API |

---

## Operator quick reference

```bash
# Full local vetting (Tiers 1–3)
go test ./...

# Or against a running API
go run ./cmd/api
python3 evals/run_marketing_mvp_benchmark.py --base-url http://127.0.0.1:8080
```

---

## References

- `docs/vertical/marketing-mvp-benchmark.md` — latest Tier 3 report
- `docs/vertical/marketing-brand-voice-spec.md` — ENG-81 behavior under test
- Linear: ENG-91 (publish), ENG-87 (semantic), ENG-56/76 (finance, blocked at M4)
