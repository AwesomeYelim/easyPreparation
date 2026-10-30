"use client";

import Stage from "./Stage";
import { palette, sequence, type Item } from "./data";

// Coded replicas of the app screens (Pro Console, Display, mobile remote).
// Each one is drawn at a fixed design size and scaled by <Stage>.

const N = sequence.length;
export const at = (i: number) => sequence[((i % N) + N) % N];
const pad2 = (n: number) => String(n).padStart(2, "0");
export const mmss = (s: number) => `${pad2(Math.floor(s / 60))}:${pad2(s % 60)}`;

/* ------------------------------------------------------------- Display */

type SlideMode = "projector" | "overlay" | "stage";

function SlideBody({ item, index, t, mode }: { item: Item; index: number; t: number; mode: SlideMode }) {
  if (mode === "stage") {
    const next = at(index + 1);
    return (
      <div className="flex h-full flex-col bg-black p-14 text-white">
        <div className="flex items-center justify-between text-[28px] text-[#8a8a8a]">
          <span>무대 모니터</span>
          <span className="num rounded-lg bg-[#1f2937] px-5 py-2 text-[40px] font-bold text-[#facc15]">
            {mmss(t)}
          </span>
        </div>
        <div className="mt-auto">
          <p className="text-[30px] text-[#7fb3e6]">지금</p>
          <p className="text-[96px] font-extrabold leading-tight">
            {item.title} <span className="text-[#9ca3af]">{item.detail !== "-" && item.detail}</span>
          </p>
          <p className="mt-6 text-[34px] text-[#9ca3af]">
            다음 · {next.title} {next.detail !== "-" && next.detail}
          </p>
        </div>
      </div>
    );
  }

  if (mode === "overlay") {
    return (
      <div className="relative h-full bg-[radial-gradient(900px_500px_at_30%_30%,#3b4a63,#121826)]">
        {/* stand-in for the camera feed */}
        <div className="absolute bottom-0 left-[38%] h-[62%] w-[24%] rounded-t-[180px] bg-[#0b1220]/70" />
        <div className="absolute bottom-0 left-[44%] h-[24%] w-[12%] -translate-y-[250%] rounded-full bg-[#0b1220]/70" />
        <div className="absolute inset-x-16 bottom-14 rounded-xl bg-black/75 px-10 py-7 text-white backdrop-blur">
          <p className="text-[30px] font-semibold text-[#7fb3e6]">
            {item.title} {item.detail !== "-" && `· ${item.detail}`}
          </p>
          <p className="mt-2 text-[44px] font-bold leading-snug">
            {item.verse ?? (item.kind === "hymn" ? `찬송가 ${item.detail}` : item.who)}
          </p>
        </div>
      </div>
    );
  }

  if (item.img) {
    return (
      <div className="flex h-full items-center justify-center bg-black">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img src={item.img} alt="찬송가 악보 슬라이드" className="h-full w-full object-contain" />
      </div>
    );
  }
  return (
    <div className="flex h-full flex-col items-center justify-center bg-black px-24 text-center text-white">
      {item.verse ? (
        <>
          <p className="text-[64px] font-bold leading-[1.5]">{item.verse}</p>
          <p className="mt-10 text-[34px] text-[#9ca3af]">{item.detail}</p>
        </>
      ) : (
        <>
          <p className="text-[120px] font-extrabold">
            {item.kind === "hymn" ? item.detail : item.title}
          </p>
          <p className="mt-4 text-[36px] text-[#9ca3af]">
            {item.kind === "hymn" ? item.who : item.detail !== "-" ? item.detail : item.who}
          </p>
        </>
      )}
    </div>
  );
}

export function Slide({ index, t = 0, mode = "projector", className = "" }: { index: number; t?: number; mode?: SlideMode; className?: string }) {
  return (
    <Stage w={1280} h={720} className={`bg-black ${className}`}>
      <div key={index} className="live-fade h-full">
        <SlideBody item={at(index)} index={index} t={t} mode={mode} />
      </div>
    </Stage>
  );
}

/* ----------------------------------------------------------- Pro Console */

const ROW = 40;

