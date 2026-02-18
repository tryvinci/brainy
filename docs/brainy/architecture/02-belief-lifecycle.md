# Belief Lifecycle

## States
- `candidate`
- `active`
- `challenged`
- `retired`

## Lifecycle
1. Formation
- Candidate belief is created from episode/pattern synthesis.

2. Ranking
- Beliefs ranked by conviction, evidence quality, recency, and identity alignment.

3. Challenge
- Belief is marked challenged when outcome delta exceeds stop-loss threshold.

4. Revision
- Conviction and rationale updated; supporting and conflicting evidence relinked.

5. Retirement
- Belief retired when sustained underperformance or contradiction persists post-experiment.

## Belief Record Requirements
- Unique ID
- Human-readable claim
- Rank
- Conviction
- Evidence references
- Last evaluated timestamp
- Status and reason code
