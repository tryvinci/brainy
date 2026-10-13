"""OpenAI list prices used for OpMem external lane worksheets (verify before live runs)."""

from __future__ import annotations

# Source: https://openai.com/api/pricing/ (GPT-4.1 mini + text-embedding-3-small), captured 2026-09-19.
RATE_SOURCE_URL = "https://openai.com/api/pricing/"
RATE_SOURCE_NOTE = (
    "gpt-4.1-mini-2025-04-14: $0.40/1M input, $1.60/1M output; "
    "text-embedding-3-small: $0.02/1M tokens"
)

from opmem_external.pins import (  # noqa: E402
    PINNED_CHAT_MODEL,
    PINNED_EMBED_MODEL,
    RATE_CHAT_INPUT_PER_M,
    RATE_CHAT_OUTPUT_PER_M,
    RATE_EMBED_PER_M,
)

__all__ = [
    "RATE_SOURCE_URL",
    "RATE_SOURCE_NOTE",
    "PINNED_CHAT_MODEL",
    "PINNED_EMBED_MODEL",
    "RATE_CHAT_INPUT_PER_M",
    "RATE_CHAT_OUTPUT_PER_M",
    "RATE_EMBED_PER_M",
]
