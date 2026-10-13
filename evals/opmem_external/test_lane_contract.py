#!/usr/bin/env python3
"""Contract tests for OpMem external lanes (no live vendor APIs)."""
from __future__ import annotations

import json
import os
import unittest
from pathlib import Path
from unittest import mock

# Ensure evals/ is on path when run as module or script
import sys

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))

from opmem_external import LANE_ADAPTERS  # noqa: E402
from opmem_external.pins import SPEND_CEILING_USD  # noqa: E402
from opmem_external.worksheet import build_worksheet  # noqa: E402
from run_opmem import run_task  # noqa: E402


class LaneContractTests(unittest.TestCase):
    def setUp(self) -> None:
        self._env = mock.patch.dict(os.environ, {}, clear=False)
        self._env.start()
        os.environ.pop("OPENAI_API_KEY", None)
        os.environ.pop("ZEP_API_KEY", None)
        os.environ.pop("LETTA_APP_SERVER_TOKEN", None)
        os.environ.pop("DATABASE_URL", None)

    def tearDown(self) -> None:
        self._env.stop()

    def test_namespace_is_deterministic(self) -> None:
        adapter = LANE_ADAPTERS["mem0-oss"](dry_run=True, run_id="fixed-run")
        adapter.begin_task("cor01_basic_revision")
        ns1 = adapter.namespace(("t1", "u1"))
        adapter2 = LANE_ADAPTERS["mem0-oss"](dry_run=True, run_id="fixed-run")
        adapter2.begin_task("cor01_basic_revision")
        ns2 = adapter2.namespace(("t1", "u1"))
        self.assertEqual(ns1, ns2)

    def test_dry_run_remember_recall_no_credentials(self) -> None:
        adapter = LANE_ADAPTERS["langmem"](dry_run=True, run_id="r1")
        self.assertTrue(adapter.available())
        adapter.begin_task("sup01_basic_forget")
        ids = adapter.remember(("t1", "u1"), "door code is 4455")
        self.assertEqual(len(ids), 1)
        hits = adapter.recall(("t1", "u1"), "door code")
        self.assertGreaterEqual(len(hits), 1)

    def test_zep_records_model_constraint(self) -> None:
        adapter = LANE_ADAPTERS["zep"](dry_run=True)
        stats = adapter.dry_run_stats()
        self.assertTrue(stats.product_constraints)

    def test_worksheet_within_ceilings_dry_run(self) -> None:
        fixture_dir = ROOT.parent / "fixtures" / "opmem"
        tasks = [json.loads(p.read_text()) for p in sorted(fixture_dir.glob("*.json"))]
        stats = []
        for name, factory in LANE_ADAPTERS.items():
            adapter = factory(dry_run=True, run_id="worksheet-test")
            for task in tasks:
                run_task(adapter, task)
            stats.append(adapter.dry_run_stats())
        ws = build_worksheet(stats, task_count=len(tasks), step_count=47)
        total = next(row for row in ws.lanes if row["lane"] == "_total_openai_estimate")
        self.assertLessEqual(total["projected_usd"], SPEND_CEILING_USD["total"])
        for row in ws.lanes:
            if row["lane"] in SPEND_CEILING_USD and not row["lane"].startswith("_"):
                self.assertLessEqual(row["projected_usd"], SPEND_CEILING_USD[row["lane"]])

    def test_raw_log_does_not_contain_env_secrets(self) -> None:
        with mock.patch.dict(
            os.environ,
            {"OPENAI_API_KEY": "sk-test-secret-value-12345", "ZEP_API_KEY": "zep-secret"},
            clear=False,
        ):
            adapter = LANE_ADAPTERS["mem0-oss"](dry_run=True, run_id="log-test")
            adapter.begin_task("dup01_idempotent_remember")
            adapter.remember(("t1", "u1"), "hello")
            log_path = Path(self._testMethodName + ".json")
            try:
                adapter.write_raw_log(log_path)
                text = log_path.read_text()
                self.assertNotIn("sk-test-secret", text)
                self.assertNotIn("zep-secret", text)
            finally:
                log_path.unlink(missing_ok=True)

    def test_live_mode_requires_credentials(self) -> None:
        adapter = LANE_ADAPTERS["letta"](dry_run=False)
        self.assertFalse(adapter.available())


if __name__ == "__main__":
    unittest.main()
