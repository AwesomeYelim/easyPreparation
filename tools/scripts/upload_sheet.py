"""
로컬 파일을 Google Drive에 업로드 (신규 생성 또는 기존 파일 업데이트)
usage: python upload_sheet.py <파일경로> [--folder <폴더ID>] [--name <파일명>] [--key <레지스트리키>]
출력: Google Drive URL (stdout)

  --key KEY   레지스트리 키. 지정 시:
                - 키가 있으면 → 기존 파일 내용 업데이트 (파일 ID / URL 유지)
                - 키가 없으면 → 신규 생성 후 키 저장

인증: tools/cache/.gtoken 파일 (python tools/auth.py 로 초기 인증)
"""

import sys
import os
import argparse
import json
import re
import time
import urllib.parse
import urllib.request
import urllib.error
from datetime import datetime


# ── OAuth 토큰 ────────────────────────────────────────────

_TOKEN_URI   = "https://oauth2.googleapis.com/token"
_TOKEN_CACHE = os.path.realpath(
    os.path.join(os.path.dirname(__file__), "..", "cache", ".gtoken")
)


def _save_token(token: dict) -> None:
    os.makedirs(os.path.dirname(_TOKEN_CACHE), exist_ok=True)
    with open(_TOKEN_CACHE, "w", encoding="utf-8") as f:
        json.dump(token, f)


def _get_client_creds(token: dict) -> tuple[str, str]:
    """.gtoken 또는 환경변수에서 client_id / client_secret 반환."""
    # .gtoken에 저장된 값 우선 (auth.py --setup으로 설정한 경우)
    cid     = token.get("client_id", "").strip()
    csecret = token.get("client_secret", "").strip()
    if cid and csecret:
        return cid, csecret
    # 환경변수 fallback
    cid     = os.environ.get("GOOGLE_CLIENT_ID", "").strip()
    csecret = os.environ.get("GOOGLE_CLIENT_SECRET", "").strip()
    return cid, csecret


def _refresh_token(token: dict) -> str:
    """refresh_token으로 새 access_token 획득."""
    client_id, client_secret = _get_client_creds(token)
    if not client_id or not client_secret:
        return None

    body = urllib.parse.urlencode({
        "grant_type":    "refresh_token",
        "refresh_token": token["refresh_token"],
        "client_id":     client_id,
        "client_secret": client_secret,
    }).encode()
    req = urllib.request.Request(
        _TOKEN_URI, data=body,
        headers={"Content-Type": "application/x-www-form-urlencoded"},
    )
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            new_data = json.loads(resp.read())
        token["access_token"] = new_data["access_token"]
        token["expiry_date"]  = int(time.time() * 1000) + new_data.get("expires_in", 3600) * 1000
        _save_token(token)
        return token["access_token"]
    except (urllib.error.HTTPError, urllib.error.URLError):
        return None


def get_access_token(force: bool = False) -> str:
    """
    유효한 access_token 반환. 만료(또는 force=True) 시 자동 갱신.
    credentials는 .gtoken 내 client_id/secret 또는 환경변수에서 읽음.
    """
    try:
        with open(_TOKEN_CACHE, "r", encoding="utf-8") as f:
            token = json.load(f)

        if not force and token.get("expiry_date", 0) >= (time.time() * 1000 + 300_000):
            return token["access_token"]

        result = _refresh_token(token)
        if result:
            return result
    except (FileNotFoundError, json.JSONDecodeError):
        pass

    raise RuntimeError(
        "Google 인증 필요:\n"
        "  run.sh token refresh  (또는 직접: tools/auth.py --paste)"
    )


# ── Drive 업로드 / 업데이트 ───────────────────────────────

