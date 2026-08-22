# Mem0 harness audit — 2026-08-22

**Not a same-pin score.** This is a faithfulness check of `evals/public/backends/mem0.py` and `evals/competitors/mem0_adapter.py` against Mem0's current Platform API and their public LoCoMo runner ([mem0ai/memory-benchmarks](https://github.com/mem0ai/memory-benchmarks)).

A Brainy lead over a handicapped Platform integration is not a fair win. Our frozen Mem0 Platform 1×30 pin is **11/30 (36.7%)**. Their published LoCoMo is **92.5%** at top_k=200 on the managed platform. Do not mix those rows.

## What their published protocol actually is

Source: [Memory Evaluation](https://docs.mem0.ai/core-concepts/memory-evaluation) and `benchmarks/common/mem0_client.py` / `benchmarks/locomo/run.py` (fetched 2026-08-22).

| Knob | Published / their runner | Our harness before this audit |
| --- | --- | --- |
| Search API | `POST /v3/memories/search/` | `POST /v2/memories/search/` |
| Add API | `POST /v3/memories/` + poll `/v1/event/{id}/` | `POST /v1/memories/` |
| Default `top_k` | **200** | **10** in the adapter; smoke defaulted **30** unless `--eval-lane industry-search` or `--top-k` |
| Ingest chunk | **1 turn** (`CHUNK_SIZE = 1`) | **8 turns** (shared Brainy batch) |
| Session time | unix `timestamp` on add from LoCoMo `session_*_date_time` | metadata dropped (`_ = metadata`) |
| Rerank | default `false` | omitted (same) |
| Graph | built-in entity boost on Platform v3 score | not requested; v2 may not fuse the same signals |
| Index wait | per-add event `SUCCEEDED` | `min_indexed=40` then first search hit |

Their LoCoMo runner is search → shared answerer → shared judge, single-pass, no agentic loops. That is our **industry-search** lane, not product `POST /recall`.

## Why 11/30 can be an artifact

1. **top_k 30 vs 200.** Same-pin docs froze `--top-k 30`. Their headline is top_200. Industry Brainy already defaults 200 when `--eval-lane industry-search` is set. Mem0 often did not get that default.
2. **v2 search vs v3 hybrid.** Current docs: v3 fuses semantic + BM25 + entity and can apply temporal reasoning. v2 is the old path.
3. **No session timestamps.** LoCoMo temporal gold is session-dated. Their runner stamps `timestamp` on every add. We discarded `observed_at`. That is a plausible cause of Mem0 temporal **2/16** vs published temporal **92.0**.
4. **Batch ingest.** Eight turns per add vs one turn changes extraction. Fair Mem0 ingest should match their CHUNK_SIZE.
5. **`min_indexed=40`.** A LoCoMo conversation writes hundreds of memories. Returning after 40 indexed rows scores an under-built store.

## Fair same-pin recipe (this repo)

Run both systems through `evals/public/locomo/run_smoke.py` / `run_s0.py`:

- Dataset SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`
- Judge temp 0, same answerer/judge models
- **Mem0:** `--system mem0 --top-k 200` (now the Mem0 default when `--top-k` is omitted)
- **Brainy industry:** `--eval-lane industry-search` (top-k 200)
- **Brainy product:** `--eval-lane product-recall` labeled separately (top-k 30 `/recall`) — not the Mem0-protocol row
- Do not compare either row to Mem0 blog 92.5 as the same pin

Adapter changes that implement this recipe (harness only, no product `/recall` behavior):

- Search v3 with v2 fallback on HTTP 404
- Add v3 + event wait, with v1 fallback
- Pass LoCoMo `observed_at` as unix `timestamp`
- Mem0 ingest chunk = 1
- Mem0 default top_k = 200
- Index wait `min_indexed=200`

## Still not claimed

- Org/project IDs (`MEM0_ORGANIZATION_ID`, `MEM0_PROJECT_ID`) — their cloud client sends them when set; we still omit them. If a later 180 pin looks weak, set those env vars and re-run before calling the integration handicapped again.
- Their proprietary extraction model vs our gpt-oss judge/answerer. Same-pin uses **our** judge; that is required by SOP and is not a Mem0-API bug.
- n=1540 / 1×30 remasure. This file is the audit. Scores come after the next frozen run.
