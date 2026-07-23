# Staging deploy & post-deploy benchmarks

**Host:** Render (Blueprint)  
**Source of truth:** [`render.yaml`](../render.yaml) on branch `dev`  
**Goal:** Public staging URL for dogfood + HTTP benchmarks before API launch.

---

## Architecture

| Service | Role |
| --- | --- |
| `brainy-api-staging` | Web — Docker image, health `/healthz` |
| `brainy-worker-staging` | Worker — same image, `BRAINY_WORKER_MODE=loop` |
| `brainy-staging-db` | Postgres 16 (basic-256mb) — migrations on API boot |

pgvector is **optional** (migration v9 no-ops if the extension is missing). Staging still runs hybrid with the deterministic local embedder.

---

## One-time: apply Blueprint

1. Push `render.yaml` on `dev` (already in repo after this change).
2. Open: [Apply Blueprint (tryvinci/brainy)](https://dashboard.render.com/select-repo?type=blueprint&repo=https://github.com/tryvinci/brainy)
3. Select branch **`dev`** (staging tracks staging branch).
4. Set secret env (Dashboard → `brainy-api-staging`):
   - `BRAINY_API_KEYS` — optional until you flip auth on  
     Example (auth later): `*:sk_staging_bench,demo:sk_demo`
5. Click **Apply**. Wait for DB + API + worker healthy (~5–10 min first deploy).
6. Copy the API URL (e.g. `https://brainy-api-staging.onrender.com`).
7. **Provider extract (conversational long-memory):** on `brainy-worker-staging`, set Dashboard secrets:
   - `BRAINY_PROVIDER_BASE_URL` — OpenAI-compatible base (e.g. CF AI Gateway `/compat`)
   - `BRAINY_PROVIDER_API_KEY`
   - `BRAINY_PROVIDER_MODEL`
   - optional `BRAINY_PROVIDER_TIMEOUT` (Blueprint default `45s`)

   Leave empty to keep deterministic-only worker extract. Sync `/ingest` on the API is always deterministic.
8. **Dense embeddings + reranking (SOTA path):** on both `brainy-api-staging` and
   `brainy-worker-staging`, set a strong hosted embeddings endpoint:
   - `BRAINY_EMBEDDING_BASE_URL`, `BRAINY_EMBEDDING_API_KEY`, `BRAINY_EMBEDDING_MODEL`

   **Cloudflare AI Gateway (Workers AI) — verified 2026-07-23:** `/compat/embeddings`
   works when the model uses the `workers-ai/` provider prefix (bare `@cf/...`
   returns `Invalid provider`; OpenAI embeddings need wholesale credits):

   ```bash
   # smoke the gateway first
   curl -s "$LLM_BASE_URL/embeddings" \
     -H "Authorization: Bearer $LLM_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{"model":"workers-ai/@cf/baai/bge-base-en-v1.5","input":"hello"}'
   # expect HTTP 200 + 768-d vector

   # then set on BOTH Render services (API + worker):
   # BRAINY_EMBEDDING_BASE_URL=<same as LLM_BASE_URL / gateway …/compat>
   # BRAINY_EMBEDDING_API_KEY=<same gateway key>
   # BRAINY_EMBEDDING_MODEL=workers-ai/@cf/baai/bge-base-en-v1.5
   ```

   Setting `BRAINY_EMBEDDING_MODEL` auto-enables entity ranking. Optionally A/B
   `BRAINY_IDF_RANKING=true`. Then re-measure LOCOMO with a **comparable
   answerer/judge** and re-tune the boost stack *there* — never against the
   30-question smoke (see `docs/research/path-to-sota.md`,
   `docs/benchmarks/entity-linking-ab.md`).

CLI alternative (if already logged into Render):

```bash
render blueprints validate
# then Apply via Dashboard, or:
# render blueprints apply   # interactive
```

Approximate cost: Postgres basic-256mb + 2× starter web/worker (not free tier — staging stays awake).

---

## Auth modes

| Mode | Env | When |
| --- | --- | --- |
| **Open dogfood** (default Blueprint) | `BRAINY_REQUIRE_API_KEY=false` | Internal benches / quick curl |
| **Keyed staging** | `REQUIRE=true` + `BRAINY_API_KEYS=*:sk_...` | Before sharing with partners |

Wildcard tenant `*` authenticates without binding `tenant_id` (needed for OpMem, which synthesizes many tenants).

---

## Smoke

```bash
export BRAINY_BASE_URL=https://brainy-api-staging.onrender.com   # your URL
curl -s "$BRAINY_BASE_URL/healthz"
```

With keys:

```bash
export BRAINY_API_KEY=sk_staging_bench
curl -s -H "Authorization: Bearer $BRAINY_API_KEY" \
  -X POST "$BRAINY_BASE_URL/ingest" \
  -H 'Content-Type: application/json' \
  -d '{"tenant_id":"demo","subject_id":"u1","source_type":"conversation","messages":[{"role":"user","content":"Prefer short answers."}]}'
```

---

## Post-deploy benchmarks (API must be up)

```bash
export BRAINY_BASE_URL=https://brainy-api-staging.onrender.com
# only if BRAINY_REQUIRE_API_KEY=true:
# export BRAINY_API_KEY=sk_staging_bench

python3 evals/run_eval.py --base-url "$BRAINY_BASE_URL"
python3 evals/run_vertical_eval.py --base-url "$BRAINY_BASE_URL"
python3 evals/run_opmem.py --systems verbatim,brainy --base-url "$BRAINY_BASE_URL"
python3 evals/run_marketing_mvp_benchmark.py --base-url "$BRAINY_BASE_URL"
python3 evals/run_hybrid_eval.py --base-url "$BRAINY_BASE_URL"
```

**Pass criteria:** all runners exit 0 (OpMem reports task failures diagnostically; infra errors fail the run).

Eval scripts pick up `BRAINY_API_KEY` automatically (`evals/httputil.py`).

---

## Local substitute (pre-staging)

```bash
docker compose up --build -d
python3 evals/run_marketing_mvp_benchmark.py --base-url http://127.0.0.1:8080
```

CI: `.github/workflows/docker-smoke.yml` + `go test ./internal/api/...`.

---

## Deploy loop after first apply

1. Merge work to **`dev`**.
2. Render auto-deploys (if Blueprint linked to `dev`) or **Manual Deploy** in Dashboard.
3. Re-run the benchmark block above against `BRAINY_BASE_URL`.

Promote to production only after Track C checklist leftovers (ToS, backups) — keep using **`dev`** → staging until then. Do not point the Blueprint at `main` until you intend prod.
