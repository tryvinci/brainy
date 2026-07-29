# Mem0 gap crush checklist (git-tracked plan)

Living plan — no Linear required. Issues: GitHub #50–#57 · Detail: [mem0-parity-gaps.md](./mem0-parity-gaps.md)

| ID | Gap | Status | Evidence |
| --- | --- | --- | --- |
| #50 C1 | Attribute atoms at ingest | **Shipped** | `attribute_atoms.go`; remeasure noisy |
| #51 M1 | Fair Mem0 LOCOMO same-pin | **Measured** | Mem0 12/30 vs Brainy peak 19/30; MH 6/10 vs 2–3/10 |
| #52 C4 | Multi-span synthesis | In progress | evals answerer merge |
| #53 C2 | IDF/entity default-on | Open | gated flags |
| #54 C3 | Temporal plans | Open | |
| #55 C5 | Budget + latency SLO | Open | |
| #56 A1 | Supersession v2 | Open | ENG-86 remainder |
| #57 P1 | OpMem Paper 1 | Open | |
| MKT | Marketing Mem0 counter-run | **Done (empirical)** | Brainy 15/16 vert; Mem0 4/16; parity both 4/4 |

## How to true-compare marketing

```bash
# Declared gaps only (old behavior)
python3 evals/run_marketing_mvp_benchmark.py --base-url "$BRAINY_BASE_URL"

# Empirical Mem0 counter-run on the same fixture JSON
export MEM0_API_KEY=…
python3 evals/run_marketing_mvp_benchmark.py --base-url "$BRAINY_BASE_URL" \
  --systems brainy,mem0 \
  --json-out docs/vertical/marketing-mvp-vs-mem0.json \
  --md-out docs/vertical/marketing-mvp-vs-mem0.md
```

Mem0 will fail fixtures that require pack primitives / lifecycle — that is the moat, now **measured**.
