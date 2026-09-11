"""One resolved API client shared by ingest, status, search, and /recall."""
from __future__ import annotations

from typing import Any

from httputil import get_json, post_json


class BoundAPIClient:
    """Pin the API URL once. Do not re-read BRAINY_BASE_URL per call."""

    def __init__(
        self,
        base_url: str,
        *,
        timeout_s: float = 60.0,
        recall_timeout_s: float = 120.0,
    ) -> None:
        url = (base_url or "").strip().rstrip("/")
        if not url:
            raise ValueError("API URL is required")
        self.base_url = url
        self.timeout_s = float(timeout_s)
        self.recall_timeout_s = float(recall_timeout_s)

    def post_json(
        self,
        path: str,
        payload: dict[str, Any],
        *,
        timeout: float | None = None,
        retries: int = 5,
    ) -> dict[str, Any]:
        return post_json(
            self.base_url,
            path,
            payload,
            timeout=self.timeout_s if timeout is None else timeout,
            retries=retries,
        )

    def get_json(
        self,
        path: str,
        params: dict[str, str] | None = None,
        *,
        timeout: float | None = None,
    ) -> dict[str, Any]:
        return get_json(
            self.base_url,
            path,
            params or {},
            timeout=self.timeout_s if timeout is None else timeout,
        )

    def healthz(self) -> str:
        import urllib.request

        from httputil import auth_headers

        req = urllib.request.Request(
            f"{self.base_url}/healthz",
            headers=auth_headers(),
            method="GET",
        )
        with urllib.request.urlopen(req, timeout=15) as resp:
            return resp.read().decode("utf-8")

    def runtime(self) -> dict[str, Any]:
        from public.runtime_manifest import fetch_runtime

        return fetch_runtime(self.base_url)

    def job(self, job_id: str) -> dict[str, Any]:
        return self.get_json(f"/jobs/{job_id}", {}, timeout=30)

    def jobs_status(self, tenant_id: str, subject_id: str) -> dict[str, Any]:
        return self.get_json(
            "/jobs/status",
            {"tenant_id": tenant_id, "subject_id": subject_id},
            timeout=30,
        )

    def search(
        self,
        tenant_id: str,
        subject_id: str,
        query: str,
        *,
        timeout: float | None = None,
    ) -> dict[str, Any]:
        return self.get_json(
            "/memories/search",
            {"tenant_id": tenant_id, "subject_id": subject_id, "q": query},
            timeout=self.timeout_s if timeout is None else timeout,
        )

    def recall(
        self,
        tenant_id: str,
        subject_id: str,
        question: str,
        *,
        mode: str = "answer",
        top_k: int = 30,
        timeout: float | None = None,
        retries: int = 2,
    ) -> dict[str, Any]:
        return self.post_json(
            "/recall",
            {
                "tenant_id": tenant_id,
                "subject_id": subject_id,
                "q": question,
                "mode": mode,
                "top_k": top_k,
            },
            timeout=self.recall_timeout_s if timeout is None else timeout,
            retries=retries,
        )
