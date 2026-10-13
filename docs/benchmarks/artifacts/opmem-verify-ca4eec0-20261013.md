# OpMem local verification — no paid APIs — 2026-10-13

**Branch / HEAD:** `pr/opmem-manuscript-5ada` at **`ca4eec0`** (docs-only commits after pin SHA `20533f2`).  
**Purpose:** Re-run embedded-harness experiments without `MEM0_API_KEY` to confirm the **20533f2** Brainy/verbatim pin still holds on current tree. **Does not** re-run Mem0 Platform or Phase~2 external lanes.

## Commands

```bash
unset MEM0_API_KEY
export OPMEM_MANUSCRIPT_OUT=/tmp/opmem-verify-ca4eec0
go test ./internal/api/ -run 'TestOpMemBenchmarkAgainstHTTPServer|TestOpMemPaperExperiments' -count=1
```

## Result (Brainy + verbatim only)

| System | Score | Match `opmem-manuscript-20533f2/opmem-primary.json` |
| --- | ---: | --- |
| brainy | **13/13** | yes |
| verbatim | **10/13** | yes |
| mem0 | (skipped) | pin unchanged (`opmem-pin-20260918.json`) |

## PDF build

```bash
make -C docs/research/opmem/paper pdf
pdftotext docs/research/opmem/paper/main.pdf - | rg 'TODO|\?\?'  # none
```

Pinned evidence under `docs/benchmarks/artifacts/opmem-manuscript-20533f2/` and `opmem-pin-20260918.json` is **unchanged** by this pass.
