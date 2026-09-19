# OpMem manuscript pin — 2026-09-18

**Git SHA:** `20533f2ca83e02265663d5984b49b13de8499231`  
**Protocol:** [PUBLICATION_READINESS.md](../../research/opmem/PUBLICATION_READINESS.md)  
**PDF:** [main.pdf](../../research/opmem/paper/main.pdf) (`make -C docs/research/opmem/paper pdf`)

## Primary three-system run (embedded API + Mem0 Platform)

| System | Overall |
| --- | ---: |
| Brainy (`/memories/search`) | **13/13** |
| Verbatim baseline | **10/13** |
| Mem0 Platform | **10/13** |

Mem0 Platform: **completed** at negligible API volume (~47 scripted steps). Paid vendor runs elsewhere require a cost estimate and explicit approval first (not a categorical spend ban).

Machine-readable: `opmem-pin-20260918.json` (copy of `opmem-manuscript-20533f2/opmem-primary.json`).

## Lane ablations (isolated run)

| Configuration | Overall |
| --- | ---: |
| `brainy` (search + correct) | **13/13** |
| `brainy-recall` (`POST /recall`) | **9/13** |
| `brainy-supersede` (`/supersede` revise) | **11/13** |

Artifact: `opmem-manuscript-20533f2/opmem-lane-ablation-isolated.json`.  
`brainy-recall` correction tasks hit HTTP 409 on `/correct` in this harness (documented threat).

## Stability (5× repeat)

Valid runs only (`infrastructure_errors=0`): **2/5** at **13/13**.  
Runs reporting **9/10** with **3 infrastructure errors** are **invalid/incomplete** (errored tasks dropped from denominator—not comparable to 13/13). See `opmem-stability.json` and manuscript Section 6.

## Reproduce

```bash
./scripts/run-opmem-manuscript.sh docs/benchmarks/artifacts/opmem-manuscript-$(git rev-parse --short HEAD)
python3 scripts/generate-opmem-paper-tex.py
make -C docs/research/opmem/paper pdf
```

## Pending external systems (blocked)

Mem0 OSS, Zep, Letta, LangMem: **no OpMem adapter, credentials, or line-item cost worksheet** yet. Cost-estimation plan: manuscript Section 6 (`sec:cost-plan`) — 47 steps × mapped vendor billing units per system before approval.

- v1 ~30 task expansion (separate track)
