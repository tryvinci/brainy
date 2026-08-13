# LME-20 product-recall — Wave D item ledger (2026-08-13)

**Pin:** `lme20-product-recall-pr1-20260812` (integrity, not quality)  
**Accuracy:** **0/20** · jobs **4829=4829** failed=0 · `/recall` 20/20  
**Dataset SHA:** `d6f21ea9d60a0d56f34a05b609c79c88a451d2ae03597821ea3d5a9678c3a442`  
**Source:** local gitignored run json (answers + top-k mined here; no secrets committed)

This is a **failure-class histogram**, not a quality claim. 0/20 is not empty ingest.

## Headline

| Signal | Value |
| --- | --- |
| Mean search hits | **24.45** (range 6–30) |
| Gold string in top-k | **6/20** (strict); several more have a related token at a bad rank |
| Empty retrieval | **0/20** |
| Primary labels | **EVIDENCE_COVERAGE_MISS 12** · **READER_MISS 6** · **ABSTENTION_MISS 2** |

Pattern: `/recall` often returns **assistant continuation / listicles / marketing copy**. Top hits are frequently `conversation_episode` of assistant turns. Example: gold **4 weeks**; rank-1 already says **3 weeks** inside a congratulations episode; the answer is still congratulations boilerplate.

`single-session-user` is 0/3 **without** missing graph edges. PR6–PR8 are **not** the LME driver.

## Histogram by type

| Type | n | 0/n | Dominant class |
| --- | ---: | --- | --- |
| temporal-reasoning | 6 | 0/6 | coverage + temporal/history |
| multi-session | 5 | 0/5 | coverage / reader (wrong span in packet) |
| knowledge-update | 3 | 0/3 | superseded current drowned by assistant episode; 1 abstention |
| single-session-user | 3 | 0/3 | coverage (facts not in top-k) |
| single-session-assistant | 2 | 0/2 | assistant content present as episodes/listicles; reader misses |
| single-session-preference | 1 | 0/1 | assistant boilerplate ranked first |

## Item labels

Sanitized: gold vs rank-1 vs `/recall` answer (truncated). Primary taxonomy from `docs/benchmarks/artifacts/failure-ledger/README.json`.

| ID | Type | Primary | Gold in top-k | Notes |
| --- | --- | --- | --- | --- |
| `7a87bd0c` | knowledge-update | EVIDENCE_COVERAGE_MISS | no | gold `4 weeks`; rank-1 assistant “3 weeks” congratulations |
| `00ca467f` | multi-session | READER_MISS | rank 3 | gold `2`; answer unrelated listicle |
| `4388e9dd` | single-session-assistant | EVIDENCE_COVERAGE_MISS | no | gold Andy shirt; answer furniture ad |
| `6b168ec8` | single-session-user | EVIDENCE_COVERAGE_MISS | no | gold `three` bikes; credit-advice episode on top |
| `d24813b1` | single-session-preference | EVIDENCE_COVERAGE_MISS | no | lemon-poppy preference drowned; “You're welcome!” |
| `gpt4_31ff4165` | multi-session | READER_MISS | rank 1 | gold `4` devices; answer is a question |
| `bc8a6e93` | single-session-user | EVIDENCE_COVERAGE_MISS | no | gold lemon blueberry cake; bake-minutes listicle |
| `d01c6aa8` | temporal-reasoning | EVIDENCE_COVERAGE_MISS | no | gold `27`; resume-template episode |
| `gpt4_15e38248` | multi-session | EVIDENCE_COVERAGE_MISS | no | gold `4` furniture acts; shop copy |
| `61f8c8f8` | multi-session | EVIDENCE_COVERAGE_MISS | no | gold `10 minutes`; template `* Time:` |
| `0ddfec37_abs` | knowledge-update | ABSTENTION_MISS | no | should say insufficient; did not abstain |
| `gpt4_e414231f` | temporal-reasoning | EVIDENCE_COVERAGE_MISS | no | gold `road bike`; answer mountain-bike aside |
| `b9cfe692` | temporal-reasoning | READER_MISS | rank 21 | gold `5.5 weeks`; drowned |
| `7024f17c` | multi-session | READER_MISS | rank 20 | gold `0.5 hours`; ID-proof listicle |
| `gpt4_70e84552_abs` | temporal-reasoning | ABSTENTION_MISS | no | should abstain; Hashicorp copy |
| `bcbe585f` | temporal-reasoning | READER_MISS | rank 5 | gold `4` weeks ago; answer is rock-climbing |
| `c960da58` | single-session-user | READER_MISS | rank 3 | gold `20` playlists; birding copy |
| `8c18457d` | temporal-reasoning | EVIDENCE_COVERAGE_MISS | no | gold `7 days`; wake-up-time listicle |
| `59524333` | knowledge-update | EVIDENCE_COVERAGE_MISS | no | gold `6:00 pm`; 6 hits, gym days not clock time |
| `6ae235be` | single-session-assistant | EVIDENCE_COVERAGE_MISS | related FCC in rank-1 | gold refining processes; answer is family-arrangement copy |

## Decision gate (blocks architecture order)

Mixed temporal + packet-narrowing + episode pollution. Default Wave 1 order from this histogram:

1. **PR5 first** — `bindPacketFromHopResults` replaces `pkt.Items` / `Contents` / `MemoryIDs`. Hybrid reader then only sees hop-narrowed / first-statement text. Split **ContextEvidence** (broad search, token-budgeted) vs **ProofChain** (hops + `hop_join_proven`).
2. **PR3** — temporal-reasoning 0/6 and knowledge-update 0/3. `Recall` currently searches **before** intents, so `IncludeHistorical` is not set for ago/before/when unless the client sends it. Auto-supersede still hides prior state from default search. Reuse mig-16; add `temporal_score` (not recency +0.05).
3. **PR4** — LME subjects have ~240 extracts; ranking + evidence-set at a **fixed** context-token budget. Untyped episodes currently get **+0.1 `episode_boost`**, which makes assistant pollution worse. Wire `MaxEvidenceTokens`. Qualify candidate pools 30/50/100/200 at one budget. Keep `CandidateOverfetch` cap.
4. **PR9 indicated (inverted)** — assistant turns are stored as `conversation_episode` and drown user facts. Evidence plane still captures `actor_role`. Do **not** persist assistant boilerplate as recall-primary episodes; still extract assistant-stated **facts**.
5. **PR6–PR8 deferred** — LME single-session fails without missing edges. LoCoMo 3×90 MH 22.2% remains a later gate after Wave 1 pins.

Kill list unchanged: no fusion-constant fishing, no graph DB, no category dictionaries, no unbounded top-k, no LoCoMo/LME-named product rules, no SOTA language.

## LoCoMo 1×30 remasure

Run `locomo-pr2-dev-1x30-20260813` against the merged PR2 local stack (commit `24be5ab`), dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, `BRAINY_USE_RECALL=1`. Pin honestly when the json lands (even if it dips vs post-cutover 15/30). Do not blend with Gate 0 18/30.
