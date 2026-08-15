# Security Policy

## Reporting a vulnerability

**Do not** open a public GitHub issue for undisclosed security bugs.

1. Prefer a [private GitHub security advisory](https://github.com/tryvinci/brainy/security/advisories/new).
2. Or email **s@siddhant.site** with:
   - Description and impact
   - Steps to reproduce
   - Affected endpoints or components
   - Whether you have a fix

We aim to acknowledge reports within **5 business days**.

## Supported versions

| Line | Supported |
| --- | --- |
| `dev` (staging) | Active |
| `main` (production tags) | Active for the latest release |
| Older tags | Best-effort |

This is a developer-preview HTTP service. Do not expose an instance to the
public internet without authentication (`BRAINY_API_KEYS` /
`BRAINY_REQUIRE_API_KEY=true` or `BRAINY_ENV=production`).

## Scope

In scope: the API (`cmd/api`), worker (`cmd/worker`), Postgres store, and
auth/tenant isolation.

Out of scope: running without auth on a public bind address; third-party LLM
providers; datasets under `datasets/` (gitignored).
