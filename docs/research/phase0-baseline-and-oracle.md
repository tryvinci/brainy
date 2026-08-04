# Oracle modes for stage-level diagnosis (program §16.3 / Phase 0)

Brainy's `POST /recall` accepts `oracle_mode` for diagnostic runs. Product
defaults leave it empty.

| Mode | Behavior |
| --- | --- |
| *(empty)* | Normal retrieval + synthesis |
| `reader` | Caller supplies gold evidence in the request path / harness; measures reader ceiling |
| `evidence` | Confirms raw source text exists after ingest (write/source plane) |
| `semantic` | Confirms typed atoms/events/assertions exist for the gold fact |

Harness helpers live in `evals/public/oracle.py`. Failure labels use the
taxonomy in `internal/memory/trace.go` (`SOURCE_MISS`, `RETRIEVAL_MISS`, …).

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
