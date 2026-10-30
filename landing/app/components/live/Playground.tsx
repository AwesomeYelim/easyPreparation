"use client";

import { useEffect, useRef, useState, type ReactNode } from "react";
import { IconBook, IconDoc, IconMusic, IconScreen } from "../icons";
import Stage from "./Stage";
import { useTicker } from "./hooks";
import { sequence } from "./data";
import { BulletinMock, DisplayMock, PhoneMock, RemoteMock, Slide } from "./mocks";

// Raycast-style playground: one big app frame, a dock that swaps what runs
// inside it, and keyboard control (← → for slides, digits + Enter for hymns).
// Plays on its own until the visitor touches anything.

const IconPhone = () => (
  <svg width={26} height={26} viewBox="0 0 24 24" fill="none" stroke="#7fb3e6" strokeWidth={1.7} strokeLinecap="round" aria-hidden>
    <rect x="7" y="2.5" width="10" height="19" rx="2.5" />
    <path d="M11 18.5h2" />
  </svg>
);

const verses = [
  { n: 9, ko: "참 빛 곧 세상에 와서 각 사람에게 비취는 빛이 있었나니", en: "That was the true Light, which lighteth every man that cometh into the world." },
  { n: 10, ko: "그가 세상에 계셨으며 세상은 그로 말미암아 지은바 되었으되 세상이 그를 알지 못하였고", en: "He was in the world, and the world was made by him, and the world knew him not." },
  { n: 11, ko: "자기 땅에 오매 자기 백성이 영접지 아니하였으나", en: "He came unto his own, and his own received him not." },
  { n: 12, ko: "영접하는 자 곧 그 이름을 믿는 자들에게는 하나님의 자녀가 되는 권세를 주셨으니", en: "But as many as received him, to them gave he power to become the sons of God, even to them that believe on his name." },
];

/* ---------------------------------------------------------------- Hymn */

// Real score pages shipped for the demo (/public/hymns/{n}/1-4.jpg). The app
// itself pulls all 645 from the asset server; the landing page carries a few.
const HYMNS: Record<number, string> = {
  28: "복의 근원 강림하사",
  96: "예수님은 누구신가",
  288: "예수를 나의 구주 삼고",
  524: "갈 길을 밝히 보이시니",
};
const DEMO_HYMNS = Object.keys(HYMNS).map(Number);
const PAGES = 4;

function HymnPage({ number, page, className = "" }: { number: number; page: number; className?: string }) {
  if (HYMNS[number]) {
    return (
      // eslint-disable-next-line @next/next/no-img-element
      <img src={`/hymns/${number}/${(page % PAGES) + 1}.jpg`} alt={`찬송가 ${number}장 ${HYMNS[number]} 악보 ${(page % PAGES) + 1}`} className={`object-contain ${className}`} />
    );
  }
  return (
    <div className={`flex flex-col items-center justify-center gap-3 bg-black px-6 text-center text-white ${className}`}>
      <p className="text-[clamp(28px,8cqw,120px)] font-extrabold leading-none">{number}장</p>
      <p className="text-[clamp(12px,2.6cqw,34px)] text-[#9ca3af]">앱에서는 645곡 악보가 전부 자동으로 뜹니다.</p>
      <p className="text-[clamp(11px,2.2cqw,28px)] text-[#7fb3e6]">체험판 악보: {DEMO_HYMNS.join(" · ")}장</p>
    </div>
  );
}

