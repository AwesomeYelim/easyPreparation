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


def open_checked(req: urllib.request.Request, timeout: int, step: str):
    # 400 같은 오류가 어느 단계인지, 서버가 준 사유와 함께 보이게 해
    try:
        return urllib.request.urlopen(req, timeout=timeout)
    except urllib.error.HTTPError as err:
        detail = err.read().decode("utf-8", "replace")[:300]
        raise OSError(f"{step} 실패 HTTP {err.code}: {detail}") from err


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
    with open_checked(urllib.request.Request(TOKEN_URL, data=body), 30, "토큰 갱신") as resp:
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
    with open_checked(start, 60, "업로드 세션 시작") as resp:
        session_url = resp.headers["Location"]
    with video.open("rb") as fh:
        put = urllib.request.Request(session_url, data=fh, method="PUT", headers={
            "Content-Length": str(size),
            "Content-Type": "video/mp4",
        })
        with open_checked(put, 1800, "파일 전송") as resp:
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


def run_queue(queue_path: Path) -> int:
    # 큐 JSON 의 status 가 ready 인 항목만 올리고, 성공하면 uploaded 와 video_id 를 적어
    data = json.loads(queue_path.read_text(encoding="utf-8-sig"))
    failed = 0
    for item in data["items"]:
        if item.get("status") != "ready":
            continue
        video = Path(item["file"])
        if not video.is_absolute():
            video = (queue_path.parent / video).resolve()
        try:
            prev = already_uploaded(video)
            video_id = prev or upload(video, item["title"], item["description"], item["tags"], item.get("privacy", "private"))
        except (ValueError, OSError, KeyError, urllib.error.URLError) as err:
            print(f"{item['key']} 실패: {err}", file=sys.stderr)
            failed += 1
            continue
        if not prev:
            record(video, video_id, item["title"], item.get("privacy", "private"))
        item["status"] = "uploaded"
        item["video_id"] = video_id
        queue_path.write_text(json.dumps(data, ensure_ascii=False, indent=2), encoding="utf-8")
        print(f"{item['key']} https://youtu.be/{video_id}")
    return 1 if failed else 0


def main(argv: list[str]) -> int:
    if len(argv) == 2 and argv[0] == "--queue":
        return run_queue(Path(argv[1]))
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
