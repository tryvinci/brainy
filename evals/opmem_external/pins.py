"""Pinned models and spend ceilings for OpMem external lane comparisons."""

from __future__ import annotations

PINNED_CHAT_MODEL = "gpt-4.1-mini-2025-04-14"
PINNED_EMBED_MODEL = "text-embedding-3-small"

# Official list prices (USD per 1M tokens) for worksheet generation — 2025-04 OpenAI public rates.
RATE_CHAT_INPUT_PER_M = 0.40
RATE_CHAT_OUTPUT_PER_M = 1.60
RATE_EMBED_PER_M = 0.02

# Owner-approved spend authority (USD); live runs must stay at or below projected totals.
SPEND_CEILING_USD: dict[str, float] = {
    "mem0-oss": 1.0,
    "letta": 1.0,
    "langmem": 1.0,
    "zep": 0.0,
    "total": 3.0,
}

LANE_ENV_REQUIRED: dict[str, tuple[str, ...]] = {
    "mem0-oss": ("OPENAI_API_KEY",),  # local Qdrant; no MEM0_API_KEY
    "zep": ("ZEP_API_KEY",),
    # LETTA_APP_SERVER_TOKEN optional for local `letta server --no-secure` (see external-lanes-setup.md).
    "letta": ("OPENAI_API_KEY",),
    "langmem": ("OPENAI_API_KEY", "DATABASE_URL"),
}
