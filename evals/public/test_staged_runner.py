from __future__ import annotations

import pathlib
import sys
import tempfile
import unittest
from unittest.mock import patch

ROOT = pathlib.Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "evals"))

from public.judge import llm_judge, _product_recall_answer  # noqa: E402
from public.llm import LLMConfig  # noqa: E402
from public.protocol import PROTOCOL_V1, PROTOCOL_V2, JUDGMENT_TRANSPORT, JUDGMENT_UNRESOLVED  # noqa: E402
from public.staged_runner import (  # noqa: E402
    IdentityMismatch,
    RunSpec,
    RunStore,
    StagedRunner,
    StagedRunError,
    product_answer_v2,
    store_identity_from_runtime,
)


def _cfg() -> LLMConfig:
    return LLMConfig(base_url="http://judge.example/v1", api_key="x", model="judge", json_mode=True)


def _question(qid: str = "conv-26-q1", **extra) -> dict:
    row = {
        "id": qid,
        "sample_id": "conv-26",
        "subject_id": "conv-26",
        "question": "What hobbies does Riley have?",
        "ground_truth": "pottery",
        "group": "single-hop",
        "category_id": 4,
    }
    row.update(extra)
    return row


def _runtime(*, fallbacks: int = 0, database: str = "brainy_n1540", active: bool = True) -> dict:
    return {
        "database": database,
        "signatures": {"api": "hash|local|32", "worker": "hash|local|32"},
        "fallbacks": {"api_embedder": fallbacks, "api_extractor": 0, "worker_total": fallbacks},
        "ann": {"active": active, "has_pgvector": active, "dim_histogram": {"32": 10}},
        "worker": {"extractor": {"version": "provider-v5-ops"}},
    }


class FakeRecallClient:
    def __init__(self, base_url: str = "http://127.0.0.1:18200") -> None:
        self.base_url = base_url
        self.calls: list[dict] = []
        self.body = {"answer": "pottery", "abstained": False}
        self.fail: Exception | None = None

    def recall(self, tenant_id, subject_id, question, *, mode="answer", top_k=30, **_kwargs):
        self.calls.append(
            {
                "tenant_id": tenant_id,
                "subject_id": subject_id,
                "question": question,
                "mode": mode,
                "top_k": top_k,
            }
        )
        if self.fail:
            raise self.fail
        return dict(self.body)

    def runtime(self) -> dict:
        return _runtime()


def _runner(tmp: pathlib.Path, **kwargs) -> StagedRunner:
    spec = RunSpec(
        run_id="test-run",
        artifact_root=tmp,
        api_url="http://127.0.0.1:18200",
        tenant_prefix="locomo-n1540-bbe55f7",
        dataset_hash="abc123",
        protocol=PROTOCOL_V2,
        product_sha="bbe55f7",
        harness_sha="harness",
        judge_model="judge@example",
        skip_ingest=kwargs.pop("skip_ingest", True),
        qualification_profile=kwargs.pop("qualification_profile", False),
        subjects=["conv-26"],
        drain_timeout_s=kwargs.pop("drain_timeout_s", 2.0),
    )
    client = FakeRecallClient()
    defaults = dict(
        runtime_loader=lambda: _runtime(),
        enqueue_fn=lambda: {"jobs_expected": 2, "job_ids": ["j1", "j2"]},
        job_poll_fn=lambda: {
            "jobs_expected": 2,
            "jobs_completed": 2,
            "jobs_failed": 0,
            "open": 0,
        },
        search_fn=lambda prefix, subject, q: [{"id": "m1", "content": "Riley likes pottery"}],
        answer_fn=lambda question: {
            "id": question["id"],
            "status": "ok",
            "answer": "pottery",
            "model": "brainy-recall+answer",
            "latency_ms": 10.0,
        },
        judge_fn=lambda payload: {
            "id": payload["question"]["id"],
            "judgment": "CORRECT",
            "score": 1.0,
            "reason": "match",
            "model": "judge",
        },
    )
    defaults.update(kwargs)
    return StagedRunner(spec, client, RunStore(tmp), **defaults)  # type: ignore[arg-type]


