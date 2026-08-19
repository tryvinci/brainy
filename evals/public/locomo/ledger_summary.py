"""Stage-oracle failure-ledger histogram for S0 / freeze reports."""
from __future__ import annotations

import json
import pathlib
from collections import Counter
from typing import Any


def summarize_ledger(path: str | pathlib.Path) -> dict[str, Any]:
    p = pathlib.Path(path)
    if not p.exists():
        return {"total": 0, "by_primary": {}, "by_group": {}}
    by_primary: Counter[str] = Counter()
    by_group: Counter[str] = Counter()
    total = 0
    with p.open(encoding="utf-8") as handle:
        for line in handle:
            line = line.strip()
            if not line:
                continue
            try:
                row = json.loads(line)
            except json.JSONDecodeError:
                continue
            total += 1
            primary = str(row.get("primary") or row.get("stage") or "UNKNOWN")
            by_primary[primary] += 1
            flags = row.get("flags") or {}
            group = str(flags.get("group") or row.get("group") or "unknown")
            by_group[f"{group}:{primary}"] += 1
    return {
        "total": total,
        "by_primary": dict(by_primary.most_common()),
        "by_group": dict(by_group.most_common()),
    }