function HymnMock({ typed, number, page, onPick }: { typed: string; number: number; page: number; onPick: (n: number) => void }) {
  const pages = PAGES;
  const p = page % pages;
  return (
    <div className="grid h-full grid-cols-[360px_1fr] bg-[#1c1c1c] font-sans text-[#e5e5e5]">
      <div className="flex flex-col gap-4 border-r border-[#2e2e2e] p-7">
        <p className="text-[15px] font-bold text-white">찬송가 · 새찬송가 645곡</p>
        <div className="flex h-14 items-center rounded-xl border-2 border-[#3b82f6] bg-[#111] px-4 text-[26px] font-bold text-white">
          <span className="num">{typed || <span className="text-[#555]">번호 입력</span>}</span>
          <span className="live-caret ml-0.5 h-7 w-[2px] bg-[#60a5fa]" />
          <span className="ml-auto text-[15px] font-medium text-[#7a7a7a]">Enter</span>
        </div>
        <div key={number} className="live-fade flex flex-col gap-2 text-[15px]">
          <p className="text-[#8a8a8a]">불러온 곡</p>
          <div className="flex flex-col gap-1 rounded-lg bg-[#26324a] px-4 py-3">
            <span className="flex items-center">
              <span className="num text-[22px] font-extrabold text-white">{number}장</span>
              {HYMNS[number] && <span className="ml-auto rounded bg-[#064e3b] px-2 py-0.5 text-[12px] text-[#6ee7b7]">악보 캐시됨</span>}
            </span>
            <span className="text-[14px] text-[#93c5fd]">{HYMNS[number] ?? "체험판에 없는 곡"}</span>
          </div>
          <p className="mt-2 text-[#8a8a8a]">체험판 악보</p>
          {DEMO_HYMNS.filter((n) => n !== number).map((n) => (
            <button key={n} type="button" onClick={() => onPick(n)} className="flex items-center gap-3 rounded-lg px-4 py-2 text-left text-[#c4c4c4] hover:bg-white/5">
              <span className="num w-12 font-semibold">{n}장</span>
              <span className="truncate text-[13px] text-[#8a8a8a]">{HYMNS[n]}</span>
            </button>
          ))}
        </div>
        <div className="mt-auto flex gap-2">
          {Array.from({ length: pages }, (_, i) => (
            <span key={i} className={`h-1.5 flex-1 rounded-full ${i === p ? "bg-[#3b82f6]" : "bg-[#333]"}`} />
          ))}
        </div>
      </div>
      <div className="flex items-center justify-center p-8">
        <div key={`${number}-${p}`} className="live-fade w-full overflow-hidden rounded-xl border border-[#2e2e2e] bg-black [container-type:inline-size]">
          <HymnPage number={number} page={p} className="aspect-video h-full w-full" />
        </div>
      </div>
    </div>
  );
}

/* ----------------------------------------------------------- Scripture */

function ScriptureMock({ verse }: { verse: number }) {
  const v = verses[verse % verses.length];
  return (
    <div className="flex h-full flex-col gap-5 bg-[#1c1c1c] p-8 font-sans text-[#e5e5e5]">
      <div className="flex items-center gap-3 text-[15px]">
        <span className="rounded-lg border border-[#3a3a3a] px-4 py-2 font-bold text-white">요한복음</span>
        <span className="rounded-lg border border-[#3a3a3a] px-4 py-2">1장</span>
        <span className="num rounded-lg border border-[#3b82f6] px-4 py-2 text-[#93c5fd]">{v.n}절</span>
        <span className="ml-auto rounded-full bg-[#26324a] px-4 py-1.5 text-[#93c5fd]">비교 모드</span>
        <span className="rounded-full border border-[#3a3a3a] px-4 py-1.5">개역한글</span>
        <span className="rounded-full border border-[#3a3a3a] px-4 py-1.5">KJV</span>
      </div>
      <div className="flex-1 overflow-hidden rounded-xl border border-[#2e2e2e]">
        <Stage w={1280} h={620} className="bg-black">
          <div key={v.n} className="live-fade grid h-full grid-cols-2 gap-16 px-20 py-16 text-white">
            <div className="flex flex-col justify-center">
              <p className="mb-6 text-[26px] text-[#7fb3e6]">개역한글 · 요 1:{v.n}</p>
              <p className="text-[50px] font-bold leading-[1.5]">{v.ko}</p>
            </div>
            <div className="flex flex-col justify-center border-l border-white/10 pl-16">
              <p className="mb-6 text-[26px] text-[#e8aa5a]">KJV · John 1:{v.n}</p>
              <p className="text-[44px] leading-[1.5] text-[#d1d5db]">{v.en}</p>
            </div>
          </div>
        </Stage>
      </div>
    </div>
  );
}

/* ----------------------------------------------------------- Playground */

type Mode = { key: string; label: string; icon: ReactNode; title: string; desc: string };

