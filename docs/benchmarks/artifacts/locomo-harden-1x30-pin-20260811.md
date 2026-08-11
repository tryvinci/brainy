# LoCoMo 1×30 — V3 hardening local pin — 2026-08-11

**Branch:** `pr/v3-hardening-qualify-a6c7` (PRs #93–#97)  
**Host:** local API+worker (subject-ordered claims, hop join coverage, truthful hybrid status)  
**Flags:** async, `BRAINY_USE_RECALL=1`, top_k=30

## Scores

| Metric | Value |
| --- | ---: |
| Overall | **14/30 (46.7%)** |
| multi-hop | 5/10 |
| temporal | 7/16 |
| open-domain | 2/4 |

## vs Gate 0 / V3 early (no tuning)

| Pin | Overall | MH | OD |
| --- | ---: | ---: | ---: |
| Local V3 early | 16/30 | 5/10 | 1/4 |
| Staging Gate 0 | 18/30 | 5/10 | 1/4 |
| **Harden local** | **14/30** | 5/10 | 2/4 |

Honest: overall dipped vs Gate 0 / V3 early. Expected risk from stricter `hop_join_proven` MH coverage (lexical bridge no longer counts as proven). Not a SOTA claim; MH+OD remain open.

Safe report stub only — see run JSON locally under `docs/benchmarks/runs/` (gitignored).
