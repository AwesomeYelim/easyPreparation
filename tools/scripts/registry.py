"""
문서 레지스트리 — Drive 파일 ID + 생성 이력 로컬 캐시
저장 위치: tools/cache/doc_registry.json  (gitignore, 로컬 전용)

스키마 v2:
{
  "key": {
    "file_id": "Drive파일ID",
    "url":     "Drive URL",
    "history": [
      {
        "v":      1,
        "at":     "2026-03-24T10:00:00",
        "git":    "1c0ef97a",
        "author": "홍예림",
        "url":    "https://..."
      }
    ]
  }
}

v1 호환: 값이 문자열이면 file_id만 있는 구 형식으로 자동 변환.
"""

import json
import os
import subprocess
from datetime import datetime

_REGISTRY_PATH = os.path.join(os.path.dirname(__file__), "..", "cache", "doc_registry.json")


# ── 내부 헬퍼 ─────────────────────────────────────────────

def _load() -> dict:
    if os.path.exists(_REGISTRY_PATH):
        with open(_REGISTRY_PATH, encoding="utf-8") as f:
            return json.load(f)
    return {}


def _save(data: dict) -> None:
    os.makedirs(os.path.dirname(_REGISTRY_PATH), exist_ok=True)
    with open(_REGISTRY_PATH, "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=2)


def _to_entry(raw) -> dict:
    """v1 string → v2 entry dict 투명 변환."""
    if isinstance(raw, str):
        return {"file_id": raw, "url": f"https://drive.google.com/file/d/{raw}", "history": []}
    return raw


def _git_info() -> tuple[str, str]:
    """(short_hash, author) — 실패 시 빈 문자열."""
    try:
        h = subprocess.check_output(
            ["git", "rev-parse", "--short", "HEAD"],
            stderr=subprocess.DEVNULL, text=True
        ).strip()
    except Exception:
        h = ""
    try:
        a = subprocess.check_output(
            ["git", "config", "user.name"],
            stderr=subprocess.DEVNULL, text=True
        ).strip()
    except Exception:
        a = ""
    return h, a


# ── 공개 API ──────────────────────────────────────────────

def get_entry(key: str) -> dict | None:
    """키에 해당하는 전체 엔트리 반환. 없으면 None."""
    raw = _load().get(key)
    return _to_entry(raw) if raw is not None else None


def get_id(key: str) -> str | None:
    """키에 해당하는 Drive 파일 ID 반환. 없으면 None."""
    entry = get_entry(key)
    return entry["file_id"] if entry else None


def save_id(key: str, file_id: str) -> None:
    """키 → 파일 ID 저장 (하위 호환 — save_entry 사용 권장)."""
    entry = get_entry(key) or {"file_id": "", "url": "", "history": []}
    entry["file_id"] = file_id
    data = _load()
    data[key] = entry
    _save(data)


def save_entry(key: str, file_id: str, url: str,
               git_hash: str = None, author: str = None) -> None:
    """엔트리 저장 + 이력 추가.

    git_hash / author 미지정 시 git 환경에서 자동 수집.
    """
    data = _load()
    entry = _to_entry(data.get(key)) if key in data else {"file_id": "", "url": "", "history": []}
    entry["file_id"] = file_id
    entry["url"] = url

    if git_hash is None or author is None:
        auto_hash, auto_author = _git_info()
        git_hash  = git_hash  if git_hash  is not None else auto_hash
        author    = author    if author    is not None else auto_author

    v = len(entry["history"]) + 1
    entry["history"].append({
        "v":      v,
        "at":     datetime.now().strftime("%Y-%m-%dT%H:%M:%S"),
        "git":    git_hash,
        "author": author,
        "url":    url,
    })

    data[key] = entry
    _save(data)


def history(key: str) -> list:
    """키의 이력 목록 반환."""
    entry = get_entry(key)
    return entry["history"] if entry else []


def list_all() -> dict:
    """전체 레지스트리 반환 (v2 형식 정규화)."""
    return {k: _to_entry(v) for k, v in _load().items()}


def remove(key: str) -> None:
    """키 삭제."""
    data = _load()
    data.pop(key, None)
    _save(data)


def print_history(key: str = None) -> None:
    """이력 출력. key=None 이면 전체 요약."""
    all_data = list_all()
    if not all_data:
        print("레지스트리가 비어 있습니다.")
        return

    if key:
        entry = all_data.get(key)
        if not entry:
            print(f"❌ 키를 찾을 수 없습니다: {key}")
            return
        hist = entry.get("history", [])
        print(f"\n📋 [{key}]  버전 이력  (총 {len(hist)}건)")
        print(f"   Drive URL : {entry['url']}")
        print(f"   File ID   : {entry['file_id']}")
        print()
        if not hist:
            print("  (이력 없음 — v1 포맷으로 저장된 항목)")
            return
        for h in reversed(hist):
            git_tag = h['git'][:7] if h.get('git') else "─"
            print(f"  v{h['v']:>3}  {h['at']}  git:{git_tag:<7}  {h.get('author') or '─'}")
            print(f"         {h['url']}")
    else:
        print(f"\n📋 문서 레지스트리  ({len(all_data)}건)\n")
        fmt = "  {:<38} {:>3}버전  {}  git:{}"
        for k, entry in sorted(all_data.items()):
            hist = entry.get("history", [])
            last = hist[-1] if hist else None
            last_at  = last["at"][:16]  if last else "─" * 16
            last_git = last["git"][:7]  if last and last.get("git") else "─"
            print(fmt.format(k, len(hist), last_at, last_git))
            print(f"    {entry['url']}")
        print()
