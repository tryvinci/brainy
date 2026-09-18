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

Machine-readable: `opmem-pin-20260918.json` (copy of `opmem-manuscript-20533f2/opmem-primary.json`).

## Lane ablations (isolated run)

| Configuration | Overall |
| --- | ---: |
| `brainy` (search + correct) | **13/13** |
| `brainy-recall` (`POST /recall`) | **9/13** |
| `brainy-supersede` (`/supersede` revise) | **11/13** |

Artifact: `opmem-manuscript-20533f2/opmem-lane-ablation-isolated.json`.  
`brainy-recall` correction tasks hit HTTP 409 on `/correct` in this harness (documented threat).

## Reproduce

```bash
./scripts/run-opmem-manuscript.sh docs/benchmarks/artifacts/opmem-manuscript-$(git rev-parse --short HEAD)
python3 scripts/generate-opmem-paper-tex.py
make -C docs/research/opmem/paper pdf
```

## Still blocked

- Mem0 OSS, Zep, Letta, LangMem (no adapter or no approved spend)
- v1 ~30 task expansion
