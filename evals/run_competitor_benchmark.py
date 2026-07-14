#!/usr/bin/env python3
"""Run parity fixtures against Brainy and Mem0 side-by-side."""
from __future__ import annotations

import argparse
import json
import pathlib
import sys
import urllib.error

ROOT = pathlib.Path(__file__).resolve().parent
sys.path.insert(0, str(ROOT))

import time

from competitors.mem0_adapter import Mem0Adapter, run_fixture as run_mem0_fixture
from run_eval import _namespace_tenants, run_fixture as run_brainy_fixture


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--brainy-url", default="http://127.0.0.1:8080")
    parser.add_argument("--fixture-dir", default="fixtures/parity")
    parser.add_argument("--json-out", default="docs/benchmarks/competitor-parity-latest.json")
    args = parser.parse_args()

    fixture_dir = pathlib.Path(args.fixture_dir)
    if not fixture_dir.is_absolute():
        fixture_dir = ROOT.parent / fixture_dir
    fixtures = sorted(fixture_dir.glob("*.json"))
    if not fixtures:
        print(json.dumps({"passed": False, "error": "no fixtures"}, indent=2))
        return 1

    mem0 = Mem0Adapter()
    mem0_enabled = mem0.available()
    nonce = str(int(time.time()))
    results = []
    overall = True

    for path in fixtures:
        entry = {"fixture": path.stem, "brainy": None, "mem0": None}
        try:
            entry["brainy"] = run_brainy_fixture(args.brainy_url.rstrip("/"), path, nonce)
        except urllib.error.URLError as exc:
            entry["brainy"] = {"passed": False, "errors": [str(exc)]}
        if not entry["brainy"].get("passed"):
            overall = False

        if mem0_enabled:
            fixture = _namespace_tenants(json.loads(path.read_text(encoding="utf-8")), nonce)
            try:
                entry["mem0"] = run_mem0_fixture(mem0, fixture)
            except Exception as exc:  # noqa: BLE001
                entry["mem0"] = {"passed": False, "errors": [f"{type(exc).__name__}: {exc}"]}
        else:
            entry["mem0"] = {"passed": None, "skipped": True, "reason": "MEM0_API_KEY not set"}

        results.append(entry)

    report = {
        "passed": overall,
        "mem0_enabled": mem0_enabled,
        "brainy_url": args.brainy_url,
        "results": results,
    }

    out = pathlib.Path(args.json_out)
    if not out.is_absolute():
        out = ROOT.parent / out
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(report, indent=2))
    return 0 if overall else 1


if __name__ == "__main__":
    sys.exit(main())
