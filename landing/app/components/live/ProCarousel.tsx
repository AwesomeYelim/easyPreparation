"use client";

import { useRef, type ReactNode } from "react";
import { useTicker } from "./hooks";
import { mmss } from "./mocks";

// Raycast-style extension cards: horizontal scroller, each card running a
// tiny loop of the Pro automation it describes.

function Card({ title, desc, tint, children }: { title: string; desc: string; tint: string; children: ReactNode }) {
  return (
    <div
      className="flex h-[400px] w-[300px] shrink-0 snap-start flex-col overflow-hidden rounded-2xl border border-white/[0.08] sm:w-[340px]"
      style={{ background: `radial-gradient(420px 260px at 50% 0%, ${tint}, transparent 70%), #0f1117` }}
    >
      <div className="flex flex-col gap-2 p-6">
        <div className="flex items-center gap-2">
          <h3 className="text-[17px] font-bold text-white">{title}</h3>
          <span className="ml-auto rounded bg-accent px-1.5 text-[10px] font-extrabold text-bg">PRO</span>
        </div>
        <p className="text-sm leading-[1.6] text-muted">{desc}</p>
      </div>
      <div className="mx-6 border-t border-white/[0.06]" />
      <div className="flex flex-1 items-center justify-center p-6">{children}</div>
    </div>
  );
}

const scenes = ["전주", "찬양", "성경봉독", "설교", "광고"];
const titles = ["참 빛 곧 세상에 와서", "예수님을 믿으면", "자기 땅에 오매"];
const presets = [
  { x: 50, y: 30, w: 30, label: "강단" },
  { x: 20, y: 55, w: 34, label: "성가대" },
  { x: 50, y: 50, w: 86, label: "전경" },
];

export default function ProCarousel() {
  const ref = useRef<HTMLDivElement>(null);
  const t = useTicker(ref);
  const scroll = (d: number) => ref.current?.scrollBy({ left: d * 360, behavior: "smooth" });
  const scene = Math.floor(t / 2) % scenes.length;
  const preset = presets[Math.floor(t / 3) % presets.length];
  const countdown = 60 - (t % 64);

  return (
    <div className="flex flex-col gap-6">
      <div ref={ref} className="no-scrollbar -mx-6 flex snap-x snap-mandatory gap-5 overflow-x-auto scroll-px-6 px-6 pb-2">
        <Card title="OBS 씬 자동 전환" desc="순서가 넘어가면 그 항목에 맞는 씬으로 바뀝니다." tint="rgba(127,179,230,0.28)">
          <div className="flex w-full flex-col gap-2">
            {scenes.map((s, i) => (
              <div key={s} className={`flex items-center gap-3 rounded-lg px-4 py-2.5 text-sm transition-colors duration-500 ${i === scene ? "bg-accent-deep/40 text-white" : "bg-white/[0.03] text-dim"}`}>
                <span className={`h-2 w-2 rounded-full ${i === scene ? "bg-[#ef4444]" : "bg-white/15"}`} />
                {s}
                {i === scene && <span className="ml-auto text-[11px] font-bold text-[#fca5a5]">PROGRAM</span>}
              </div>
            ))}
          </div>
        </Card>

        <Card title="예배 시간 카운트다운" desc="시간이 되면 카운트다운 후 송출을 시작합니다." tint="rgba(232,170,90,0.25)">
          {countdown > 0 ? (
            <div className="text-center">
              <p className="num text-[64px] font-extrabold text-white">{mmss(countdown)}</p>
              <p className="text-sm text-dim">11:00 주일예배 송출까지</p>
            </div>
          ) : (
            <p className="live-pop flex items-center gap-2 text-[40px] font-extrabold text-[#ef4444]">
              <span className="live-blink h-4 w-4 rounded-full bg-[#ef4444]" /> LIVE
            </p>
          )}
        </Card>

        <Card title="YouTube 라이브 · 썸네일" desc="방송과 썸네일을 본문 제목으로 자동 생성합니다." tint="rgba(239,68,68,0.22)">
          <div className="w-full">
            <div key={Math.floor(t / 4)} className="live-fade relative flex aspect-video flex-col justify-end overflow-hidden rounded-lg bg-[linear-gradient(135deg,#1e3a5f,#0b1220_60%,#3b2a14)] p-4">
              <span className="absolute left-3 top-3 rounded bg-[#ef4444] px-1.5 text-[10px] font-bold text-white">LIVE</span>
              <p className="text-[11px] text-accent-light">주일예배 · 요한복음 1장</p>
              <p className="text-lg font-extrabold leading-snug text-white">{titles[Math.floor(t / 4) % titles.length]}</p>
            </div>
            <p className="num mt-3 text-sm text-dim">
              <span className="text-[#ef4444]">●</span> 시청 중 {(48 + ((t * 7) % 90)).toLocaleString("ko-KR")}명
            </p>
          </div>
        </Card>

        <Card title="PTZ 카메라 프리셋" desc="찬양엔 성가대, 설교엔 강단. 카메라가 순서를 따라갑니다." tint="rgba(110,231,183,0.18)">
          <div className="relative aspect-[4/3] w-full rounded-lg border border-white/10 bg-[radial-gradient(circle_at_50%_20%,#1f2a3d,#0b0f17)]">
            <span className="absolute left-1/2 top-[22%] h-6 w-14 -translate-x-1/2 rounded bg-white/10" />
            <span className="absolute left-[12%] top-[48%] h-8 w-[30%] rounded bg-white/[0.07]" />
            <span className="absolute bottom-[10%] left-[10%] right-[10%] h-[18%] rounded bg-white/[0.05]" />
            <span
              className="absolute -translate-x-1/2 -translate-y-1/2 rounded border-2 border-[#6ee7b7] transition-all duration-700 ease-in-out"
              style={{ left: `${preset.x}%`, top: `${preset.y}%`, width: `${preset.w}%`, aspectRatio: "16 / 9" }}
            >
              <span className="absolute -top-5 left-0 text-[11px] font-bold text-[#6ee7b7]">{preset.label}</span>
            </span>
          </div>
        </Card>

        <Card title="자동 스케줄러" desc="매주 같은 시간, 같은 순서. 켜 두기만 하면 됩니다." tint="rgba(169,205,240,0.2)">
          <div className="grid w-full grid-cols-7 gap-1.5 text-center text-[11px]">
            {["월", "화", "수", "목", "금", "토", "주"].map((d, i) => (
              <div key={d} className="flex flex-col gap-1.5">
                <span className="text-dim">{d}</span>
                <span className={`flex h-14 items-center justify-center rounded-md ${i === 2 || i === 6 ? "bg-accent-deep/40 text-white" : "bg-white/[0.03] text-faint"} ${i === 6 && t % 2 ? "ring-1 ring-accent" : ""}`}>
                  {i === 2 ? "19:30" : i === 6 ? "11:00" : "-"}
                </span>
              </div>
            ))}
          </div>
        </Card>
      </div>
      <div className="flex justify-end gap-2">
        <button type="button" aria-label="이전" onClick={() => scroll(-1)} className="flex h-11 w-11 items-center justify-center rounded-full border border-white/10 bg-white/[0.04] text-muted transition hover:text-white">‹</button>
        <button type="button" aria-label="다음" onClick={() => scroll(1)} className="flex h-11 w-11 items-center justify-center rounded-full border border-white/10 bg-white/[0.04] text-muted transition hover:text-white">›</button>
      </div>
    </div>
  );
}
