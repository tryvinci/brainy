"""Aggregate OpenAI usage for external lane runs (secrets redacted in logs)."""

from __future__ import annotations

import json
import re
from dataclasses import dataclass, field
from pathlib import Path

_SECRET_RE = re.compile(r"(sk-[A-Za-z0-9_-]{10,}|api[_-]?key[\"']?\s*[:=]\s*[\"'][^\"']+)", re.I)


@dataclass
class UsageTotals:
    chat_input_tokens: int = 0
    chat_output_tokens: int = 0
    embed_tokens: int = 0
    estimated_usd: float = 0.0
    http_requests: int = 0

    def to_dict(self) -> dict:
        return {
            "chat_input_tokens": self.chat_input_tokens,
            "chat_output_tokens": self.chat_output_tokens,
            "embed_tokens": self.embed_tokens,
            "estimated_usd": round(self.estimated_usd, 6),
            "http_requests": self.http_requests,
        }


@dataclass
class LaneTelemetry:
    lane: str
    usage: UsageTotals = field(default_factory=UsageTotals)
    events: list[dict] = field(default_factory=list)

    def add_openai_usage(self, *, input_tokens: int = 0, output_tokens: int = 0, embed_tokens: int = 0, usd: float = 0.0) -> None:
        self.usage.chat_input_tokens += input_tokens
        self.usage.chat_output_tokens += output_tokens
        self.usage.embed_tokens += embed_tokens
        self.usage.estimated_usd += usd

    def log_http(self, op: str, status: int, body_preview: str) -> None:
        self.usage.http_requests += 1
        redacted = _SECRET_RE.sub("***REDACTED***", body_preview)[:500]
        self.events.append({"op": op, "status": status, "body_preview": redacted})

    def write(self, path: Path) -> None:
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(
            json.dumps({"lane": self.lane, "usage": self.usage.to_dict(), "events": self.events}, indent=2) + "\n",
            encoding="utf-8",
        )
