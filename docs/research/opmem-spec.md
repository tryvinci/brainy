# OpMem: Operational Memory Correctness Benchmark (v0)

**Status:** Draft harness (paper topic 2 from [paper-topics.md](./paper-topics.md))
**Artifacts:** `fixtures/opmem/`, `evals/run_opmem.py`, `evals/opmem_adapters.py`

## Motivation

Existing agent-memory benchmarks (LoCoMo, LongMemEval, HaluMem, LMEB) measure
*retrieval quality*: does the system recall the right fact? None measure whether a
memory system *behaves correctly as a system*:

- Does a forgotten (suppressed) memory ever leak back into retrieval?
- Does a correction stick, even when the original content is re-ingested?
- Is tenant and subject isolation airtight, including for delete operations?
- Does the latest version of a changed fact outrank the stale one?
- Is ingestion idempotent, or do duplicates pollute recall?

These are the failure modes that matter in production and in regulated domains.
OpMem makes each one a small, deterministic, pass/fail task.

## Design

### Neutral operation set

Tasks are scripts over four operations that every memory system exposes in some form:

| Op | Meaning | Brainy mapping | Verbatim baseline | Mem0 mapping |
| --- | --- | --- | --- | --- |
| `remember` | ingest content for an actor | `POST /ingest` | append entry | `POST /v1/memories/` |
| `recall` | ranked retrieval for an actor | `GET /memories/search` | token-overlap scan | `POST /v2/memories/search/` |
| `revise` | correct a prior memory | `POST /memories/{id}/correct` | replace content | `PUT /v1/memories/{id}/` |
| `forget` | remove a prior memory | `POST /memories/{id}/suppress` | delete entry | `DELETE /v1/memories/{id}/` |

An *actor* is a `(tenant, subject)` pair, so isolation is testable. `revise` and
`forget` reference the step id of an earlier `remember`; adapters track which
system-side memory ids each step produced.

### Task schema

```json
{
  "name": "sup01_basic_forget",
  "category": "suppression",
  "description": "A forgotten memory must not appear in later recall.",
  "actors": { "a1": { "tenant": "t1", "subject": "u1" } },
  "steps": [
    { "op": "remember", "id": "m1", "actor": "a1", "content": "..." },
    { "op": "recall", "actor": "a1", "query": "...", "expect": { "min_results": 1 } },
    { "op": "forget", "actor": "a1", "target": "m1" },
    { "op": "recall", "actor": "a1", "query": "...", "expect": { "max_results": 0 } }
  ]
}
```

`actors` is optional; an unlisted actor name defaults to tenant `t1`, subject =
actor name. Adapters namespace tenants per run so tasks are hermetic on shared
backends.

Recall assertions (all content matching is case-insensitive substring):

| Key | Meaning |
| --- | --- |
| `min_results` / `max_results` | result-count bounds |
| `top_contains` | first result must contain substring |
| `top_not_contains` | first result (if any) must not contain substring |
| `must_include` | each substring appears in at least one result |
| `must_exclude` | no result contains any of these substrings |
| `unique_contents` | no two results share identical normalized content |

### Task taxonomy (v0: 12 tasks)

| Category | Tasks | Probes |
| --- | --- | --- |
| suppression | `sup01` basic forget, `sup02` targeted forget, `sup03` durable forget | forget leaks; collateral damage; resurrection via re-ingestion |
| correction | `cor01` basic revision, `cor02` correction stickiness, `cor03` revised retrievable | revision visibility; stale content winning after re-ingestion; retrieval by new terms |
| isolation | `iso01` subject, `iso02` tenant, `iso03` forget isolation | cross-actor reads; cross-tenant reads; delete affecting a lookalike memory of another actor |
| staleness | `upd01` stale fact, `upd02` preference change | latest version of a changed fact/preference must outrank the stale one |
| idempotency | `dup01` idempotent remember | duplicate ingestion polluting recall |

Contents are deliberately domain-neutral (door codes, wifi names, launch dates,
language preferences) — no vertical pack semantics are required. The phrasing is
chosen so deterministic extractors classify them (never-constraints, `X is Y`
facts, `I prefer` preferences, `My name is` profile statements).

### Declared semantic contracts

Two tasks encode policy choices that the spec makes explicit rather than leaving
implicit:

1. **Durable forget (`sup03`):** an explicitly forgotten memory must not silently
   resurrect when the same content is re-ingested. Rationale: `forget` is an
   operator/governance action; passive re-observation should not override it.
   (A system may legitimately want an explicit "reinstate" path — but not a
   silent one.)
2. **Staleness (`upd01`/`upd02`):** when the same fact or preference is restated
   with a new value, recall must rank the newer value first. No explicit
   supersede call is made — this probes the system's own conflict handling.

These are the tasks most likely to discriminate between systems, including ours.
A system failing them is a *finding*, not a harness bug.

## Scoring and reporting

Each task is pass/fail per system; a task with an infrastructure error (HTTP
failure, adapter exception) is reported as `error`, separately from `fail`.
The report aggregates per category: `passed/total`. There is no partial credit —
operational correctness is binary by design.

The runner exits non-zero only on infrastructure errors, so it can run in CI as
a harness-health check while task results remain diagnostic data.

## Usage

```bash
# offline: runner + fixtures against the in-process verbatim baseline
python3 evals/run_opmem.py --systems verbatim

# against a live Brainy API
python3 evals/run_opmem.py --systems verbatim,brainy --base-url http://127.0.0.1:8080

# include Mem0 (requires MEM0_API_KEY)
MEM0_API_KEY=... python3 evals/run_opmem.py --systems verbatim,brainy,mem0

# write the JSON report
python3 evals/run_opmem.py --systems verbatim,brainy --json-out docs/benchmarks/opmem-latest.json
```

CI runs the harness end-to-end via `TestOpMemBenchmarkAgainstHTTPServer`
(`go test ./internal/api/`).

## Roadmap to the paper

- v0 (this): 12 tasks, 3 adapters (Brainy, verbatim baseline, Mem0), binary scoring.
- v1: add adapters for Zep, Letta, LangMem; expand to ~30 tasks including
  lifecycle/expiry (needs a neutral expiry attribute), batched/async ingestion
  visibility, and concurrent-write races.
- Paper: results matrix across 5–6 systems, failure taxonomy, and the two
  semantic contracts above argued as normative requirements for production
  memory systems.