_CONVERT = {
    ".xlsx": ("application/vnd.google-apps.spreadsheet",
              "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"),
    ".xls":  ("application/vnd.google-apps.spreadsheet",
              "application/vnd.ms-excel"),
    ".pptx": ("application/vnd.google-apps.presentation",
              "application/vnd.openxmlformats-officedocument.presentationml.presentation"),
    ".ppt":  ("application/vnd.google-apps.presentation",
              "application/vnd.ms-powerpoint"),
    ".docx": ("application/vnd.google-apps.document",
              "application/vnd.openxmlformats-officedocument.wordprocessingml.document"),
    ".doc":  ("application/vnd.google-apps.document",
              "application/msword"),
}

_MIME_URL = {
    "application/vnd.google-apps.spreadsheet": "https://docs.google.com/spreadsheets/d/{}",
    "application/vnd.google-apps.presentation": "https://docs.google.com/presentation/d/{}",
    "application/vnd.google-apps.document":     "https://docs.google.com/document/d/{}",
}


def _build_multipart(metadata: dict, src_mime: str, file_data: bytes) -> tuple[bytes, str]:
    boundary = "==mcp_upload_boundary=="
    meta_bytes = json.dumps(metadata).encode("utf-8")
    body = (
        f"--{boundary}\r\n"
        f"Content-Type: application/json; charset=UTF-8\r\n\r\n"
    ).encode() + meta_bytes + (
        f"\r\n--{boundary}\r\n"
        f"Content-Type: {src_mime}\r\n\r\n"
    ).encode() + file_data + f"\r\n--{boundary}--".encode()
    return body, boundary


def _file_url(file_id: str, mime_type: str) -> str:
    tmpl = _MIME_URL.get(mime_type, "https://drive.google.com/file/d/{}")
    return tmpl.format(file_id)


def upload(local_path: str, name: str = None, folder_id: str = None) -> str:
    """
    로컬 파일을 Google Drive에 신규 업로드.
    오피스 포맷(.pptx/.docx/.xlsx)은 Google 포맷으로 자동 변환.
    반환값: Drive URL
    """
    if not os.path.exists(local_path):
        raise FileNotFoundError(f"파일을 찾을 수 없습니다: {local_path}")

    access_token = get_access_token()
    file_name = name or os.path.basename(local_path)
    ext = os.path.splitext(local_path)[1].lower()

    if ext in _CONVERT:
        target_mime, src_mime = _CONVERT[ext]
    else:
        target_mime, src_mime = None, "application/octet-stream"

    with open(local_path, "rb") as f:
        file_data = f.read()

    metadata: dict = {"name": file_name}
    if target_mime:
        metadata["mimeType"] = target_mime
    if folder_id:
        metadata["parents"] = [folder_id]

    body, boundary = _build_multipart(metadata, src_mime, file_data)
    req = urllib.request.Request(
        "https://www.googleapis.com/upload/drive/v3/files?uploadType=multipart&fields=id,mimeType&supportsAllDrives=true",
        data=body,
        headers={
            "Authorization": f"Bearer {access_token}",
            "Content-Type": f"multipart/related; boundary={boundary}",
        },
        method="POST",
    )
    with urllib.request.urlopen(req) as resp:
        result = json.loads(resp.read())

    return _file_url(result["id"], result.get("mimeType", ""))


def update(file_id: str, local_path: str, name: str = None) -> str:
    """
    Drive의 기존 파일 내용을 새 로컬 파일로 교체.
    파일 ID / URL이 유지됨.
    반환값: Drive URL (변경 없음)
    """
    if not os.path.exists(local_path):
        raise FileNotFoundError(f"파일을 찾을 수 없습니다: {local_path}")

    access_token = get_access_token()
    file_name = name or os.path.basename(local_path)
    ext = os.path.splitext(local_path)[1].lower()
    _, src_mime = _CONVERT.get(ext, (None, "application/octet-stream"))

    with open(local_path, "rb") as f:
        file_data = f.read()

    # 업데이트 시 parents / mimeType 생략 (기존 설정 유지)
    metadata = {"name": file_name}
    body, boundary = _build_multipart(metadata, src_mime, file_data)
    req = urllib.request.Request(
        f"https://www.googleapis.com/upload/drive/v3/files/{file_id}"
        "?uploadType=multipart&fields=id,mimeType&supportsAllDrives=true",
        data=body,
        headers={
            "Authorization": f"Bearer {access_token}",
            "Content-Type": f"multipart/related; boundary={boundary}",
        },
        method="PATCH",
    )
    with urllib.request.urlopen(req) as resp:
        result = json.loads(resp.read())

    return _file_url(result["id"], result.get("mimeType", ""))


