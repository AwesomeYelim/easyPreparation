"use client";

/**
 * 모바일 리모컨 (/mobile) — 기존 Go 내장 mobile-remote.html 의 React 이식판.
 * 같은 네트워크의 폰·태블릿이 QR 로 접속한다. 서버 주소는 접속한 주소(origin)를 그대로 쓴다.
 *
 * 기능: Display 실시간 미리보기 · 이전/다음 · 항목 점프 · 섹션 점프 · 서브페이지 진행바 ·
 *       자동 타이머 · 방송 시작/종료(확인 창) · 터치로 순서 바꾸기 · 좌우 스와이프 · 키보드 ·
 *       QR 표시 · Display/Stage 열기 · 첫 방문 안내 · 오프라인 표시 · PWA(서비스워커)
 */

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useDisplaySession } from "@/lib/useDisplaySession";

const C = {
  primary: "#002045",
  secondary: "#0051d5",
  surface: "#f9f9ff",
  low: "#f0f3ff",
  container: "#e7eefe",
  high: "#e2e8f8",
  onSurface: "#151c27",
  variant: "#43474e",
  outline: "#74777f",
  outlineVariant: "#c4c6cf",
  error: "#ba1a1a",
};

type Tab = "sequence" | "settings";

function Icon({ name, fill = false, size = 22, className = "", style }: { name: string; fill?: boolean; size?: number; className?: string; style?: React.CSSProperties }) {
  return (
    <span
      className={`material-symbols-outlined ${className}`}
      style={{ fontSize: size, fontVariationSettings: `'FILL' ${fill ? 1 : 0}, 'wght' 400, 'GRAD' 0, 'opsz' 24`, ...style }}
    >
      {name}
    </span>
  );
}

/** 1920×1080 Display 를 폭에 맞춰 축소해 보여 준다 */
function DisplayPreview({ offline }: { offline: boolean }) {
  const outerRef = useRef<HTMLDivElement>(null);
  const [scale, setScale] = useState(0.2);
  useEffect(() => {
    const el = outerRef.current;
    if (!el) return;
    const update = () => setScale(el.clientWidth / 1920);
    update();
    const ro = new ResizeObserver(update);
    ro.observe(el);
    return () => ro.disconnect();
  }, []);
  return (
    <div ref={outerRef} className="relative w-full overflow-hidden bg-black rounded-xl" style={{ height: Math.round(1080 * scale) }}>
      <div style={{ position: "absolute", top: 0, left: 0, width: 1920, height: 1080, transform: `scale(${scale})`, transformOrigin: "top left", pointerEvents: "none" }}>
        <iframe src="/display" title="Display 미리보기" style={{ width: 1920, height: 1080, border: "none", display: "block" }} />
      </div>
      {offline && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/60">
          <span className="text-white text-xs font-bold tracking-[0.1em]">OFFLINE</span>
        </div>
      )}
    </div>
  );
}

