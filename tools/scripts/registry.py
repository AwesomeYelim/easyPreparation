"""
문서 레지스트리 — Drive 파일 ID 로컬 캐시
저장 위치: tools/output/doc_registry.json  (gitignore, 로컬 전용)

사용법:
    from registry import get_id, save_id

    file_id = get_id("onboarding_guide")   # 없으면 None
    save_id("onboarding_guide", "1gCrp...")
"""

import json
import os

_REGISTRY_PATH = os.path.join(os.path.dirname(__file__), "..", "output", "doc_registry.json")


def _load() -> dict:
    if os.path.exists(_REGISTRY_PATH):
        with open(_REGISTRY_PATH, encoding="utf-8") as f:
            return json.load(f)
    return {}


def _save(data: dict) -> None:
    os.makedirs(os.path.dirname(_REGISTRY_PATH), exist_ok=True)
    with open(_REGISTRY_PATH, "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=2)


def get_id(key: str) -> str | None:
    """키에 해당하는 Drive 파일 ID 반환. 없으면 None."""
    return _load().get(key)


def save_id(key: str, file_id: str) -> None:
    """키 → 파일 ID 저장."""
    data = _load()
    data[key] = file_id
    _save(data)


def list_all() -> dict:
    """전체 레지스트리 반환."""
    return _load()


def remove(key: str) -> None:
    """키 삭제."""
    data = _load()
    data.pop(key, None)
    _save(data)
