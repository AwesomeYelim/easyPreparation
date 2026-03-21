#!/usr/bin/env python3
"""
tools/gdrive_mcp.py — Google Drive/Sheets/Docs/Slides MCP Server

Node.js tools/gdrive-mcp/index.js + @presto-ai/google-workspace-mcp 통합 대체.

인증: tools/output/.gtoken (OAuth refresh token 파일)
초기 인증: python tools/auth.py
"""

import json
import os
import re
import time
import urllib.parse
import urllib.request
from pathlib import Path

from googleapiclient.discovery import build
from google.oauth2.credentials import Credentials
from mcp.server.fastmcp import FastMCP

# ── 상수 ──────────────────────────────────────────────────────────────────────

TOOLS_DIR  = Path(__file__).parent
TOKEN_PATH = TOOLS_DIR / "output" / ".gtoken"
TOKEN_URI  = "https://oauth2.googleapis.com/token"

mcp = FastMCP("gdrive")

# ── 인증 ──────────────────────────────────────────────────────────────────────

def _load_token() -> dict:
    if not TOKEN_PATH.exists():
        raise FileNotFoundError(
            f"OAuth 토큰 파일 없음: {TOKEN_PATH}\n"
            "초기 인증: python tools/auth.py"
        )
    return json.loads(TOKEN_PATH.read_text(encoding="utf-8"))


def _save_token(token: dict) -> None:
    TOKEN_PATH.parent.mkdir(parents=True, exist_ok=True)
    TOKEN_PATH.write_text(json.dumps(token), encoding="utf-8")


def _refresh(token: dict) -> str:
    client_id     = os.environ.get("GOOGLE_CLIENT_ID", "").strip()
    client_secret = os.environ.get("GOOGLE_CLIENT_SECRET", "").strip()
    if not client_id or not client_secret:
        raise RuntimeError(
            "환경변수 GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET 가 설정되지 않았습니다."
        )
    body = urllib.parse.urlencode({
        "grant_type":    "refresh_token",
        "refresh_token": token["refresh_token"],
        "client_id":     client_id,
        "client_secret": client_secret,
    }).encode()
    req = urllib.request.Request(
        TOKEN_URI, data=body,
        headers={"Content-Type": "application/x-www-form-urlencoded"},
    )
    with urllib.request.urlopen(req) as resp:
        data = json.loads(resp.read())
    token["access_token"] = data["access_token"]
    token["expiry_date"]  = int(time.time() * 1000) + data.get("expires_in", 3600) * 1000
    _save_token(token)
    return token["access_token"]


def _creds() -> Credentials:
    token      = _load_token()
    is_expired = token.get("expiry_date", 0) < (time.time() * 1000 + 300_000)
    access_token = _refresh(token) if is_expired else token["access_token"]
    return Credentials(token=access_token)


def _svc(name: str, version: str):
    return build(name, version, credentials=_creds(), cache_discovery=False)


def _id(url_or_id: str) -> str:
    m = re.search(r"[-\w]{25,}", url_or_id)
    return m.group(0) if m else url_or_id


# ── People / Auth ──────────────────────────────────────────────────────────────

@mcp.tool()
def people_getMe() -> dict:
    """인증된 사용자 프로필 조회 (토큰 갱신 확인용)"""
    svc = _svc("people", "v1")
    res = svc.people().get(
        resourceName="people/me",
        personFields="names,emailAddresses",
    ).execute()
    return {
        "name":  res.get("names", [{}])[0].get("displayName", ""),
        "email": res.get("emailAddresses", [{}])[0].get("value", ""),
    }


@mcp.tool()
def auth_refreshToken() -> dict:
    """OAuth 액세스 토큰을 강제 갱신합니다."""
    token = _load_token()
    _refresh(token)
    token = _load_token()  # 갱신된 값 재로드
    return {"refreshed": True, "expiry_date": token["expiry_date"]}


# ── Drive ─────────────────────────────────────────────────────────────────────

_MIME = {
    "spreadsheet": "application/vnd.google-apps.spreadsheet",
    "presentation": "application/vnd.google-apps.presentation",
    "folder":       "application/vnd.google-apps.folder",
    "doc":          "application/vnd.google-apps.document",
}

_EXT_MIME = {
    ".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
    ".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
    ".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    ".pdf":  "application/pdf",
    ".png":  "image/png",
    ".jpg":  "image/jpeg",
    ".jpeg": "image/jpeg",
    ".zip":  "application/zip",
    ".json": "application/json",
    ".txt":  "text/plain",
}


