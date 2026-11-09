"""설교 라이브 쇼츠 도구. fetch: 자동 자막 받기, render: 사양 JSON 으로 9:16 쇼츠 렌더."""
from __future__ import annotations

import argparse
import json
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path

SKILL_DIR = Path(__file__).resolve().parent
HEADER_ASS = SKILL_DIR / "templates" / "header.ass"
LOGO_PNG = SKILL_DIR / "templates" / "logo.png"
FONTS_DIR = "'C\\:/Windows/Fonts'"

BAND = 540
AREA_H = 1380
MIN_GAP = 3.0
MIN_CUT = 3.0
MAX_SHARE = 1 / 3


@dataclass
class Cue:
    start: float
    end: float
    text: str


@dataclass
class Cutaway:
    file: str
    start: float
    end: float
    ss: float
    fadeout: bool


@dataclass
class Spec:
    src: str
    out: str
    workdir: Path
    segments: list[tuple[float, float]]
    title: tuple[str, str]
    cues: list[Cue]
    cutaways: list[Cutaway]
    pulses: list[float]

    @staticmethod
    def load(path: Path) -> "Spec":
        raw = json.loads(path.read_text(encoding="utf-8"))
        title = raw["title"]
        if len(title) != 2:
            raise ValueError("title 은 2줄이어야 해")
        return Spec(
            src=raw["src"],
            out=raw["out"],
            workdir=Path(raw["workdir"]),
            segments=[(float(a), float(b)) for a, b in raw["segments"]],
            title=(title[0], title[1]),
            cues=[Cue(float(c["start"]), float(c["end"]), c["text"]) for c in raw["cues"]],
            cutaways=[
                Cutaway(
                    file=c["file"],
                    start=float(c["start"]),
                    end=float(c["end"]),
                    ss=float(c.get("ss", 1.0)),
                    fadeout=bool(c.get("fadeout", True)),
                )
                for c in raw.get("cutaways", [])
            ],
            pulses=[float(p) for p in raw.get("pulses", [])],
        )


def run(cmd: list[str], cwd: Path | None = None) -> None:
    done = subprocess.run(cmd, cwd=cwd)
    if done.returncode != 0:
        raise RuntimeError(f"명령 실패({done.returncode}): {cmd[0]}")


def hms(sec: float) -> str:
    s = int(sec)
    return f"{s // 3600:02d}:{s % 3600 // 60:02d}:{s % 60:02d}"


def ass_time(sec: float) -> str:
    cs = round(max(sec, 0.0) * 100)
    return f"{cs // 360000}:{cs % 360000 // 6000:02d}:{cs % 6000 // 100:02d}.{cs % 100:02d}"


def fetch(url: str, out_dir: Path) -> Path:
    subs = out_dir / "subs"
    subs.mkdir(parents=True, exist_ok=True)
    run(["yt-dlp", "--write-auto-sub", "--sub-lang", "ko", "--sub-format", "json3",
         "--skip-download", "-o", str(subs / "sermon"), url])
    src = subs / "sermon.ko.json3"
    if not src.exists():
        raise FileNotFoundError(f"자막 파일이 없어: {src}")
    data = json.loads(src.read_text(encoding="utf-8"))
    lines: list[str] = []
    for ev in data.get("events", []):
        segs = ev.get("segs")
        if not segs:
            continue
        text = "".join(s.get("utf8", "") for s in segs).replace("\n", " ").strip()
        if text:
            lines.append(f"{hms(ev['tStartMs'] / 1000)} {text}")
    out = subs / "sermon.txt"
    out.write_text("\n".join(lines) + "\n", encoding="utf-8")
    return out


def clip_time(spec: Spec, s: float) -> float:
    acc = 0.0
    for a, b in spec.segments:
        if a - 1e-6 <= s <= b + 1e-6:
            return acc + (s - a)
        acc += b - a
    raise ValueError(f"구간 밖 시각이야: {s}")


def total_length(spec: Spec) -> float:
    return sum(b - a for a, b in spec.segments)


def build_ass(spec: Spec, total: float) -> str:
    head = HEADER_ASS.read_text(encoding="utf-8").rstrip() + "\n\n"
    head += "[Events]\nFormat: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n\n"
    pop = r"{\fscx85\fscy85\t(0,110,\fscx100\fscy100)}"
    rows: list[str] = []
    for c in spec.cues:
        text = c.text.replace("{Y}", r"{\c&H00E6FF&}").replace("{/}", r"{\c&HFFFFFF&}")
        rows.append(f"Dialogue: 0,{ass_time(clip_time(spec, c.start))},{ass_time(clip_time(spec, c.end))},Sub,,0,0,0,,{pop}{text}")
    end = ass_time(total)
    rows.append(f"Dialogue: 1,{ass_time(0)},{end},T1,,0,0,0,,{{\\pos(540,325)\\fad(250,0)}}{spec.title[0]}")
    rows.append(f"Dialogue: 1,{ass_time(0.4)},{end},T2,,0,0,0,,{{\\pos(540,450)\\fad(250,0)\\fscx85\\fscy85\\t(0,150,\\fscx100\\fscy100)}}{spec.title[1]}")
    return head + "\n".join(rows) + "\n"


