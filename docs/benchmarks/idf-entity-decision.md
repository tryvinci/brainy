# GAP-C2 decision: IDF + entity ranking defaults

**Date:** 2026-07-29  
**Decision:** Keep `BRAINY_IDF_RANKING` and `BRAINY_ENTITY_RANKING` **opt-in** (default OFF).

## Evidence

- Entity ranking with dense embeddings regresses LOCOMO smoke (11/30 vs 13/30 entity-off) — `entity-linking-ab.md`.
- IDF-weighted coverage similarly regresses without a full boost-stack re-tune.
- Extraction/persistence of entities remains always-on; only ranking fusion stays gated.

## Completion criteria for GAP-C2

Gap closed as an **engineering decision with evidence**, not by force-enabling a
regressing default. Re-open when a staging A/B on a ≥3-convo pin shows lift on
multi-hop **and** non-regression on OpMem 12/12 + marketing vertical.

## How to experiment

```bash
BRAINY_ENTITY_RANKING=true BRAINY_IDF_RANKING=true
```
