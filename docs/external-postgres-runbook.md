# External Postgres Runbook

## Goal

Run the Brainy API and worker against an operator-managed Postgres instance instead of the embedded test harness.

## Required Environment

Populate:

- `BRAINY_DATABASE_URL`
- `BRAINY_HTTP_ADDR`
- `BRAINY_WORKER_MODE`
- `BRAINY_WORKER_POLL_INTERVAL`

Example:

```bash
export BRAINY_DATABASE_URL='postgres://brainy:brainy@localhost:5432/brainy?sslmode=disable'
export BRAINY_HTTP_ADDR='127.0.0.1:8080'
export BRAINY_WORKER_MODE='loop'
export BRAINY_WORKER_POLL_INTERVAL='2s'
```

## Start Sequence

1. Start Postgres and ensure the target database exists.
2. Start the API:

```bash
go run ./cmd/api
```

3. Start the worker:

```bash
go run ./cmd/worker
```

## Smoke Flow

1. Synchronous ingest:

```bash
curl -X POST http://127.0.0.1:8080/ingest \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id":"t1","subject_id":"u1","source_type":"conversation","messages":[{"role":"user","content":"I prefer concise answers."}]}'
```

2. Async ingest:

```bash
curl -X POST http://127.0.0.1:8080/ingest/async \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id":"t1","subject_id":"u1","source_type":"conversation","messages":[{"role":"user","content":"I prefer detailed summaries."}]}'
```

3. Search:

```bash
curl 'http://127.0.0.1:8080/memories/search?tenant_id=t1&subject_id=u1&q=How%20should%20I%20respond'
```

4. Correct:

```bash
curl -X POST 'http://127.0.0.1:8080/memories/<memory_id>/correct?tenant_id=t1&subject_id=u1' \
  -H 'Content-Type: application/json' \
  -d '{"content":"Prefers detailed answers"}'
```

5. Suppress:

```bash
curl -X POST 'http://127.0.0.1:8080/memories/<memory_id>/suppress?tenant_id=t1&subject_id=u1'
```

## Expected Results

- API and worker boot can both call the migration runner safely because migrations are serialized with a database advisory lock
- synchronous ingest returns created/updated/deduped counts
- async ingest returns `ingest_id` and `job_id`
- loop-mode worker eventually processes pending jobs
- stale in-progress jobs are reclaimable after the lease window if a worker crashes mid-processing
- correction changes later search output
- duplicate corrections that would collide with an existing canonical memory return a conflict instead of silently corrupting state
- suppression removes later search results

## Verification (CI parity)

From repo root with the API running (`go run ./cmd/api`):

```bash
go test ./...
python3 evals/run_eval.py --base-url http://127.0.0.1:8080
python3 evals/correction_stickiness_eval.py --base-url http://127.0.0.1:8080
```

All three must pass before merge. GitHub Actions runs `go test ./...` on push/PR to `dev` and `main`.

## Marketing vertical smoke (optional)

With `vertical=marketing` on ingest, records receive pack labels and primitive rank weights from `packs/marketing/v1/pack.yaml`:

```bash
curl -X POST http://127.0.0.1:8080/ingest \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id":"t1","subject_id":"brand-acme","vertical":"marketing","source_type":"conversation","messages":[{"role":"user","content":"Never mention competitor X. We prefer warm, concise copy."}]}'

curl 'http://127.0.0.1:8080/memories/search?tenant_id=t1&subject_id=brand-acme&vertical=marketing&q=competitor'
```
