"""예배 영상 감시. 새 주일예배는 바로, 백로그는 하루 한 편씩 /autoshorts 세션을 시작한다.

설정 파일(기본 ~/.config/autoshorts/config.json)과 상태 파일(기본 ~/.local/state/autoshorts/)은 저장소 밖에 둔다.
사용: watch.py {seed,status,new,backlog} [--dry-run]
"""
from __future__ import annotations

import argparse
import json
import os
import shlex
import shutil
import subprocess
import sys
import uuid
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path

CONFIG_PATH = Path(os.environ.get("AUTOSHORTS_CONFIG", Path.home() / ".config" / "autoshorts" / "config.json"))
STATE_DIR = Path(os.environ.get("AUTOSHORTS_STATE_DIR", Path.home() / ".local" / "state" / "autoshorts"))
STATE_PATH = STATE_DIR / "state.json"
LOG_PATH = STATE_DIR / "watch.log"


@dataclass
class Config:
    channel_url: str
    start_folder: str
    model: str
    window_start: int
    window_end: int
    skip_ids: list[str]
    manual_ids: list[str]
    yt_dlp: str
    claude: str
    tmux: str

    @staticmethod
    def load() -> "Config":
        raw = json.loads(CONFIG_PATH.read_text(encoding="utf-8"))
        return Config(
            channel_url=raw["channel_url"].rstrip("/"),
            start_folder=raw["start_folder"],
            model=raw.get("model", "opus"),
            window_start=int(raw.get("window_start_hour", 13)),
            window_end=int(raw.get("window_end_hour", 22)),
            skip_ids=list(raw.get("skip_ids", [])),
            manual_ids=list(raw.get("manual_ids", [])),
            yt_dlp=raw.get("yt_dlp_path") or shutil.which("yt-dlp") or "yt-dlp",
            claude=raw.get("claude_path") or shutil.which("claude") or "claude",
            tmux=raw.get("tmux_path") or shutil.which("tmux") or "tmux",
        )


@dataclass
class Stream:
    id: str
    live: str
    title: str


class Logger:
    def __init__(self, mode: str, dry: bool) -> None:
        self.tag = f"[{mode}]" + ("[dry]" if dry else "")

    def __call__(self, msg: str) -> None:
        line = f"{datetime.now():%Y-%m-%d %H:%M:%S} {self.tag} {msg}"
        print(line, flush=True)
        STATE_DIR.mkdir(parents=True, exist_ok=True)
        with LOG_PATH.open("a", encoding="utf-8") as fh:
            fh.write(line + "\n")


def run_capture(cmd: list[str], timeout: int = 180) -> str:
    env = dict(os.environ, PYTHONUTF8="1")
    done = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=timeout, env=env)
    if done.returncode != 0:
        raise RuntimeError(f"명령 실패({done.returncode}): {Path(cmd[0]).name} {done.stderr.strip()[:200]}")
    return done.stdout


def kind_of(title: str) -> str:
    return "sermon" if "주일예배" in title else "praise"


def get_streams(cfg: Config, count: int) -> list[Stream]:
    cmd = [cfg.yt_dlp, "--flat-playlist", "--print", "%(id)s|%(live_status)s|%(title)s"]
    if count > 0:
        cmd += ["--playlist-end", str(count)]
    cmd.append(cfg.channel_url + "/streams")
    rows: list[Stream] = []
    for line in run_capture(cmd).splitlines():
        parts = line.split("|", 2)
        if len(parts) == 3:
            rows.append(Stream(parts[0], parts[1], parts[2]))
    return rows


def has_captions(cfg: Config, video_id: str, log: Logger) -> bool:
    cmd = [cfg.yt_dlp, "--skip-download", "--print", "%(automatic_captions|{})j", f"https://www.youtube.com/watch?v={video_id}"]
    try:
        out = run_capture(cmd).strip().splitlines()
        names = json.loads(out[-1]) if out else {}
    except (RuntimeError, ValueError, subprocess.TimeoutExpired) as err:
        log(f"자막 확인 실패({video_id}): {err}")
        return False
    return "ko" in names or "ko-orig" in names


def load_state() -> dict:
    if not STATE_PATH.exists():
        return {"last_backlog_date": "", "videos": []}
    return json.loads(STATE_PATH.read_text(encoding="utf-8"))


def save_state(state: dict, dry: bool) -> None:
    if dry:
        return
    STATE_DIR.mkdir(parents=True, exist_ok=True)
    STATE_PATH.write_text(json.dumps(state, ensure_ascii=False, indent=2), encoding="utf-8")


def find_video(state: dict, video_id: str) -> dict | None:
    return next((v for v in state["videos"] if v["id"] == video_id), None)


