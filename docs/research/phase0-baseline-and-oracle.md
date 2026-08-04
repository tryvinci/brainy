# Oracle modes for stage-level diagnosis (program §16.3 / Phase 0)

Brainy's `POST /recall` accepts `oracle_mode` for diagnostic runs. Product
defaults leave it empty.

| Mode | Behavior |
| --- | --- |
| *(empty)* | Normal retrieval + synthesis |
| `evidence` | Lists raw `memory_evidence` for the subject (source plane) |
| `semantic` | Confirms memories (and typed atoms when present) exist for the subject |
| `retrieval` | Confirms `SearchOpt` returns ≥1 active hit |
| `coverage` | Confirms enumerate/evidence-set yields ≥1 item |
| `reader` | Runs product `mode=answer` path tagged as reader oracle |

Unknown modes return `oracle_unsupported` (honest no-op — no silent product path).

Harness helpers:
- `evals/public/oracle.py` — failure taxonomy + `recall_body`
- `evals/public/stage_oracle.py` — ledger writer + multi-stage `probe_failure_stages`
- LoCoMo smoke: `--failure-ledger` appends JSONL for WRONG/SKIPPED with stage labels

Failure labels use the taxonomy in `internal/memory/trace.go`
(`SOURCE_MISS`, `RETRIEVAL_MISS`, …).

Baseline freeze (2026-08-01 internal):

| Axis | Signal |
| --- | ---: |
| LoCoMo 3-seed | ~49.8% |
| LoCoMo multi-hop | ~26% |
| LongMemEval-S 100 | ~4% |
| BEAM-100K / 20q | ~40% |
| OpMem | 13/13 |
| Marketing | 15/16 |
| Support | 3/3 |
| Search c=8 p50/p95 | ~2.4s / ~5.0s |

Reproduce commands: see `docs/research/master-plan-execution-status.md`.
