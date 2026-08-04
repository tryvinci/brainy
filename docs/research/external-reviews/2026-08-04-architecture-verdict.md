# External review — Architecture verdict (2026-08-04)

**Source:** External architecture assessment of Brainy five-plane / Fusion V2 mid-migration  
**Adjudicator:** Brainy coding agent (spot-checked against `main`/`dev` tip)  
**Status:** **Accepted** as course correction for remaining SOTA work

## Verdict (accepted)

> The five-plane / Postgres **target** architecture is SOTA-capable. The **current implementation** is a record-centric memory service mid-migration and is **not** yet structurally capable of SOTA conversational memory — despite useful progress in fusion, lifecycle correctness, and vertical fixtures.
>
> OpMem and vertical greens currently validate the legacy `memory_records` path more strongly than the new planes.

## Adjudication (verified)

| Finding | Verdict | Evidence |
| --- | --- | --- |
| Evidence shadows extracted `record.Content`; `source_ref` = memory ID | **Accept** | `internal/memory/evidence_plane.go` |
| Evidence hash `tenant+content` collapses across subjects | **Accept** | `ShadowWriteEvidence` unique `(tenant_id, content_hash)` |
| Fusion “bm25” is token coverage, not `ts_rank_cd` | **Accept** | `SearchOpt` → `ScoreAndRankV2` |
| `NormalizeBM25Sigmoid` unused on hot path | **Accept** | `fusion_v2.go` |
| `looksMultiHopQuery` over-fires → 400-row scans | **Accept** | ask + ≥2 bearing tokens |
| Non-128-d embeddings → full `LoadEmbeddings` | **Accept** | `SearchByEmbedding` |
| `current_state` upsert before auto-supersede | **Accept** | ingest order |
| `/recall` ignores `view`/`as_of`; coverage nominal | **Accept** | `recall.go` |
| Pack v2 entities/state machines not loaded | **Accept** | `pack.LoadFile` only `pack.yaml` |
| Reject fusion retune / graph DB / category dictionaries | **Accept** | anti-benchmax + ADR-004 |

## Accepted work sequence

1. Stage-oracle evaluation + failure ledger  
2. Evidence Plane v2 (raw capture before extract)  
3. Structured semantic extraction v3  
4. Temporal resolver + guarded current-state  
5. Retrieval store v3 (real FTS rank, embedding-dim indexes, kill scan fallbacks)  
6. Typed query planner + evidence packets  
7. Executable packs v2  

## Implementation maturity (at review time)

| Plane | Maturity |
| --- | --- |
| Source | Messages on ingest request only |
| Evidence | **Shadow of interpretation** (not raw) |
| Semantic | Text-first records + partial atoms/events |
| Projection | Unsafe last-write `current_state` MVP |
| Recall | Synthesis wrapper over generic search; intents cosmetic |

## Follow-up (this cycle)

- Artifact truthfulness pass  
- Operational oracle modes + ledger scaffolding  
- Evidence Plane v2 raw capture + subject-safe dedupe  
- FTS rank into fusion; tighten multi-hop scan; guard current_state write order  