def start_session(cfg: Config, video_id: str, title: str, dry: bool, log: Logger) -> bool:
    url = f"https://www.youtube.com/watch?v={video_id}"
    prompt = f"/autoshorts {url} 영상 제목: {title} (자동으로 시작된 세션이야. 체크포인트 1 에서 질문을 남기고 멈춰.)"
    name = "[side] 쇼츠 " + title.split("/")[0].strip()
    sid = str(uuid.uuid4())
    claude_cmd = [cfg.claude, "--session-id", sid, "--model", cfg.model, f"--remote-control={name}", prompt]
    if dry:
        log(f"세션 시작 안 함(드라이런): {name} ({video_id}) 명령={shlex.join(claude_cmd)}")
        return True
    if os.name == "nt":
        subprocess.Popen(claude_cmd, cwd=cfg.start_folder, creationflags=0x00000010)
    else:
        session = f"autoshorts-{video_id}"
        tmux_cmd = [cfg.tmux, "new-session", "-d", "-s", session, "-c", cfg.start_folder, shlex.join(claude_cmd)]
        subprocess.run(tmux_cmd, check=True)
    log(f"세션 시작: {name} ({video_id}) sid={sid}")
    return True


def mode_seed(cfg: Config, state: dict, dry: bool, log: Logger) -> None:
    streams = get_streams(cfg, 0)
    known = {v["id"] for v in state["videos"]}
    added = 0
    for s in streams:
        if s.id in known:
            continue
        kind = kind_of(s.title)
        if s.id in cfg.skip_ids:
            status = "skip"
        elif s.id in cfg.manual_ids:
            status = "manual"
        else:
            status = "pending" if kind == "sermon" else "hold-praise"
        state["videos"].append({"id": s.id, "kind": kind, "status": status, "title": s.title})
        added += 1
    save_state(state, dry)
    log(f"seed: 목록 {len(streams)}개, 새로 넣은 것 {added}개")


def mode_status(state: dict) -> None:
    counts: dict[str, int] = {}
    for v in state["videos"]:
        key = f"{v['kind']}, {v['status']}"
        counts[key] = counts.get(key, 0) + 1
    for key, n in sorted(counts.items()):
        print(f"{key}: {n}")
    nxt = next((v for v in state["videos"] if v["kind"] == "sermon" and v["status"] == "pending"), None)
    print(f"다음 백로그: {nxt['id']} | {nxt['title']}" if nxt else "다음 백로그: 없음")
    print(f"마지막 백로그 시작일: {state.get('last_backlog_date', '')}")


def mode_new(cfg: Config, state: dict, dry: bool, log: Logger) -> None:
    now = datetime.now()
    if not dry and (now.weekday() != 6 or now.hour < cfg.window_start or now.hour >= cfg.window_end):
        log(f"감시 시간(일요일 {cfg.window_start}~{cfg.window_end}시) 밖이라 종료")
        return
    for s in get_streams(cfg, 8):
        if find_video(state, s.id):
            continue
        if s.live != "was_live":
            log(f"아직 라이브 중이거나 예정: {s.id} {s.title}")
            continue
        if kind_of(s.title) == "praise":
            state["videos"].insert(0, {"id": s.id, "kind": "praise", "status": "hold-praise", "title": s.title})
            log(f"찬양 컷 대상(스킬 대기)로 기록만 함: {s.id} {s.title}")
            continue
        if not has_captions(cfg, s.id, log):
            log(f"자막 준비 전, 다음 확인 때 재시도: {s.id} {s.title}")
            continue
        entry = {"id": s.id, "kind": "sermon", "status": "pending", "title": s.title}
        if start_session(cfg, s.id, s.title, dry, log):
            entry["status"] = "started"
            entry["started_at"] = f"{now:%Y-%m-%d %H:%M}"
        state["videos"].insert(0, entry)
    save_state(state, dry)


def mode_backlog(cfg: Config, state: dict, dry: bool, log: Logger) -> None:
    today = f"{datetime.now():%Y-%m-%d}"
    if not dry and state.get("last_backlog_date") == today:
        log("오늘 이미 시작해서 종료")
        return
    for v in state["videos"]:
        if v["kind"] != "sermon" or v["status"] != "pending":
            continue
        if not has_captions(cfg, v["id"], log):
            log(f"자막 없음, 건너뜀: {v['id']} {v['title']}")
            if not dry:
                v["status"] = "no-captions"
            continue
        if start_session(cfg, v["id"], v["title"], dry, log) and not dry:
            v["status"] = "started"
            v["started_at"] = f"{datetime.now():%Y-%m-%d %H:%M}"
            state["last_backlog_date"] = today
        break
    save_state(state, dry)


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(description="예배 영상 감시")
    parser.add_argument("mode", choices=["seed", "status", "new", "backlog"])
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args(argv)
    log = Logger(args.mode, args.dry_run)
    try:
        cfg = Config.load()
        state = load_state()
        if args.mode == "seed":
            mode_seed(cfg, state, args.dry_run, log)
        elif args.mode == "status":
            mode_status(state)
        elif args.mode == "new":
            mode_new(cfg, state, args.dry_run, log)
        else:
            mode_backlog(cfg, state, args.dry_run, log)
    except (RuntimeError, ValueError, KeyError, OSError, subprocess.SubprocessError) as err:
        log(f"오류: {err}")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