function Sidebar({ index, rows = 13 }: { index: number; rows?: number }) {
  const cur = ((index % N) + N) % N;
  const offset = Math.min(Math.max(0, cur - Math.floor(rows / 2)), N - rows) * ROW;
  return (
    <div className="flex h-full flex-col border-r border-[#2e2e2e] bg-[#202020]">
      <div className="flex items-center gap-2 border-b border-[#2e2e2e] px-4 py-3 text-[13px]">
        <span className="h-2.5 w-2.5 rounded-full bg-[#ef4444]" />
        <span className="text-[#9ca3af]">OBS</span>
        <span className="ml-auto rounded border border-[#3a3a3a] px-2 py-1 text-[#9ca3af]">테스트 방송</span>
        <span className="rounded bg-[#dc2626] px-2 py-1 font-bold text-white">방송 시작</span>
      </div>
      <div className="border-b border-[#2e2e2e] px-4 py-3 text-[15px]">
        <span key={cur} className="live-fade inline-block">
          <span className="font-bold text-[#60a5fa]">{at(cur).title}</span>
          <span className="text-[#8a8a8a]"> — {at(cur).detail}</span>
        </span>
      </div>
      <div className="flex items-center gap-2 border-b border-[#2e2e2e] px-4 py-3">
        <span className="flex h-9 flex-1 items-center justify-center rounded-md border border-[#3a3a3a] bg-[#262626] text-[14px] font-bold text-white">◀ PREV</span>
        <span className="num w-14 text-center text-[13px] text-[#8a8a8a]">{cur + 1} / {N}</span>
        <span key={cur} className="live-press flex h-9 flex-1 items-center justify-center rounded-md border border-[#3a3a3a] bg-[#262626] text-[14px] font-bold text-white">NEXT ▶</span>
      </div>
      <div className="relative flex-1 overflow-hidden">
        <div className="transition-transform duration-500 ease-out" style={{ transform: `translateY(-${offset}px)` }}>
          {sequence.map((it, i) => {
            const on = i === cur;
            return (
              <div
                key={i}
                className={`flex items-center gap-3 border-l-[3px] px-4 text-[15px] transition-colors duration-300 ${on ? "border-[#3b82f6] bg-[#26324a]" : "border-transparent"}`}
                style={{ height: ROW }}
              >
                <span className="num w-5 text-right text-[12px] text-[#6b6b6b]">{i + 1}</span>
                <span className={on ? "font-semibold text-white" : "text-[#e5e5e5]"}>{it.title}</span>
                <span className="truncate text-[#7a7a7a]">{it.detail !== "-" && it.detail}</span>
                <span
                  className={`ml-auto h-6 w-7 shrink-0 rounded-md border ${it.kind === "hymn" || it.kind === "prayer" ? "border-[#3b82f6] bg-[#1e3a8a]/50" : "border-[#10b981] bg-[#064e3b]/50"}`}
                />
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

function Timeline({ index, t, period }: { index: number; t: number; period: number }) {
  const cur = ((index % N) + N) % N;
  const p = ((t % period) + 1) / period;
  return (
    <div className="border-t border-[#2e2e2e] bg-[#1a1a1a]">
      <div className="flex items-center gap-3 px-4 py-2 text-[11px] font-semibold tracking-[0.12em] text-[#7a7a7a]">
        SERVICE TIMELINE
        <span className="ml-auto">AUTO ▶</span>
        <span className="num text-[#60a5fa]">● {mmss(t)}</span>
      </div>
      <div className="flex h-[52px]">
        {sequence.map((it, i) => (
          <div key={i} className={`relative flex-1 overflow-hidden border-r border-[#2a2a2a] px-2 pt-1.5 text-[12px] ${i === cur ? "bg-[#1e3a5f] text-[#93c5fd]" : i < cur ? "text-[#6b6b6b]" : "text-[#a3a3a3]"}`}>
            <p className="truncate">{it.title}</p>
            <p className="truncate text-[10px] opacity-70">{it.detail !== "-" && it.detail}</p>
            {i === cur && (
              <span className="absolute bottom-0 left-0 h-[3px] bg-[#3b82f6] transition-[width] duration-1000 ease-linear" style={{ width: `${p * 100}%` }} />
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

export function ConsoleMock({ index, t, period, clockBase = 10 * 3600 + 58 * 60 }: { index: number; t: number; period: number; clockBase?: number }) {
  const cur = ((index % N) + N) % N;
  const clock = clockBase + t;
  const hh = pad2(Math.floor(clock / 3600) % 24);
  return (
    <div className="flex h-full flex-col bg-[#1c1c1c] font-sans text-[#e5e5e5]">
      <div className="flex h-12 items-center gap-4 border-b border-[#2e2e2e] px-4 text-[14px]">
        <span className="rounded-md bg-[#2563eb] px-1.5 text-[12px] font-extrabold italic text-white">ep</span>
        <span className="font-bold text-white">easyPreparation</span>
        <span className="text-[#6b6b6b]">Pro Console</span>
        <span className="flex items-center gap-1.5 text-[#ef4444]">
          <span className="live-blink h-2 w-2 rounded-full bg-[#ef4444]" /> ON AIR
        </span>
        <span className="num rounded border border-[#3a3a3a] px-2 py-0.5 text-[13px]">
          {hh}:{mmss(clock % 3600)}
        </span>
        <span className="ml-6 border-b-2 border-[#3b82f6] py-3 font-semibold text-[#60a5fa]">Bulletin</span>
        <span className="text-[#9ca3af]">Hymns</span>
        <span className="text-[#9ca3af]">Scripture</span>
        <span className="ml-auto rounded-md bg-[#3b82f6] px-4 py-1.5 font-bold text-white">Broadcast</span>
      </div>
      <div className="flex min-h-0 flex-1">
        <div className="flex w-[52px] flex-col items-center gap-5 border-r border-[#2e2e2e] pt-4">
          <span className="h-8 w-8 rounded-lg bg-[#26324a]" />
          <span className="h-5 w-5 rounded bg-[#3a3a3a]" />
          <span className="h-5 w-5 rounded bg-[#3a3a3a]" />
        </div>
        <div className="w-[300px]">
          <Sidebar index={cur} />
        </div>
        <div className="flex min-w-0 flex-1 flex-col gap-4 px-7 py-6">
          <div className="flex items-center gap-3">
            <h4 className="text-[24px] font-extrabold text-white">Worship Sequence</h4>
            <span className="ml-auto rounded-lg border border-[#3a3a3a] px-3 py-1.5 text-[13px]">주일예배 ▾</span>
            <span className="rounded-lg border border-[#3a3a3a] px-3 py-1.5 text-[13px]">PDF 다운로드</span>
            <span className="rounded-lg bg-[#3b82f6] px-3 py-1.5 text-[13px] font-bold text-white">프로젝터에 보내기</span>
          </div>
          <div className="flex-1 overflow-hidden rounded-xl border border-[#2e2e2e] bg-[#212121] px-5 py-3">
            <p className="mb-1 text-[13px] tracking-[0.2em] text-[#8a8a8a]">생성된 예배 내용</p>
            {sequence.slice(0, 12).map((it, i) => (
              <div key={i} className={`grid grid-cols-[1fr_1.3fr_1fr] items-center border-b border-[#2a2a2a] py-[7px] text-[14px] transition-colors duration-300 ${i === cur ? "text-[#93c5fd]" : ""}`}>
                <span className="font-semibold">{it.title}</span>
                <span className="truncate text-center text-[#9ca3af]">{it.detail}</span>
                <span className="text-right text-[#9ca3af]">{it.who}</span>
              </div>
            ))}
          </div>
        </div>
        <div className="flex w-[280px] flex-col gap-3 border-l border-[#2e2e2e] px-4 py-5">
          <p className="text-[15px] font-bold text-white">Studio</p>
          <div className="flex gap-3 text-[12px] text-[#8a8a8a]">
            <span className="text-[#60a5fa]">미리보기</span><span>화면</span><span>배경</span><span>OBS</span>
          </div>
          <p className="mt-2 text-[12px] text-[#8a8a8a]">Display 화면 실시간 미리보기</p>
          <div className="overflow-hidden rounded-lg border border-[#3b82f6]/60">
            <Slide index={cur} />
          </div>
          <p className="text-[12px] text-[#8a8a8a]">다음</p>
          <div className="overflow-hidden rounded-lg border border-[#2e2e2e] opacity-70">
            <Slide index={cur + 1} />
          </div>
        </div>
      </div>
      <Timeline index={cur} t={t} period={period} />
    </div>
  );
}

/* --------------------------------------------------------- Mobile remote */

export function PhoneMock({ index }: { index: number }) {
  const cur = ((index % N) + N) % N;
  const list = [cur - 1, cur, cur + 1, cur + 2].filter((i) => i >= 0 && i < N);
  return (
    <div className="flex h-full flex-col bg-[#f5f7fc] font-sans text-[#0f1b3d]">
      <div className="flex h-16 items-center gap-4 border-b border-[#dde3ef] px-6">
        <span className="text-[22px]">☰</span>
        <span className="text-[22px] font-extrabold">easyPreparation</span>
        <span className="ml-auto h-2.5 w-2.5 rounded-full bg-[#22c55e]" />
      </div>
      <div className="mx-5 mt-5 overflow-hidden rounded-2xl border-2 border-[#3b82f6]">
        <Slide index={cur} />
      </div>
      <div className="mx-5 mt-4 grid grid-cols-3 gap-3 text-center text-[12px] font-bold tracking-wide">
        <span className="rounded-2xl bg-[#3b6ff5] py-4 text-white">STAGE<br />DISPLAY</span>
        <span className="rounded-2xl border border-[#cfd6e4] bg-white py-4">AUTO<br />TIMER</span>
        <span className="rounded-2xl border border-[#cfd6e4] bg-white py-4">● LIVE</span>
      </div>
      <div className="mx-5 mt-6 flex items-end justify-between">
        <span className="text-[28px] font-extrabold">Sequence</span>
        <span className="num text-[18px] font-bold text-[#2563eb]">{cur + 1} / {N}</span>
      </div>
      <div className="mx-5 mt-3 flex flex-col gap-2.5">
        {list.map((i) => (
          <div key={i} className={`flex items-center gap-4 rounded-2xl px-4 py-3 transition-colors duration-300 ${i === cur ? "bg-[#dbe6ff] ring-2 ring-[#3b6ff5]" : "bg-[#e9eef8]"}`}>
            <span className="num flex h-11 w-11 items-center justify-center rounded-full bg-white text-[15px] font-bold">{pad2(i + 1)}</span>
            <span>
              <span className="block text-[18px] font-bold">{sequence[i].title}</span>
              <span className="block text-[14px] text-[#5b6680]">{sequence[i].detail}</span>
            </span>
          </div>
        ))}
      </div>
      <div className="mx-5 mb-5 mt-auto grid grid-cols-[1fr_2fr] gap-3">
        <span className="rounded-2xl border border-[#cfd6e4] bg-[#e9eef8] py-4 text-center text-[18px] font-bold">← 이전</span>
        <span key={cur} className="live-press rounded-2xl bg-[#0b57d0] py-4 text-center text-[18px] font-bold text-white">다음 →</span>
      </div>
    </div>
  );
}

/* ------------------------------------------------------ Bulletin editor */

/** `step` grows by one per tick: chips get picked, rows fill in, PDF is ready. */
export function BulletinMock({ step }: { step: number }) {
  const total = sequence.length;
  const picked = Math.min(step, total);
  const done = step >= total;
  const nextChip = done ? "" : sequence[picked].title;
  return (
    <div className="flex h-full flex-col gap-5 bg-[#1c1c1c] px-9 py-7 font-sans text-[#e5e5e5]">
      <div className="flex items-center gap-3">
        <h4 className="text-[28px] font-extrabold text-white">Worship Sequence</h4>
        <span className="ml-auto rounded-lg border border-[#3a3a3a] px-4 py-2 text-[15px]">주일예배 ▾</span>
        <span className="rounded-lg border border-[#3a3a3a] px-4 py-2 text-[15px]">주보 시안 ▾</span>
        <span className={`rounded-lg px-4 py-2 text-[15px] font-bold transition-colors duration-300 ${done ? "live-press bg-[#3b82f6] text-white" : "border border-[#3a3a3a]"}`}>
          {done ? "✓ PDF 생성 완료" : "PDF 다운로드"}
        </span>
      </div>
      <div className="grid min-h-0 flex-1 grid-cols-[1.1fr_1fr] gap-5">
        <div className="flex flex-col gap-5">
          <div className="rounded-xl border border-[#2e2e2e] bg-[#212121] p-5">
            <p className="mb-3 text-[14px] tracking-[0.2em] text-[#8a8a8a]">예배 순서 선택하기</p>
            <div className="flex flex-wrap gap-2.5">
              {palette.map((c) => (
                <span key={c} className={`rounded-full border px-4 py-1.5 text-[16px] font-semibold transition-colors duration-300 ${c === nextChip ? "border-[#3b82f6] bg-[#1e3a8a]/60 text-white" : "border-[#3a3a3a] bg-[#262626]"}`}>
                  {c}
                </span>
              ))}
            </div>
          </div>
          <div className="flex-1 rounded-xl border border-[#2e2e2e] bg-[#212121] p-5">
            <p className="mb-3 text-[14px] tracking-[0.2em] text-[#8a8a8a]">선택된 예배 순서</p>
            <div className="flex flex-wrap gap-2.5">
              {sequence.slice(0, picked).map((it, i) => (
                <span key={i} className="live-pop rounded-full border border-[#3a3a3a] bg-[#262626] px-3.5 py-1.5 text-[15px] font-semibold">
                  <span className="mr-1.5 text-[#6b6b6b]">⠿</span>{it.title}
                </span>
              ))}
            </div>
          </div>
        </div>
        <div className="overflow-hidden rounded-xl border border-[#2e2e2e] bg-[#212121] px-6 py-5">
          <p className="mb-2 text-[14px] tracking-[0.2em] text-[#8a8a8a]">생성된 예배 내용</p>
          {sequence.slice(0, picked).map((it, i) => (
            <div key={i} className="live-fade grid grid-cols-[1fr_1.3fr_1fr] border-b border-[#2a2a2a] py-[9px] text-[16px]">
              <span className="font-bold text-white">{it.title}</span>
              <span className="truncate text-center text-[#9ca3af]">{it.detail}</span>
              <span className="text-right text-[#9ca3af]">{it.who}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

/* --------------------------------------------- Display: three outputs */

export function DisplayMock({ index, t }: { index: number; t: number }) {
  const label = "mb-2 text-[15px] font-semibold text-[#9ca3af]";
  return (
    <div className="grid h-full grid-cols-[1.75fr_1fr] gap-6 bg-[#0e0e10] p-8 font-sans">
      <div className="flex flex-col justify-center">
        <p className={label}>프로젝터 · /display</p>
        <div className="overflow-hidden rounded-xl border border-[#2e2e2e]">
          <Slide index={index} t={t} />
        </div>
      </div>
      <div className="flex flex-col justify-center gap-6">
        <div>
          <p className={label}>방송 오버레이 · OBS</p>
          <div className="overflow-hidden rounded-xl border border-[#2e2e2e]">
            <Slide index={index} t={t} mode="overlay" />
          </div>
        </div>
        <div>
          <p className={label}>무대 모니터</p>
          <div className="overflow-hidden rounded-xl border border-[#2e2e2e]">
            <Slide index={index} t={t} mode="stage" />
          </div>
        </div>
      </div>
    </div>
  );
}

/* ----------------------------------------------- Phone driving the screen */

export function RemoteMock({ index }: { index: number }) {
  return (
    <div className="flex h-full items-center justify-center gap-10 bg-[radial-gradient(700px_400px_at_30%_50%,#1b2a44,#0e0e10)] px-10 font-sans">
      <div className="h-[660px] w-[305px] shrink-0 overflow-hidden rounded-[40px] border-[8px] border-[#1c2740] shadow-[0_24px_48px_rgba(0,0,0,0.55)]">
        <Stage w={390} h={844}>
          <PhoneMock index={index} />
        </Stage>
      </div>
      <div className="flex flex-col items-center gap-2 text-[#7fb3e6]">
        <span className="live-wave text-[34px]">)))</span>
        <span className="text-[14px] text-[#9ca3af]">같은 Wi-Fi</span>
      </div>
      <div className="flex-1">
        <p className="mb-2 text-[15px] font-semibold text-[#9ca3af]">프로젝터 화면</p>
        <div className="overflow-hidden rounded-xl border border-[#2e2e2e]">
          <Slide index={index} />
        </div>
      </div>
    </div>
  );
}
