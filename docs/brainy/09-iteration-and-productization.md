# Iteration and Productization Framework

## Guarded Rollout Criteria

A release candidate can advance only if all checks pass:
- Public track score is non-decreasing versus last baseline.
- Cognitive track challenge-detection score >= 0.9.
- No benchmark regression above thresholds in `docs/brainy/benchmarks/regression-thresholds.yaml`.
- Audit log integrity and rollback controls pass governance tests.

## Regression Thresholds
- Public-track average score drop > 0.05 is blocking.
- Cognitive-track score drop > 0.02 is blocking.
- Retrieval latency increase > 20 percent is warning; > 35 percent is blocking.

## Failure-to-Fix Loop
1. Convert failed benchmark to a tracked hypothesis update.
2. Implement minimal architecture fix.
3. Re-run benchmark suite with archived config.
4. Record final disposition in change log.

## Release Cadence
- Weekly internal quality checkpoints.
- Bi-weekly architecture review and hypothesis pruning.
- Quarterly moat review against competitor movement.
