# Path to a competitive conversational memory system (2026-08-14)

**Status:** course correction — execute this, do not treat Wave 1 as the SOTA bet  
**Tips:** `main` = `dev` = `bd987fa` (Wave 1 + pins on production after explicit merge)  
**Does not claim:** SOTA, beats-Mem0, or a LoCoMo/LME target score

Wave 1 (budgets, `temporal_score`, hop packet split, skip phatic assistant turns) was **retrieval/ranking efficiency around a transcript index**. That is not how Mem0, Graphiti/Zep, or the LoCoMo/LME papers win conversational QA. This note replaces “reader next, defer PR6–PR8” with the representation path those systems actually use.

## Direct answers

**Will reading papers get us there?** Papers are necessary **inputs**, not the work. The relevant ones are already known:

| Source | What to take | What not to take |
| --- | --- | --- |
| [Mem0 paper (arXiv:2504.19413)](https://arxiv.org/abs/2504.19413) + [2026 algorithm](https://mem0.ai/blog/state-of-ai-agent-memory-2026) | Extract **atomic facts** from dialogue (ADD-only); retrieve **those facts**; entity linking as a retrieval signal; temporal ranking over dated facts | Managed-platform 92.5 LoCoMo as a comparable pin; Neo4j; bench-specific prompts |
| Graphiti / Zep (ENG-69) | **Episode = provenance**. **Entity + edge = retrieval unit**. Validity windows on facts | Graph DB as required substrate (ADR-004) |
| LoCoMo / LongMemEval papers | What MH/temporal questions *require* (attribute join, list aggregation, current vs past) | Product rules named after the bench |
| HippoRAG / A-MEM / MemGPT (ENG-71, still backlog) | Later: entity-centric walk, memory evolution, working-vs-archival blocks | Do not start here while the index is still chat turns |

We do **not** need another survey before writing code. ENG-54/ENG-71 should annotate this path, not block it. The July gap doc already named the conversational hole: ingest does not emit **atomic attribute facts** ([mem0-parity-gaps.md](./mem0-parity-gaps.md) GAP-C1). Wave 1 did not close that.

**Is the work “diff our system vs popular systems”?** Yes — and we already did the archaeology. The **gaping** gap is not BM25 weights or candidate 50 vs 100.

| System | Retrieval unit | Brainy today |
| --- | --- | --- |
| Mem0 | Extracted fact sentences + entity collection | Mix of **conversation_episode transcripts** + some facts; episodes are recall-primary |
| Graphiti | Entities + relation edges; episodes stored separately | `memory_entity_links` hub only; **no relation table**; hops = first linked memory ID |
| Brainy (ops/vertical) | Governed records, lifecycle, packs | **Already ahead** (OpMem 13/13, marketing 17/17) — keep |

Wave 1 ledgers saying `READER_MISS` with “coverage supported” were **misleading**. Oracle “supported” means the gold **substring appears in a retrieved chat turn**. Mem0 never asks the reader to parse “Yeah, Caroline, Yep…” because Sweden / single / pottery are **stored as facts**. Treating that as a reader problem is how we slid into efficiency PRs.

**What it will take (architecture, not calendar):**

```text
transcript (immutable evidence)
  → atomic facts + preferences + dated events     [Mem0 ADD]
  → canonical entities + aliases                  [PR6]
  → relation edges with validity windows          [PR7 / Graphiti]
  → search facts/edges by default                 [this PR]
  → hop along relations for MH                    [PR8]
  → answer from structured values, not chat dump
  → keep OpMem / vertical governance as the exceed axis
```

Credibility floor remains measured LoCoMo 3×90 + LME-20 **quality** under frozen pins — not 1×30 smoke and not “75% promised.”

## Why Wave 1 felt like efficiency

| Shipped | What it actually is |
| --- | --- |
| PR4 MaxEvidenceTokens / pool 30–200 / episode −0.15 | Ranking around the same episode corpus |
| PR3 temporal_score + IncludeHistorical | Ranking signal; did move local temporal 1/16 → 9/16 **with hybrid reader on** |
| PR5 ContextEvidence vs ProofChain | Packet layout for a reader still fed **dialogue** |
| PR9 skip phatic assistant episodes | Filter; user turns still become recall-primary episodes |
| Deferred PR6–PR8 | **Wrong deferral.** MH 3/10 vs Mem0 7/10 is missing **facts/entities/edges**, not a graph DB |

Local Wave 1 LoCoMo **14/30** (MH **3/10**) is not an improvement vs Gate 0 18/30. Attribute atoms were even tagged `primitive=episode`, so they took the episode penalty. That is a representation bug, not a reader bug.

## Execution sequence (this is the program now)

### R1 — Fact-primary recall (this PR)

Mem0/Graphiti: search **memories**, not **utterances**.

- Keep `conversation_episode` in the store as **provenance**.
- Default `Search` / `/recall` **drops episodes when any non-episode candidate exists**.
- If the only hits are episodes, keep them and mark `episode_fallback` on the trace (honest empty-representation signal).
- Stop tagging attribute atoms and dated provider facts as `episode`.
- Enumerate skips episode values so list answers are not “Yeah, Caroline”.

**Exit:** unit tests + OpMem/marketing non-reg. LoCoMo 1×30 remasure is a **measurement**, not a merge gate. If MH stays READER_MISS on remaining chat, extract is still thin → R1b. If coverage oracles flip to miss, that is the real representation hole → R1b/R2.

### R1b — Extract atomic facts (GAP-C1)

Provider prompt already asks for one attribute per memory. Prove it on a **held-out** LoCoMo conversation (not tuning conv-26). Inspect stored rows: count facts vs episodes per subject. If provider facts lack identity/origin/activity, fix extract schema / merge, not fusion.

Inspect Mem0 OSS `mem0/configs/prompts.py` (ADD-only fact sentences) as **ADAPT**, not a verbatim copy.

### R2 — Canonical entities (program PR6)

Postgres entity + alias table. Link facts to entities. Query entity matching boosts fact retrieval (Mem0 `{collection}_entities`, Graphiti nodes). **REJECT** Neo4j.

### R3 — Relation memory (program PR7)

`(subject_entity, predicate, object_entity|value, valid_from, valid_to, evidence_ids)`. This is what LoCoMo MH actually is: *Melanie — activity — pottery*, *Caroline — origin — Sweden*. Graphiti edges, not hop heuristics.

### R4 — Relation hops (program PR8)

`hop[i].output == hop[i+1].input` over R3 edges. Only after R3 exists.

### R5 — Answer from structured values

Enumerate/answer over fact `value` / edge objects. Hybrid reader stays bound to the fact packet. Still no LoCoMo-named product rules.

### Measure (PR10, still no claim)

Same-pin Brainy vs Mem0 LoCoMo 1×30 then 3×90; LME-20 **quality** only after representation pins move off 0/20 for a reason other than “reader continued the chat.” OpMem/marketing must stay green.

## Kill list (unchanged)

No fusion-constant fishing, no graph DB default, no category dictionaries, no unbounded top-k, no LoCoMo/LME-named product rules, no SOTA / beats-Mem0 language, no treating 1×30 as the qualification.

## Linear

- ENG-168 conversational long-memory epic — this path is the engineering response  
- ENG-176 multi-hop synthesis — after R1–R3, not instead of them  
- ENG-69 Graphiti temporal fact model — input to R3  
- ENG-71 academic survey — annotate, do not block R1  
- ENG-60 graph layer — **Postgres graph-shaped** (ADR-004), not “build Neo4j”
