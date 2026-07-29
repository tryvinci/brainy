# Benchmark runs log (holdout / phase gates)

Per master-plan §W1 holdout policy: tuning = LOCOMO convs 1–3; validation = 4–10 at most once per phase gate.

| Date | Commit | Purpose | Notes |
| --- | --- | --- | --- |
| 2026-07-29 | pre-W1 peak | History | LOCOMO 1×30 peak **19/30** MH **5/10** with LOCOMO-shaped atoms (see locomo-smoke.md) |
| 2026-07-29 | P0 de-overfit | Re-baseline E1 | After removing LOCOMO surface-forms from product; expect smoke drop |

Reproduce re-baseline:

```bash
python -m public.locomo.run_smoke --conversations 1 --questions 30 --top-k 30 \
  --answerer-model "$LLM_MODEL" --judge-model "$LLM_MODEL" \
  --run-id locomo-p0-deoverfit-baseline
```