def cut_chain(idx: int, name: str, cut: Cutaway, t0: float) -> str:
    dur = cut.end - cut.start
    fade_out = f",fade=t=out:st={dur - 0.3:.2f}:d=0.3:alpha=1" if cut.fadeout else ""
    return (
        f"[{idx}:v]fps=30,scale=1080:1920:flags=lanczos,crop=1080:{AREA_H}:0:{(1920 - AREA_H) // 2},"
        f"scale=1296:1656:flags=lanczos,trim=start={cut.ss}:duration={dur + 0.4:.2f},setpts=PTS-STARTPTS,"
        f"zoompan=z='1.0+0.0009*on':x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)':d=1:s=1080x{AREA_H}:fps=30,"
        f"eq=saturation=0.95:contrast=1.04:brightness=-0.02,format=yuva420p,"
        f"fade=t=in:st=0:d=0.3:alpha=1{fade_out},setpts=PTS+{t0:.2f}/TB[{name}];"
    )


def build_filter(spec: Spec, total: float, ass_name: str) -> tuple[str, str]:
    n = len(spec.segments)
    logo_idx = n
    parts: list[str] = []
    if n == 1:
        src_v, audio = "[0:v]", "0:a"
    else:
        for i in range(n):
            parts.append(f"[{i}:v]fps=30,setpts=PTS-STARTPTS[v{i}];[{i}:a]asetpts=PTS-STARTPTS[a{i}];")
        joined = "".join(f"[v{i}][a{i}]" for i in range(n))
        parts.append(f"{joined}concat=n={n}:v=1:a=1[vv][aa];")
        src_v, audio = "[vv]", "[aa]"

    pulse = "+".join(
        f"0.04*max(0,1-(on/30-{clip_time(spec, t):.2f})*4)*gte(on/30,{clip_time(spec, t):.2f})" for t in spec.pulses
    ) or "0"
    parts.append(
        f"{src_v}crop=450:575:759:110,scale=2160:2760:flags=lanczos,"
        f"zoompan=z='1+0.07*on/{int(total * 30)}+{pulse}':x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)':d=1:s=1080x{AREA_H}:fps=30,"
        f"eq=contrast=1.06:saturation=1.12,unsharp=5:5:0.7,pad=1080:1920:0:{BAND}:black[z0];"
    )

    prev = "z0"
    for i, cut in enumerate(spec.cutaways):
        idx = logo_idx + 1 + i
        parts.append(cut_chain(idx, f"cut{i}", cut, clip_time(spec, cut.start)))
        parts.append(f"[{prev}][cut{i}]overlay=0:{BAND}:eof_action=pass[z{i + 1}];")
        prev = f"z{i + 1}"

    parts.append(
        f"[{logo_idx}:v]scale=240:-1[lg];color=c=0x00E5FF:s=1080x10:r=30[bar];"
        f"[{prev}][lg]overlay=24:16[tmp];"
        f"[tmp][bar]overlay=x='-W+W*t/{total:.2f}':y=0:eval=frame:shortest=1[c];"
        f"[c]ass={ass_name}:fontsdir={FONTS_DIR}[v]"
    )
    return "".join(parts), audio


def check_rules(spec: Spec, total: float) -> list[str]:
    warns: list[str] = []
    cuts = sorted(spec.cutaways, key=lambda c: c.start)
    share = 0.0
    for i, c in enumerate(cuts):
        dur = c.end - c.start
        share += dur
        if dur < MIN_CUT:
            warns.append(f"컷어웨이 {i + 1}번이 {dur:.1f}초로 짧아")
        if i > 0 and c.start - cuts[i - 1].end < MIN_GAP:
            warns.append(f"컷어웨이 {i}번과 {i + 1}번 사이가 {MIN_GAP:.0f}초 미만이야")
    if share > total * MAX_SHARE:
        warns.append(f"스톡 비중이 {share / total:.0%}로 1/3 초과야")
    if cuts and clip_time(spec, cuts[-1].end) < total - 0.5:
        warns.append("마지막이 스톡으로 끝나지 않아")
    return warns


def render(spec_path: Path) -> Path:
    spec = Spec.load(spec_path)
    total = total_length(spec)
    spec.workdir.mkdir(parents=True, exist_ok=True)
    ass_name = "clip.ass"
    (spec.workdir / ass_name).write_text(build_ass(spec, total), encoding="utf-8-sig")
    for w in check_rules(spec, total):
        print(f"경고: {w}", file=sys.stderr)

    fc, audio = build_filter(spec, total, ass_name)
    cmd = ["ffmpeg", "-v", "error", "-y"]
    for a, b in spec.segments:
        cmd += ["-ss", f"{a:.3f}", "-t", f"{b - a:.3f}", "-i", spec.src]
    cmd += ["-loop", "1", "-i", str(LOGO_PNG)]
    for c in spec.cutaways:
        cmd += ["-i", c.file]
    cmd += ["-filter_complex", fc, "-map", "[v]", "-map", audio, "-t", f"{total:.3f}",
            "-c:v", "libx264", "-crf", "19", "-preset", "medium", "-c:a", "aac", "-b:a", "160k", spec.out]
    run(cmd, cwd=spec.workdir)
    return Path(spec.out)


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(description="설교 라이브 쇼츠 도구")
    sub = parser.add_subparsers(dest="cmd", required=True)
    p_fetch = sub.add_parser("fetch", help="자동 자막을 받아 줄별 시각 텍스트로 저장")
    p_fetch.add_argument("url")
    p_fetch.add_argument("--out", default=".", help="작업 폴더 (subs/ 아래에 저장)")
    p_render = sub.add_parser("render", help="사양 JSON 으로 쇼츠 렌더")
    p_render.add_argument("spec")
    args = parser.parse_args(argv)
    try:
        if args.cmd == "fetch":
            print(fetch(args.url, Path(args.out)))
        else:
            print(render(Path(args.spec)))
    except (RuntimeError, ValueError, FileNotFoundError, KeyError) as err:
        print(f"오류: {err}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
