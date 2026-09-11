# Handoff — n=1540 product `/recall` + next-cycle plan (2026-09-11)

For a reviewer writing the **analysis and next-steps plan**. This is operational truth, not a pin.

## Ask from the operator

1. Fast-forward **`main`** to current **`dev`** (done).
2. Run **full LoCoMo n=1540** product `POST /recall` on that SHA.
3. Treat the result as **true product scoring** even if some conversations scored while extract was still draining (explicit: “it's okay to list it as true product scoring”).
4. Do **not** resume leftover-covering from the skip-ingest 180 remaining ledger.

## Landed (code)

| Ref | SHA | Notes |
| --- | --- | --- |
| `origin/main` | **`bbe55f7`** | FF from `c1aa732` (P66 pin docs) on 2026-09-04 |
| `origin/dev` | **`bbe55f7`** | Same tip |
| Product | PR **#184** merge | S1 compiler (will-be / told-about / you're-doing) + S2 think-about ranking + S3 think-about / would-member hops; provider **`provider-v5-ops`** |
| Do not ship | leftover-covering `pr/locomo-180-p29-1e9e` and PRs **#133, #131, #143, #145** | Kill list |

`go test ./...` passed on `bbe55f7` when #184 merged.

## Frozen pins (do not overwrite with a partial 180)

| Pin | Product SHA | Score | Notes |
| --- | --- | --- | --- |
| **P84 (bar)** | `d7d55ab` | **152/180** | MH 28/33 · **OD 4/11 (trail)** · SH 87/98 · temporal 33/38. Frozen tenant `diag-mh-135`, **no re-extract**. |
| P66 on old `main` | `205c47d` | 149/180 | Superseded on `main` by the FF, not by a new 180 pin. |
| Full LoCoMo | `1b5ab3e` | **175/1540 = 11.4%** | **Current published product `/recall` pin** until a finished n=1540 JSON exists. |
| Industry S0 | same 180 tenant | **62/180** | Do not average with product. |
| 1×30 conv-26 | `1b5ab3e` | **21/30 (70%)** | MH 10/10 · OD **0/4**. Measurement, not qualification. |

Pin gate that was used for 180 increments: **≥153/180 and unique leftover losses = none** vs P84. Unique 0/0 is not a pin. **162/180** is 90% on that sample only; public LoCoMo 90% is **n=1540**.

## Live n=1540 run (this VM, 2026-09-11)

This is the run to **finish**, then file a cycle-closeout with **correct/1540 by group**.

| Item | Value |
| --- | --- |
| SHA | `bbe55f7` |
| Lane | product `/recall`, `BRAINY_RECALL_LLM=1` |
| Dataset | `locomo10.json` SHA `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4` (1540 scored cats 1–4) |
| API | loopback port 18200 · DB `brainy_n1540` |
| Tenant prefix | `locomo-n1540-bbe55f7` (tenants `…-conv-26` … `…-conv-50`) |
| Harness | enqueue all 10 convs with `wait_jobs=False`, drain extract, then `run_smoke --skip-ingest --questions 0 --eval-lane product-recall` |
| Log | `/tmp/n1540-bbe55f7.log` |
| Script | `/tmp/run_n1540_complete.sh` (also copied under `evals/tools/`) |
| As of 2026-09-11 ~07:12 UTC | **~577 / 1571** extract jobs completed, **10 in progress** (one per conversation), **~984 pending** |

Expected next: extract drain (~tens of minutes if the VM stays awake at ~10 parallel jobs), then skip-ingest scoring (~1540 `POST /recall` + LLM judge). Outputs:

- `docs/benchmarks/runs/locomo-full-n1540-bbe55f7-s0-complete.json`
- `docs/benchmarks/runs/locomo-full-n1540-bbe55f7-s0-complete.md`
- failure ledger beside that run id

If the VM sleeps, Postgres/API/workers die and `/tmp` logs vanish. Restore from this note; **do not** check out leftover-covering branches.

## Harness pitfalls (already hit)

1. **Product-recall base URL override.** Cloud injects a staging API host. `_product_recall_answer` posts `/recall` there. The first n=1540 attempt (`3dc7ea`) ingested locally and scored staging — almost all scored-category WRONG. Point the product-recall env at the local API (port 18200) and set `BRAINY_USE_RECALL=1` in the harness process. Local API logs must show `POST /recall`.

2. **Per-conversation ingest wait + FIFO extract.** `ClaimNextExtractionJob` serializes jobs per `(tenant_id, subject_id)`. One worker `concurrency=10` still parked idle goroutines, so only ~3 jobs ran. Fix used here: **10 worker processes, `BRAINY_WORKER_CONCURRENCY=1`**, after enqueueing all 10 subjects.

3. **8h `wait_until_jobs_done` then search-settle.** `ingest_conversation` catches timeout and, outside publish mode, scores when search looks stable. An earlier run (`06117e` / `locomo-1fe3a191`) **started scoring conv-43 with ~120 extract jobs still pending**. Operator accepted listing that class of run as product scoring. Prefer drain-to-zero before skip-ingest when possible; do not hide a partial-extract caveat in the closeout.

4. **Progress ticks are not the score.** `run_smoke` prints every CORRECT and every 10th item. Adversarial is excluded from n=1540. Do not quote tick ratios as 11.4% replacement.

5. **Local hash embedder.** This VM has no pgvector; hosted 768-d embedder refuses to boot. Search/recall still work; ANN fail-closed is off.

## Honesty constraints for the plan

- Do **not** write SOTA / beats-Mem0.
- Do **not** replace README **11.4%** until n=1540 JSON exists; then replace it with the new **correct/1540**, labeled SHA + lane.
- Do **not** publish 152/180 or 162/180 as full LoCoMo.
- Trail axis stays **open-domain** (P84 4/11).
- Full n=1540 was requested now (operator override of “S6 only” for this measurement). Still do not treat 180 leftover-covering as S1–S5.
- Mem0 Platform fair 180: last note was quota **2026-09-01**; confirm live quota before claiming a same-pin.
- Industry on the old 180 tenant is still **62/180**.

## Kill list (unchanged)

No new leftover-covering `looks*Query` from the remaining-28 ledger. No Xeonoblade spelling map. Do not steal Deborah onto Jolene / John’s Wolves onto Tim. Do not invent Sunday / study-together / Witcher six months. Do not synthesize LGBTQ **No** from absence. No fusion fishing, graph DB default, category dictionaries, unbounded top-k, LoCoMo/LME-named product rules.

True-product next (after the n=1540 number is in): **S1 compiler / WRITE + re-ingest on a new tenant**, **S2 structured answers** (OD membership/hypothesis without leftover covering), **S5 industry**. See [sota-execution-plan.md](./sota-execution-plan.md) and [competitive/benchmax-audit-2026-08-25.md](./competitive/benchmax-audit-2026-08-25.md).

## What the reviewer should produce

A plan that:

1. Reads the finished n=1540 JSON (or says “still running” with jobs remaining) and files a **cycle-closeout** in order: Landed → Own pins (n=1540 by MH/OD/SH/temporal) → Competitor compare (detailed, in `cycle-closeout.md`) → Why → Next.
2. Names dips as dips. Does not mix 180 covering scores with n=1540.
3. Allocates the next product work from the **failure ledger stages** (WRITE / RETRIEVAL / PROOF / READER), not from leftover query shapes.
4. Leaves GTM/README bake-off-free; evals may name competitors.

## Resume commands (this VM)

```bash
sudo service postgresql start
# API :18200 + 10x /tmp/brainy-worker (CONCURRENCY=1), env in /tmp/brainy_n1540.env
# Do not commit that env file (it may contain injected keys).
# Drain: sudo -u postgres psql -d brainy_n1540 -c "SELECT status, count(*) FROM extraction_jobs GROUP BY 1;"
# Then skip-ingest score (product-recall env must be the local API, not staging):
PYTHONPATH=/workspace/evals BRAINY_USE_RECALL=1 \
  python3 evals/tools/n1540_run_product_recall.sh
```

Want `main` updated with pin docs only after the JSON exists and the closeout is honest. Do not FF leftover-covering into `main`.
