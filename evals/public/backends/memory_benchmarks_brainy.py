"""Brainy backend for mem0ai/memory-benchmarks-style UnifiedResult runs (Lane A).

Drop-in alongside evals/public backends; speaks Brainy /ingest + /memories/search
(+ optional /recall) so progress is measurable on third-party harnesses.
"""
from __future__ import annotations

import pathlib
import sys
import time

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "evals"))

from httputil import get_json, post_json  # noqa: E402


class BrainyMemoryBenchBackend:
    name = "brainy"

    def __init__(self, base_url: str, *, use_recall: bool = False) -> None:
        self.base_url = base_url.rstrip("/")
        self.use_recall = use_recall

    def add(self, user_id: str, messages: list[dict], metadata: dict | None = None) -> dict:
        payload = {
            "tenant_id": f"mbench-{user_id}",
            "subject_id": user_id,
            "source_type": "conversation",
            "messages": messages,
        }
        if metadata:
            payload["metadata"] = metadata
        return post_json(self.base_url, "/ingest", payload, timeout=60)

    def search(self, user_id: str, query: str, top_k: int = 30) -> list[dict]:
        body = get_json(
            self.base_url,
            "/memories/search",
            {
                "tenant_id": f"mbench-{user_id}",
                "subject_id": user_id,
                "q": query,
                "limit": str(top_k),
            },
            timeout=60,
        )
        return body.get("results", [])

    def recall(self, user_id: str, query: str, mode: str = "context", top_k: int = 30) -> dict:
        return post_json(
            self.base_url,
            "/recall",
            {
                "tenant_id": f"mbench-{user_id}",
                "subject_id": user_id,
                "q": query,
                "mode": mode,
                "top_k": top_k,
            },
            timeout=120,
        )

    def search_timed(self, user_id: str, query: str, top_k: int = 30) -> tuple[list[dict], float]:
        t0 = time.perf_counter()
        if self.use_recall:
            out = self.recall(user_id, query, mode="context", top_k=top_k)
            results = out.get("memories") or []
        else:
            results = self.search(user_id, query, top_k=top_k)
        return results, (time.perf_counter() - t0) * 1000.0
