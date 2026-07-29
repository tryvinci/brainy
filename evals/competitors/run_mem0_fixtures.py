"""Run Brainy-shaped eval fixtures against Mem0 Platform (empirical counter-run).

Used by the marketing MVP benchmark so Mem0 cells are measured, not assumed.
"""
from __future__ import annotations

import json
import pathlib
import time
import uuid

from competitors.mem0_adapter import Mem0Adapter, mem0_user_id


def _namespace(obj, nonce: str):
    if isinstance(obj, dict):
        out = {}
        for key, value in obj.items():
            if key == "tenant_id" and isinstance(value, str) and value:
                out[key] = f"mem0-{nonce}-{value}"
            else:
                out[key] = _namespace(value, nonce)
        return out
    if isinstance(obj, list):
        return [_namespace(item, nonce) for item in obj]
    return obj


def _check(
    results: list[dict],
    expectations: dict,
    errors: list[str],
    *,
    strict_schema: bool = False,
) -> bool:
    passed = True
    if len(results) < expectations.get("min_results", 0):
        if expectations.get("min_results", 0) > 0:
            passed = False
            errors.append("search result count below expected minimum")
    if "max_results" in expectations and len(results) > expectations["max_results"]:
        passed = False
        errors.append("search result count above expected maximum")
    if results and "first_content_contains" in expectations:
        needle = expectations["first_content_contains"].lower()
        if needle not in results[0].get("content", "").lower():
            passed = False
            errors.append("first search result content mismatch")
    # Schema fields Brainy exposes (kind / explain.primitive) are optional for
    # Mem0 unless strict_schema=True (used for vertical moat fixtures).
    if strict_schema:
        if expectations.get("first_explain_primitive") and results:
            if not results[0].get("explain", {}).get("primitive"):
                passed = False
                errors.append("first search result explain.primitive mismatch (unsupported on Mem0)")
        if expectations.get("first_kind") and results:
            kind = results[0].get("kind") or ""
            if kind != expectations["first_kind"]:
                passed = False
                errors.append("first search result kind mismatch")
    return passed


