# Cutover Plan

## Goal

Replace the active root implementation with the Go rebuild while keeping the archived Python prototype recoverable until the destructive-change gate passes.

## Destructive-Change Gate

Do not permanently remove archived prototype content until all of the following are true:

1. `go test ./...` passes.
2. The thin-slice service path passes deterministic ingest -> persist -> search verification.
3. Duplicate ingest idempotency is proven.
4. One correction or suppression flow changes later retrieval results.
5. The replacement docs and runbook describe how to reproduce the current verification evidence.

## Archive Policy

Archived prototype path:

- `archive/brainy-python-prototype/`

Rules:

- no production Go code may import or depend on the archive
- archive remains inspectable during the rebuild
- any later deletion must be a distinct logical change after the destructive-change gate passes

## Replacement Order

1. Archive prototype
2. Bootstrap Go module and root docs
3. Implement thin-slice contracts
4. Add persistence verification
5. Add parity fixtures and eval harness
6. Remove no-longer-needed transitional root files only after proof exists
