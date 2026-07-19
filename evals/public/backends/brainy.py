from __future__ import annotations

import pathlib
import re
import sys
import time

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT))

from httputil import get_json, post_json  # noqa: E402


class BrainyBackend:
    """HTTP backend against a live Brainy API (local or staging)."""

    name = "brainy"

    def __init__(
        self,
        base_url: str,
        tenant_prefix: str = "public",
        *,
        async_ingest: bool = False,
        async_timeout_s: float = 120.0,
        async_poll_s: float = 1.0,
    ) -> None:
        self.base_url = base_url.rstrip("/")
        self.tenant_prefix = tenant_prefix
        self.async_ingest = async_ingest
        self.async_timeout_s = async_timeout_s
        self.async_poll_s = async_poll_s

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
        *,
        wait: bool | None = None,
    ) -> list[str]:
        """Ingest one or more messages (atomic turns preferred per product docs).

        When ``async_ingest`` is enabled, posts to ``/ingest/async`` (worker /
        provider extract path) and optionally waits until a probe token from the
        batch is searchable.
        """
        tenant = self._tenant(user_id)
        payload: dict = {
            "tenant_id": tenant,
            "subject_id": user_id,
            "source_type": "conversation",
            "messages": messages,
        }
        if metadata:
            payload["metadata"] = metadata

        if self.async_ingest:
            response = post_json(self.base_url, "/ingest/async", payload, timeout=60)
            should_wait = self.async_ingest if wait is None else wait
            if should_wait:
                probe = _probe_token(messages)
                if probe:
                    self.wait_until_searchable(user_id, probe, min_results=1)
            return []

        response = post_json(self.base_url, "/ingest", payload)
        return [m["memory_id"] for m in response.get("memories", [])]

    def wait_until_searchable(
        self,
        user_id: str,
        query: str,
        *,
        min_results: int = 1,
        timeout_s: float | None = None,
    ) -> list[dict]:
        """Poll search until worker extract has made memories findable."""
        return self.wait_until_any_searchable(
            user_id,
            [query],
            min_results=min_results,
            timeout_s=timeout_s,
        )

    def wait_until_any_searchable(
        self,
        user_id: str,
        queries: list[str],
        *,
        min_results: int = 1,
        timeout_s: float | None = None,
    ) -> list[dict]:
        """Poll until any probe query returns enough hits (provider may rewrite text)."""
        probes = [q.strip() for q in queries if q and str(q).strip()]
        if not probes:
            probes = ["conversation"]
        deadline = time.time() + (self.async_timeout_s if timeout_s is None else timeout_s)
        last: list[dict] = []
        last_query = probes[0]
        while time.time() < deadline:
            for query in probes:
                last_query = query
                last, _ = self.recall(user_id, query, top_k=max(min_results, 10))
                if len(last) >= min_results:
                    # Brief settle so later FIFO jobs can finish after first hit.
                    time.sleep(max(self.async_poll_s, 1.0))
                    settled, _ = self.recall(user_id, query, top_k=max(min_results, 10))
                    return settled or last
            time.sleep(self.async_poll_s)
        raise TimeoutError(
            f"async extract not searchable within timeout "
            f"(probes={probes!r}, last_query={last_query!r}, got {len(last)} results). "
            "Is the worker running with provider/deterministic extract?"
        )

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


_STOP = {
    "the",
    "a",
    "an",
    "and",
    "or",
    "to",
    "of",
    "in",
    "on",
    "for",
    "is",
    "it",
    "i",
    "you",
    "we",
    "they",
    "he",
    "she",
    "was",
    "were",
    "be",
    "with",
    "that",
    "this",
    "at",
    "as",
    "from",
    "by",
    "my",
    "me",
    "our",
    "your",
}


def _probe_token(messages: list[dict]) -> str:
    """Pick a distinctive token from the last message for readiness polling."""
    for message in reversed(messages):
        content = str(message.get("content") or "")
        # Prefer year / date-like / longer tokens.
        candidates = re.findall(r"[A-Za-z0-9]{3,}", content)
        ranked = sorted(
            (c for c in candidates if c.lower() not in _STOP),
            key=lambda c: (c.isdigit(), len(c)),
            reverse=True,
        )
        if ranked:
            return ranked[0]
    return ""
