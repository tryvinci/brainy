## Staging deploy & post-deploy eval

After deploying API + Postgres + worker to a staging host:

```bash
export BRAINY_BASE_URL=https://staging.example.com
python3 evals/run_marketing_mvp_benchmark.py --base-url "$BRAINY_BASE_URL"
python3 evals/run_eval.py --base-url "$BRAINY_BASE_URL"
python3 evals/run_vertical_eval.py --base-url "$BRAINY_BASE_URL"
```

**Pass criteria (Gate M2):** all three runners exit 0; commit or archive the generated `marketing-mvp-benchmark.md` on release tags.

**Docker local smoke (pre-staging):**

```bash
docker compose up --build -d
python3 evals/run_marketing_mvp_benchmark.py --base-url http://127.0.0.1:8080
```

Track progress: [`docs/vertical/execution-plan.md`](vertical/execution-plan.md) · Linear ENG-96.
