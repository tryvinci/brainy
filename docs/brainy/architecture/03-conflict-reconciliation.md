# Conflict Reconciliation Protocol

## Principle
Default to complement-first reconciliation before contradiction-first elimination.

## Protocol
1. Detect conflict between active beliefs.
2. Attempt complement decomposition: identify non-overlapping contexts where both beliefs hold.
3. If unresolved, create experiment plan with explicit evaluation metric and success criteria.
4. Execute experiment and record outcomes.
5. Update beliefs:
- preserve both with scoped applicability, or
- downgrade one belief, or
- retire one belief.

## Operational Rules
- No silent conflict deletion.
- Every resolution emits an audit event.
- Unresolved conflicts remain query-visible.
