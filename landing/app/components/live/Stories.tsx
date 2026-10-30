"use client";

import { useRef, type ReactNode } from "react";
import Stage from "./Stage";
import { useTicker } from "./hooks";
import { ConsoleMock, PhoneMock, Slide } from "./mocks";

// Cursor-style feature rows: copy on one side, an app window floating over a
// public-domain painting on the other, with the work happening line by line.

type Line = { text: string; meta?: string; done?: boolean };

const paintings = {
  inwang: { src: "/art/inwang.jpg", credit: "정선, 〈인왕제색도〉 1751" },
  morning: { src: "/art/morning.jpg", credit: "C. D. Friedrich, 〈리젠산맥의 아침〉 1810" },
  sierra: { src: "/art/sierra.jpg", credit: "A. Bierstadt, 〈시에라네바다〉 1868" },
};

function Canvas({ art, children }: { art: keyof typeof paintings; children: ReactNode }) {
  const p = paintings[art];
  return (
    <figure className="flex flex-col gap-2">
      <div className="relative overflow-hidden rounded-2xl border border-white/[0.08]">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img src={p.src} alt="" loading="lazy" className="absolute inset-0 h-full w-full object-cover" />
        <div className="absolute inset-0 bg-[#0b1220]/25" />
        <div className="relative px-5 py-8 sm:px-10 sm:py-12">{children}</div>
      </div>
      <figcaption className="text-right text-[11px] text-faint">{p.credit} · 퍼블릭 도메인</figcaption>
    </figure>
  );
}

function Window({ title, children, className = "" }: { title: string; children: ReactNode; className?: string }) {
  return (
    <div className={`overflow-hidden rounded-xl border border-white/[0.12] bg-[#131417]/95 shadow-[0_30px_70px_rgba(0,0,0,0.55)] backdrop-blur ${className}`}>
      <div className="flex h-8 items-center gap-1.5 border-b border-white/[0.06] px-3">
        <span className="h-2.5 w-2.5 rounded-full bg-white/15" />
        <span className="h-2.5 w-2.5 rounded-full bg-white/15" />
        <span className="h-2.5 w-2.5 rounded-full bg-white/15" />
        <span className="min-w-0 flex-1 truncate pr-8 text-center text-[11px] text-dim">{title}</span>
      </div>
      {children}
    </div>
  );
}

/** Lines appear one per second while on screen, hold, then replay. */
function StepLog({ lines, footer }: { lines: Line[]; footer?: (done: boolean) => ReactNode }) {
  const ref = useRef<HTMLDivElement>(null);
  const t = useTicker(ref);
  const cycle = lines.length + 5;
  const shown = t === 0 ? lines.length : Math.min(lines.length, t % cycle);
  const done = shown === lines.length;
  return (
    <div ref={ref} className="flex min-h-[300px] flex-col gap-2.5 p-5 font-mono text-[13px]">
      {lines.slice(0, shown).map((l, i) => {
        const last = i === shown - 1 && !done;
        return (
          <div key={`${i}-${Math.floor(t / cycle)}`} className="live-fade flex items-baseline gap-2.5">
            <span className={l.done ? "text-[#6ee7b7]" : last ? "text-[#facc15]" : "text-[#60a5fa]"}>
              {l.done ? "✓" : last ? "●" : "›"}
            </span>
            <span className={l.done ? "font-semibold text-white" : "text-[#d4d4d8]"}>{l.text}</span>
            {l.meta && <span className="ml-auto shrink-0 text-[11px] text-faint">{l.meta}</span>}
          </div>
        );
      })}
      {!done && <span className="live-caret ml-5 h-4 w-[7px] bg-[#60a5fa]" />}
      {footer?.(done)}
    </div>
  );
}

function Row({ title, sub, children, flip }: { title: string; sub: string; children: ReactNode; flip?: boolean }) {
  return (
    <div className={`grid items-center gap-8 md:gap-14 ${flip ? "md:grid-cols-[1.45fr_1fr]" : "md:grid-cols-[1fr_1.45fr]"}`}>
      <div className={`flex flex-col gap-3 ${flip ? "md:order-2" : ""}`}>
        <h3 className="text-[26px] font-bold leading-[1.3] tracking-[-0.01em] text-white md:text-[30px]">{title}</h3>
        <p className="text-[17px] leading-[1.7] text-dim">{sub}</p>
      </div>
      <div className={flip ? "md:order-1" : ""}>{children}</div>
    </div>
  );
}