def run_fixture(
    adapter: Mem0Adapter,
    fixture_path: pathlib.Path,
    nonce: str,
    *,
    strict_schema: bool = False,
) -> dict:
    fixture = _namespace(json.loads(fixture_path.read_text(encoding="utf-8")), nonce)
    name = fixture.get("name") or fixture_path.stem
    if fixture.get("skip"):
        return {
            "fixture": name,
            "provider": "mem0",
            "passed": True,
            "skipped": True,
            "reason": fixture.get("skip_reason", "deferred"),
        }

    errors: list[str] = []
    passed = True
    id_by_content: dict[str, str] = {}

    ingest_payloads = fixture.get("ingests")
    if ingest_payloads is None:
        ingest_payloads = [fixture["ingest"]]

    last_user = ""
    for payload in ingest_payloads:
        tenant_id = payload["tenant_id"]
        subject_id = payload["subject_id"]
        user_id = mem0_user_id(tenant_id, subject_id)
        last_user = user_id
        messages = payload.get("messages") or []
        if not messages:
            continue
        adapter.add_messages(user_id, messages)
        listed = adapter.wait_until_ready(user_id, min_count=1, timeout_s=60)
        for item in listed:
            if isinstance(item, dict) and item.get("id"):
                content = (item.get("memory") or item.get("content") or "").strip().lower()
                if content:
                    id_by_content[content] = item["id"]

    search = fixture["search"]
    search_user = mem0_user_id(search["tenant_id"], search["subject_id"])
    raw = adapter.search(search_user, search["q"], top_k=10)
    if not raw:
        adapter.wait_until_ready(search_user, min_count=1, timeout_s=30)
        raw = adapter.search(search_user, search["q"], top_k=10)
    results = adapter.normalize_results(raw)

    expect = fixture.get("expect", {})
    if not _check(results, expect, errors, strict_schema=strict_schema):
        passed = False

    # created_at_least: Mem0 does not return created counts; approximate via list size.
    if expect.get("created_at_least"):
        listed = adapter.list_memories(search_user or last_user)
        if len(listed) < expect["created_at_least"]:
            passed = False
            errors.append("created count below expected minimum")

    if fixture.get("repeat_ingest_expect"):
        # Mem0 ADD-only: re-ingest may create duplicates — that is a measured outcome.
        first = ingest_payloads[0]
        user_id = mem0_user_id(first["tenant_id"], first["subject_id"])
        before = len(adapter.list_memories(user_id))
        adapter.add_messages(user_id, first.get("messages") or [])
        time.sleep(2)
        after = len(adapter.list_memories(user_id))
        # Brainy expects dedupe; Mem0 typically grows — fail if fixture requires dedupe.
        if after > before and fixture["repeat_ingest_expect"].get("deduped_at_least", 0) > 0:
            passed = False
            errors.append("repeat ingest did not dedupe as expected")

    if fixture.get("correct"):
        correct = fixture["correct"]
        # Prefer memory from first ingest content match; else top search hit.
        memory_id = ""
        if results:
            memory_id = results[0].get("memory_id") or ""
        if not memory_id and id_by_content:
            memory_id = next(iter(id_by_content.values()))
        if not memory_id:
            passed = False
            errors.append("correct step missing ingest memories")
        else:
            try:
                adapter._request("PUT", f"/v1/memories/{memory_id}/", {"text": correct["content"]})
                time.sleep(2)
            except Exception as exc:  # noqa: BLE001 — surface as fixture fail
                passed = False
                errors.append(f"correct failed: {exc}")
            after_search = fixture.get("search_after_correct", fixture["search"])
            after_user = mem0_user_id(after_search["tenant_id"], after_search["subject_id"])
            adapter.wait_until_ready(after_user, min_count=1, timeout_s=30)
            after_raw = adapter.search(after_user, after_search["q"], top_k=10)
            after_results = adapter.normalize_results(after_raw)
            if not _check(
                after_results,
                fixture.get("expect_after_correct", {}),
                errors,
                strict_schema=strict_schema,
            ):
                passed = False

    if fixture.get("suppress_after_search"):
        if not results:
            passed = False
            errors.append("cannot suppress without initial search result")
        else:
            memory_id = results[0].get("memory_id") or ""
            try:
                adapter._request("DELETE", f"/v1/memories/{memory_id}/")
                time.sleep(2)
            except Exception as exc:  # noqa: BLE001
                passed = False
                errors.append(f"suppress failed: {exc}")
            after_raw = adapter.search(search_user, search["q"], top_k=10)
            after_results = adapter.normalize_results(after_raw)
            max_after = fixture.get("expect_after_suppress", {}).get("max_results", 0)
            if len(after_results) > max_after:
                passed = False
                errors.append("suppression did not remove result from later search")

    return {
        "fixture": name,
        "provider": "mem0",
        "search_result_count": len(results),
        "passed": passed,
        "errors": errors,
        "top_content": results[0]["content"] if results else "",
    }


def run_suite(
    adapter: Mem0Adapter,
    fixture_dir: pathlib.Path,
    *,
    strict_schema: bool = False,
) -> tuple[bool, list[dict]]:
    fixtures = sorted(fixture_dir.glob("*.json"))
    if not fixtures:
        return False, [{"fixture": "_suite", "passed": False, "errors": ["no fixtures found"]}]
    nonce = uuid.uuid4().hex[:8]
    results = []
    overall = True
    for path in fixtures:
        try:
            result = run_fixture(adapter, path, nonce, strict_schema=strict_schema)
        except Exception as exc:  # noqa: BLE001
            result = {
                "fixture": path.stem,
                "provider": "mem0",
                "passed": False,
                "errors": [f"{type(exc).__name__}: {exc}"],
            }
        results.append(result)
        if not result.get("skipped"):
            overall = overall and bool(result.get("passed"))
    return overall, results
