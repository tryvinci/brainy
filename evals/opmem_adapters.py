#!/usr/bin/env python3
"""System adapters for the OpMem operational-correctness benchmark.

Each adapter maps the neutral operation set (remember / recall / revise /
forget) onto one memory system. Actors are (tenant, subject) tuples; adapters
namespace them per run so tasks stay hermetic on shared backends.
"""
from __future__ import annotations

import pathlib
import sys
import time
import urllib.parse

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from httputil import get_json, post_json  # noqa: E402


class VerbatimBaseline:
    """Naive verbatim store: append on remember, token-overlap recall,
    exact CRUD for revise/forget. A stand-in for raw-RAG memory."""

    name = "verbatim"

    def available(self) -> bool:
        return True

    def begin_task(self, task_name: str) -> None:
        self._entries: dict[tuple[str, str], list[dict]] = {}
        self._seq = 0

    def remember(self, actor: tuple[str, str], content: str) -> list[str]:
        self._seq += 1
        entry_id = f"v{self._seq}"
        self._entries.setdefault(actor, []).append({"id": entry_id, "content": content, "seq": self._seq})
        return [entry_id]

    def recall(self, actor: tuple[str, str], query: str) -> list[dict]:
        query_tokens = _tokenize(query)
        scored = []
        for entry in self._entries.get(actor, []):
            content_tokens = set(_tokenize(entry["content"]))
            matched = sum(1 for token in query_tokens if token in content_tokens)
            if matched:
                scored.append((matched / len(query_tokens), entry["seq"], entry))
        scored.sort(key=lambda item: (-item[0], -item[1]))
        return [{"id": entry["id"], "content": entry["content"]} for _, _, entry in scored]

    def forget(self, actor: tuple[str, str], memory_ids: list[str]) -> None:
        entries = self._entries.get(actor, [])
        self._entries[actor] = [entry for entry in entries if entry["id"] not in memory_ids]

    def revise(self, actor: tuple[str, str], memory_ids: list[str], content: str) -> None:
        for entry in self._entries.get(actor, []):
            if entry["id"] == memory_ids[0]:
                entry["content"] = content
                return


class BrainyAdapter:
    """HTTP adapter for a live Brainy API."""

    name = "brainy"

    def __init__(self, base_url: str) -> None:
        self.base_url = base_url.rstrip("/")
        self._run_nonce = str(int(time.time()))

    def available(self) -> bool:
        return True

    def begin_task(self, task_name: str) -> None:
        self._ns = f"opmem-{self._run_nonce}-{task_name}"

    def _scope(self, actor: tuple[str, str]) -> tuple[str, str]:
        tenant, subject = actor
        return f"{self._ns}-{tenant}", subject

    def remember(self, actor: tuple[str, str], content: str) -> list[str]:
        tenant, subject = self._scope(actor)
        response = post_json(self.base_url, "/ingest", {
            "tenant_id": tenant,
            "subject_id": subject,
            "source_type": "conversation",
            "messages": [{"role": "user", "content": content}],
        })
        return [m["memory_id"] for m in response.get("memories", [])]

    def recall(self, actor: tuple[str, str], query: str) -> list[dict]:
        tenant, subject = self._scope(actor)
        body = get_json(self.base_url, "/memories/search", {
            "tenant_id": tenant,
            "subject_id": subject,
            "q": query,
        })
        return [{"id": r["memory_id"], "content": r["content"]} for r in body.get("results", [])]

    def forget(self, actor: tuple[str, str], memory_ids: list[str]) -> None:
        tenant, subject = self._scope(actor)
        query = urllib.parse.urlencode({"tenant_id": tenant, "subject_id": subject})
        for memory_id in memory_ids:
            post_json(self.base_url, f"/memories/{memory_id}/suppress?{query}", {})

    def revise(self, actor: tuple[str, str], memory_ids: list[str], content: str) -> None:
        tenant, subject = self._scope(actor)
        query = urllib.parse.urlencode({"tenant_id": tenant, "subject_id": subject})
        post_json(
            self.base_url,
            f"/memories/{memory_ids[0]}/correct?{query}",
            {"content": content, "source_text": content},
        )


class Mem0OpAdapter:
    """Best-effort adapter for the Mem0 platform API (requires MEM0_API_KEY)."""

    name = "mem0"
    write_delay_seconds = 2

    def __init__(self, api_key: str | None = None) -> None:
        from competitors.mem0_adapter import Mem0Adapter

        self._client = Mem0Adapter(api_key=api_key)
        self._run_nonce = str(int(time.time()))

    def available(self) -> bool:
        return self._client.available()

    def begin_task(self, task_name: str) -> None:
        self._ns = f"opmem-{self._run_nonce}-{task_name}"

    def _user_id(self, actor: tuple[str, str]) -> str:
        tenant, subject = actor
        return f"{self._ns}::{tenant}::{subject}"

    def remember(self, actor: tuple[str, str], content: str) -> list[str]:
        response = self._client.add_messages(self._user_id(actor), [{"role": "user", "content": content}])
        time.sleep(self.write_delay_seconds)
        ids = _extract_mem0_ids(response)
        if not ids:
            # Fallback: locate the stored memory by searching its own content.
            ids = [r["id"] for r in self.recall(actor, content)[:1]]
        return ids

    def recall(self, actor: tuple[str, str], query: str) -> list[dict]:
        raw = self._client.search(self._user_id(actor), query)
        normalized = self._client.normalize_results(raw)
        return [{"id": r["memory_id"], "content": r["content"]} for r in normalized]

    def forget(self, actor: tuple[str, str], memory_ids: list[str]) -> None:
        for memory_id in memory_ids:
            self._client._request("DELETE", f"/v1/memories/{memory_id}/")
        time.sleep(self.write_delay_seconds)

    def revise(self, actor: tuple[str, str], memory_ids: list[str], content: str) -> None:
        self._client._request("PUT", f"/v1/memories/{memory_ids[0]}/", {"text": content})
        time.sleep(self.write_delay_seconds)


def _extract_mem0_ids(response: dict | list) -> list[str]:
    items = response if isinstance(response, list) else response.get("results", [])
    ids = []
    for item in items:
        if isinstance(item, dict) and item.get("id"):
            ids.append(item["id"])
    return ids


def _tokenize(value: str) -> list[str]:
    for ch in ",.?!:;":
        value = value.replace(ch, " ")
    return [token for token in value.lower().split() if token]