const modes: Mode[] = [
  { key: "bulletin", label: "주보", icon: <IconDoc />, title: "주보 편집.", desc: "순서를 고르면 표가 채워지고, 인쇄용 PDF가 바로 나옵니다." },
  { key: "hymn", label: "찬송", icon: <IconMusic />, title: "찬송 번호만.", desc: "번호를 치고 Enter — 악보가 슬라이드로 뜹니다. 여기선 28 · 96 · 288 · 524장을, 앱에선 645곡 전부를." },
  { key: "bible", label: "성경", icon: <IconBook />, title: "성경 비교 모드.", desc: "두 번역판을 나란히. ← → 로 절을 넘깁니다." },
  { key: "display", label: "송출", icon: <IconScreen />, title: "화면 세 가지.", desc: "프로젝터 · 방송 오버레이 · 무대 모니터가 같은 순서로 움직입니다." },
  { key: "remote", label: "리모컨", icon: <IconPhone />, title: "폰으로 넘기기.", desc: "같은 Wi-Fi의 스마트폰이 리모컨이 됩니다." },
];

function Key({ children, onClick, wide, submit }: { children: ReactNode; onClick?: () => void; wide?: boolean; submit?: boolean }) {
  return (
    <button
      type={submit ? "submit" : "button"}
      onClick={onClick}
      className={`inline-flex h-8 items-center justify-center rounded-md border border-white/15 bg-white/[0.06] px-2 text-xs font-semibold text-ink shadow-[inset_0_-2px_0_rgba(255,255,255,0.06)] transition hover:border-accent hover:text-white active:translate-y-px ${wide ? "min-w-[56px]" : "min-w-[32px]"}`}
    >
      {children}
    </button>
  );
}

/**
 * Phone-width version: the whole console scaled down is unreadable, so show
 * only the output of each mode at real size.
 */
function MobileView({ mode, beat, bulletinStep, hymn, t }: { mode: number; beat: number; bulletinStep: number; hymn: number; t: number }) {
  const v = verses[beat % verses.length];
  const shown = Math.min(sequence.length, Math.max(4, bulletinStep));
  const body: ReactNode[] = [
    <div key="b" className="flex flex-col divide-y divide-white/[0.06] rounded-xl border border-white/[0.08] bg-[#17181c] px-4">
      <p className="py-3 text-xs tracking-[0.2em] text-dim">생성된 예배 내용</p>
      {sequence.slice(0, shown).slice(-6).map((it, i) => (
        <div key={`${it.title}-${i}`} className="live-fade grid grid-cols-[1fr_auto] gap-2 py-2.5 text-sm">
          <span className="font-semibold text-white">{it.title} <span className="font-normal text-dim">{it.detail !== "-" && it.detail}</span></span>
          <span className="text-dim">{it.who}</span>
        </div>
      ))}
      <p className="py-3 text-xs text-accent">{shown >= sequence.length ? "✓ 주보 PDF 생성 완료" : `${shown} / ${sequence.length} 항목`}</p>
    </div>,
    <div key="h" className="flex flex-col gap-2">
      <p className="text-sm text-muted">
        <b className="num text-white">{hymn}장</b> {HYMNS[hymn] ?? ""}
        <span className="num float-right text-dim">{(beat % PAGES) + 1} / {PAGES}</span>
      </p>
      <div key={`${hymn}-${beat % PAGES}`} className="live-fade overflow-hidden rounded-xl border border-white/[0.08] bg-black [container-type:inline-size]">
        <HymnPage number={hymn} page={beat} className="aspect-[4/3] w-full" />
      </div>
    </div>,
    <div key="s" className="flex flex-col gap-3">
      <div key={v.n} className="live-fade rounded-xl border border-white/[0.08] bg-black p-5">
        <p className="mb-2 text-xs text-accent">개역한글 · 요 1:{v.n}</p>
        <p className="text-lg font-bold leading-[1.6] text-white">{v.ko}</p>
      </div>
      <div key={`en-${v.n}`} className="live-fade rounded-xl border border-white/[0.08] bg-black p-5">
        <p className="mb-2 text-xs text-[#e8aa5a]">KJV · John 1:{v.n}</p>
        <p className="text-base leading-[1.6] text-[#d1d5db]">{v.en}</p>
      </div>
    </div>,
    <div key="d" className="flex flex-col gap-3">
      <p className="text-xs text-dim">프로젝터</p>
      <div className="overflow-hidden rounded-xl border border-white/[0.08]">
        <Slide index={[13, 1, 8, 3][beat % 4]} />
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div>
          <p className="mb-1.5 text-xs text-dim">방송 오버레이</p>
          <div className="overflow-hidden rounded-lg border border-white/[0.08]"><Slide index={[13, 1, 8, 3][beat % 4]} mode="overlay" /></div>
        </div>
        <div>
          <p className="mb-1.5 text-xs text-dim">무대 모니터</p>
          <div className="overflow-hidden rounded-lg border border-white/[0.08]"><Slide index={[13, 1, 8, 3][beat % 4]} t={t % 60} mode="stage" /></div>
        </div>
      </div>
    </div>,
    <div key="r" className="mx-auto w-[260px] overflow-hidden rounded-[30px] border-[6px] border-[#1c2740] shadow-[0_24px_48px_rgba(0,0,0,0.55)]">
      <Stage w={390} h={844}>
        <PhoneMock index={9 + beat} />
      </Stage>
    </div>,
  ];
  return (
    <div key={mode} className="live-fade w-full sm:hidden">
      {body[mode]}
    </div>
  );
}