@mcp.tool()
def drive_listFiles(
    folderId: str = None,
    mimeType: str = None,
    pageSize: int = 20,
) -> dict:
    """Drive 폴더의 파일/폴더 목록을 조회합니다."""
    svc = _svc("drive", "v3")
    q = ["trashed = false"]
    if folderId:
        q.append(f"'{folderId}' in parents")
    if mimeType and mimeType in _MIME:
        q.append(f"mimeType = '{_MIME[mimeType]}'")

    res = svc.files().list(
        q=" and ".join(q),
        fields="files(id,name,mimeType,modifiedTime,webViewLink)",
        pageSize=min(pageSize, 100),
        orderBy="modifiedTime desc",
        supportsAllDrives=True,
        includeItemsFromAllDrives=True,
    ).execute()

    files = res.get("files", [])
    return {
        "files": [
            {
                "id":           f["id"],
                "name":         f["name"],
                "type":         next((k for k, v in _MIME.items() if v == f["mimeType"]), f["mimeType"]),
                "modifiedTime": f.get("modifiedTime"),
                "url":          f.get("webViewLink"),
            }
            for f in files
        ],
        "count": len(files),
    }


@mcp.tool()
def drive_search(query: str, mimeType: str = None, pageSize: int = 20) -> dict:
    """Google Drive에서 이름으로 파일을 검색합니다."""
    svc = _svc("drive", "v3")
    q = ["trashed = false", f"name contains '{query}'"]
    if mimeType and mimeType in _MIME:
        q.append(f"mimeType = '{_MIME[mimeType]}'")

    res = svc.files().list(
        q=" and ".join(q),
        fields="files(id,name,mimeType,modifiedTime,webViewLink)",
        pageSize=min(pageSize, 100),
        orderBy="modifiedTime desc",
        supportsAllDrives=True,
        includeItemsFromAllDrives=True,
    ).execute()

    files = res.get("files", [])
    return {
        "files": [{"id": f["id"], "name": f["name"], "url": f.get("webViewLink")} for f in files],
        "count": len(files),
    }


@mcp.tool()
def drive_findFolder(name: str, parentId: str = None) -> dict:
    """이름으로 Drive 폴더를 찾습니다."""
    svc = _svc("drive", "v3")
    q = [
        "trashed = false",
        f"mimeType = '{_MIME['folder']}'",
        f"name = '{name}'",
    ]
    if parentId:
        q.append(f"'{parentId}' in parents")

    res = svc.files().list(
        q=" and ".join(q),
        fields="files(id,name,webViewLink)",
        pageSize=10,
        supportsAllDrives=True,
        includeItemsFromAllDrives=True,
    ).execute()

    files = res.get("files", [])
    if not files:
        return {"found": False, "folder": None}
    f = files[0]
    return {"found": True, "folder": {"id": f["id"], "name": f["name"], "url": f.get("webViewLink")}}


@mcp.tool()
def drive_uploadFile(
    localPath: str,
    folderId: str,
    fileName: str = None,
    mimeType: str = None,
) -> dict:
    """로컬 파일을 Google Drive 폴더에 업로드합니다."""
    from googleapiclient.http import MediaFileUpload

    if not os.path.exists(localPath):
        raise FileNotFoundError(f"파일을 찾을 수 없습니다: {localPath}")

    svc       = _svc("drive", "v3")
    folder_id = _id(folderId)
    file_name = fileName or os.path.basename(localPath)
    ext       = os.path.splitext(localPath)[1].lower()
    mime      = mimeType or _EXT_MIME.get(ext, "application/octet-stream")

    res = svc.files().create(
        supportsAllDrives=True,
        body={"name": file_name, "parents": [folder_id]},
        media_body=MediaFileUpload(localPath, mimetype=mime),
        fields="id,name,webViewLink",
    ).execute()

    return {"fileId": res["id"], "name": res["name"], "webViewLink": res.get("webViewLink")}


@mcp.tool()
def drive_createFolder(name: str, parentId: str = None) -> dict:
    """Google Drive에 폴더를 생성합니다."""
    svc  = _svc("drive", "v3")
    meta = {"name": name, "mimeType": _MIME["folder"]}
    if parentId:
        meta["parents"] = [_id(parentId)]

    res = svc.files().create(
        supportsAllDrives=True,
        body=meta,
        fields="id,name,webViewLink",
    ).execute()
    return {"folderId": res["id"], "name": res["name"], "webViewLink": res.get("webViewLink")}