export default function MobileRemote() {
  const [tab, setTab] = useState<Tab>("sequence");
  const [expanded, setExpanded] = useState<Set<number>>(new Set());
  const [confirm, setConfirm] = useState<null | { start: boolean }>(null);
  const [qrOpen, setQrOpen] = useState(false);
  const [qrNonce, setQrNonce] = useState(0);
  const [hint, setHint] = useState(false);
  const [toast, setToast] = useState("");
  const [dragFrom, setDragFrom] = useState<number | null>(null);
  const [dropOver, setDropOver] = useState<number | null>(null);

  const listRef = useRef<HTMLDivElement>(null);
  const scrollRef = useRef<HTMLDivElement>(null);
  const toastTimer = useRef<ReturnType<typeof setTimeout>>();

  const showToast = useCallback((msg: string) => {
    setToast(msg);
    if (toastTimer.current) clearTimeout(toastTimer.current);
    toastTimer.current = setTimeout(() => setToast(""), 2200);
  }, []);

  const scrollToActive = useCallback(() => {
    const el = listRef.current?.querySelector<HTMLElement>("[data-active='true']");
    el?.scrollIntoView({ behavior: "smooth", block: "center" });
  }, []);

  // 진행 상태·조작 — PC 순서 패널(ProSequencePanel)과 공통 (lib/useDisplaySession)
  // 진행 중 현재 항목이 바뀌면 목록 가운데로 스크롤 (WS 로 바뀔 때만)
  const session = useDisplaySession({
    pollMs: 10000,
    onIdxChange: () => setTimeout(scrollToActive, 0),
  });
  const {
    connected: isOpen, items, idx, subIdx, subTotal, timerOn, streamStatus, streamBusy,
    schedCountdown, loadingMsg,
  } = session;
  const live = streamStatus.active;

  const sessionNavigate = session.navigate;
  const navigate = useCallback((dir: "prev" | "next") => {
    sessionNavigate(dir).catch(() => showToast("연결 오류"));
  }, [sessionNavigate, showToast]);

  const jumpTo = (i: number) => session.jump(i).catch(() => showToast("연결 오류"));
  const jumpToSection = (i: number, sp: number) => session.jump(i, sp).catch(() => showToast("연결 오류"));
  const toggleTimer = () => session.toggleTimer().catch(() => showToast("연결 오류"));

  const askStream = () => {
    if (streamBusy) { showToast("처리 중..."); return; }
    setConfirm({ start: !live });
  };

  const doStream = async (start: boolean) => {
    try {
      const r = await session.setStream(start);
      if (start) showToast(r.ok ? "방송이 시작되었습니다" : "방송 시작 실패: " + (r.error || "알 수 없는 오류"));
      else showToast("방송이 종료되었습니다");
    } catch {
      showToast("연결 오류");
    }
  };

  const reorder = (from: number, to: number) => {
    session.reorder(from, to).catch(() => showToast("순서 변경 실패"));
  };

  const toggleExpand = (i: number) =>
    setExpanded((prev) => {
      const n = new Set(prev);
      if (n.has(i)) n.delete(i); else n.add(i);
      return n;
    });

  // ── 전역 제스처·키보드 ──
  useEffect(() => {
    // 핀치 줌 방지 (iOS Safari 포함)
    const prevent = (e: Event) => e.preventDefault();
    const multiTouch = (e: TouchEvent) => { if (e.touches.length > 1) e.preventDefault(); };
    document.addEventListener("gesturestart", prevent, { passive: false });
    document.addEventListener("gesturechange", prevent, { passive: false });
    document.addEventListener("touchmove", multiTouch, { passive: false });

    // 좌우 스와이프 (목록 스크롤 영역 밖에서) → 이전/다음
    let sx = 0, sy = 0, moved = false;
    const ts = (e: TouchEvent) => { sx = e.touches[0].clientX; sy = e.touches[0].clientY; moved = false; };
    const tm = (e: TouchEvent) => {
      if (e.touches.length === 1 && (Math.abs(e.touches[0].clientX - sx) > 10 || Math.abs(e.touches[0].clientY - sy) > 10)) moved = true;
    };
    const te = (e: TouchEvent) => {
      if (!moved) return;
      const dx = e.changedTouches[0].clientX - sx;
      const dy = e.changedTouches[0].clientY - sy;
      if (Math.abs(dy) > Math.abs(dx) || Math.abs(dx) < 60) return;
      const r = scrollRef.current?.getBoundingClientRect();
      if (r && sy >= r.top && sy <= r.bottom) return;
      if (dx < 0) { navigate("next"); showToast("다음 슬라이드"); }
      else { navigate("prev"); showToast("이전 슬라이드"); }
    };
    document.addEventListener("touchstart", ts, { passive: true });
    document.addEventListener("touchmove", tm, { passive: true });
    document.addEventListener("touchend", te, { passive: true });

    const kd = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA") return;
      if (e.key === "ArrowRight" || e.key === " ") { e.preventDefault(); navigate("next"); }
      if (e.key === "ArrowLeft") { e.preventDefault(); navigate("prev"); }
      if (e.key === "Escape") { setQrOpen(false); setConfirm(null); }
    };
    document.addEventListener("keydown", kd);

    return () => {
      document.removeEventListener("gesturestart", prevent);
      document.removeEventListener("gesturechange", prevent);
      document.removeEventListener("touchmove", multiTouch);
      document.removeEventListener("touchstart", ts);
      document.removeEventListener("touchmove", tm);
      document.removeEventListener("touchend", te);
      document.removeEventListener("keydown", kd);
    };
  }, [navigate, showToast]);

  // 첫 방문 안내 + PWA 서비스워커
  useEffect(() => {
    try { if (!localStorage.getItem("ep_remote_hint_seen")) setHint(true); } catch {}
    if ("serviceWorker" in navigator) navigator.serviceWorker.register("/mobile/sw.js").catch(() => {});
  }, []);

  // ── 터치로 순서 바꾸기 (손잡이를 누른 채 끌기) ──
  const itemIdxAt = (x: number, y: number): number | null => {
    for (const el of document.elementsFromPoint(x, y)) {
      const v = (el as HTMLElement).dataset?.idx;
      if (v !== undefined) return parseInt(v, 10);
    }
    return null;
  };
  const onHandleTouchStart = (i: number) => (e: React.TouchEvent) => {
    e.stopPropagation();
    setDragFrom(i);
  };
  useEffect(() => {
    if (dragFrom === null) return;
    const move = (e: TouchEvent) => {
      e.preventDefault();
      const t = e.touches[0];
      const over = itemIdxAt(t.clientX, t.clientY);
      setDropOver(over !== null && over !== dragFrom ? over : null);
    };
    const end = (e: TouchEvent) => {
      const t = e.changedTouches[0];
      const to = itemIdxAt(t.clientX, t.clientY);
      if (to !== null && to !== dragFrom) reorder(dragFrom, to);
      setDragFrom(null);
      setDropOver(null);
    };
    document.addEventListener("touchmove", move, { passive: false });
    document.addEventListener("touchend", end);
    return () => {
      document.removeEventListener("touchmove", move);
      document.removeEventListener("touchend", end);
    };
  }, [dragFrom]); // eslint-disable-line react-hooks/exhaustive-deps

  // 현재 서브페이지가 속한 섹션 (startPage <= subIdx 인 마지막)
  const activeSectionSp = useMemo(() => {
    const secs = items[idx]?.sections;
    if (!Array.isArray(secs)) return -1;
    let sp = -1;
    secs.forEach((s, si) => {
      const p = typeof s.startPage === "number" ? s.startPage : si;
      if (p <= subIdx) sp = p;
    });
    return sp;
  }, [items, idx, subIdx]);

  const qrUrl = typeof window !== "undefined" ? `${window.location.origin}/mobile` : "";

  return (
    <div className="flex flex-col select-none" style={{ height: "100dvh", background: C.surface, color: C.onSurface, fontFamily: "Inter, 'Noto Sans KR', sans-serif" }}>
      {/* 상단 바 */}
      <header className="flex items-center justify-between px-4 h-14 flex-shrink-0" style={{ background: C.surface }}>
        <button className="flex items-center gap-2" onClick={scrollToActive} aria-label="현재 항목으로">
          <Icon name="menu" style={{ color: C.primary }} />
          <span className="text-[17px] font-black tracking-tight" style={{ color: C.primary }}>easyPreparation</span>
        </button>
        <span className="flex items-center gap-1.5 text-[11px] font-semibold" style={{ color: isOpen ? "#1a7f37" : C.error }} data-testid="m-ws">
          <span className="w-2 h-2 rounded-full" style={{ background: isOpen ? "#2ea043" : C.error }} />
          {isOpen ? "연결됨" : "연결 끊김"}
        </span>
      </header>

      {/* 위쪽 고정: 미리보기(작게) + 빠른 제어(LIVE 등) + 순서 머리 — 목록만 스크롤된다 */}
      {tab === "sequence" && (
        <div className="flex-shrink-0 px-4 pb-2" style={{ background: C.surface }} data-testid="m-top">
          <div className="flex gap-2 items-stretch">
            <div className="flex-1 min-w-0"><DisplayPreview offline={!isOpen} /></div>
            <div className="flex flex-col gap-1.5 w-[96px] flex-shrink-0">
              <QuickBtn icon="monitor" fill label="Stage" onClick={() => window.open("/display", "_blank")} />
              <QuickBtn
                icon={timerOn ? "pause_circle" : "play_circle"}
                label={timerOn ? "Timer ON" : "Timer"}
                active={timerOn}
                onClick={toggleTimer}
                testId="m-timer"
              />
              <QuickBtn
                icon={streamBusy ? "hourglass_empty" : "radio_button_checked"}
                label={streamBusy ? "처리 중" : "LIVE"}
                live={live && !streamBusy}
                onClick={askStream}
                testId="m-stream"
              />
            </div>
          </div>

          {/* 예약 송출 카운트다운 */}
          {schedCountdown && (
            <div className="mt-2 flex items-center justify-center gap-2 py-2 rounded-xl text-white" style={{ background: C.error }}>
              <span className="text-[12px] font-semibold">{schedCountdown.label}</span>
              <span className="text-[16px] font-bold font-mono tracking-widest">
                {String(schedCountdown.minutes).padStart(2, "0")}:{String(schedCountdown.seconds).padStart(2, "0")}
              </span>
            </div>
          )}

          <div className="flex items-baseline justify-between pt-3">
            <span className="text-[20px] font-black tracking-tight" style={{ color: C.primary }}>Sequence</span>
            <span className="text-[12px] font-semibold" style={{ color: C.variant }}>
              {loadingMsg || (items.length > 0 ? `${idx + 1} / ${items.length}` : "")}
            </span>
          </div>
        </div>
      )}

      {/* 스크롤 영역 (순서 목록 / 설정) */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto min-h-0 pt-1">
        {tab === "sequence" ? (
          <>
            {items.length === 0 ? (
              <div className="text-center text-[13px] py-10 leading-relaxed" style={{ color: C.variant }}>
                <Icon name="queue_music" size={40} style={{ color: C.outlineVariant }} />
                <div>예배 순서가 없습니다.<br />주보 탭에서 순서를 전송해주세요.</div>
              </div>
            ) : (
              <div ref={listRef} className="flex flex-col gap-2 px-4 pb-4" data-testid="m-list">
                {items.map((item, i) => {
                  const active = i === idx;
                  const badge = active ? "Current" : i === idx + 1 ? "Next" : "";
                  const secs = Array.isArray(item.sections) ? item.sections : [];
                  const pages = Array.isArray(item.pages) ? item.pages : [];
                  const hasDetail = secs.length > 0 || pages.length > 0 || (!!item.obj && item.obj !== "-");
                  const open = expanded.has(i);
                  const pct = active && subTotal > 1 ? Math.round((subIdx / (subTotal - 1)) * 100) : 0;
                  return (
                    <div key={`${item.key || item.title}-${i}`}>
                      <div
                        data-idx={i}
                        data-active={active}
                        className="relative flex items-center justify-between rounded-xl transition-transform active:scale-[0.99]"
                        style={{
                          padding: 16,
                          background: dropOver === i ? "rgba(0,81,213,0.08)" : active ? "#fff" : C.low,
                          borderLeft: active ? `4px solid ${C.secondary}` : "4px solid transparent",
                          boxShadow: active ? "0 2px 12px rgba(0,32,69,0.1)" : undefined,
                          opacity: dragFrom === i ? 0.5 : 1,
                        }}
                      >
                        <button className="flex items-center gap-3 min-w-0 flex-1 text-left" onClick={() => jumpTo(i)}>
                          <span
                            className="w-9 h-9 rounded-full flex items-center justify-center text-[12px] font-bold flex-shrink-0"
                            style={{ background: active ? "rgba(0,81,213,0.1)" : C.high, color: active ? C.secondary : C.variant }}
                          >
                            {String(i + 1).padStart(2, "0")}
                          </span>
                          <span className="flex flex-col gap-0.5 min-w-0">
                            <span className="text-[9px] font-bold tracking-[0.1em] uppercase" style={{ color: active ? C.secondary : C.variant }}>{badge || " "}</span>
                            <span className="text-[15px] font-bold truncate" style={{ color: C.primary }}>{item.title}</span>
                            {item.obj && item.obj !== "-" && <span className="text-[12px] truncate" style={{ color: C.variant }}>{item.obj}</span>}
                          </span>
                        </button>
                        {hasDetail && (
                          <button onClick={() => toggleExpand(i)} aria-label="자세히" className="p-1">
                            <Icon name="expand_more" style={{ color: C.outline, transform: open ? "rotate(180deg)" : undefined, transition: "transform .15s" }} />
                          </button>
                        )}
                        <span className="p-1 touch-none" onTouchStart={onHandleTouchStart(i)} aria-label="끌어서 순서 바꾸기">
                          <Icon name={active ? "equalizer" : "drag_indicator"} fill={active} style={{ color: active ? C.secondary : C.outline }} />
                        </span>
                        {active && subTotal > 1 && (
                          <div className="absolute left-0 right-0 bottom-0 h-[3px] rounded-b-xl overflow-hidden" style={{ background: C.container }}>
                            <div className="h-full transition-[width] duration-200" style={{ width: `${pct}%`, background: C.secondary }} />
                          </div>
                        )}
                      </div>

                      {open && hasDetail && (
                        <div className="mx-2 mt-1 px-3 py-2 rounded-lg bg-white">
                          {secs.length > 0 ? (
                            secs.map((sec, si) => {
                              const sp = typeof sec.startPage === "number" ? sec.startPage : si;
                              const isCur = active && sp === activeSectionSp;
                              const done = active && sp >= 0 && sp < activeSectionSp;
                              return (
                                <button
                                  key={si}
                                  onClick={() => jumpToSection(i, sp)}
                                  className="flex items-start gap-2.5 w-full text-left px-2 py-2 border-b last:border-b-0"
                                  style={{ borderColor: C.outlineVariant, background: isCur ? "rgba(0,81,213,0.08)" : undefined, opacity: done ? 0.5 : 1 }}
                                >
                                  <span className="text-[11px] font-bold min-w-[18px]" style={{ color: isCur ? C.secondary : C.variant }}>{sec.label || si + 1}</span>
                                  <span className="text-[13px] leading-snug" style={{ color: C.onSurface }}>{sec.text}</span>
                                </button>
                              );
                            })
                          ) : pages.length > 0 ? (
                            <div className="text-[12px]" style={{ color: C.variant }}>{pages.length}페이지</div>
                          ) : (
                            <span className="text-[12px]" style={{ color: C.variant }}>{item.obj}</span>
                          )}
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </>
        ) : (
          <div className="p-4">
            <div className="text-[22px] font-black tracking-tight mb-4" style={{ color: C.primary }}>Settings</div>
            <div className="flex flex-col gap-2.5">
              <SettingLink icon="monitor" title="Display 열기" desc="프로젝터용 슬라이드 화면" href="/display" />
              <SettingLink icon="desktop_windows" title="Stage Monitor 열기" desc="무대 모니터용 화면 (예배팀·설교자)" href="/display/stage" />
            </div>
          </div>
        )}
      </div>

      {/* 이전 / 다음 */}
      {tab === "sequence" && (
        <div className="flex gap-2 px-4 py-2 flex-shrink-0" style={{ background: C.surface }}>
          <button
            onClick={() => navigate("prev")}
            className="flex-1 h-[52px] rounded-xl flex items-center justify-center gap-1 text-[13px] font-bold active:opacity-80"
            style={{ background: C.high, color: C.primary }}
          >
            <Icon name="arrow_back" size={18} /> 이전
          </button>
          <button
            onClick={() => navigate("next")}
            className="flex-[2] h-[52px] rounded-xl flex items-center justify-center gap-1 text-[13px] font-bold text-white active:opacity-90"
            style={{ background: C.secondary, boxShadow: "0 2px 8px rgba(0,81,213,0.3)" }}
            data-testid="m-next"
          >
            다음 <Icon name="arrow_forward" size={18} />
          </button>
        </div>
      )}

      {/* 하단 탭 */}
      <nav className="flex-shrink-0 border-t" style={{ borderColor: C.outlineVariant, background: "#fff", paddingBottom: "env(safe-area-inset-bottom)" }}>
        <div className="grid grid-cols-4 h-16">
          <TabBtn icon="queue_music" label="Sequence" active={tab === "sequence"} onClick={() => setTab("sequence")} />
          <TabBtn icon="radio" label="Live" active={false} onClick={() => { setTab("sequence"); askStream(); }} />
          <TabBtn icon="settings_remote" label="Controls" active={false} onClick={() => { setQrNonce(Date.now()); setQrOpen(true); }} />
          <TabBtn icon="settings" label="Settings" active={tab === "settings"} onClick={() => setTab("settings")} />
        </div>
      </nav>

      {/* QR */}
      {qrOpen && (
        <Modal onClose={() => setQrOpen(false)}>
          <div className="text-[17px] font-extrabold" style={{ color: C.primary }}>모바일 접속 주소</div>
          <div className="text-[12px] leading-relaxed" style={{ color: C.variant }}>같은 WiFi에 연결된 기기에서<br />스캔하면 리모컨을 사용할 수 있습니다</div>
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src={`/mobile/qr.png?t=${qrNonce}`} alt="QR" className="w-48 h-48 mx-auto" />
          <div className="text-[12px] font-mono break-all" style={{ color: C.secondary }}>{qrUrl}</div>
          <button onClick={() => setQrOpen(false)} className="py-3 rounded-xl text-[14px] font-bold" style={{ background: C.high, color: C.primary }}>닫기</button>
        </Modal>
      )}

      {/* 방송 확인 */}
      {confirm && (
        <Modal onClose={() => setConfirm(null)}>
          <div className="text-[17px] font-extrabold" style={{ color: C.primary }}>{confirm.start ? "방송 시작" : "방송 종료"}</div>
          <div className="text-[13px]" style={{ color: C.variant }}>OBS 스트리밍을 {confirm.start ? "시작" : "종료"}하시겠습니까?</div>
          <div className="grid grid-cols-2 gap-2">
            <button onClick={() => setConfirm(null)} className="py-3 rounded-xl text-[14px] font-bold" style={{ background: C.high, color: C.primary }}>취소</button>
            <button
              onClick={() => { const s = confirm.start; setConfirm(null); doStream(s); }}
              className="py-3 rounded-xl text-[14px] font-bold text-white"
              style={{ background: confirm.start ? C.secondary : C.error }}
              data-testid="m-confirm-ok"
            >
              {confirm.start ? "시작" : "종료"}
            </button>
          </div>
        </Modal>
      )}

      {/* 첫 방문 안내 */}
      {hint && (
        <Modal onClose={() => {}}>
          <Icon name="swipe_left" size={40} style={{ color: C.secondary }} />
          <div className="text-[17px] font-extrabold" style={{ color: C.primary }}>슬라이드 제어</div>
          <div className="text-[13px] leading-relaxed" style={{ color: C.variant }}>좌우로 스와이프하거나<br />이전 / 다음 버튼으로<br />슬라이드를 이동하세요.</div>
          <button
            onClick={() => { setHint(false); try { localStorage.setItem("ep_remote_hint_seen", "1"); } catch {} }}
            className="py-3 rounded-xl text-[14px] font-bold text-white"
            style={{ background: C.secondary }}
          >
            시작하기
          </button>
        </Modal>
      )}

      {/* 토스트 */}
      <div
        className="fixed left-1/2 px-5 py-2.5 rounded-full text-[13px] font-medium text-white whitespace-nowrap pointer-events-none transition-all duration-200 z-[200]"
        style={{ bottom: 150, background: "rgba(21,28,39,0.92)", opacity: toast ? 1 : 0, transform: `translateX(-50%) translateY(${toast ? 0 : 16}px)` }}
      >
        {toast}
      </div>
    </div>
  );
}

function QuickBtn({ icon, label, onClick, active, live, fill, testId }: { icon: string; label: string; onClick: () => void; active?: boolean; live?: boolean; fill?: boolean; testId?: string }) {
  const bg = live ? C.error : active ? C.secondary : C.low;
  const fg = live || active ? "#fff" : C.primary;
  return (
    <button
      onClick={onClick}
      data-testid={testId}
      className="flex-1 flex items-center justify-center gap-1.5 px-2 rounded-xl transition-transform active:scale-95"
      style={{ background: bg, color: fg }}
    >
      <Icon name={icon} size={18} fill={fill || live || active} />
      <span className="text-[10px] font-bold tracking-[0.06em] uppercase whitespace-nowrap">{label}</span>
    </button>
  );
}

function TabBtn({ icon, label, active, onClick }: { icon: string; label: string; active: boolean; onClick: () => void }) {
  return (
    <button onClick={onClick} className="flex flex-col items-center justify-center gap-0.5" style={{ color: active ? C.secondary : C.outline }}>
      <Icon name={icon} fill={active} />
      <span className="text-[10px] font-semibold">{label}</span>
    </button>
  );
}

function SettingLink({ icon, title, desc, href }: { icon: string; title: string; desc: string; href: string }) {
  return (
    <button
      onClick={() => window.open(href, "_blank")}
      className="flex items-center gap-3 p-4 rounded-xl border text-left"
      style={{ background: C.low, borderColor: C.outlineVariant, color: C.primary }}
    >
      <Icon name={icon} fill style={{ color: C.secondary }} />
      <span>
        <span className="block text-[14px] font-bold">{title}</span>
        <span className="block text-[11px] mt-0.5" style={{ color: C.variant }}>{desc}</span>
      </span>
    </button>
  );
}

function Modal({ children, onClose }: { children: React.ReactNode; onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-[150] flex items-center justify-center bg-black/40 p-6" onClick={onClose}>
      <div className="w-full max-w-[300px] bg-white rounded-2xl p-6 flex flex-col gap-3 text-center shadow-2xl" onClick={(e) => e.stopPropagation()}>
        {children}
      </div>
    </div>
  );
}
