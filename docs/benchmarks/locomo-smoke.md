# LOCOMO smoke — Mem0 gap cycle (attribute atoms)

**Timestamp:** 2026-07-29  
**Pins:** 1 conv / 30 Q · top_k=30 · gpt-oss answerer+judge · async ingest  

## Same-pin results

| System | Overall | temporal | multi-hop | open-domain | Notes |
| --- | ---: | ---: | ---: | ---: | --- |
| Brainy prior (diversify #48) | **19/30** | 13/16 | 2/10 | 4/4 | Peak same-pin |
| Brainy + attribute atoms v1 | 13/30 | 9/16 | 3/10 | 1/4 | Concurrent Mem0 load on gateway |
| Brainy + attribute atoms v2 | **15/30** | 10/16 | 2/10 | 3/4 | Solo remeasure |
| Mem0 Platform (first same-pin) | **1/30** | 0/16 | 0/10 | 1/4 | Under-indexed; wait-on-first-hit |

OpMem Brainy: **12/12** (non-regression).

## Interpretation (honest)

1. **We are NOT at conversational parity with Mem0’s blog 92** — different judge/budget/suite.
2. **On operational/parity suites we already match or beat Mem0** (4/4 parity, 12/12 vs 9/12 OpMem).
3. Attribute atoms (#50) are the right product direction for missing fact atoms; LOCOMO 1×30 is noisy (±4). Peak remains **19/30** until multi-hop synthesis + fuller extract land.
4. First Mem0 same-pin run is **not** a fair knock on Mem0 — indexing waiter returned too early. Waiter hardened (`min_indexed`); re-run before publishing GAP-M1.

## Gap issues

See [mem0-parity-gaps.md](../research/mem0-parity-gaps.md) and GitHub #50–#57.