# ── Sheets ────────────────────────────────────────────────────────────────────

@mcp.tool()
def sheets_create(title: str, folderId: str = None) -> dict:
    """Google Sheets 스프레드시트를 생성합니다."""
    sheets = _svc("sheets", "v4")
    drive  = _svc("drive", "v3")

    res     = sheets.spreadsheets().create(body={"properties": {"title": title}}).execute()
    file_id = res["spreadsheetId"]
    if folderId:
        drive.files().update(
            fileId=file_id, addParents=folderId, removeParents="root", fields="id,parents"
        ).execute()

    return {
        "spreadsheetId": file_id,
        "title":         title,
        "url":           f"https://docs.google.com/spreadsheets/d/{file_id}",
    }


@mcp.tool()
def sheets_addSheet(spreadsheetId: str, sheetTitle: str) -> dict:
    """스프레드시트에 새 시트(탭)를 추가합니다."""
    svc = _svc("sheets", "v4")
    sid = _id(spreadsheetId)
    res = svc.spreadsheets().batchUpdate(
        spreadsheetId=sid,
        body={"requests": [{"addSheet": {"properties": {"title": sheetTitle}}}]},
    ).execute()
    props = res["replies"][0]["addSheet"]["properties"]
    return {"sheetId": props["sheetId"], "title": props["title"]}


@mcp.tool()
def sheets_getMetadata(spreadsheetId: str) -> dict:
    """스프레드시트의 제목과 시트 목록을 가져옵니다."""
    svc = _svc("sheets", "v4")
    sid = _id(spreadsheetId)
    res = svc.spreadsheets().get(
        spreadsheetId=sid,
        fields="spreadsheetId,properties.title,sheets.properties",
    ).execute()
    return {
        "spreadsheetId": res["spreadsheetId"],
        "title":         res["properties"]["title"],
        "sheets": [
            {
                "sheetId":     s["properties"]["sheetId"],
                "title":       s["properties"]["title"],
                "index":       s["properties"]["index"],
                "rowCount":    s["properties"].get("gridProperties", {}).get("rowCount"),
                "columnCount": s["properties"].get("gridProperties", {}).get("columnCount"),
            }
            for s in res.get("sheets", [])
        ],
    }


@mcp.tool()
def sheets_getRange(spreadsheetId: str, range: str) -> dict:
    """스프레드시트의 특정 범위 값을 읽어옵니다."""
    svc    = _svc("sheets", "v4")
    sid    = _id(spreadsheetId)
    res    = svc.spreadsheets().values().get(spreadsheetId=sid, range=range).execute()
    values = res.get("values", [])
    return {"range": res.get("range"), "values": values, "rowCount": len(values)}


@mcp.tool()
def sheets_updateRange(spreadsheetId: str, range: str, values: list) -> dict:
    """스프레드시트의 특정 범위에 값을 씁니다 (기존 값 덮어씀)."""
    svc = _svc("sheets", "v4")
    sid = _id(spreadsheetId)
    svc.spreadsheets().values().update(
        spreadsheetId=sid,
        range=range,
        valueInputOption="USER_ENTERED",
        body={"values": values},
    ).execute()
    return {"updated": True, "range": range}


@mcp.tool()
def sheets_appendRows(spreadsheetId: str, range: str, values: list) -> dict:
    """스프레드시트 마지막 행 아래에 새 행을 추가합니다."""
    svc = _svc("sheets", "v4")
    sid = _id(spreadsheetId)
    res = svc.spreadsheets().values().append(
        spreadsheetId=sid,
        range=range,
        valueInputOption="USER_ENTERED",
        insertDataOption="INSERT_ROWS",
        body={"values": values},
    ).execute()
    upd = res.get("updates", {})
    return {"appended": True, "updatedRange": upd.get("updatedRange"), "updatedRows": upd.get("updatedRows")}


# ── Slides ────────────────────────────────────────────────────────────────────

@mcp.tool()
def slides_create(title: str, folderId: str = None) -> dict:
    """Google Slides 프레젠테이션을 생성합니다."""
    slides = _svc("slides", "v1")
    drive  = _svc("drive", "v3")

    res     = slides.presentations().create(body={"title": title}).execute()
    file_id = res["presentationId"]
    if folderId:
        drive.files().update(
            fileId=file_id, addParents=folderId, removeParents="root", fields="id,parents"
        ).execute()

    return {
        "presentationId": file_id,
        "title":          title,
        "url":            f"https://docs.google.com/presentation/d/{file_id}",
    }


