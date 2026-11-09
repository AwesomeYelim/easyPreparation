"""쇼츠 업로드. 사용: upload.py <영상> --title T --desc-file D [--tags a,b,c] [--privacy private]

인증 파일(저장소 밖): AUTOSHORTS_GOOGLE_DIR 또는 ~/.config/autoshorts/ 아래
google_oauth.json (클라이언트 id/secret), youtube_token.json (refresh_token 포함)
"""
from __future__ import annotations

import argparse
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request
from datetime import datetime
from pathlib import Path

AUTH_DIR = Path(os.environ.get("AUTOSHORTS_GOOGLE_DIR", Path.home() / ".config" / "autoshorts"))
STATE_PATH = Path(os.environ.get("AUTOSHORTS_STATE_DIR", Path.home() / ".local" / "state" / "autoshorts")) / "uploads.json"
TOKEN_URL = "https://oauth2.googleapis.com/token"
UPLOAD_URL = "https://www.googleapis.com/upload/youtube/v3/videos?uploadType=resumable&part=snippet,status"


def load_client() -> dict:
    raw = json.loads((AUTH_DIR / "google_oauth.json").read_text(encoding="utf-8"))
    return raw.get("installed") or raw.get("web") or raw


def access_token() -> str:
    client = load_client()
    token_path = AUTH_DIR / "youtube_token.json"
    token = json.loads(token_path.read_text(encoding="utf-8"))
    body = urllib.parse.urlencode({
        "client_id": client["client_id"],
        "client_secret": client["client_secret"],
        "refresh_token": token["refresh_token"],
        "grant_type": "refresh_token",
    }).encode()
    with urllib.request.urlopen(urllib.request.Request(TOKEN_URL, data=body), timeout=30) as resp:
        return json.loads(resp.read())["access_token"]


def check_meta(title: str, tags: list[str]) -> None:
    if not title or len(title) > 100:
        raise ValueError(f"제목은 1~100자여야 해 (지금 {len(title)}자)")
    if sum(len(t) + 1 for t in tags) > 500:
        raise ValueError("태그 합계가 500자를 넘어")


def upload(video: Path, title: str, desc: str, tags: list[str], privacy: str) -> str:
    check_meta(title, tags)
    token = access_token()
    meta = {
        "snippet": {"title": title, "description": desc, "tags": tags, "categoryId": "29"},
        "status": {"privacyStatus": privacy, "selfDeclaredMadeForKids": False},
    }
    size = video.stat().st_size
    start = urllib.request.Request(UPLOAD_URL, data=json.dumps(meta).encode("utf-8"), method="POST", headers={
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json; charset=UTF-8",
        "X-Upload-Content-Length": str(size),
        "X-Upload-Content-Type": "video/mp4",
    })
    with urllib.request.urlopen(start, timeout=60) as resp:
        session_url = resp.headers["Location"]
    with video.open("rb") as fh:
        put = urllib.request.Request(session_url, data=fh, method="PUT", headers={
            "Content-Length": str(size),
            "Content-Type": "video/mp4",
        })
        with urllib.request.urlopen(put, timeout=1800) as resp:
            return json.loads(resp.read())["id"]


def record(video: Path, video_id: str, title: str, privacy: str) -> None:
    STATE_PATH.parent.mkdir(parents=True, exist_ok=True)
    rows = json.loads(STATE_PATH.read_text(encoding="utf-8")) if STATE_PATH.exists() else []
    rows.append({"file": str(video), "id": video_id, "title": title, "privacy": privacy, "at": f"{datetime.now():%Y-%m-%d %H:%M}"})
    STATE_PATH.write_text(json.dumps(rows, ensure_ascii=False, indent=2), encoding="utf-8")


def already_uploaded(video: Path) -> str:
    if not STATE_PATH.exists():
        return ""
    for row in json.loads(STATE_PATH.read_text(encoding="utf-8")):
        if row["file"] == str(video):
            return row["id"]
    return ""


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(description="쇼츠 업로드")
    parser.add_argument("video")
    parser.add_argument("--title", required=True)
    parser.add_argument("--desc-file", required=True)
    parser.add_argument("--tags", default="")
    parser.add_argument("--privacy", default="private", choices=["private", "unlisted", "public"])
    parser.add_argument("--check-only", action="store_true", help="제목·태그 검사만 하고 올리지 않아")
    args = parser.parse_args(argv)
    video = Path(args.video).resolve()
    tags = [t.strip() for t in args.tags.split(",") if t.strip()]
    desc = Path(args.desc_file).read_text(encoding="utf-8")
    try:
        check_meta(args.title, tags)
        if args.check_only:
            print("검사 통과")
            return 0
        prev = already_uploaded(video)
        if prev:
            print(f"이미 올린 파일이야: https://youtu.be/{prev}")
            return 0
        video_id = upload(video, args.title, desc, tags, args.privacy)
    except (ValueError, OSError, KeyError, urllib.error.URLError) as err:
        print(f"오류: {err}", file=sys.stderr)
        return 1
    record(video, video_id, args.title, args.privacy)
    print(f"https://youtu.be/{video_id}")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
