"""OpenAI-compatible chat client for answerer + judge.

Works with OpenAI, Ollama, vLLM, LM Studio, Together, OpenRouter, etc.

Env:
  LLM_BASE_URL   default https://api.openai.com/v1
  LLM_API_KEY    optional; falls back to OPENAI_API_KEY; local Ollama accepts \"ollama\"
  LLM_MODEL      optional default model name
  LLM_JSON_MODE  \"1\" to send response_format=json_object (OpenAI / some hosts only)
"""
from __future__ import annotations

import json
import os
import re
import time
import urllib.error
import urllib.request
from dataclasses import dataclass


DEFAULT_BASE_URL = "https://api.openai.com/v1"


@dataclass(frozen=True)
class LLMConfig:
    base_url: str
    api_key: str
    model: str
    json_mode: bool
    timeout_s: float = 180.0

    @property
    def label(self) -> str:
        host = self.base_url.rstrip("/")
        return f"{self.model}@{host}"


def resolve_config(
    *,
    model: str = "",
    base_url: str = "",
    api_key: str = "",
    json_mode: bool | None = None,
) -> LLMConfig | None:
    """Return config if an LLM endpoint is usable; else None (lexical fallback)."""
    base = (base_url or os.environ.get("LLM_BASE_URL") or DEFAULT_BASE_URL).rstrip("/")
    key = (
        api_key
        or os.environ.get("LLM_API_KEY", "").strip()
        or os.environ.get("OPENAI_API_KEY", "").strip()
    )
    local = _is_local(base)
    if not key and local:
        key = "ollama"
    if not key and not local:
        return None

    chosen_model = (
        model
        or os.environ.get("LLM_MODEL", "").strip()
        or ("llama3.1" if local else "gpt-4o-mini")
    )
    if json_mode is None:
        env_flag = os.environ.get("LLM_JSON_MODE", "").strip().lower()
        if env_flag in {"1", "true", "yes"}:
            json_mode = True
        elif env_flag in {"0", "false", "no"}:
            json_mode = False
        else:
            # OpenAI hosts support json_object; most OSS/Ollama do not.
            json_mode = "openai.com" in base

    return LLMConfig(base_url=base, api_key=key, model=chosen_model, json_mode=json_mode)


def _is_local(base: str) -> bool:
    lowered = base.lower()
    return any(
        host in lowered
        for host in ("127.0.0.1", "localhost", "0.0.0.0", "host.docker.internal")
    )


def chat_completion(
    messages: list[dict[str, str]],
    config: LLMConfig,
    *,
    temperature: float = 0.0,
    force_json: bool = False,
) -> str:
    url = f"{config.base_url}/chat/completions"
    payload: dict = {
        "model": config.model,
        "temperature": temperature,
        "messages": messages,
        # Reasoning models (e.g. gpt-oss) spend tokens on `reasoning` before
        # `content`; a low default leaves content null.
        "max_tokens": 2048,
    }
    use_json = force_json and config.json_mode
    if use_json:
        payload["response_format"] = {"type": "json_object"}

    body = _post(url, payload, config)
    message = (body.get("choices") or [{}])[0].get("message") or {}
    content = message.get("content")
    if isinstance(content, list):
        # Some providers return content parts.
        content = "".join(
            part.get("text", "") if isinstance(part, dict) else str(part) for part in content
        )
    if content is None or str(content).strip() == "" or str(content).strip().lower() == "none":
        # Last-resort: some gateways put usable text only in reasoning fields.
        content = message.get("reasoning_content") or message.get("reasoning") or ""
    return str(content).strip()


def _post(
    url: str,
    payload: dict,
    config: LLMConfig,
    *,
    _dropped_json_mode: bool = False,
    _attempt: int = 0,
) -> dict:
    data = json.dumps(payload).encode("utf-8")
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {config.api_key}",
        # CF AI Gateway WAF (error 1010) blocks non-browser/curl UAs from some
        # egress IPs; curl UA keeps chat/completions reachable for evals.
        "User-Agent": "curl/8.5.0",
    }
    req = urllib.request.Request(url, data=data, headers=headers, method="POST")
    try:
        with urllib.request.urlopen(req, timeout=config.timeout_s) as resp:
            return json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode("utf-8", errors="replace")[:500]
        detail_l = detail.lower()
        # Retry once without response_format if the host rejected json_object.
        if (
            not _dropped_json_mode
            and payload.get("response_format")
            and exc.code in {400, 404, 422}
            and ("response_format" in detail_l or "json_object" in detail_l)
        ):
            retry = dict(payload)
            retry.pop("response_format", None)
            return _post(url, retry, config, _dropped_json_mode=True, _attempt=_attempt)
        # Transient gateway / WAF / rate-limit (Cloudflare 1010 shows as 403).
        if exc.code in {403, 408, 429, 500, 502, 503, 504} and _attempt < 5:
            time.sleep(1.5 * (2**_attempt))
            return _post(
                url,
                payload,
                config,
                _dropped_json_mode=_dropped_json_mode,
                _attempt=_attempt + 1,
            )
        raise RuntimeError(f"LLM HTTP {exc.code} at {url}: {detail}") from exc
    except TimeoutError as exc:
        if _attempt < 5:
            time.sleep(1.5 * (2**_attempt))
            return _post(
                url,
                payload,
                config,
                _dropped_json_mode=_dropped_json_mode,
                _attempt=_attempt + 1,
            )
        raise RuntimeError(f"LLM timeout at {url}") from exc


def parse_judgment_json(content: str) -> dict:
    """Extract judgment JSON from model output (handles fences / extra prose)."""
    text = content.strip()
    try:
        return json.loads(text)
    except json.JSONDecodeError:
        pass
    fence = re.search(r"```(?:json)?\s*(\{.*?\})\s*```", text, re.DOTALL | re.IGNORECASE)
    if fence:
        return json.loads(fence.group(1))
    brace = re.search(r"\{[^{}]*\"judgment\"[^{}]*\}", text, re.DOTALL | re.IGNORECASE)
    if brace:
        return json.loads(brace.group(0))
    raise ValueError(f"no judgment JSON in model output: {text[:200]}")
