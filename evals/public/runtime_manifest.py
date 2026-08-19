"""Fetch and pin GET /runtime without ever storing API keys."""

from __future__ import annotations

import json
import os
from typing import Any

import urllib.request


_SECRET_KEYS = {
    "api_key",
    "apikey",
    "authorization",
    "token",
    "secret",
    "password",
    "bearer",
}


def fetch_runtime(base_url: str, *, timeout: float = 15.0) -> dict[str, Any]:
    url = base_url.rstrip("/") + "/runtime"
    req = urllib.request.Request(url, method="GET")
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        payload = json.loads(resp.read().decode("utf-8"))
    return sanitize_runtime(payload)


def sanitize_runtime(payload: Any) -> Any:
    """Drop secret-shaped fields. Never persist an API key."""
    if isinstance(payload, dict):
        out = {}
        for key, value in payload.items():
            lowered = str(key).lower()
            if lowered in _SECRET_KEYS or "api_key" in lowered or lowered.endswith("_key"):
                continue
            out[key] = sanitize_runtime(value)
        return out
    if isinstance(payload, list):
        return [sanitize_runtime(item) for item in payload]
    if isinstance(payload, str):
        stripped = payload.strip()
        if stripped.startswith("sk-") or stripped.lower().startswith("bearer "):
            raise RuntimeError("runtime payload looks like it contains a secret")
        return payload
    return payload


def attach_runtime_extras(extras: dict[str, Any], runtime: dict[str, Any]) -> dict[str, Any]:
    out = dict(extras)
    out["runtime"] = sanitize_runtime(runtime) if isinstance(runtime, dict) else {}
    publish = os.environ.get("BRAINY_EVAL_PUBLISH", "").lower() in {"1", "true", "yes", "on"}
    if publish:
        out["fail_closed"] = True
        out["publish_mode"] = True
    return out
