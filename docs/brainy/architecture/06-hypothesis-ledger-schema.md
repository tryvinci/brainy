# Hypothesis Ledger Schema

## Purpose
Track every architecture bet with evidence, confidence, risk, and falsifiability.

## Schema

```json
{
  "id": "HYP-001",
  "title": "Short claim",
  "description": "Long-form hypothesis statement",
  "component": "ingestion|consolidation|belief_graph|retrieval|reflection|governance|benchmark",
  "confidence": 0.0,
  "evidence_refs": ["url-or-doc-ref"],
  "risk_level": "low|medium|high",
  "falsification_test": "test/scenario that would prove this wrong",
  "owner": "team-or-person",
  "status": "proposed|active|validated|rejected",
  "created_at": "ISO8601",
  "updated_at": "ISO8601"
}
```

## Rules
- No implementation without a linked active hypothesis.
- Rejected hypotheses remain immutable history.
