"""Mem0 Platform backend for public LOCOMO same-pin compares (GAP-M1)."""
from __future__ import annotations

import pathlib
import sys
import time
import uuid

ROOT = pathlib.Path(__file__).resolve().parents[3]
sys.path.insert(0, str(ROOT / "evals"))

from competitors.mem0_adapter import Mem0Adapter, mem0_user_id  # noqa: E402


class Mem0Backend:
    """Thin LOCOMO backend wrapping Mem0 Platform search/add."""

    name = "mem0"

    def __init__(self, *, async_timeout_s: float = 180.0) -> None:
        self._client = Mem0Adapter()
        if not self._client.available():
            raise RuntimeError("MEM0_API_KEY required for Mem0Backend")
        self.async_timeout_s = async_timeout_s
        self.async_ingest = True  # platform add is async; smoke waiter uses this flag
        self._run_ns = uuid.uuid4().hex[:10]

    def _user(self, user_id: str) -> str:
        # Namespace per run so concurrent smokes do not collide.
        return mem0_user_id(f"locomo-{self._run_ns}", user_id)

    def remember_messages(
        self,
        user_id: str,
        messages: list[dict],
        metadata: dict | None = None,
        *,
        wait: bool | None = None,
    ) -> list[str]:
        _ = metadata
        uid = self._user(user_id)
        cleaned = []
        for m in messages:
            content = (m.get("content") or "").strip()
            if not content:
                continue
            role = m.get("role") or "user"
            cleaned.append({"role": role, "content": content})
        if not cleaned:
            return []
        self._client.add_messages(uid, cleaned)
        if wait is False:
            return []
        self._client.wait_until_ready(uid, min_count=1, timeout_s=min(60.0, self.async_timeout_s))
        return []

    def wait_until_any_searchable(
        self,
        user_id: str,
        probes: list[str],
        *,
        min_results: int = 1,
        timeout_s: float | None = None,
        settle_polls: int = 2,
    ) -> bool:
        uid = self._user(user_id)
        deadline = time.time() + (timeout_s or self.async_timeout_s)
        probes = [p for p in probes if p][:8] or ["the"]
        while time.time() < deadline:
            for probe in probes:
                hits = self._client.search(uid, probe, top_k=max(min_results, 5))
                if len(hits) >= min_results:
                    # brief settle
                    for _ in range(max(1, settle_polls)):
                        time.sleep(1.0)
                    return True
            time.sleep(2.0)
        return False

    def recall(self, user_id: str, query: str, top_k: int = 10, timeout: float = 60) -> tuple[list[dict], float]:
        _ = timeout
        uid = self._user(user_id)
        t0 = time.perf_counter()
        raw = self._client.search(uid, query, top_k=top_k)
        latency_ms = (time.perf_counter() - t0) * 1000.0
        results = []
        for item in self._client.normalize_results(raw)[:top_k]:
            results.append(
                {
                    "memory_id": item.get("memory_id") or "",
                    "content": item.get("content") or "",
                    "score": float(item.get("score") or 0),
                    "observed_at": "",
                }
            )
        return results, latency_ms
