from __future__ import annotations

from typing import Protocol


class MemoryBackend(Protocol):
    name: str

    def remember(self, user_id: str, content: str) -> list[str]:
        ...

    def recall(self, user_id: str, query: str, top_k: int = 10) -> tuple[list[dict], float]:
        """Return (results[{id,content,score}], latency_ms)."""
        ...
