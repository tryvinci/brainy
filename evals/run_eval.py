from __future__ import annotations

import argparse
import json
import pathlib
import sys
import urllib.error
import urllib.parse

ROOT = pathlib.Path(__file__).resolve().parent
sys.path.insert(0, str(ROOT))

from httputil import get_json, post_json  # noqa: E402


def check_expectations(
    search_results: list[dict],
    expectations: dict,
    errors: list[str],
) -> bool:
    passed = True
    if len(search_results) < expectations.get("min_results", 0):
        if expectations.get("min_results", 0) > 0:
            passed = False
            errors.append("search result count below expected minimum")
    if "max_results" in expectations and len(search_results) > expectations["max_results"]:
        passed = False
        errors.append("search result count above expected maximum")
    if search_results and "first_kind" in expectations:
        if search_results[0].get("kind") != expectations["first_kind"]:
            passed = False
            errors.append("first search result kind mismatch")
    if search_results and "first_content_contains" in expectations:
        if expectations["first_content_contains"] not in search_results[0].get("content", ""):
            passed = False
            errors.append("first search result content mismatch")
    if search_results and "first_explain_primitive" in expectations:
        primitive = expectations["first_explain_primitive"]
        if search_results[0].get("explain", {}).get("primitive") != primitive:
            passed = False
            errors.append("first search result explain.primitive mismatch")
    return passed


def run_fixture(base_url: str, fixture_path: pathlib.Path) -> dict:
    fixture = json.loads(fixture_path.read_text(encoding="utf-8"))
    if fixture.get("skip"):
        return {
            "fixture": fixture["name"],
            "passed": True,
            "skipped": True,
            "reason": fixture.get("skip_reason", "deferred"),
        }

    ingest_payloads = fixture.get("ingests")
    if ingest_payloads is None:
        ingest_payloads = [fixture["ingest"]]

    last_ingest: dict = {}
    total_created = 0
    total_deduped = 0
    for payload in ingest_payloads:
        last_ingest = post_json(base_url, "/ingest", payload)
        total_created += last_ingest.get("created", 0)
        total_deduped += last_ingest.get("deduped", 0)

    search_response = get_json(base_url, "/memories/search", fixture["search"])
    search_results = search_response.get("results", [])

    result = {
        "fixture": fixture["name"],
        "ingest_created": total_created,
        "ingest_deduped": total_deduped,
        "search_result_count": len(search_results),
        "passed": True,
        "errors": [],
    }

    expectations = fixture.get("expect", {})
    if total_created < expectations.get("created_at_least", 1):
        result["passed"] = False
        result["errors"].append("created count below expected minimum")

    if not check_expectations(search_results, expectations, result["errors"]):
        result["passed"] = False

    repeat_expect = fixture.get("repeat_ingest_expect")
    if repeat_expect:
        repeat_response = post_json(base_url, "/ingest", ingest_payloads[0])
        result["repeat_deduped"] = repeat_response.get("deduped", 0)
        if repeat_response.get("deduped", 0) < repeat_expect.get("deduped_at_least", 1):
            result["passed"] = False
            result["errors"].append("repeat ingest did not dedupe as expected")

    correct_spec = fixture.get("correct")
    if correct_spec:
        memories = last_ingest.get("memories", [])
        memory_index = correct_spec.get("memory_index", 0)
        if not memories:
            result["passed"] = False
            result["errors"].append("correct step missing ingest memories")
        else:
            memory_id = memories[memory_index]["memory_id"]
            first_ingest = ingest_payloads[0]
            tenant_id = correct_spec.get("tenant_id", first_ingest["tenant_id"])
            subject_id = correct_spec.get("subject_id", first_ingest["subject_id"])
            post_json(
                base_url,
                f"/memories/{memory_id}/correct?tenant_id={urllib.parse.quote(tenant_id)}&subject_id={urllib.parse.quote(subject_id)}",
                {
                    "content": correct_spec["content"],
                    "source_text": correct_spec.get("source_text", correct_spec["content"]),
                },
            )
            after_search = get_json(
                base_url,
                "/memories/search",
                fixture.get("search_after_correct", fixture["search"]),
            )
            after_results = after_search.get("results", [])
            after_expect = fixture.get("expect_after_correct", {})
            if not check_expectations(after_results, after_expect, result["errors"]):
                result["passed"] = False

    if fixture.get("suppress_after_search"):
        if not search_results:
            result["passed"] = False
            result["errors"].append("cannot suppress without initial search result")
        else:
            suppress_target = search_results[0]["memory_id"]
            tenant_id = fixture["search"]["tenant_id"]
            subject_id = fixture["search"]["subject_id"]
            post_json(
                base_url,
                f"/memories/{suppress_target}/suppress?tenant_id={urllib.parse.quote(tenant_id)}&subject_id={urllib.parse.quote(subject_id)}",
                {},
            )
            search_after = get_json(base_url, "/memories/search", fixture["search"])
            search_after_results = search_after.get("results", [])
            result["search_after_suppress_count"] = len(search_after_results)
            max_after = fixture.get("expect_after_suppress", {}).get("max_results", 0)
            if len(search_after_results) > max_after:
                result["passed"] = False
                result["errors"].append("suppression did not remove result from later search")

    return result


def run_suite(base_url: str, fixture_dir: pathlib.Path) -> tuple[bool, list[dict]]:
    fixtures = sorted(fixture_dir.glob("*.json"))
    if not fixtures:
        return False, [{"passed": False, "error": "no fixtures found"}]

    results = []
    overall = True
    for fixture_path in fixtures:
        try:
            result = run_fixture(base_url.rstrip("/"), fixture_path)
        except urllib.error.URLError as exc:
            return False, [{"passed": False, "error": f"request failed for {fixture_path.name}: {exc}"}]
        results.append(result)
        if not result.get("skipped"):
            overall = overall and result["passed"]

    return overall, results


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-url", default="http://127.0.0.1:8080")
    parser.add_argument("--fixture-dir", default="fixtures/parity")
    args = parser.parse_args()

    overall, results = run_suite(args.base_url, pathlib.Path(args.fixture_dir))
    print(json.dumps({"passed": overall, "results": results}, indent=2))
    return 0 if overall else 1


if __name__ == "__main__":
    sys.exit(main())