function PdfThumb() {
  return (
    <div className="live-pop mt-2 flex items-center gap-3 rounded-lg border border-white/10 bg-white/[0.04] p-2.5">
      <div className="flex h-14 w-11 flex-col gap-1 rounded-sm bg-white p-1.5">
        <span className="h-1 w-6 rounded bg-[#765c2d]" />
        {[8, 7, 8, 6, 8].map((w, i) => (
          <span key={i} className="h-[2px] rounded bg-[#c7c7c7]" style={{ width: `${w * 4}px` }} />
        ))}
      </div>
      <div className="font-sans">
        <p className="text-[13px] font-semibold text-white">주보_주일예배.pdf</p>
        <p className="text-[11px] text-dim">A4 · 2면 · 인쇄 준비 완료</p>
      </div>
    </div>
  );
}

function SyncedWindows() {
  const ref = useRef<HTMLDivElement>(null);
  const t = useTicker(ref);
  const index = 3 + Math.floor(t / 3);
  return (
    <div ref={ref} className="relative pb-[18%] pr-[16%]">
      <Window title="제어판 — Pro Console">
        <Stage w={1200} h={750}>
          <ConsoleMock index={index} t={t} period={3} clockBase={10 * 3600 + 59 * 60} />
        </Stage>
      </Window>
      <Window title="프로젝터" className="absolute bottom-0 left-[8%] w-[46%]">
        <Slide index={index} />
      </Window>
      <div className="absolute bottom-[4%] right-0 w-[22%] overflow-hidden rounded-[18px] border-[4px] border-[#1c2740] shadow-[0_24px_48px_rgba(0,0,0,0.6)]">
        <Stage w={390} h={844}>
          <PhoneMock index={index} />
        </Stage>
      </div>
    </div>
  );
}

export default function Stories() {
  return (
    <div className="flex flex-col gap-24 md:gap-32">
      <Row title="토요일 저녁, 주보 끝." sub="순서를 고르고 찬송 번호와 본문만 넣으면 됩니다. 악보 확인, 본문 삽입, 인쇄용 PDF까지 알아서 이어집니다.">
        <Canvas art="inwang">
          <Window title="주보 — 주일예배" className="mx-auto max-w-[560px]">
            <StepLog
              lines={[
                { text: "주일예배 순서 불러옴", meta: "15개 항목" },
                { text: "찬송 28 · 96 · 524장 악보 확인", meta: "캐시됨" },
                { text: "요한복음 1:9-14 본문 삽입", meta: "개역한글" },
                { text: "교회소식 3건 정리" },
                { text: "PDF 생성 · 시안 '클래식'", meta: "A4 2면" },
                { text: "완료", meta: "2.1초", done: true },
              ]}
              footer={(done) => done && <PdfThumb />}
            />
          </Window>
        </Canvas>
      </Row>

      <Row flip title="주일 아침, 이미 준비돼 있습니다." sub="토요일에 만든 순서가 그대로 예배 화면이 됩니다. 제어판에서 한 번 넘기면 프로젝터와 폰 리모컨이 함께 움직입니다.">
        <Canvas art="morning">
          <SyncedWindows />
        </Canvas>
      </Row>

      <Row title="방송도 순서를 따라갑니다." sub="예배 시간이 되면 카운트다운, YouTube 라이브, 씬 전환, 카메라 이동까지. 방송팀은 지켜보기만 하면 됩니다.">
        <Canvas art="sierra">
          <Window title="방송 자동화 — Pro" className="mx-auto max-w-[560px]">
            <StepLog
              lines={[
                { text: "OBS 연결됨 · 씬 6개", meta: "10:50" },
                { text: "YouTube 라이브 예약 · 썸네일 생성", meta: "10:52" },
                { text: "카운트다운 시작", meta: "10:59" },
                { text: "송출 시작 · 씬 '전주'", meta: "11:00" },
                { text: "찬송 28장 → 씬 '찬양' · PTZ 프리셋 2", meta: "11:06" },
                { text: "말씀 → 씬 '설교' · PTZ 프리셋 1", meta: "11:24" },
                { text: "예배 순서와 동기화 중", meta: "LIVE", done: true },
              ]}
            />
          </Window>
        </Canvas>
      </Row>
    </div>
  );
}
