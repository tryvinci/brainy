# Iteration and Productization Framework

## Guarded Rollout Criteria

A release candidate can advance only if all checks pass:
- Public track score is non-decreasing versus last baseline.
- Cognitive track challenge-detection score >= 0.9.
- No benchmark regression above thresholds in `docs/brainy/benchmarks/regression-thresholds.yaml`.
- Audit log integrity and rollback controls pass governance tests.

## Vertical memory vetting (Go rebuild)

The Python prototype benchmarks above are legacy. The **Go vertical runtime** uses fixture-driven vetting documented in [`docs/vertical/marketing-vetting-gate.md`](../vertical/marketing-vetting-gate.md).

| Gate | Requirement | Blocks |
| --- | --- | --- |
| **M1** | Tiers 0–3 green (`go test ./...`, marketing MVP benchmark) | Treating ENG-93 as “done product” |
| **M2** | M1 + published `dev` / CI on origin | External reproducibility |
| **M3** | M2 + all marketing use-case eval seeds + semantic non-regression | Finance implementation, second vertical packs |
| **M4** | M3 + architecture sign-off | Finance pack merge to main |

**Policy:** No finance pack (`ENG-56`, `ENG-76`) or second vertical until **Gate M3**. Research docs only before that.

## Regression Thresholds
- Public-track average score drop > 0.05 is blocking.
- Cognitive-track score drop > 0.02 is blocking.
- Retrieval latency increase > 20 percent is warning; > 35 percent is blocking.
- Marketing parity + vertical fixture pass rate must remain 100% on PRs to `dev` / `main`.

## Failure-to-Fix Loop
1. Convert failed benchmark to a tracked hypothesis update.
2. Implement minimal architecture fix.
3. Re-run benchmark suite with archived config.
4. Record final disposition in change log.

## Release Cadence
- Weekly internal quality checkpoints.
- Bi-weekly architecture review and hypothesis pruning.
- Quarterly moat review against competitor movement.
