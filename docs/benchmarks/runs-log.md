# Benchmark runs log (holdout / phase gates)

Per master-plan §W1 holdout policy: tuning = LOCOMO convs 1–3; validation = 4–10 at most once per phase gate.

| Date | Commit | Purpose | Notes |
| --- | --- | --- | --- |
| 2026-07-29 | pre-W1 peak | History | LOCOMO 1×30 peak **19/30** MH **5/10** with LOCOMO-shaped atoms (see locomo-smoke.md) |
| 2026-07-29 | P0 de-overfit | Re-baseline E1 | After removing LOCOMO surface-forms from product; expect smoke drop |
| 2026-07-29 | `ece8d52` | P0 re-baseline recorded | LOCOMO 1×30 **16/30** (MH 4/10, temp 9/16, open 3/4); OpMem **12/12**. Pre-W1 peak was 19/30 MH 5/10 with hacks. |
| 2026-07-31 | `57e3dbc` | P4 full LoCoMo dry run (seed 0) started | `run_full.py --conversations 10 --questions 0 --top-k 50`; staging worker SIGTERM~60s workaround = local drain worker on external DB |
| 2026-07-31 | `57e3dbc` | W6 latency load | `latency_load.py` c=8 → p50 **2403**/p95 **4997** ms (SLO miss under load); artifact `latency-load-20260731T065251Z.json` |
| 2026-07-31 | `b5552d3` | P4 full LoCoMo seed 0 **complete** | **49.4% (761/1540)** cats 1–4; MH 25.2% (71/282), temp 54.8%, open 38.5%, single 56.7%; search p50/p95 2017/3447 ms. Artifact `docs/benchmarks/artifacts/locomo-full-publish-s0-2a6a04.md`. Seeds 1–2 started. |

Reproduce re-baseline:

```bash
python -m public.locomo.run_smoke --conversations 1 --questions 30 --top-k 30 \
  --answerer-model "$LLM_MODEL" --judge-model "$LLM_MODEL" \
  --run-id locomo-p0-deoverfit-baseline
```
