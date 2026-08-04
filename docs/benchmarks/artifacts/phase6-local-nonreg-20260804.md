# Phase 6 local non-regression (post Fusion V2 / evidence plane)

- Date: 2026-08-04T00:18:19Z
- Commit: 631a2934069ee3a779f6273798d154d742183fa4
- Branch: merged to `dev` as 4423109 (PR #77)

## Commands
```
go test ./internal/memory/ ./internal/jobs/ ./internal/pack/ ./internal/api/ ./internal/store/postgres/ ./internal/config/ -count=1
```

## Results
- OpMem HTTP harness: PASS
- Marketing MVP harness: PASS (incl. bv02 suppress)
- Vertical marketing suite: PASS (incl. mk_v2_01)
- Support fixtures loaded via pack eval_fixtures path when run
- Unit: fusion_v2 / predicate_policy / evidence_set / intents: PASS

## Not re-run here (ops/staging)
- Full LoCoMo 3-seed (baseline freeze ~49.8%)
- LongMemEval-S 100
- Latency load @ c=8
