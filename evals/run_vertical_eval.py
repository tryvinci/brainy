#!/usr/bin/env python3
"""Run marketing vertical eval fixtures against a live Brainy API."""
from __future__ import annotations

import argparse
import json
import pathlib
import sys

from run_eval import run_suite


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--base-url", default="http://127.0.0.1:8080")
    parser.add_argument("--fixture-dir", default="fixtures/vertical/marketing")
    args = parser.parse_args()

    overall, results = run_suite(args.base_url, pathlib.Path(args.fixture_dir))
    print(json.dumps({"passed": overall, "vertical": "marketing", "results": results}, indent=2))
    return 0 if overall else 1


if __name__ == "__main__":
    sys.exit(main())
