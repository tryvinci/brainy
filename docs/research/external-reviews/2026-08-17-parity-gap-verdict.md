# External review — current-SHA parity and implementation gap

**Date:** 2026-08-17
**Source:** uploaded deep-research report ("Brainy competitive parity and implementation-gap review")
**Adjudicator:** coding agent on `pr/full-recall-dip-review-a6c7`
**Prompt used:** [2026-08-17-full-recall-self-review-prompt.md](./2026-08-17-full-recall-self-review-prompt.md) (all eight questions)
**Report archive:** [2026-08-17-parity-gap-review.md](./2026-08-17-parity-gap-review.md)
**Pinned to:** docs SHA `8492ad3`, product SHA `1b5ab3e` (live remasure). This is the **live** review.

A same-day Wave 1 archaeology report pinned to `bd987fa` was already adjudicated: [2026-08-17-competitive-archaeology-verdict.md](./2026-08-17-competitive-archaeology-verdict.md). Keep that file as **historical** (do not re-queue P0-P7). **Live next-work is this verdict**, not that one.

Spot-checked against current code: `write_policy.go`, `fusion_v2.go`, `planner.go`, `recall.go`, `relations.go`, `entities.go`, `hop_executor.go`. Claims held.

## Verdict (1 paragraph)

**Keep** the representation-first course (facts for recall, relations for reasoning, episodes for proof, projections for truth). **Adjust** the immediate sequence: Brainy already has more of the competitor checklist than gap docs sometimes implied (append-only core writes, multi-signal fusion, overfetch-to-120, ContextEvidence/ProofChain split, `memory_relations` v1, LME publish integrity). The product answer path is currently much worse than the information surface beneath it, so the first PR is **R5A structured-first `/recall`** (retire `firstStatementFromPacket` as a normal factual strategy), not "fix OD" as a bench-named PR, and not a reader-prompt sprint. Do **not** treat **11.4% to ~50%** as a measured ceiling on current storage: July **49.4/49.8%** was a different stack; a **current-SHA search+harness diagnostic** is required to size that lever. Remaining work after R5A is typed packets, compiler coverage that generalizes past conv-26, canonical entity IDs, relation V2, hop joins on those IDs, then frozen dual-path qualification. Do not re-queue R0-R4. Do not land proposed `memory_*_v2` DDL in R5A. Do not put calendar-week estimates into PoR.

## Answers to prompt questions

1. **Course:** Keep representation-first. The two-gap split is **directionally** right: (a) a real product answer-path gap (`firstStatementFromPacket` / enumerate / abstain on items where gold-bearing material is in the packet); (b) a deeper representation gap (compiler coverage does not generalize; 1x30 21/30 vs full SH 10.5%; LME multi-session 0/5; even July search+harness was 49.8% vs Mem0 Platform 92.5% n=1540 top-k 200). **Modify** the numeric ceiling: 11.4% to ~50% is not proven on the current SHA. Slogan/enumerate/abstain failures are still causal on a meaningful subset. Structured-first lands first because an answerer that fails on already-retrieved facts hides later representation work.

2. **Published metric (`/recall` vs search+harness):** **Accept two lanes.** README bake-off row stays product `POST /recall` (n=1540, cats 1-4; current freeze 175/1540 = 11.4%). A separately labeled industry-format row is search to shared answerer to shared judge, n=1540, top-k 200, 3 seeds, report actual retrieved tokens. Do not mix 11.4 with Mem0 92.5 as identical paths. 49.8% is that industry lane on an **old** stack, not current.

3. **First product PR (cite-facts vs R5 OD):** **R5A structured-first `/recall`.** Failure class: `READER_MISS` / `ANSWER_PATH_MISS`. Scalar answers consume typed fact/atom/relation values; lists enumerate structured values; hops consume proof-chain outputs; abstain only after structured support is assessed. **Modify:** R5-on-OD remains a **diagnostic** (1x30 OD 0/4 vs Mem0 3/4) inside that PR's test plan, not the product PR name. Do not restore OD/SH by stuffing episodes into top-k.

4. **Fair Mem0 stack:** Eventual n=1540 compare: same dataset hash, top-k/candidate output 200, same answerer model and prompt, same judge (temp 0), 3 seeds, mean/p50/p95 retrieved tokens, latency, cost, category accuracy. Label Mem0 Platform vs OSS. Their published 92.5 remains a vendor-published quality bar, not a current same-pin. Top-k 200 on our `/recall` is not a substitute for citing facts.

5. **Skipped suites (LME-500 / BEAM 1M):** **Accept skip.** LME-20 is 4/20 (multi-session 0/5); BEAM 100K is 8/20. LME-500 becomes useful when LME-20 is no longer failing the same holes and a larger sample would change which hypothesis we pick. BEAM 1M/10M becomes useful when scale itself is the active uncertainty. Do not use expensive scale benches as architecture substitutes.

