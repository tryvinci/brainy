#!/usr/bin/env python3
"""Mem0 Platform API adapter for parity fixture comparison."""
from __future__ import annotations

import io
import json
import os
import time
import urllib.error
import urllib.parse
import urllib.request
from datetime import datetime, timezone


_LOCOMO_DATE_FORMATS = (
    "%I:%M %p on %d %B, %Y",
    "%I:%M %p on %d %b, %Y",
    "%Y-%m-%dT%H:%M:%S%z",
    "%Y-%m-%dT%H:%M:%S",
    "%Y-%m-%d",
)


def observed_at_to_epoch(value: str | int | float | None) -> int | None:
    """Convert LoCoMo session dates / ISO / unix timestamps to epoch seconds."""
    if value is None or value == "":
        return None
    if isinstance(value, bool):
        return None
    if isinstance(value, (int, float)):
        n = int(value)
        return n if n > 0 else None
    text = str(value).strip()
    if not text:
        return None
    if text.isdigit():
        n = int(text)
        return n if n > 0 else None
    for fmt in _LOCOMO_DATE_FORMATS:
        try:
            parsed = datetime.strptime(text, fmt)
            if parsed.tzinfo is None:
                parsed = parsed.replace(tzinfo=timezone.utc)
            return int(parsed.timestamp())
        except ValueError:
            continue
    return None


class Mem0Adapter:
    name = "mem0"

    def __init__(
        self,
        api_key: str | None = None,
        base_url: str = "https://api.mem0.ai",
        *,
        event_timeout_s: float = 300.0,
    ) -> None:
        self.api_key = api_key or os.environ.get("MEM0_API_KEY", "")
        self.base_url = base_url.rstrip("/")
        self.search_path = "/v3/memories/search/"
        self.add_path = "/v3/memories/add/"
        self.event_timeout_s = float(event_timeout_s)

    def available(self) -> bool:
        return bool(self.api_key)

    def _request(self, method: str, path: str, payload: dict | None = None, *, timeout: float = 60) -> dict | list:
        headers = {
            "Authorization": f"Token {self.api_key}",
            "Content-Type": "application/json",
        }
        data = None if payload is None else json.dumps(payload).encode("utf-8")
        req = urllib.request.Request(f"{self.base_url}{path}", data=data, headers=headers, method=method)
        try:
            with urllib.request.urlopen(req, timeout=timeout) as resp:
                body = resp.read().decode("utf-8")
                if not body:
                    return {}
                return json.loads(body)
        except urllib.error.HTTPError as exc:
            detail = ""
            try:
                detail = exc.read().decode("utf-8", errors="replace")[:800]
            except Exception:
                detail = ""
            reason = f"{exc.reason}: {detail}" if detail else str(exc.reason)
            raise urllib.error.HTTPError(exc.url, exc.code, reason, exc.hdrs, io.BytesIO(b"")) from None

    def _request_with_fallback(
        self,
        method: str,
        path: str,
        fallback: str,
        payload: dict | None = None,
        *,
        timeout: float = 60,
    ) -> tuple[str, dict | list]:
        try:
            return path, self._request(method, path, payload, timeout=timeout)
        except urllib.error.HTTPError as exc:
            if exc.code != 404:
                raise
            return fallback, self._request(method, fallback, payload, timeout=timeout)

    def _wait_for_event(self, event_id: str, *, timeout_s: float | None = None) -> dict:
        timeout_s = self.event_timeout_s if timeout_s is None else float(timeout_s)
        deadline = time.time() + timeout_s
        last: dict | list = {}
        while time.time() < deadline:
            last = self._request("GET", f"/v1/event/{event_id}/", timeout=30)
            if not isinstance(last, dict):
                time.sleep(1.0)
                continue
            status = str(last.get("status") or "").upper()
            if status == "SUCCEEDED":
                return last
            if status == "FAILED":
                raise RuntimeError(f"mem0 event {event_id} failed: {last.get('error') or last}")
            time.sleep(0.5)
        raise TimeoutError(f"mem0 event {event_id} timed out after {timeout_s:.0f}s")

    def add_messages(
        self,
        user_id: str,
        messages: list[dict],
        *,
        timestamp: int | None = None,
        metadata: dict | None = None,
        wait_event: bool = True,
    ) -> dict:
        payload: dict = {"messages": messages, "user_id": user_id}
        if timestamp is not None:
            payload["timestamp"] = int(timestamp)
        if metadata:
            payload["metadata"] = metadata
        path, response = self._request_with_fallback(
            "POST",
            "/v3/memories/add/",
            "/v1/memories/",
            payload,
            timeout=120,
        )
        self.add_path = path
        if not wait_event or not isinstance(response, dict):
            return response if isinstance(response, dict) else {"results": response}
        event_id = str(response.get("event_id") or "").strip()
        if event_id:
            try:
                return self._wait_for_event(event_id)
            except TimeoutError:
                # One extra poll: the event may have landed after the wait window.
                last = self._request("GET", f"/v1/event/{event_id}/", timeout=30)
                if isinstance(last, dict) and str(last.get("status") or "").upper() == "SUCCEEDED":
                    return last
                raise
        return response

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

    def search(self, user_id: str, query: str, top_k: int = 10, *, rerank: bool = False) -> list[dict]:
        payload = {
            "query": query,
            "filters": {"user_id": user_id},
            "top_k": top_k,
            "rerank": bool(rerank),
        }
        path, response = self._request_with_fallback(
            "POST",
            "/v3/memories/search/",
            "/v2/memories/search/",
            payload,
        )
        self.search_path = path
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
                    "kind": item.get("metadata", {}).get("kind", "fact") if isinstance(item.get("metadata"), dict) else "fact",
                    "content": content,
                    "score": float(item.get("score") or item.get("similarity") or 0),
                    "created_at": item.get("created_at") or item.get("updated_at") or "",
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
