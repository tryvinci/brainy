from __future__ import annotations

import json
import pathlib
import random
import urllib.request
from collections import defaultdict

from ..proveability import sha256_file
from ..schema import CATEGORIES_TO_SCORE, CATEGORY_NAMES

LOCOMO_DATASET_URL = (
    "https://raw.githubusercontent.com/snap-research/locomo/main/data/locomo10.json"
)
LOCOMO_UPSTREAM = "https://github.com/snap-research/locomo"
LOCOMO_PAPER = "https://aclanthology.org/2024.acl-long.747/"


def default_dataset_path(root: pathlib.Path | None = None) -> pathlib.Path:
    base = root or pathlib.Path(__file__).resolve().parents[3]
    return base / "datasets" / "locomo" / "locomo10.json"


def ensure_dataset(path: pathlib.Path | None = None) -> tuple[pathlib.Path, str]:
    target = path or default_dataset_path()
    target.parent.mkdir(parents=True, exist_ok=True)
    if not target.exists():
        with urllib.request.urlopen(LOCOMO_DATASET_URL, timeout=120) as resp:
            target.write_bytes(resp.read())
    return target, sha256_file(target)


def load_conversations(path: pathlib.Path) -> list[dict]:
    data = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(data, list):
        raise ValueError("expected LOCOMO JSON list")
    return data


def iter_sessions(conversation: dict) -> list[dict]:
    """Yield sessions with optional client-style event metadata.

    Each item: {session_id, observed_at, turns: [(speaker, text[, image_urls]), ...]}.
    observed_at is the LOCOMO session_*_date_time string when present —
    realistic client metadata, not a product special-case.
    """
    convo = conversation.get("conversation") or {}
    session_nums: list[int] = []
    for key, value in convo.items():
        if not key.startswith("session_"):
            continue
        if key.endswith("_date_time"):
            continue
        if not isinstance(value, list):
            continue
        suffix = key[len("session_") :]
        if suffix.isdigit():
            session_nums.append(int(suffix))
    session_nums = sorted(set(session_nums))

    sessions: list[dict] = []
    for num in session_nums:
        turns: list[tuple] = []
        for turn in convo.get(f"session_{num}") or []:
            if not isinstance(turn, dict):
                continue
            speaker = str(turn.get("speaker") or "user")
            text = str(turn.get("text") or "").strip()
            alts: list[str] = []
            for key in ("query", "blip_caption"):
                cap = str(turn.get(key) or "").strip()
                if not cap:
                    continue
                blob = f"{text} {' '.join(alts)}".lower()
                if cap.lower() not in blob:
                    alts.append(cap)
            if alts:
                extra = " ".join(f"[{a}]" for a in alts)
                text = f"{text} {extra}".strip() if text else extra
            raw_img = turn.get("img_url") or turn.get("image") or []
            urls: list[str] = []
            if isinstance(raw_img, str) and raw_img.strip():
                urls = [raw_img.strip()]
            elif isinstance(raw_img, list):
                urls = [str(u).strip() for u in raw_img if str(u).strip()]
            if text or urls:
                if urls:
                    turns.append((speaker, text, urls))
                else:
                    turns.append((speaker, text))
        if not turns:
            continue
        date_raw = convo.get(f"session_{num}_date_time")
        observed = str(date_raw).strip() if date_raw else ""
        sessions.append(
            {
                "session_id": f"session_{num}",
                "observed_at": observed,
                "turns": turns,
            }
        )
    return sessions


def iter_session_turns(conversation: dict) -> list[tuple[str, str]]:
    """Flatten all sessions into (speaker, text) turns."""
    turns: list[tuple[str, str]] = []
    for session in iter_sessions(conversation):
        for turn in session["turns"]:
            if not turn:
                continue
            speaker = turn[0]
            text = turn[1] if len(turn) > 1 else ""
            turns.append((str(speaker), str(text)))
    return turns


def iter_questions(conversation: dict) -> list[dict]:
    out = []
    for idx, qa in enumerate(conversation.get("qa") or []):
        if not isinstance(qa, dict):
            continue
        out.append(
            {
                "id": f"q{idx}",
                "question": str(qa.get("question") or ""),
                "answer": str(qa.get("answer") or ""),
                "category": qa.get("category"),
                "evidence": qa.get("evidence") or [],
            }
        )
    return out


def category_group(cat_id) -> str:
    try:
        cat_id_int = int(cat_id) if cat_id is not None else 0
    except (TypeError, ValueError):
        cat_id_int = 0
    return CATEGORY_NAMES.get(cat_id_int, f"cat-{cat_id}")


def scored_question_pool(conversations: list[dict]) -> list[dict]:
    """Flatten scored LoCoMo items (categories 1–4) with conversation identity."""
    pool: list[dict] = []
    for conv_idx, conversation in enumerate(conversations):
        sample_id = str(conversation.get("sample_id") or f"c{conv_idx}")
        for qa in iter_questions(conversation):
            try:
                cat_id = int(qa.get("category")) if qa.get("category") is not None else 0
            except (TypeError, ValueError):
                cat_id = 0
            if cat_id not in CATEGORIES_TO_SCORE:
                continue
            row = dict(qa)
            row["sample_id"] = sample_id
            row["conv_idx"] = conv_idx
            row["group"] = category_group(cat_id)
            pool.append(row)
    return pool


def stratified_questions(pool: list[dict], n: int, seed: int = 1) -> list[dict]:
    """Proportional SH/MH/temporal/OD sample. Stable under the same seed."""
    if n <= 0 or n >= len(pool):
        return list(pool)
    by_group: dict[str, list[dict]] = defaultdict(list)
    for q in pool:
        by_group[str(q.get("group") or "unknown")].append(q)
    rng = random.Random(seed)
    for rows in by_group.values():
        rng.shuffle(rows)
    total = len(pool)
    allot: dict[str, int] = {}
    remaining = n
    groups = sorted(by_group.keys())
    for i, g in enumerate(groups):
        if i == len(groups) - 1:
            allot[g] = remaining
        else:
            k = max(1, round(n * len(by_group[g]) / total))
            k = min(k, len(by_group[g]), remaining - (len(groups) - i - 1))
            allot[g] = k
            remaining -= k
    out: list[dict] = []
    for g in groups:
        out.extend(by_group[g][: allot[g]])
    rng.shuffle(out)
    return out[:n]
