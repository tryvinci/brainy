# LOCOMO smoke — `locomo-smoke-pr22`

**Timestamp:** 2026-07-19T11:20:24Z  
**Brainy:** `http://127.0.0.1:8080` (commit `b61f815d1a904dab4d6164dd371b43815af8e6de`)  
**Dataset:** [https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json](https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json)  
**SHA256:** `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`  
**Answerer:** `gpt-oss-120b` @ Cloudflare AI Gateway (Workers AI)  
**Judge:** `gpt-oss-120b` @ Cloudflare AI Gateway (Workers AI, temp=0.0)  
**Schema:** mem0ai/memory-benchmarks@UnifiedResult-1.0

## Scores (categories 1–4; adversarial excluded)

| Metric | Value |
| --- | ---: |
| Overall | 0.133 (4/30) |
| Search p50 ms | 18.0 |
| Search p95 ms | 20.6 |

| Category | Acc | n |
| --- | ---: | ---: |
| multi-hop | 0.000 | 10 |
| open-domain | 0.500 | 4 |
| temporal | 0.125 | 16 |

## Proveability

Pins present (dataset SHA, judge model, brainy URL/commit).

## Notes (same-pin remeasure)

| Run | Commit | Overall | Retrieval-miss (of 30) |
| --- | --- | ---: | ---: |
| Baseline (pre ENG-171) | `e8b3bb8` era | **2/30 (6.7%)** | 28 |
| Post ENG-171 | `1dc738c` | **4/30 (13.3%)** | 25 |
| Post ENG-172/92/173 (PR #22) | `b61f815` | **4/30 (13.3%)** | — |

Pins held fixed: 1 conversation / 30 Q / `gpt-oss-120b` judge+answerer / same dataset SHA.

**Honest read:** PR #22 did **not** move the smoke score (still 4/30). This is expected — the
smoke ingests via the **synchronous `/ingest` (deterministic) path**, so PR #22's headline
change (OpenAI-compatible **provider extraction**) is not exercised here; only the ranking /
`observed_at` changes are, and they did not flip any of these 30 questions. Provider extraction
was validated separately end-to-end (async ingest → worker → memories with date/event slots);
see the PR walkthrough. Remaining smoke failures are still dominated by retrieval-miss of the GT
span, unchanged from the ENG-171 run.

> The lower search latency (p50 18 ms vs 422 ms in the prior run) reflects a **local** brainy
> backend on this run rather than the staging network round-trip — it is not a product change.

## Staging confirmation (provider extraction)

After the staging worker was configured with `BRAINY_PROVIDER_*` (Render dashboard) and
redeployed, async provider extraction was confirmed live on `brainy-api-staging.onrender.com`:
`/healthz` → `ok`; an async ingest with dates produced provider memories — a dated fact stored
as an `episode` (`On 2021-06-01 I joined Globex as VP of Growth`, migration v10 `observed_at`
behavior) and a preference rephrased to third person (`Prefers async standups over live
meetings`, the provider tell vs the deterministic raw turn).

## Environment note (provider extractor fix)

Running provider extraction against `gpt-oss-120b` on the Cloudflare AI Gateway surfaced a
truncation bug: the extractor did not send `max_tokens`, so the gateway's small default (256
completion tokens) truncated the JSON for this reasoning model (`finish_reason: length` →
`unexpected end of JSON input`). Fixed by bounding the completion (`max_tokens`) in
`internal/memory/provider_extractor.go`.

## Outlinks

- Dataset upstream: https://github.com/snap-research/locomo
- Paper: https://aclanthology.org/2024.acl-long.747/
- Harness peer: https://github.com/mem0ai/memory-benchmarks
