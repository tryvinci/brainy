"""Dry-run token estimation (no vendor calls)."""

from __future__ import annotations

from dataclasses import dataclass, field

from opmem_external.pins import (
    PINNED_CHAT_MODEL,
    PINNED_EMBED_MODEL,
    RATE_CHAT_INPUT_PER_M,
    RATE_CHAT_OUTPUT_PER_M,
    RATE_EMBED_PER_M,
)


def estimate_tokens(text: str) -> int:
    """Conservative heuristic when tiktoken is not a repo dependency."""
    stripped = (text or "").strip()
    if not stripped:
        return 0
    return max(1, len(stripped) // 4)


@dataclass
class TokenLedger:
    chat_input_tokens: int = 0
    chat_output_tokens: int = 0
    embed_tokens: int = 0
    chat_model: str = PINNED_CHAT_MODEL
    embed_model: str = PINNED_EMBED_MODEL

    def add_chat(self, prompt: str, completion: str = "") -> None:
        self.chat_input_tokens += estimate_tokens(prompt)
        self.chat_output_tokens += estimate_tokens(completion)

    def add_embed(self, text: str) -> None:
        self.embed_tokens += estimate_tokens(text)

    def estimated_usd(self) -> float:
        return (
            self.chat_input_tokens * RATE_CHAT_INPUT_PER_M / 1_000_000
            + self.chat_output_tokens * RATE_CHAT_OUTPUT_PER_M / 1_000_000
            + self.embed_tokens * RATE_EMBED_PER_M / 1_000_000
        )

    def to_dict(self) -> dict:
        return {
            "chat_model": self.chat_model,
            "embed_model": self.embed_model,
            "chat_input_tokens": self.chat_input_tokens,
            "chat_output_tokens": self.chat_output_tokens,
            "embed_tokens": self.embed_tokens,
            "estimated_usd": round(self.estimated_usd(), 6),
        }


@dataclass
class OperationCounts:
    remember: int = 0
    recall: int = 0
    revise: int = 0
    forget: int = 0
    reset: int = 0
    http_calls: int = 0

    def to_dict(self) -> dict:
        return {
            "remember": self.remember,
            "recall": self.recall,
            "revise": self.revise,
            "forget": self.forget,
            "reset": self.reset,
            "http_calls": self.http_calls,
        }


@dataclass
class LaneDryRunStats:
    lane: str
    ops: OperationCounts = field(default_factory=OperationCounts)
    tokens: TokenLedger = field(default_factory=TokenLedger)
    product_constraints: list[str] = field(default_factory=list)
    credentials_present: bool = False
    missing_credentials: list[str] = field(default_factory=list)

    def to_dict(self) -> dict:
        return {
            "lane": self.lane,
            "operations": self.ops.to_dict(),
            "tokens": self.tokens.to_dict(),
            "product_constraints": self.product_constraints,
            "credentials_present": self.credentials_present,
            "missing_credentials": self.missing_credentials,
        }
