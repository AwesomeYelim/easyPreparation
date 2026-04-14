#!/usr/bin/env bash
# tools/start_mcp.sh — gdrive_mcp.py 크로스플랫폼 실행 래퍼
# .mcp.json에서 호출됨. Windows/Linux/macOS 공통.

TOOLS_DIR="$(cd "$(dirname "$0")" && pwd)"

if [ -x "$TOOLS_DIR/.venv/Scripts/python.exe" ]; then
    exec "$TOOLS_DIR/.venv/Scripts/python.exe" "$TOOLS_DIR/gdrive_mcp.py"
else
    exec "$TOOLS_DIR/.venv/bin/python" "$TOOLS_DIR/gdrive_mcp.py"
fi
