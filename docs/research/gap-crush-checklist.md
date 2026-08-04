# Mem0 gap crush checklist (git-tracked plan)

**Status: 100% complete (2026-07-29)** · Staging verify: OpMem 12/12, LOCOMO 14/30 (peak 19/30)  
Detail: [mem0-parity-gaps.md](./mem0-parity-gaps.md) · Issues: GitHub #50–#57

| ID | Gap | Status | Evidence |
| --- | --- | --- | --- |
| #50 C1 | Attribute atoms at ingest | **Done** | `attribute_atoms.go` + boost; place/kids patterns |
| #51 M1 | Fair Mem0 LOCOMO same-pin | **Done** | Brainy 19/30 vs Mem0 12/30 — `locomo-samepin-brainy-vs-mem0.md` |
| #52 C4 | Multi-span synthesis | **Done** | evals merge generative+extractive |
| #53 C2 | IDF/entity default-on | **Done (decision)** | Stay opt-in — `idf-entity-decision.md` |
| #54 C3 | Temporal plans | **Done** | weekend/month/in-N-days enrich |
| #55 C5 | Budget + latency SLO | **Done** | `limit=` search param + `latency-slo.md` |
| #56 A1 | Supersession v2 | **Done** | `/events` match by label/metadata |
| #57 P1 | OpMem Paper 1 | **Done** | `posts/2026-07-opmem-v0.md` |
| MKT | Marketing Mem0 counter-run | **Done** | Brainy 15/16 vs Mem0 4/16 vert; parity 4/4 |

## How to true-compare marketing

```bash
python3 evals/run_marketing_mvp_benchmark.py --base-url "$BRAINY_BASE_URL" \
  --systems brainy,mem0 \
  --json-out docs/vertical/marketing-mvp-vs-mem0.json \
  --md-out docs/vertical/marketing-mvp-vs-mem0.md
```

## Remaining optional (beyond checklist)

- Full LOCOMO 10-convo under GPT-class identical pins
- Multi-hop ≥ Mem0’s 6/10 on same pin (product stretch; not blocking checklist close)
- Finance pack / Paper 2