export default function Playground() {
  const ref = useRef<HTMLDivElement>(null);
  const t = useTicker(ref);
  const [mode, setMode] = useState(1);
  const [manual, setManual] = useState(false);
  const [step, setStep] = useState(0);
  const [typed, setTyped] = useState("");
  const [hymn, setHymn] = useState(524);
  const [hover, setHover] = useState<number | null>(null);
  const visible = useRef(false);

  // Before the visitor interacts: cycle modes and "type" a hymn number.
  const autoMode = Math.floor(t / 9) % modes.length;
  const m = manual ? mode : (1 + autoMode) % modes.length;
  const beat = manual ? step : Math.floor((t % 9) / 2.5);
  const autoHymn = DEMO_HYMNS[(DEMO_HYMNS.indexOf(524) + Math.floor(t / 45)) % DEMO_HYMNS.length];
  const autoTyped = String(autoHymn).slice(0, Math.max(0, (t % 9) - 1));
  const shownTyped = manual ? typed : autoTyped.length < String(autoHymn).length ? autoTyped : "";
  const shownHymn = manual ? hymn : autoHymn;

  const take = () => {
    if (!manual) {
      setManual(true);
      setMode(m);
      setStep(0);
    }
  };
  const go = (d: number) => {
    take();
    setStep((s) => Math.max(0, s + d));
  };
  const pick = (i: number) => {
    setManual(true);
    setMode(i);
    setStep(0);
  };
  const commitHymn = (v = typed) => {
    const n = parseInt(v, 10);
    if (n >= 1 && n <= 645) {
      setHymn(n);
      setStep(0);
    }
    setTyped("");
  };

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const io = new IntersectionObserver(([e]) => (visible.current = e.isIntersecting), { threshold: 0.4 });
    io.observe(el);
    return () => io.disconnect();
  }, []);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (!visible.current || e.metaKey || e.ctrlKey || e.altKey) return;
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA") return;
      if (e.key === "ArrowRight") go(1);
      else if (e.key === "ArrowLeft") go(-1);
      else if (/^[0-9]$/.test(e.key)) {
        if (!manual || mode !== 1) pick(1);
        setTyped((v) => (v + e.key).slice(-3));
      } else if (e.key === "Backspace") setTyped((v) => v.slice(0, -1));
      else if (e.key === "Enter" && typed) commitHymn();
      else return;
      e.preventDefault();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  });

  const screens = [
    <BulletinMock key="b" step={manual ? 6 + step * 3 : (t % 9) * 2} />,
    <HymnMock key="h" typed={shownTyped} number={shownHymn} page={beat} onPick={(n) => { pick(1); commitHymn(String(n)); }} />,
    <ScriptureMock key="s" verse={beat} />,
    <DisplayMock key="d" index={[13, 1, 8, 3][beat % 4]} t={t % 60} />,
    <RemoteMock key="r" index={9 + beat} />,
  ];
  const cur = modes[m];

  return (
    <div ref={ref} className="flex flex-col items-center gap-7">
      <MobileView mode={m} beat={beat} bulletinStep={manual ? 6 + step * 3 : (t % 9) * 2} hymn={shownHymn} t={t} />
      <div className="hidden w-full rounded-[18px] border border-white/[0.12] bg-white/[0.03] p-2 shadow-[0_40px_120px_rgba(20,60,120,0.35),0_0_0_1px_rgba(0,0,0,0.5)] backdrop-blur sm:block">
        <div className="overflow-hidden rounded-[12px] border border-white/[0.08] bg-[#0d0d0f]">
          <div className="flex h-9 items-center gap-5 border-b border-white/[0.06] bg-black/40 px-4 text-[12px] text-[#a3a3a3]">
            <span className="font-bold text-white">easyPreparation</span>
            <span className="hidden sm:inline">파일</span>
            <span className="hidden sm:inline">편집</span>
            <span className="hidden sm:inline">예배</span>
            <span className="hidden sm:inline">창</span>
            <span className="ml-auto flex items-center gap-1.5 text-[#ef4444]">
              <span className="live-blink h-1.5 w-1.5 rounded-full bg-[#ef4444]" />
              ON AIR
            </span>
            <span className="hidden sm:inline">주일 오전 10:58</span>
          </div>
          <Stage w={1200} h={700}>
            <div key={m} className="live-fade h-full">
              {screens[m]}
            </div>
          </Stage>
        </div>
      </div>

      {/* Dock */}
      <div className="relative flex items-end gap-2 rounded-2xl border border-white/[0.1] bg-white/[0.04] p-2 shadow-[0_10px_40px_rgba(0,0,0,0.5)] backdrop-blur">
        {modes.map((md, i) => (
          <button
            key={md.key}
            type="button"
            aria-label={md.label}
            aria-pressed={i === m}
            onClick={() => pick(i)}
            onMouseEnter={() => setHover(i)}
            onMouseLeave={() => setHover(null)}
            className={`relative flex h-12 w-12 items-center justify-center rounded-xl border transition duration-200 sm:h-14 sm:w-14 ${i === m ? "-translate-y-1 border-accent/50 bg-accent-deep/25" : "border-white/[0.08] bg-white/[0.04] hover:-translate-y-1 hover:bg-white/[0.08]"}`}
          >
            {md.icon}
            {hover === i && (
              <span className="absolute -top-10 whitespace-nowrap rounded-md border border-white/10 bg-[#15161a] px-2.5 py-1 text-xs font-semibold text-white">
                {md.label}
              </span>
            )}
            {i === m && <span className="absolute -bottom-1.5 h-1 w-1 rounded-full bg-accent" />}
          </button>
        ))}
      </div>

      <p key={m} className="live-fade max-w-[560px] text-center text-[15px] leading-[1.7] text-dim">
        <b className="font-bold text-white">{cur.title}</b> {cur.desc}
      </p>

      <div className="flex flex-wrap items-center justify-center gap-2 text-xs text-dim">
        <Key onClick={() => go(-1)}>←</Key>
        <Key onClick={() => go(1)}>→</Key>
        <span className="mr-3">슬라이드 넘기기</span>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            pick(1);
            commitHymn();
          }}
          className="flex items-center gap-2"
        >
          <input
            value={typed}
            onChange={(e) => {
              take();
              setTyped(e.target.value.replace(/\D/g, "").slice(0, 3));
            }}
            onFocus={() => pick(1)}
            inputMode="numeric"
            placeholder="524"
            aria-label="찬송 번호"
            className="num h-8 w-16 rounded-md border border-white/15 bg-white/[0.06] px-2 text-center text-xs font-semibold text-white placeholder:text-faint focus:border-accent focus:outline-none"
          />
          <Key wide submit>Enter</Key>
          <span>찬송 불러오기</span>
        </form>
        <span className="flex w-full items-center justify-center gap-1.5 sm:w-auto">
          <span className="text-faint">체험판</span>
          {DEMO_HYMNS.map((n) => (
            <Key key={n} onClick={() => { pick(1); commitHymn(String(n)); }}>{n}</Key>
          ))}
        </span>
      </div>
    </div>
  );
}
