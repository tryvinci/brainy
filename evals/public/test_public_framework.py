from __future__ import annotations

import pathlib
import sys
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "evals"))

from public.judge import lexical_judge  # noqa: E402
from public.llm import parse_judgment_json, resolve_config  # noqa: E402
from public.proveability import RunManifest, require_pins, sha256_file  # noqa: E402
from public.schema import (  # noqa: E402
    CATEGORIES_TO_SCORE,
    EvalItem,
    JudgmentData,
    compute_metrics,
)


class ProveabilityTests(unittest.TestCase):
    def test_require_pins_complete(self) -> None:
        m = RunManifest(
            benchmark="locomo-smoke",
            dataset_url="https://example.com/d.json",
            dataset_sha256="abc",
            brainy_url="https://brainy.example",
            judge_model="gpt-4o-mini",
            judge_temperature=0.0,
        )
        self.assertEqual(require_pins(m), [])

    def test_require_pins_gaps(self) -> None:
        gaps = require_pins(RunManifest())
        self.assertTrue(any("dataset_url" in g for g in gaps))
        self.assertTrue(any("judge_model" in g for g in gaps))

    def test_sha256_file(self) -> None:
        path = pathlib.Path(__file__).with_name("_tmp_hash.txt")
        path.write_text("hello\n", encoding="utf-8")
        try:
            digest = sha256_file(path)
            self.assertEqual(len(digest), 64)
        finally:
            path.unlink(missing_ok=True)


class SchemaTests(unittest.TestCase):
    def test_compute_metrics_excludes_adversarial(self) -> None:
        items = [
            EvalItem(
                id="a",
                group="single-hop",
                judgment=JudgmentData(judgment="CORRECT", score=1.0),
                extras={"category_id": 4},
            ),
            EvalItem(
                id="b",
                group="adversarial",
                judgment=JudgmentData(judgment="CORRECT", score=1.0),
                extras={"category_id": 5},
            ),
            EvalItem(
                id="c",
                group="multi-hop",
                judgment=JudgmentData(judgment="WRONG", score=0.0),
                extras={"category_id": 1},
            ),
        ]
        metrics = compute_metrics(items, CATEGORIES_TO_SCORE)
        self.assertEqual(metrics.total, 2)
        self.assertEqual(metrics.correct, 1)
        self.assertAlmostEqual(metrics.overall_accuracy, 0.5)


class JudgeTests(unittest.TestCase):
    def test_lexical_substring(self) -> None:
        r = lexical_judge("She went on 7 May 2023", "7 May 2023")
        self.assertEqual(r.judgment, "CORRECT")
        self.assertEqual(r.model, "lexical-overlap-v0")

    def test_lexical_miss(self) -> None:
        r = lexical_judge("I don't know", "7 May 2023")
        self.assertEqual(r.judgment, "WRONG")


class LLMConfigTests(unittest.TestCase):
    def test_local_ollama_without_key(self) -> None:
        cfg = resolve_config(base_url="http://127.0.0.1:11434/v1", model="llama3.1")
        self.assertIsNotNone(cfg)
        assert cfg is not None
        self.assertEqual(cfg.api_key, "ollama")
        self.assertEqual(cfg.model, "llama3.1")
        self.assertFalse(cfg.json_mode)

    def test_remote_requires_key(self) -> None:
        cfg = resolve_config(base_url="https://api.openai.com/v1", api_key="", model="gpt-4o-mini")
        # Empty key + non-local => unavailable unless env provides one.
        # Clear env for this assertion.
        import os

        old_o = os.environ.pop("OPENAI_API_KEY", None)
        old_l = os.environ.pop("LLM_API_KEY", None)
        try:
            cfg = resolve_config(base_url="https://api.openai.com/v1", api_key="", model="x")
            self.assertIsNone(cfg)
        finally:
            if old_o is not None:
                os.environ["OPENAI_API_KEY"] = old_o
            if old_l is not None:
                os.environ["LLM_API_KEY"] = old_l

    def test_parse_judgment_fenced(self) -> None:
        parsed = parse_judgment_json('Here you go:\n```json\n{"judgment":"CORRECT","reason":"match"}\n```')
        self.assertEqual(parsed["judgment"], "CORRECT")


class DatasetParserTests(unittest.TestCase):
    def test_iter_session_turns(self) -> None:
        from public.locomo.dataset import iter_questions, iter_session_turns, iter_sessions

        conv = {
            "conversation": {
                "speaker_a": "A",
                "session_1_date_time": "1 May 2023",
                "session_1": [
                    {"speaker": "A", "text": "hi"},
                    {"speaker": "B", "text": "hello"},
                ],
                "session_2": [{"speaker": "A", "text": "later"}],
            },
            "qa": [{"question": "q?", "answer": "a", "category": 2}],
        }
        turns = iter_session_turns(conv)
        self.assertEqual(len(turns), 3)
        sessions = iter_sessions(conv)
        self.assertEqual(len(sessions), 2)
        self.assertEqual(sessions[0]["observed_at"], "1 May 2023")
        qs = iter_questions(conv)
        self.assertEqual(qs[0]["category"], 2)


if __name__ == "__main__":
    unittest.main()
