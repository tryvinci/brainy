from __future__ import annotations

import argparse
import json
import pathlib
import sys
import urllib.error
import urllib.parse
import urllib.request


def post_json(base_url: str, path: str, payload: dict) -> dict:
    request = urllib.request.Request(
        f"{base_url}{path}",
        data=json.dumps(payload).encode("utf-8"),
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(request) as response:
        return json.loads(response.read().decode("utf-8"))


def get_json(base_url: str, path: str, params: dict[str, str]) -> dict:
    query = urllib.parse.urlencode(params)
    with urllib.request.urlopen(f"{base_url}{path}?{query}") as response:
        return json.loads(response.read().decode("utf-8"))


def run_fixture(base_url: str, fixture_path: pathlib.Path) -> dict:
    fixture = json.loads(fixture_path.read_text(encoding="utf-8"))
    ingest_response = post_json(base_url, "/ingest", fixture["ingest"])
    search_response = get_json(base_url, "/memories/search", fixture["search"])

    result = {
        "fixture": fixture["name"],
        "ingest_created": ingest_response.get("created", 0),
        "ingest_deduped": ingest_response.get("deduped", 0),
        "search_result_count": len(search_response.get("results", [])),
        "passed": True,
        "errors": [],
    }

    expectations = fixture["expect"]
    if ingest_response.get("created", 0) < expectations.get("created_at_least", 1):
        result["passed"] = False
        result["errors"].append("created count below expected minimum")

    results = search_response.get("results", [])
    if len(results) < expectations.get("min_results", 1):
        result["passed"] = False
        result["errors"].append("search result count below expected minimum")
    elif "first_kind" in expectations and results[0].get("kind") != expectations["first_kind"]:
        result["passed"] = False
        result["errors"].append("first search result kind mismatch")
    elif "first_content_contains" in expectations and expectations["first_content_contains"] not in results[0].get("content", ""):
        result["passed"] = False
        result["errors"].append("first search result content mismatch")

    repeat_expect = fixture.get("repeat_ingest_expect")
    if repeat_expect:
        repeat_response = post_json(base_url, "/ingest", fixture["ingest"])
        result["repeat_deduped"] = repeat_response.get("deduped", 0)
        if repeat_response.get("deduped", 0) < repeat_expect.get("deduped_at_least", 1):
            result["passed"] = False
            result["errors"].append("repeat ingest did not dedupe as expected")

    if fixture.get("suppress_after_search"):
        if not results:
            result["passed"] = False
            result["errors"].append("cannot suppress without initial search result")
        else:
            suppress_target = results[0]["memory_id"]
            tenant_id = fixture["search"]["tenant_id"]
            subject_id = fixture["search"]["subject_id"]
            post_json(
                base_url,
                f"/memories/{suppress_target}/suppress?tenant_id={urllib.parse.quote(tenant_id)}&subject_id={urllib.parse.quote(subject_id)}",
                {},
            )
            search_after = get_json(base_url, "/memories/search", fixture["search"])
            result["search_after_suppress_count"] = len(search_after.get("results", []))
            if len(search_after.get("results", [])) != 0:
                result["passed"] = False
                result["errors"].append("suppression did not remove result from later search")

    return result


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-url", default="http://127.0.0.1:8080")
    parser.add_argument("--fixture-dir", default="fixtures/parity")
    args = parser.parse_args()

    fixture_dir = pathlib.Path(args.fixture_dir)
    fixtures = sorted(fixture_dir.glob("*.json"))
    if not fixtures:
        print(json.dumps({"passed": False, "error": "no fixtures found"}, indent=2))
        return 1

    results = []
    overall = True
    for fixture_path in fixtures:
        try:
            result = run_fixture(args.base_url.rstrip("/"), fixture_path)
        except urllib.error.URLError as exc:
            print(json.dumps({"passed": False, "error": f"request failed for {fixture_path.name}: {exc}"}, indent=2))
            return 1
        results.append(result)
        overall = overall and result["passed"]

    print(json.dumps({"passed": overall, "results": results}, indent=2))
    return 0 if overall else 1


if __name__ == "__main__":
    sys.exit(main())
