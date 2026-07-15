from __future__ import annotations

import pathlib
import sys
import time

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT))

from httputil import get_json, post_json  # noqa: E402


class BrainyBackend:
    """HTTP backend against a live Brainy API (local or staging)."""

    name = "brainy"

    def __init__(self, base_url: str, tenant_prefix: str = "public") -> None:
        self.base_url = base_url.rstrip("/")
        self.tenant_prefix = tenant_prefix

    def _tenant(self, user_id: str) -> str:
        return f"{self.tenant_prefix}-{user_id}"

    def remember(self, user_id: str, content: str) -> list[str]:
        """Ingest a single text blob as one message (legacy helper)."""
        return self.remember_messages(user_id, [{"role": "user", "content": content}])

    def remember_messages(
        self,
        user_id: str,
        messages: list[dict],
        metadata: dict | None = None,
    ) -> list[str]:
        """Ingest one or more messages (atomic turns preferred per product docs)."""
        tenant = self._tenant(user_id)
        payload: dict = {
            "tenant_id": tenant,
            "subject_id": user_id,
            "source_type": "conversation",
            "messages": messages,
        }
        if metadata:
            payload["metadata"] = metadata
        response = post_json(self.base_url, "/ingest", payload)
        return [m["memory_id"] for m in response.get("memories", [])]

    def recall(self, user_id: str, query: str, top_k: int = 10) -> tuple[list[dict], float]:
        tenant = self._tenant(user_id)
        started = time.perf_counter()
        body = get_json(
            self.base_url,
            "/memories/search",
            {
                "tenant_id": tenant,
                "subject_id": user_id,
                "q": query,
            },
        )
        latency_ms = (time.perf_counter() - started) * 1000.0
        results = []
        for item in body.get("results", [])[:top_k]:
            results.append(
                {
                    "id": item.get("memory_id", ""),
                    "content": item.get("content", ""),
                    "score": float(item.get("score") or 0),
                }
            )
        return results, latency_ms
