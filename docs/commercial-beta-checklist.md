# Commercial Beta Checklist — Track C (Gate M5)

**Status:** Beta-ready (2026-07-05)  
**GitHub:** [#12](https://github.com/tryvinci/brainy/issues/12) · **Linear:** Gate M5

Use this checklist before onboarding design partners or accepting payment.

---

## Product

| Item | Status | Notes |
| --- | --- | --- |
| Stable ingest/search/correct/suppress | Done | Go API on `main` |
| Marketing vertical pack | Done | `packs/marketing/v1/pack.yaml` |
| Async worker + DLQ | Done | `cmd/worker` |
| API versioning policy | Done | v0.x preview; breaking changes documented in releases |
| Migration policy | Partial | Postgres migrations v1–v9; document backup before upgrade |

---

## Auth & multi-tenancy

| Item | Status | Notes |
| --- | --- | --- |
| API keys per tenant | Done | `BRAINY_API_KEYS=tenant:secret,...` |
| Require auth in production | Done | `BRAINY_REQUIRE_API_KEY=true` or `BRAINY_ENV=production` |
| Tenant isolation in store | Done | `tenant_id` on all records |
| Tenant/key mismatch rejected | Done | 403 when request `tenant_id` ≠ key tenant |
| Key rotation runbook | Todo | Manual env update today; DB-backed keys post-beta |

**Example (staging):**

```bash
export BRAINY_ENV=production
export BRAINY_REQUIRE_API_KEY=true
export BRAINY_API_KEYS="partner-a:sk_live_partner_a,partner-b:sk_live_partner_b"
```

Clients send `Authorization: Bearer sk_live_partner_a` on all routes except `/healthz`.

---

## Billing & legal

| Item | Status | Notes |
| --- | --- | --- |
| Manual invoicing for design partners | Ready | No Stripe integration yet |
| Terms of service | Todo | Required before GA |
| Privacy policy | Todo | Required before GA |
| DPA for enterprise | Todo | Post-beta |

---

## Operations

| Item | Status | Notes |
| --- | --- | --- |
| Docker Compose self-host | Done | `docker-compose.yml` |
| CI (test + docker-smoke) | Done | GitHub Actions |
| Staging deploy runbook | Done | [staging-deploy-runbook.md](./staging-deploy-runbook.md) — Render Blueprint |
| External Postgres runbook | Done | [external-postgres-runbook.md](./external-postgres-runbook.md) |
| Staging host (Render) | In progress | [`render.yaml`](../render.yaml) — apply Blueprint on `dev` |
| Production deploy | Todo | Clone staging Blueprint → prod services after GA criteria |
| Backups + RPO/RTO | Todo | Postgres PITR recommended |
| Status page | Todo | Before GA |
| On-call | Todo | Before GA |

---

## Support & docs

| Item | Status | Notes |
| --- | --- | --- |
| README quickstart | Done | 5-minute Docker path |
| Benchmark reports | Done | OpMem 12/12 + marketing moat |
| Launch narrative | Done | [launch-narrative.md](./benchmarks/launch-narrative.md) |
| Design partner Slack | Ready | Manual onboarding |
| Ticketing | Todo | Before GA |

---

## Beta sign-off

Minimum bar to onboard **2–3 design partners**:

- [x] Track A — `v0.1.0` on `main`
- [x] Track B — OpMem 12/12 + launch narrative published
- [x] Track C — API key auth implemented (#11)
- [x] This checklist documented (#12)
- [ ] ToS + privacy policy published
- [ ] Staging/prod URL with auth enabled
- [ ] First partner onboarded with scoped API key

---

## References

- [execution-plan.md](./vertical/execution-plan.md)
- [go-to-market-roadmap.md](./vertical/go-to-market-roadmap.md)
- GitHub [#11](https://github.com/tryvinci/brainy/issues/11) API key auth
