from __future__ import annotations

import argparse
import json
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


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-url", default="http://127.0.0.1:8080")
    args = parser.parse_args()
    base = args.base_url.rstrip("/")

    tenant_id = "fixture-correction"
    subject_id = "user-correction"

    # Step 1: ingest an original memory
    ingest_resp = post_json(base, "/ingest", {
        "tenant_id": tenant_id,
        "subject_id": subject_id,
        "source_type": "conversation",
        "messages": [{"role": "user", "content": "I prefer Python."}],
    })
    created = ingest_resp.get("created", 0)
    if created < 1:
        print(json.dumps({"passed": False, "error": "ingest failed", "ingest": ingest_resp}, indent=2))
        return 1

    memory_id = ingest_resp["memories"][0]["memory_id"]

    # Step 2: search before correction
    search_before = get_json(base, "/memories/search", {
        "tenant_id": tenant_id,
        "subject_id": subject_id,
        "q": "What language",
    })
    before_results = search_before.get("results", [])
    if not before_results or before_results[0]["content"] != "Prefers Python":
        print(json.dumps({"passed": False, "error": "initial search did not return original memory", "search": search_before}, indent=2))
        return 1

    original_score = before_results[0].get("score", 0)

    # Step 3: correct the memory
    correct_resp = post_json(
        base,
        f"/memories/{memory_id}/correct?tenant_id={urllib.parse.quote(tenant_id)}&subject_id={urllib.parse.quote(subject_id)}",
        {"content": "Prefers Ruby", "source_text": "Actually I prefer Ruby."},
    )
    if correct_resp.get("content") != "Prefers Ruby":
        print(json.dumps({"passed": False, "error": "correction failed", "correct": correct_resp}, indent=2))
        return 1

    # Step 4: search after correction
    search_after = get_json(base, "/memories/search", {
        "tenant_id": tenant_id,
        "subject_id": subject_id,
        "q": "What language",
    })
    after_results = search_after.get("results", [])
    if not after_results:
        print(json.dumps({"passed": False, "error": "search after correction returned no results", "search": search_after}, indent=2))
        return 1

    corrected_content = after_results[0]["content"]
    corrected_score = after_results[0].get("score", 0)
    corrected_explain = after_results[0].get("explain", {})

    if corrected_content != "Prefers Ruby":
        print(json.dumps({"passed": False, "error": f"correction did not stick; got {corrected_content}", "search": search_after}, indent=2))
        return 1

    if corrected_score <= original_score:
        print(json.dumps(
            {"passed": False, "error": f"corrected memory score did not improve: {corrected_score} <= {original_score}", "search": search_after},
            indent=2,
        ))
        return 1

    if not corrected_explain.get("corrected"):
        print(json.dumps({"passed": False, "error": "corrected memory missing 'corrected' explain flag", "search": search_after}, indent=2))
        return 1

    # Step 5: re-ingest original content and verify correction still wins
    reingest_resp = post_json(base, "/ingest", {
        "tenant_id": tenant_id,
        "subject_id": subject_id,
        "source_type": "conversation",
        "messages": [{"role": "user", "content": "I prefer Python."}],
    })
    search_final = get_json(base, "/memories/search", {
        "tenant_id": tenant_id,
        "subject_id": subject_id,
        "q": "What language",
    })
    final_results = search_final.get("results", [])
    if not final_results or final_results[0]["content"] != "Prefers Ruby":
        print(json.dumps(
            {"passed": False, "error": "correction lost after re-ingesting original content", "search": search_final},
            indent=2,
        ))
        return 1

    print(json.dumps({
        "passed": True,
        "original_score": original_score,
        "corrected_score": corrected_score,
        "explain": corrected_explain,
    }, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
