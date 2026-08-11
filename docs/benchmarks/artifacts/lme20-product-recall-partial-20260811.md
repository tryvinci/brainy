# LME-20 product-recall — partial proof — 2026-08-11

**Harness:** `--publish --product-recall` on `pr/v3-hardening-qualify-a6c7` / PR #95  
**Queue precheck:** idle (DB)  
**Host:** local API+worker (hardening build)

## Proof observed

| Item | Type | Result | answer_path |
| --- | --- | --- | --- |
| 7a87bd0c | knowledge-update | WRONG | **`/recall`** |

Product `/recall` path is fail-closed and recorded. Full 20-item score deferred while LoCoMo harden 1×30 uses the worker; LME-20 resume after that pin (subject-ordered claims serialize large same-subject haystacks — expected slow drain).

**Not a publishable LME accuracy claim.**
