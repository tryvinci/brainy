# External review — representation-path additions

**Date:** 2026-08-14  
**Source:** external reviewer (POV additions to “Path to a competitive conversational memory system”)  
**Adjudicator:** coding agent on `pr/fact-primary-recall-a6c7`  
**Prompt used:** reviewer comments on [sota-representation-path.md](../sota-representation-path.md) (first draft)

## Verdict (1 paragraph)

**Keep** the representation-first thesis; **adjust** sequencing and R1c. Wave 1 remains infrastructure, not the SOTA bet. The next bet is still compile-then-retrieve, not reader-first and not retrieval-tuning-first. The first draft’s default Search rule — drop episodes when any non-episode candidate exists — is too aggressive before the atomic compiler proves coverage. Replace it with fact-priority plus bounded episode fallback on incomplete representation. Land a representation coverage oracle (R0) before interpreting another LoCoMo category regression. Treat R1+R1b as one milestone: an interaction must compile into useful semantic memory before the transcript loses recall priority.

## Answers to prompt questions

1. **Course:** Keep representation-first; amend R1 from “drop episodes” to “facts first, episodes on insufficiency.”
2. **Primary gap (e.g. MH):** WRITE_MISS / thin compiler and missing entity-relation substrate — not reader-only, and not “episodes still in the index” as a complete diagnosis.
3. **Next 3–7 PRs:** R0 oracle → R1a primitive semantics → R1b atomic compiler → R1c coverage-gated fact-primary recall → R2 entities → R3 relation projection → R4 ID joins → R5 structured-first answer → R6 freeze/qualify.
4. **Claims allowed vs forbidden:** unchanged honesty rules; 1×30 diagnostic only; representation audit is a merge gate before bench score.
5. **Kill list confirmation:** plus no hard episode-suppression before R1b coverage; no parallel fact/graph extractors; no “assistant is not memory.”

## Findings table

| Finding | Accept / Modify / Reject | Code evidence | Action |
| --- | --- | --- | --- |
| Thesis: representation-first, not reader/retrieval-tuning-first | **Accept** | Wave 1 14/30 MH 3/10; oracles “supported” on chat turns | Keep as program header |
| R1 hard-drop when any fact exists is too aggressive | **Accept** | `dropProvenanceEpisodes` deleted all episodes if `primary > 0` | Coverage-gated fallback; trace `representation_status` |
| R1 + R1b are one milestone (semantic coverage, not fact count) | **Accept** | `primitive != episode` is insufficient | Document; do not declare R1 done on tagging alone |
| Representation coverage oracle / stage taxonomy | **Accept** | `stage_oracle.py` maps leftover misses to `READER_MISS`; semantic oracle counted any memory including episodes | R0: facts-only representation probe; gold-aware WRITE_MISS vs READER_MISS |
| Atomic facts need structure (subject/predicate/value/temporal/evidence) | **Accept** | Extract still sentence-shaped | R1b schema; share compiler with R3 |
| Relations are a projection of entity-valued facts | **Accept** | No relation table today | R3 = projection, not a second extractor |
| Durable assistant facts; skip only phatic | **Accept** | PR9 `isPhaticAssistantText`; `assistant_stated` already exists | Keep PR9; compiler must capture actions/commitments/decisions |
| Canonical entities = IDs, aliases, ranked resolution | **Accept** | Hop V2 used first linked memory ID | R2 |
| Keep Wave 1 temporal work; score dated facts | **Accept** | `temporal_score` on transcripts; local temporal 9/16 | R1b/R3 inputs; do not rip out PR3 |
| Hop proof = entity-ID join | **Accept** | Earlier `hop_join_proven` too permissive | R4 invariant |
| R5 structured-first, not structured-only | **Accept** | Reader still fed dialogue | Packet = structured values + provenance snippets |
| Facts/entities/relations recall-primary; episodes evidence-primary | **Accept** | First draft said “drop episodes” | Language + Search behavior |
| Representation health merge gate before LoCoMo | **Accept** | 1×30 used as informal gate | Audit protocol in R0/R6 |
| OSS vs Platform split (Mem0, Graphiti vs Zep) | **Accept** | Competitive stubs already hinted | Make explicit in competitive SOP |
| Do not ship R1c hard suppression before R1b coverage | **Accept** | This was the first-draft Search behavior | Sequencing + code change |

## Accepted next sequence

1. **R0** WRITE_MISS/READER_MISS taxonomy + representation probe — eval/trace artifact  
2. **R1a** stop tagging facts/atoms as episodes — unit tests  
3. **R1b** atomic compiler + held-out coverage audit — audit report (merge gate)  
4. **R1c** fact-primary + episode fallback on incomplete coverage — Search trace fields  
5. **R2** canonical entities — Postgres identity + ranked resolution  
6. **R3** relation projection from entity-valued facts  
7. **R4** `hop[i].output_entity_id == hop[i+1].input_entity_id`  
8. **R5** structured-first answer with provenance snippets  
9. **R6** 1×30 diagnostic → 3×90 → LME-20 quality (scores after representation gates)

## Rejected / deferred

- Unconditional episode drop as default Search  
- Declaring the representation milestone complete because `primitive != episode`  
- Separate relation extractor in parallel with the fact compiler (unless later measurement requires it)  
- Treating 1×30 as qualification  
- Graph DB, fusion fishing, category dictionaries, LoCoMo-named product rules  

## Claims discipline check

- [x] No invented LME / SOTA scores
- [x] Gate 0 vs harden pins not blended
- [x] MH% cited with the matching pin (1×30 vs 3×90)
- [x] Dip honesty preserved where scores fell

## Artifact diffs required

- [x] assessment pack
- [x] codebase-graph md/json
- [x] program-execution-status
- [x] PoR (`sota-representation-path.md`)
- [x] external-reviews/README.md priority pointer

## Linked PRs / commits

- GitHub PR targeting `dev` on `pr/fact-primary-recall-a6c7`
- ENG-168