class ProtocolJudgeTests(unittest.TestCase):
    def test_v2_malformed_judge_is_unresolved_not_substring_correct(self) -> None:
        with patch("public.judge.chat_completion", return_value="not-json definitely pottery"):
            result = llm_judge("She likes pottery", "pottery", "hobbies?", _cfg(), protocol=PROTOCOL_V2)
        self.assertEqual(result.judgment, JUDGMENT_UNRESOLVED)

    def test_v1_malformed_judge_may_substring_correct(self) -> None:
        with patch("public.judge.chat_completion", return_value="not-json"):
            result = llm_judge("She likes pottery", "pottery", "hobbies?", _cfg(), protocol=PROTOCOL_V1)
        self.assertEqual(result.judgment, "CORRECT")
        self.assertIn("deterministic", result.reason)

    def test_v2_product_recall_always_mode_answer(self) -> None:
        client = FakeRecallClient()
        with patch.dict(
            "os.environ",
            {"BRAINY_USE_RECALL": "1", "BRAINY_EVAL_PROTOCOL": PROTOCOL_V2},
            clear=False,
        ):
            answer, model = _product_recall_answer(
                "What hobbies does Riley have?",
                tenant_id="t",
                subject_id="u",
                protocol=PROTOCOL_V2,
                client=client,
            )
        self.assertEqual(answer, "pottery")
        self.assertEqual(model, "brainy-recall+answer")
        self.assertEqual(client.calls[0]["mode"], "answer")

    def test_v2_http_failure_is_transport_not_not_in_memory(self) -> None:
        client = FakeRecallClient()
        client.fail = RuntimeError("connection reset")
        with patch.dict("os.environ", {"BRAINY_USE_RECALL": "1"}, clear=False):
            answer, model = _product_recall_answer(
                "What hobbies does Riley have?",
                tenant_id="t",
                subject_id="u",
                protocol=PROTOCOL_V2,
                client=client,
            )
        self.assertEqual(answer, "")
        self.assertEqual(model, "brainy-recall+transport")

    def test_v2_does_not_use_context_block(self) -> None:
        client = FakeRecallClient()
        client.body = {"answer": "", "context_block": "secret dump", "abstained": False}
        with patch.dict("os.environ", {"BRAINY_USE_RECALL": "1"}, clear=False):
            answer, model = _product_recall_answer(
                "q",
                tenant_id="t",
                subject_id="u",
                protocol=PROTOCOL_V2,
                client=client,
            )
        self.assertEqual(answer, "")
        self.assertEqual(model, "brainy-recall+empty")

    def test_product_answer_v2_retries_then_keeps_transport_row(self) -> None:
        client = FakeRecallClient()
        client.fail = TimeoutError("deadline")
        row = product_answer_v2(
            client, _question(), tenant_id="t", subject_id="u", max_attempts=2  # type: ignore[arg-type]
        )
        self.assertEqual(row["status"], JUDGMENT_TRANSPORT)
        self.assertEqual(len(client.calls), 2)


