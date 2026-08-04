# Program execution status — SOTA end-to-end (2026-08-04)

**Program of record:** [sota-end-to-end-program.md](./sota-end-to-end-program.md)  
**Baseline freeze:** [phase0-baseline-and-oracle.md](./phase0-baseline-and-oracle.md)

## Realistic adjudication of the attached guide

| Suggestion | Verdict | Action taken |
| --- | --- | --- |
| Keep Postgres; no graph DB | Accept | Unchanged architecture |
| Fusion necessary but not sufficient | Accept | Fusion V2 default-on (`BRAINY_FUSION_V2`, disable with `false`) |
| Immutable evidence plane | Accept (shadow) | `memory_evidence` migration 16 + shadow writes |
| Full bitemporal model | Accept (incremental) | atoms: `valid_from`/`recorded_at`/`retired_at` (mig 15); `valid_to` already existed |
| Predicate-specific policy (not latest-wins) | Accept | `predicate_policy.go` + stateful auto-supersede gate |
| Events + participants | Accept (MVP) | `memory_events` + `memory_event_participants` |
| Query intents + answer statuses | Accept | `AnalyzeQueryIntents`, recall `answer_status` / abstention |
| Evidence-set selection (not flat top-k) | Accept | `selectEvidenceSet` for list/multi-hop |
| Remove hot-path full subject scans | Accept (bounded) | `ListMemoriesLimited` + GetMemory admit with status filter |
| Packs v2 | Accept (scaffold) | `packs/{support,marketing}/v2/` + fixtures |
| Associative triggers / learned policy / graph DB | Defer | Gated research (Phases 7–8) |
| Benchmax / tune on convs 4–10 | Reject | Denylist + holdout policy unchanged |

## Code landed this cycle

- Phase 0: search/recall traces, failure taxonomy, oracle helpers, baseline doc
- Phase 1: Fusion V2, over-fetch, evidence-set selector, limited subject scans
- Phase 2: evidence shadow, bitemporal atom fields, current_state projection table
- Phase 3: events MVP + extraction wiring on ingest/worker
- Phase 4: intent classifier + answer statuses on `/recall`
- Phase 5: support/marketing pack v2 scaffolds + fixtures
- Phase 6: OpMem + marketing + support harness green via `go test ./internal/api/`

## Flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `BRAINY_FUSION_V2` | **on** | Mem0-style additive fusion |
| `BRAINY_ENTITY_RANKING` | off | Prior A/B regressor |
| `BRAINY_IDF_RANKING` | off | Opt-in IDF lexical |
| `BRAINY_ENSURE_FTS_INDEX` | off on API | Controlled GIN build |

## Remaining (measured / ops)

- Staging LoCoMo multi-seed re-measure under Fusion V2
- LongMemEval failure adjudication (≥95% labeled sample)
- Indexed retrieval p50/p95 SLO proof under load
- Expand OpMem / SupportBench fixture counts per program §13
- Full bitemporal reads (`as_of`, system-time) end-to-end API
- Entity aliases / reversible merge (MEM-030/031)

## Phase 6 staging smoke (post-merge to `dev`)

| Run | Result |
| --- | ---: |
| LoCoMo smoke 1 conv / 20 Q (async, Fusion V2 live) | **25% (5/20)** |
| multi-hop (n=8) | **50%** |
| temporal (n=10) | **10%** |
| Search p50 / p95 | 1627 / 2867 ms |

Artifact: `docs/benchmarks/artifacts/locomo-smoke-fusionv2-20260804.md`

Interpretation: small-n smoke; multi-hop slice improved vs full-set MH≈26% baseline, but temporal slice is weak and overall is not a publish claim. Full 3-seed re-measure still required before SOTA gates.

## Production push + progress retest (2026-08-04)

- **`main` pushed** (`13d6b3b`) with user approval. Render hosts **staging on `dev`** only (no separate prod service).
- LoCoMo smoke 2 conv / 30 Q: **50% (15/30)** — temporal **66.7%**, multi-hop **25%**, search p50/p95 ≈ 1.7s / 3.0s  
  Artifact: `docs/benchmarks/artifacts/locomo-smoke-main-progress-20260804.md`
- OpMem staging: **13/13**
- Support vertical: **4/4**; Marketing: **17/17** (after semantic-only isolation fix)
- vs prior: Fusion V2 first smoke 25% (5/20); publish-stack full LoCoMo mean still ≈49.8% (3-seed) until re-run

## External review course correction (2026-08-04) — execution

Accepted architecture verdict: five-plane **target** approved; implementation remains **record-centric mid-migration**.

Landed this cycle:

- External review archive + intake SOP (`docs/research/external-reviews/`)
- Assessment pack honesty / maturity matrix refresh
- Evidence Plane v2: raw message capture before extract; subject-safe dedupe (migration 17)
- FTS `ts_rank_cd` plumbed into Fusion V2 when available
- Bounded embedding fallback (no unbounded `LoadEmbeddings` hot path)
- Stricter multi-hop scan gate (ask + ≥3 bearing tokens)
- Current-state projection after auto-supersede on sync ingest
- Oracle modes: `evidence` operational; unsupported modes return `oracle_unsupported`
- Failure ledger scaffolding under `docs/benchmarks/artifacts/failure-ledger/`

Still open (queued): typed extract v3, temporal resolver/as_of, executable packs v2, full LME adjudication (≥100), pgvector for non-128 dims, planner/evidence packets.