@mcp.tool()
def slides_getMetadata(presentationId: str) -> dict:
    """Google Slides 프레젠테이션 메타데이터를 가져옵니다."""
    svc = _svc("slides", "v1")
    pid = _id(presentationId)
    res = svc.presentations().get(presentationId=pid).execute()
    return {
        "presentationId": res["presentationId"],
        "title":          res.get("title"),
        "slideCount":     len(res.get("slides", [])),
    }


@mcp.tool()
def slides_getText(presentationId: str) -> dict:
    """Google Slides의 모든 텍스트를 추출합니다."""
    svc = _svc("slides", "v1")
    pid = _id(presentationId)
    res = svc.presentations().get(presentationId=pid).execute()

    result = []
    for i, slide in enumerate(res.get("slides", [])):
        texts = []
        for elem in slide.get("pageElements", []):
            for te in elem.get("shape", {}).get("text", {}).get("textElements", []):
                content = te.get("textRun", {}).get("content", "").strip()
                if content:
                    texts.append(content)
        if texts:
            result.append({"slide": i + 1, "texts": texts})

    return {"presentationId": pid, "slides": result}


@mcp.tool()
def slides_updateText(presentationId: str, oldText: str, newText: str) -> dict:
    """Google Slides에서 특정 텍스트를 찾아 교체합니다."""
    svc = _svc("slides", "v1")
    pid = _id(presentationId)
    svc.presentations().batchUpdate(
        presentationId=pid,
        body={"requests": [{"replaceAllText": {
            "containsText": {"text": oldText, "matchCase": True},
            "replaceText":  newText,
        }}]},
    ).execute()
    return {"updated": True, "oldText": oldText, "newText": newText}


# ── Docs ──────────────────────────────────────────────────────────────────────

@mcp.tool()
def docs_create(title: str, folderId: str = None) -> dict:
    """Google Docs 문서를 생성합니다."""
    docs  = _svc("docs", "v1")
    drive = _svc("drive", "v3")

    res    = docs.documents().create(body={"title": title}).execute()
    doc_id = res["documentId"]
    if folderId:
        drive.files().update(
            fileId=doc_id, addParents=folderId, removeParents="root", fields="id,parents"
        ).execute()

    return {
        "documentId": doc_id,
        "title":      title,
        "url":        f"https://docs.google.com/document/d/{doc_id}",
    }


@mcp.tool()
def docs_getText(documentId: str) -> dict:
    """Google Docs 문서의 전체 텍스트를 읽어옵니다."""
    svc = _svc("docs", "v1")
    did = _id(documentId)
    res = svc.documents().get(documentId=did).execute()

    text = "".join(
        e.get("textRun", {}).get("content", "")
        for elem in res.get("body", {}).get("content", [])
        if elem.get("paragraph")
        for e in elem["paragraph"].get("elements", [])
    )
    return {"documentId": did, "title": res.get("title"), "text": text}


@mcp.tool()
def docs_appendText(documentId: str, text: str) -> dict:
    """Google Docs 문서 끝에 텍스트를 추가합니다."""
    svc = _svc("docs", "v1")
    did = _id(documentId)
    doc       = svc.documents().get(documentId=did, fields="body.content").execute()
    end_index = doc["body"]["content"][-1]["endIndex"] - 1
    svc.documents().batchUpdate(
        documentId=did,
        body={"requests": [{"insertText": {"location": {"index": end_index}, "text": text}}]},
    ).execute()
    return {"appended": True, "documentId": did}


@mcp.tool()
def docs_replaceText(documentId: str, oldText: str, newText: str) -> dict:
    """Google Docs 문서에서 특정 텍스트를 찾아 교체합니다."""
    svc = _svc("docs", "v1")
    did = _id(documentId)
    svc.documents().batchUpdate(
        documentId=did,
        body={"requests": [{"replaceAllText": {
            "containsText": {"text": oldText, "matchCase": True},
            "replaceText":  newText,
        }}]},
    ).execute()
    return {"updated": True, "oldText": oldText, "newText": newText}


# ── Entry point ───────────────────────────────────────────────────────────────

if __name__ == "__main__":
    mcp.run()