def move_to_folder(file_id: str, folder_id: str) -> None:
    """업로드된 파일을 지정 폴더로 이동"""
    access_token = get_access_token()

    # 현재 부모 조회
    req = urllib.request.Request(
        f"https://www.googleapis.com/drive/v3/files/{file_id}?fields=parents&supportsAllDrives=true",
        headers={"Authorization": f"Bearer {access_token}"},
    )
    with urllib.request.urlopen(req) as resp:
        old_parents = ",".join(json.loads(resp.read()).get("parents", []))

    url = (
        f"https://www.googleapis.com/drive/v3/files/{file_id}"
        f"?addParents={folder_id}&removeParents={old_parents}&fields=id&supportsAllDrives=true"
    )
    req = urllib.request.Request(
        url,
        data=b"{}",
        headers={
            "Authorization": f"Bearer {access_token}",
            "Content-Type": "application/json",
        },
        method="PATCH",
    )
    with urllib.request.urlopen(req):
        pass


# ── Sheets API ────────────────────────────────────────────

_SHEETS_API = "https://sheets.googleapis.com/v4/spreadsheets"


def _sheets_get(url: str, token: str) -> dict:
    req = urllib.request.Request(url, headers={"Authorization": f"Bearer {token}"})
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            return json.loads(resp.read())
    except urllib.error.HTTPError as e:
        body = e.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"Sheets API 오류 {e.code}: {body}") from e


def _extract_id(id_or_url: str) -> str:
    """URL 또는 ID에서 spreadsheetId 추출."""
    m = re.search(r"/spreadsheets/d/([a-zA-Z0-9_-]+)", id_or_url)
    return m.group(1) if m else id_or_url.strip()


def _extract_gid(url: str) -> str | None:
    """URL에서 gid 추출."""
    m = re.search(r"[#&?]gid=(\d+)", url)
    return m.group(1) if m else None


def get_sheet_names(spreadsheet_id: str, token: str) -> list[dict]:
    """시트 목록 반환: [{"title": "SA", "sheetId": 1568920909}, ...]"""
    url  = f"{_SHEETS_API}/{spreadsheet_id}?fields=sheets.properties"
    data = _sheets_get(url, token)
    return [
        {"title": s["properties"]["title"], "sheetId": s["properties"]["sheetId"]}
        for s in data.get("sheets", [])
    ]


def resolve_sheet_name(spreadsheet_id: str, token: str,
                       sheet: str = None, gid: str = None) -> str:
    """시트명 결정: 명시 > gid 검색 > 첫 번째 시트."""
    if sheet:
        return sheet
    sheets = get_sheet_names(spreadsheet_id, token)
    if not sheets:
        raise RuntimeError("시트 목록을 가져올 수 없습니다.")
    if gid:
        for s in sheets:
            if str(s["sheetId"]) == str(gid):
                return s["title"]
        raise RuntimeError(f"gid={gid} 에 해당하는 시트를 찾을 수 없습니다.")
    return sheets[0]["title"]


def fetch_sheet(spreadsheet_id: str, sheet_name: str, token: str,
                col_range: str = "A:P") -> list[list[str]]:
    """시트 전체 읽기. 빈 셀은 빈 문자열로 정규화."""
    encoded = urllib.parse.quote(f"{sheet_name}!{col_range}", safe="!:")
    url  = f"{_SHEETS_API}/{spreadsheet_id}/values/{encoded}"
    data = _sheets_get(url, token)
    rows = data.get("values", [])
    return [row + [""] * max(0, 16 - len(row)) for row in rows]


