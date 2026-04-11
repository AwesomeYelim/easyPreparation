# tools — Python AI 툴킷

문서 생성, Drive 업로드, MCP 서버.

## 실행

```bash
bash tools/setup.sh          # 초기 세팅
tools/run.sh report --type code  # 슬라이드 생성
tools/run.sh upload <file> --folder <ID> --key <key>  # Drive 업로드
```

## 스크립트 패턴

> heredoc 금지. Write 도구 → Bash 실행 → 삭제.

```bash
PYTHONUTF8=1 tools/.venv/bin/python tools/output/_tmp.py && rm tools/output/_tmp.py
```

## 경로

- 출력: `tools/output/`
- 임시: `tools/output/_tmp.py`
- 템플릿: `tools/templates/`
- R2 업로드: `tools/upload-r2.sh`