class StagedRunnerTests(unittest.TestCase):
    def test_resume_rejects_wrong_api_and_store_identity(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            questions = [_question()]
            runner = _runner(root)
            runner.preflight(questions)
            runner.spec.api_url = "http://127.0.0.1:9999"
            runner.client.base_url = "http://127.0.0.1:9999"
            with self.assertRaises(IdentityMismatch) as ctx:
                runner.preflight(questions)
            self.assertIn("api_url", str(ctx.exception))

            runner2 = _runner(root, runtime_loader=lambda: _runtime(database="otherdb"))
            with self.assertRaises(IdentityMismatch) as ctx2:
                runner2.preflight(questions)
            self.assertIn("store_identity", str(ctx2.exception))

    def test_missing_and_failed_jobs_fail_drain(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            runner = _runner(
                root,
                skip_ingest=False,
                drain_timeout_s=0.2,
                job_poll_fn=lambda: {
                    "jobs_expected": 4,
                    "jobs_completed": 1,
                    "jobs_failed": 0,
                    "open": 0,
                    "poll_s": 0.01,
                },
            )
            runner.preflight([_question()])
            runner.enqueue({})
            with self.assertRaises(StagedRunError) as ctx:
                runner.drain()
            self.assertIn("empty queue is not sufficient", str(ctx.exception))

            runner.job_poll_fn = lambda: {
                "jobs_expected": 2,
                "jobs_completed": 1,
                "jobs_failed": 1,
                "open": 0,
            }
            with self.assertRaises(StagedRunError) as ctx2:
                runner.drain()
            self.assertIn("failed", str(ctx2.exception).lower())

    def test_provider_fallback_fails_qualification(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            runner = _runner(
                pathlib.Path(tmp),
                qualification_profile=True,
                runtime_loader=lambda: _runtime(fallbacks=2, active=False),
            )
            with self.assertRaises(StagedRunError) as ctx:
                runner.preflight([_question()])
            self.assertIn("fallback", str(ctx.exception))

    def test_interrupted_run_reuses_answers_and_retries_judge(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            questions = [_question("q1"), _question("q2")]
            answers_seen: list[str] = []
            judge_calls: list[str] = []

            def answer_fn(question: dict) -> dict:
                answers_seen.append(question["id"])
                return {
                    "id": question["id"],
                    "status": "ok",
                    "answer": "pottery",
                    "model": "brainy-recall+answer",
                    "latency_ms": 1.0,
                }

            def judge_fn(payload: dict) -> dict:
                qid = payload["question"]["id"]
                judge_calls.append(qid)
                if qid == "q2" and judge_calls.count("q2") == 1:
                    return {
                        "id": qid,
                        "judgment": JUDGMENT_UNRESOLVED,
                        "score": 0.0,
                        "reason": "bad json",
                        "model": "judge",
                    }
                return {
                    "id": qid,
                    "judgment": "CORRECT",
                    "score": 1.0,
                    "reason": "ok",
                    "model": "judge",
                }

            runner = _runner(root, answer_fn=answer_fn, judge_fn=judge_fn)
            runner.preflight(questions)
            runner.enqueue({})
            runner.drain()
            runner.verify_store(runner.store.read_json("identity.json") or {})
            runner.answer(questions, {})
            runner.judge(questions, {})
            with self.assertRaises(StagedRunError) as ctx:
                runner.finalize(questions, {"protocol": PROTOCOL_V2, "product_sha": "bbe55f7"})
            self.assertIn("unresolved", str(ctx.exception).lower())

            runner.answer(questions, {})
            runner.judge(questions, {})
            result = runner.finalize(questions, {"protocol": PROTOCOL_V2, "product_sha": "bbe55f7"})
            self.assertEqual(answers_seen, ["q1", "q2"])
            self.assertEqual(judge_calls.count("q1"), 1)
            self.assertEqual(judge_calls.count("q2"), 2)
            self.assertEqual(result["metrics"]["correct"], 2)
            self.assertTrue((root / "result.json").exists())

    def test_transport_failure_blocks_finalize_and_keeps_denominator(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            questions = [_question("q1")]
            runner = _runner(
                root,
                answer_fn=lambda q: {
                    "id": q["id"],
                    "status": JUDGMENT_TRANSPORT,
                    "answer": "",
                    "model": "brainy-recall+transport",
                    "reason": "timeout",
                },
            )
            ident = runner.preflight(questions)
            runner.enqueue(ident)
            runner.drain()
            runner.verify_store(ident)
            runner.answer(questions, ident)
            runner.judge(questions, ident)
            with self.assertRaises(StagedRunError) as ctx:
                runner.finalize(questions, ident)
            self.assertIn("transport", str(ctx.exception).lower())
            self.assertIn("q1", runner.store.latest_by_id("answers.jsonl"))

    def test_verify_store_requires_searchable_index(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            runner = _runner(
                pathlib.Path(tmp),
                search_fn=lambda prefix, subject, q: [],
            )
            ident = runner.preflight([_question()])
            with self.assertRaises(StagedRunError) as ctx:
                runner.verify_store(ident)
            self.assertIn("not searchable", str(ctx.exception))

    def test_store_identity_changes_with_database(self) -> None:
        a = store_identity_from_runtime(_runtime(database="brainy_n1540"))
        b = store_identity_from_runtime(_runtime(database="other"))
        self.assertNotEqual(a, b)


class PublishJobAccountingTests(unittest.TestCase):
    def test_wait_until_jobs_done_rejects_unaccounted_jobs(self) -> None:
        from http.server import BaseHTTPRequestHandler, HTTPServer
        import threading
        from public.backends.brainy import BrainyBackend

        class Handler(BaseHTTPRequestHandler):
            def do_GET(self):  # noqa: N802
                if self.path.startswith("/jobs/job_ok"):
                    body = b'{"job_id":"job_ok","status":"completed"}'
                elif self.path.startswith("/jobs/status"):
                    body = b'{"pending":0,"in_progress":0,"failed":0,"completed":1,"open":0}'
                else:
                    self.send_response(404)
                    self.end_headers()
                    return
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(body)

            def log_message(self, *_args):  # noqa: ANN002
                return

        server = HTTPServer(("127.0.0.1", 0), Handler)
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        try:
            base = f"http://127.0.0.1:{server.server_port}"
            backend = BrainyBackend(base, async_ingest=True, publish_mode=True, async_timeout_s=2.0)
            with self.assertRaises(RuntimeError) as ctx:
                backend.wait_until_jobs_done("u1", job_ids=["job_ok", "job_missing"], timeout_s=1.0)
            msg = str(ctx.exception).lower()
            self.assertTrue("unavailable" in msg or "not all accounted" in msg)
        finally:
            server.shutdown()


if __name__ == "__main__":
    unittest.main()