def _safe(row: list, idx: int) -> str:
    try:
        return str(row[idx]).strip()
    except IndexError:
        return ""


def rows_to_cases(rows: list[list[str]], skip: int = 3) -> list[dict]:
    """시트 행 배열 → cases[] 변환. 컬럼 매핑(0-index): 1=대분류 2=중분류 3=소분류 4=ID 5=타입
    6=예상flow 7=실행1차 8=날짜 9=결과 10=담당자 11=실행2차 12=수정날짜 13=결과2차 14=담당자2차"""
    data_rows = rows[skip:]
    cases = []
    prev_category = prev_sub_category = ""

    for row in data_rows:
        id_val = _safe(row, 4)
        if not id_val:
            continue
        category     = _safe(row, 1) or prev_category
        sub_category = _safe(row, 2) or prev_sub_category
        prev_category     = category
        prev_sub_category = sub_category
        cases.append({
            "id":              id_val,
            "category":        category,
            "sub_category":    sub_category,
            "name":            _safe(row, 3),
            "type":            _safe(row, 5),
            "expected_flow":   _safe(row, 6),
            "executed_flow":   _safe(row, 7),
            "date":            _safe(row, 8),
            "result":          _safe(row, 9) or "시험 전",
            "assignee":        _safe(row, 10),
            "retest_flow":     _safe(row, 11),
            "retest_date":     _safe(row, 12),
            "retest_result":   _safe(row, 13),
            "retest_assignee": _safe(row, 14),
        })
    return cases


def convert_sheet(spreadsheet_id: str, sheet_name: str, token: str,
                  skip: int = 3, target: str = "", version: str = "") -> dict:
    """시트 → cases[] JSON dict 반환."""
    rows  = fetch_sheet(spreadsheet_id, sheet_name, token)
    cases = rows_to_cases(rows, skip=skip)
    if not target:
        target = _safe(rows[1], 1) if len(rows) > 1 else sheet_name
    if not target:
        target = sheet_name
    return {
        "target":  target,
        "version": version,
        "date":    datetime.now().strftime("%Y-%m-%d"),
        "sheet":   sheet_name,
        "cases":   cases,
    }


# ── main ──────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(
        description="로컬 파일을 Google Drive에 업로드 (신규 또는 기존 파일 업데이트)"
    )
    parser.add_argument("file", help="업로드할 로컬 파일 경로")
    parser.add_argument("--folder", default=None, metavar="FOLDER_ID",
                        help="업로드할 Drive 폴더 ID (신규 생성 시에만 적용)")
    parser.add_argument("--name", default=None,
                        help="Drive에 저장될 파일명 (기본: 원본 파일명)")
    parser.add_argument("--key", default=None, metavar="KEY",
                        help="레지스트리 키. 지정 시 기존 파일 업데이트, 없으면 신규 생성 후 저장.")
    args = parser.parse_args()

    if args.key:
        import sys
        sys.path.insert(0, os.path.dirname(__file__))
        from registry import get_id, save_entry

        existing_id = get_id(args.key)
        if existing_id:
            print(f"[registry] update: {args.key} -> {existing_id}", file=sys.stderr)
            url = update(existing_id, args.file, name=args.name)
            save_entry(args.key, existing_id, url)
        else:
            print(f"[registry] create: {args.key}", file=sys.stderr)
            url = upload(args.file, name=args.name, folder_id=args.folder)
            file_id = url.rstrip("/").split("/")[-1]
            save_entry(args.key, file_id, url)
        print(url)
    else:
        url = upload(args.file, name=args.name, folder_id=args.folder)
        print(url)


if __name__ == "__main__":
    main()