6. **Next 3-7 PRs:** R5A structured-first `/recall` to R5B typed EvidencePacket + spans to R6 Compiler Coverage V2 to R7 Canonical Entity V2 to R8 Relation V2 to R9 Hop Executor V3 (canonical ID joins; unscoped/fuzzy cannot be `typed_exact` proof) to R10 frozen dual-path qualification. See accepted sequence below. Omit week estimates from PoR. Schema DDL is R7/R8, dual-write later, not next.

7. **Claims allowed vs forbidden:** Allowed: OpMem 13/13 vs Mem0 10/13 this freeze; marketing 17/17 vs Mem0 4/17 empirical; 1x30 21/30 vs Mem0 11/30 (trail OD 0/4 vs 3/4); full `/recall` 175/1540 (11.4%) as a named dip vs July search+harness 49.8%; LME-20 4/20 jobs 4829=4829; BEAM 100K 8/20 non-reg; Mem0 92.5 / 94.4 as vendor-published, not same-pin. Forbidden: SOTA, beats-Mem0, MH solved, 70% as full LoCoMo, restoring 49.8% as current, calling 11.4% a harness glitch, mixing 11.4 with 92.5 as identical paths, LME 4/20 vs 94.4, BEAM 8/20 vs 64.1 (1M), treating ops/vertical lead as conversational lead.

8. **Kill list confirmation:** Confirmed. Plus: no reopening R0-R4 as if missing; no fusion constants (`ENTITY_BOOST_WEIGHT=0.5`); no spaCy requirement; no graph DB / Neo4j; no LoCoMo-named product rules; no episode stuffing; no reader-prompt sprint; no another full n=1540 remasure before R5A; no LME-500-as-quality; no proposed v2 DDL in R5A; no calendar-week commitments in PoR.

## Findings table

| Finding | Accept / Modify / Reject | Code evidence | Action |
| --- | --- | --- | --- |
| Keep representation-first; adjust sequence to R5A then coverage then identity | **Accept** | PoR already representation-first; `/recall` still loses on retrieved facts | Live next-work = this sequence |
| Two-gap thesis; 11.4% to ~50% is the answer-path lever | **Modify** | July 49.8% was search+harness on an old stack, not a current-SHA oracle | Directionally yes; size the ceiling with current-SHA search+harness on a stratified subset after R5A |
| Slogan / enumerate / abstain are causal, not cosmetic | **Accept** | `recall.go` `firstStatementFromPacket` (first non-question `pkt.Contents` / `TemporalAnswer`); list-cue enumerate; abstain `not in memory`; `conv-26-q83` slogan | R5A retires that as a normal factual strategy |
| Core writes are already append-only; gap is coverage not ADD-vs-UPDATE | **Accept** | `write_policy.go` `WriteMutationModeOf`: core/empty -> `append_only`; non-core vertical -> governed | Do not reopen UPDATE/DELETE for conversation |
| Hybrid retrieval is already present | **Accept** (reject "needs fusion") | `fusion_v2.go` `ScoreAndRankV2Temporal` (dense + BM25-like + entity + temporal) | Do not fusion-fish |
| "Only top-30 vs Mem0 200" | **Modify** | Default `Recall` `topK=30`, budget 4000; `CandidateOverfetch` = max(limit*4, 60) cap 200 -> typically **120** | Separate candidate pool, final context budget, proof budget; measure tokens, do not merely raise top-k |
| ContextEvidence / ProofChain missing | **Reject** as missing; **accept** quality gap | `planner.go` `EvidencePacket` has both; `ContextEvidence []string`; legacy `Contents` still drive `firstStatementFromPacket` | R5B typed objects + spans; R5A must stop answering from raw `Contents` |
| Relation memory missing | **Reject** literally; **accept** quality gap | `relations.go` `MemoryRelation.{SrcEntity,DstEntity}` normalized strings; mig v20 | R8 Relation V2 with canonical IDs, validity, provenance; copy Graphiti **semantics**, not Neo4j |
| Canonical entities already exist | **Reject** | `entities.go` quoted / proper-noun / year string keys + `memory_entity_links` hub | R7 Canonical Entity V2 (IDs, aliases, mentions, merge). Names are not unique identities (two Johns). Tenant/subject scoped |
| Typed hops are competitor-grade | **Reject** | `hop_executor.go` `resolveEntityHop` sets `res.Value = mention`; unscoped `GetCurrentState(..., pred)` fallback; atom path `if len(filtered) > 0 { picked = filtered }` keeps all predicate hits if entity filter is empty | R9: canonical IDs are hop I/O; unscoped/fuzzy may inform context, not `typed_exact` proof |
| LME integrity is the blocker | **Reject** | LME-20 4/20, jobs 4829=4829 | Integrity stays; quality is 4/20 |
| Mix 11.4 with Mem0 92.5 as one bake-off | **Reject** | Different path / top-k / harness | Two labeled lanes |
| Graph DB required for MH | **Reject** | ADR-004; hops already Postgres | Keep Postgres graph-shaped tables |
| Copy Mem0 OSS scoring / spaCy NER | **Reject** | Brainy already has fusion + entity regex; review itself says inspect not port | Kill list |
| Proposed `memory_entities_v2` / `facts_v2` / `relations_v2` DDL as next | **Modify** | Proposed, not existing | Dual-write later in R7/R8; **not** R5A |
| R5-on-OD as the named first PR | **Modify** | Full SH 88/841 (10.5%) is the mass; OD 0/4 is WRITE_MISS + answer-path | R5A general structured-first; OD is a diagnostic checkpoint |
| Skip LME-500 / BEAM 1M | **Accept** | LME-20 4/20; BEAM 100K 8/20 | Unchanged |
| Wave 1 P0-P7 / archaeology sequence as live next-work | **Reject** | R0-R4 landed; 1x30 21/30 vs Wave 1 14/30 | Archaeology verdict stays historical |
| Calendar-week estimates (4-12 eng-days) as PoR | **Reject** | Review itself: planning assumptions, not commitments | Omit from PoR |
| Histogram sum adds float bit patterns | **Accept (already fixed)** | `internal/observability/metrics.go` now CAS-adds floats | Landed on this branch; not the cycle |

