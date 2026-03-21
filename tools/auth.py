#!/usr/bin/env python3
"""
tools/auth.py — Google OAuth 초기 인증

사용법:
  python tools/auth.py           # 브라우저 자동 오픈 (Windows/macOS)
  python tools/auth.py --url     # URL 출력 (Linux 서버 1단계)
  python tools/auth.py --paste   # 리디렉션 URL 붙여넣기 (Linux 서버 2단계)

처음 한 번만 실행하면 됩니다. 이후 토큰은 자동 갱신됩니다.
"""

import argparse
import json
import os
import socket
import sys
import time
import urllib.parse
import urllib.request
import webbrowser
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path

# ── 상수 ──────────────────────────────────────────────────────────────────────
# Google Cloud Console → APIs & Services → Credentials → OAuth 2.0 Client ID
# Application type: Desktop application

_CLIENT_ID     = os.environ.get("GOOGLE_CLIENT_ID", "")
_CLIENT_SECRET = os.environ.get("GOOGLE_CLIENT_SECRET", "")

TOKEN_URI  = "https://oauth2.googleapis.com/token"
TOKEN_PATH = Path(__file__).parent / "output" / ".gtoken"
_STATE_PATH = Path(__file__).parent / "output" / ".auth_state"

SCOPES = [
    "https://www.googleapis.com/auth/drive",
    "https://www.googleapis.com/auth/documents",
    "https://www.googleapis.com/auth/presentations",
    "https://www.googleapis.com/auth/spreadsheets",
    "https://www.googleapis.com/auth/userinfo.profile",
    "https://www.googleapis.com/auth/userinfo.email",
]


# ── 공통 유틸 ─────────────────────────────────────────────────────────────────

def _check_creds() -> None:
    if _CLIENT_ID == "YOUR_CLIENT_ID":
        print("❌ auth.py에 client_id / client_secret을 채워넣으세요.")
        print("\n  Google Cloud Console → APIs & Services → Credentials")
        print("  → Create OAuth 2.0 Client ID → Desktop application")
        sys.exit(1)


def find_free_port() -> int:
    with socket.socket() as s:
        s.bind(("", 0))
        return s.getsockname()[1]


def _build_auth_url(redirect_uri: str) -> str:
    params = {
        "client_id":     _CLIENT_ID,
        "redirect_uri":  redirect_uri,
        "response_type": "code",
        "scope":         " ".join(SCOPES),
        "access_type":   "offline",
        "prompt":        "consent",
    }
    return "https://accounts.google.com/o/oauth2/v2/auth?" + urllib.parse.urlencode(params)


def exchange_code(code: str, redirect_uri: str) -> dict:
    body = urllib.parse.urlencode({
        "code":          code,
        "client_id":     _CLIENT_ID,
        "client_secret": _CLIENT_SECRET,
        "redirect_uri":  redirect_uri,
        "grant_type":    "authorization_code",
    }).encode()
    req = urllib.request.Request(
        TOKEN_URI, data=body,
        headers={"Content-Type": "application/x-www-form-urlencoded"},
    )
    try:
        with urllib.request.urlopen(req) as resp:
            data = json.loads(resp.read())
    except urllib.error.HTTPError as e:
        err = json.loads(e.read())
        print(f"\n❌ 토큰 발급 실패: {err.get('error')} — {err.get('error_description','')}")
        sys.exit(1)

    expiry = int(time.time() * 1000) + data.get("expires_in", 3600) * 1000
    return {
        "access_token":  data["access_token"],
        "refresh_token": data.get("refresh_token", ""),
        "expiry_date":   expiry,
        "client_id":     _CLIENT_ID,
        "client_secret": _CLIENT_SECRET,
    }


def save_token(token: dict) -> None:
    TOKEN_PATH.parent.mkdir(parents=True, exist_ok=True)
    TOKEN_PATH.write_text(json.dumps(token), encoding="utf-8")
    print(f"  ✅ 토큰 저장 완료: {TOKEN_PATH}")


# ── 로컬 서버 방식 (Windows/macOS) ───────────────────────────────────────────

def cmd_local_server() -> None:
    _check_creds()
    port = find_free_port()
    redirect_uri = f"http://localhost:{port}"
    auth_url = _build_auth_url(redirect_uri)

    print("\n[1] 브라우저에서 Google 로그인 후 자동으로 완료됩니다...\n")
    webbrowser.open(auth_url)

    code_holder: dict = {}

    class Handler(BaseHTTPRequestHandler):
        def do_GET(self):
            qs = urllib.parse.parse_qs(urllib.parse.urlparse(self.path).query)
            if "code" in qs:
                code_holder["code"] = qs["code"][0]
                self.send_response(200)
                self.send_header("Content-Type", "text/html; charset=utf-8")
                self.end_headers()
                self.wfile.write(
                    "<html><body><h2>Done. Close this tab.</h2></body></html>".encode()
                )
            else:
                self.send_response(400)
                self.end_headers()
        def log_message(self, *args): pass

    HTTPServer(("localhost", port), Handler).handle_request()

    if "code" not in code_holder:
        print("❌ 인증 코드를 받지 못했습니다.")
        sys.exit(1)

    print("[2] 토큰 발급 중...")
    token = exchange_code(code_holder["code"], redirect_uri)
    save_token(token)
    print("\n인증 완료. 이제 tools를 사용할 수 있습니다.")


# ── --url / --paste (Linux 서버) ──────────────────────────────────────────────

def cmd_url() -> None:
    _check_creds()
    redirect_uri = "http://localhost"
    auth_url = _build_auth_url(redirect_uri)

    _STATE_PATH.parent.mkdir(parents=True, exist_ok=True)
    _STATE_PATH.write_text(json.dumps({"redirect_uri": redirect_uri}))

    print(auth_url)
    print("\n# 위 URL을 브라우저에서 열어 인증 후:")
    print("# run.sh token refresh  (또는 직접: tools/auth.py --paste)")


def cmd_paste() -> None:
    _check_creds()

    if _STATE_PATH.exists():
        state = json.loads(_STATE_PATH.read_text())
        redirect_uri = state["redirect_uri"]
        _STATE_PATH.unlink(missing_ok=True)
    else:
        redirect_uri = "http://localhost"
        print("\n[1] 아래 URL을 브라우저에서 여세요:\n")
        print(f"  {_build_auth_url(redirect_uri)}\n")

    print("[리디렉션 URL을 붙여넣으세요]")
    print("  (예: http://localhost/?code=4/0AX...)\n")

    raw = input("  URL: ").strip()
    qs  = urllib.parse.parse_qs(urllib.parse.urlparse(raw).query)
    if "code" not in qs:
        print("❌ URL에서 code를 찾을 수 없습니다.")
        sys.exit(1)

    print("\n토큰 발급 중...")
    token = exchange_code(qs["code"][0], redirect_uri)
    save_token(token)
    print("\n인증 완료. 이제 tools를 사용할 수 있습니다.")


# ── 메인 ──────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description="Google OAuth 초기 인증")
    parser.add_argument("--url",   action="store_true", help="인증 URL 출력 (Linux 서버 1단계)")
    parser.add_argument("--paste", action="store_true", help="리디렉션 URL 붙여넣기 (Linux 서버 2단계)")
    args = parser.parse_args()

    if args.url:
        cmd_url()
    elif args.paste:
        cmd_paste()
    else:
        cmd_local_server()


if __name__ == "__main__":
    main()
