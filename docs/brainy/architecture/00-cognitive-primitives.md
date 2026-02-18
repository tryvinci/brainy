# Cognitive Primitives and State Transitions

## Canonical Primitives

1. `Principle`
- Stable first-principles rule that changes rarely.
- Example: preserve brand core promise consistency.

2. `IdentityPrior`
- Persistent preference and style prior.
- Encodes voice, risk appetite, and aesthetic direction.

3. `Episode`
- Time-bound observed event with context and provenance.
- Input to extraction and belief updates.

4. `Pattern`
- Repeated episode structure abstracted into reusable heuristic.

5. `Belief`
- Ranked hypothesis with conviction score, evidence links, and status.

6. `Outcome`
- Expected-versus-observed result used for calibration.

7. `Experiment`
- Explicit conflict-resolution mechanism when beliefs compete.

8. `TasteSignal`
- Non-functional preference signal influencing ranking and synthesis.

9. `Reflection`
- Meta-process that revises beliefs, conviction, and memory compression.

## State Transitions

- Episode -> Pattern via consolidation.
- Episode + Pattern + Principle + IdentityPrior -> Belief formation.
- Belief -> Challenged when outcomes deviate beyond stop-loss threshold.
- Challenged beliefs -> Revised or Retired after reflection/experiments.
- Reflection -> updates Pattern and TasteSignal strength.
- Principle changes require explicit governance approval path.

## Invariants
- Principles are immutable by default.
- Every belief update requires provenance and reason code.
- Contradictions can coexist until explicit reconciliation.