## Accepted next sequence

1. **R5A structured-first `/recall`** — `READER_MISS` / `ANSWER_PATH_MISS`. Consume typed fact/atom/relation values; retire `firstStatementFromPacket` as a normal factual strategy; enumerate structured lists; abstain only after structured support. Hybrid reader for synthesis only. Episodes remain provenance + fallback. **Not** a prompt sprint. **Not** "fix OD". Exit: OpMem 13/13 and marketing 17/17 stay green; 1x30 diagnostic; stratified 100-200 SH/OD/temporal subset; **current-SHA search+harness on that same subset** to size answer-path vs WRITE_MISS. Not a full n=1540 remasure.
2. **R5B typed EvidencePacket + spans** — `EVIDENCE_COVERAGE_MISS` / `PROOF_MISS`. `ContextEvidence` becomes structured objects; ProofChain stays separate; legacy `Contents` compatibility-only.
3. **R6 Compiler Coverage V2** — `WRITE_MISS` / `REPRESENTATION_MISS`. Generalize past conv-26; ADAPT Mem0 recent-session + existing-memory context and ADD-only semantics (not verbatim prompts). Durable assistant facts stay first-class. Exit: held-out representation audit, not a LoCoMo bump.
4. **R7 Canonical Entity V2** — `ENTITY_LINK_MISS`. Durable IDs, aliases, mentions, ranked resolution, merge lifecycle. Dual-write the hub. Tenant/subject scoped. Two Johns coexist.
5. **R8 Relation V2** — `RELATION_MISS`. Project entity-valued facts onto canonical-ID edges with validity and evidence spans. Dual-write v1 strings. Copy Graphiti semantics, not Neo4j.
6. **R9 Hop Executor V3** — `PROOF_MISS` / `PLANNING_MISS`. Canonical IDs are hop I/O (`hop[i].output == hop[i+1].input`). Unscoped/fuzzy cannot constitute `typed_exact` proof. Do not claim MH-solved from 1x30 10/10 while full MH is 7.4%.
7. **R10 frozen dual-path qualification** — product `/recall` and industry-format search+shared-answerer+shared-judge, labeled separately; 1x30 diagnostic then 3x90; LME-20 quality; OpMem/marketing green. Full n=1540 only after a freeze. No SOTA / beats-Mem0.

## Rejected / deferred

- Re-queue R0-R4 / Wave 1 P0-P7 as missing substrate
- v2 entity/fact/relation DDL as the next PR (R7/R8 later; dual-write)
- R5-on-OD as the product PR name
- Treating 49.8% as a current-SHA `/recall` ceiling
- Calendar-week estimates in PoR
- Fusion fishing, graph DB, category dictionaries, LoCoMo-named rules, episode top-k to restore OD/SH
- Reader-prompt sprint / LLM-over-episodes as the SOTA bet
- spaCy as a requirement; copying `ENTITY_BOOST_WEIGHT=0.5`
- Another full n=1540 remasure before R5A
- LME-500 or BEAM 1M as a quality claim
- Mixing product 11.4% with Mem0 92.5% as one number

## Claims discipline check

- [x] No invented LME / SOTA / BEAM 1M scores
- [x] Gate 0 vs harden vs remasure pins not blended
- [x] 1x30 70% not published as full LoCoMo
- [x] 11.4% labeled product `/recall`; 49.8% labeled search+harness (not current; not a proven current-SHA ceiling)
- [x] Vendor 90+ labeled n / metric / top-k / path
- [x] Dip honesty preserved where scores fell

## Artifact diffs required

- [x] assessment pack
- [ ] codebase graph md/json (no structural change this pass)
- [x] program-execution-status
- [x] PoR adoption note (`sota-representation-path.md` R5A-R10 + ceiling honesty)
- [x] external-reviews/README.md priority pointer

## Linked PRs / commits

- Adjudication lands on `pr/full-recall-dip-review-a6c7` (GitHub PR targeting `dev`)
- Live product pin: `1b5ab3e`
- Docs pin: `8492ad3`
- Report pin: current SHA (this review), not `bd987fa`
- ENG-168 remains the conversational epic; this does not reopen R0-R4
- Histogram CAS-add fix already on this branch (`22ff11d`); not the cycle
