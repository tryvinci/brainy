"""Per-lane cost worksheet from dry-run instrumentation."""

from __future__ import annotations

import json
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path

from opmem_external.pins import (
    PINNED_CHAT_MODEL,
    PINNED_EMBED_MODEL,
    RATE_CHAT_INPUT_PER_M,
    RATE_CHAT_OUTPUT_PER_M,
    RATE_EMBED_PER_M,
    SPEND_CEILING_USD,
)
from opmem_external.token_estimate import LaneDryRunStats


@dataclass
class Worksheet:
    generated_utc: str
    task_count: int
    step_count: int
    rates: dict
    pinned_models: dict
    ceilings_usd: dict
    lanes: list[dict]

    def to_dict(self) -> dict:
        return {
            "generated_utc": self.generated_utc,
            "task_count": self.task_count,
            "step_count": self.step_count,
            "rates": self.rates,
            "pinned_models": self.pinned_models,
            "ceilings_usd": self.ceilings_usd,
            "lanes": self.lanes,
        }


def build_worksheet(
    lane_stats: list[LaneDryRunStats],
    *,
    task_count: int,
    step_count: int,
) -> Worksheet:
    lanes_out = []
    total_usd = 0.0
    for stats in lane_stats:
        usd = stats.tokens.estimated_usd()
        if stats.lane == "zep":
            usd = 0.0
        total_usd += usd
        ceiling = SPEND_CEILING_USD.get(stats.lane, 0.0)
        lanes_out.append(
            {
                **stats.to_dict(),
                "projected_usd": round(usd, 6),
                "ceiling_usd": ceiling,
                "within_ceiling": usd <= ceiling,
                "live_run_allowed": False,
                "live_run_note": (
                    "Phase 1 dry-run only; enable after credentials present and "
                    "measured estimate confirmed under ceiling."
                ),
            }
        )
    lanes_out.append(
        {
            "lane": "_total_openai_estimate",
            "projected_usd": round(total_usd, 6),
            "ceiling_usd": SPEND_CEILING_USD["total"],
            "within_ceiling": total_usd <= SPEND_CEILING_USD["total"],
        }
    )
    return Worksheet(
        generated_utc=datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        task_count=task_count,
        step_count=step_count,
        rates={
            PINNED_CHAT_MODEL: {
                "input_per_m": RATE_CHAT_INPUT_PER_M,
                "output_per_m": RATE_CHAT_OUTPUT_PER_M,
            },
            PINNED_EMBED_MODEL: {"per_m": RATE_EMBED_PER_M},
        },
        pinned_models={"chat": PINNED_CHAT_MODEL, "embed": PINNED_EMBED_MODEL},
        ceilings_usd=SPEND_CEILING_USD,
        lanes=lanes_out,
    )


def write_worksheet(path: Path, worksheet: Worksheet) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(worksheet.to_dict(), indent=2) + "\n", encoding="utf-8")
