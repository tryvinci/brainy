#!/usr/bin/env python3
"""Mem0 Platform API adapter for parity fixture comparison."""
from __future__ import annotations

import json
import os
import time
import urllib.error
import urllib.parse
import urllib.request


class Mem0Adapter:
    name = "mem0"

    def __init__(self, api_key: str | None = None, base_url: str = "https://api.mem0.ai") -> None:
        self.api_key = api_key or os.environ.get("MEM0_API_KEY", "")
        self.base_url = base_url.rstrip("/")

    def available(self) -> bool:
        return bool(self.api_key)

    def _request(self, method: str, path: str, payload: dict | None = None) -> dict | list:
        headers = {
            "Authorization": f"Token {self.api_key}",
            "Content-Type": "application/json",
        }
        data = None if payload is None else json.dumps(payload).encode("utf-8")
        req = urllib.request.Request(f"{self.base_url}{path}", data=data, headers=headers, method=method)
        with urllib.request.urlopen(req, timeout=60) as resp:
            body = resp.read().decode("utf-8")
            if not body:
                return {}
            return json.loads(body)

    def add_messages(self, user_id: str, messages: list[dict]) -> dict:
        return self._request(
            "POST",
            "/v1/memories/",
            {"messages": messages, "user_id": user_id},
        )

    def list_memories(self, user_id: str) -> list[dict]:
        response = self._request("GET", f"/v1/memories/?user_id={urllib.parse.quote(user_id)}")
        if isinstance(response, list):
            return response
        return response.get("results", response.get("memories", []))

    def wait_until_ready(self, user_id: str, min_count: int = 1, timeout_s: float = 45.0) -> list[dict]:
        """Mem0 add is async (PENDING); poll until memories are searchable."""
        deadline = time.time() + timeout_s
        last: list[dict] = []
        while time.time() < deadline:
            last = self.list_memories(user_id)
            if len(last) >= min_count:
                return last
            time.sleep(2)
        return last

    def search(self, user_id: str, query: str, top_k: int = 10) -> list[dict]:
        response = self._request(
            "POST",
            "/v2/memories/search/",
            {"query": query, "filters": {"user_id": user_id}, "top_k": top_k},
        )
        if isinstance(response, list):
            return response
        return response.get("results", response.get("memories", []))

    def normalize_results(self, raw: list[dict]) -> list[dict]:
        normalized = []
        for item in raw:
            if not isinstance(item, dict):
                continue
            content = item.get("memory") or item.get("content") or item.get("text") or ""
            normalized.append(
                {
                    "memory_id": item.get("id") or item.get("memory_id") or "",
                    "kind": item.get("metadata", {}).get("kind", "fact"),
                    "content": content,
                    "score": float(item.get("score") or item.get("similarity") or 0),
                }
            )
        return normalized


def mem0_user_id(tenant_id: str, subject_id: str) -> str:
    return f"{tenant_id}::{subject_id}"


def run_fixture(adapter: Mem0Adapter, fixture: dict) -> dict:
    tenant_id = fixture["ingest"]["tenant_id"]
    subject_id = fixture["ingest"]["subject_id"]
    user_id = mem0_user_id(tenant_id, subject_id)

    ingest_payloads = fixture.get("ingests") or [fixture["ingest"]]
    expected = 0
    for payload in ingest_payloads:
        messages = payload.get("messages") or []
        if messages:
            adapter.add_messages(user_id, messages)
            expected += 1
    if expected:
        adapter.wait_until_ready(user_id, min_count=1)

    search = fixture["search"]
    search_user = mem0_user_id(search["tenant_id"], search["subject_id"])
    raw = adapter.search(search_user, search["q"])
    # One more poll if search races ahead of indexing.
    if not raw:
        adapter.wait_until_ready(search_user, min_count=1, timeout_s=20)
        raw = adapter.search(search_user, search["q"])
    results = adapter.normalize_results(raw)

    errors: list[str] = []
    expect = fixture.get("expect", {})
    passed = True
    if len(results) < expect.get("min_results", 0) and expect.get("min_results", 0) > 0:
        passed = False
        errors.append("search result count below expected minimum")
    if results and "first_content_contains" in expect:
        if expect["first_content_contains"].lower() not in results[0].get("content", "").lower():
            passed = False
            errors.append("first search result content mismatch")

    return {
        "fixture": fixture["name"],
        "provider": adapter.name,
        "search_result_count": len(results),
        "passed": passed,
        "errors": errors,
        "top_content": results[0]["content"] if results else "",
    }
