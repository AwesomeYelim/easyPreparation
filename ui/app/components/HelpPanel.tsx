"use client";

interface HelpPanelProps {
  open: boolean;
  onClose: () => void;
}

const SHORTCUTS = [
  { keys: ["Space", "→"], action: "다음 슬라이드" },
  { keys: ["←"], action: "이전 슬라이드" },
  { keys: ["F1"], action: "Bulletin 탭" },
  { keys: ["F2"], action: "Hymns 탭" },
  { keys: ["F3"], action: "Scripture 탭" },
];

const QUICK_START = [
  {
    icon: "edit_document",
    title: "예배 정보 입력",
    desc: "Bulletin 탭에서 예배 정보 입력 후 저장",
  },
  {
    icon: "cast",
    title: "프로젝터 전송",
    desc: "예배 순서 패널 → 프로젝터에 보내기",
  },
  {
    icon: "keyboard",
    title: "슬라이드 전환",
    desc: "Space 또는 → 키로 슬라이드 전환",
  },
];

export default function HelpPanel({ open, onClose }: HelpPanelProps) {
  return (
    <>
      {/* Backdrop */}
      <div
        className={`fixed inset-0 z-40 bg-black transition-opacity duration-200 ${
          open ? "opacity-40 pointer-events-auto" : "opacity-0 pointer-events-none"
        }`}
        onClick={onClose}
      />

      {/* Slide-in panel */}
      <div
        className={`fixed top-0 right-0 h-full w-80 z-50 bg-pro-surface border-l border-pro-border flex flex-col shadow-2xl transition-transform duration-200 ${
          open ? "translate-x-0" : "translate-x-full"
        }`}
      >
        {/* Header */}
        <div className="flex items-center gap-2 px-4 py-3 border-b border-pro-border flex-shrink-0">
          <span
            className="material-symbols-outlined text-pro-accent"
            style={{ fontSize: "18px" }}
          >
            help
          </span>
          <span className="text-pro-text font-bold text-sm flex-1">도움말</span>
          <button
            onClick={onClose}
            className="w-7 h-7 flex items-center justify-center rounded text-pro-text-muted hover:text-pro-text hover:bg-pro-hover transition-all"
            title="닫기"
          >
            <span className="material-symbols-outlined" style={{ fontSize: "16px" }}>
              close
            </span>
          </button>
        </div>

        {/* Scrollable content */}
        <div className="flex-1 overflow-y-auto p-4 space-y-5">
          {/* Section 1: Keyboard shortcuts */}
          <section>
            <h3 className="text-pro-text-muted text-[10px] font-semibold uppercase tracking-widest mb-2">
              키보드 단축키
            </h3>
            <div className="bg-pro-elevated rounded p-2 text-xs space-y-1.5">
              {SHORTCUTS.map(({ keys, action }) => (
                <div key={action} className="flex items-center justify-between">
                  <span className="text-pro-text-muted">{action}</span>
                  <div className="flex items-center gap-1">
                    {keys.map((k, i) => (
                      <span key={k} className="flex items-center gap-1">
                        {i > 0 && (
                          <span className="text-pro-text-dim text-[9px]">/</span>
                        )}
                        <kbd className="bg-pro-bg border border-pro-border rounded px-1 py-0.5 font-mono text-[10px] text-pro-text">
                          {k}
                        </kbd>
                      </span>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </section>

          {/* Section 2: Quick start */}
          <section>
            <h3 className="text-pro-text-muted text-[10px] font-semibold uppercase tracking-widest mb-2">
              빠른 시작 가이드
            </h3>
            <div className="space-y-2">
              {QUICK_START.map(({ icon, title, desc }, idx) => (
                <div
                  key={icon}
                  className="flex items-start gap-3 bg-pro-elevated rounded p-3"
                >
                  <div className="flex items-center justify-center w-7 h-7 rounded bg-pro-hover flex-shrink-0 mt-0.5">
                    <span
                      className="material-symbols-outlined text-pro-accent"
                      style={{ fontSize: "16px" }}
                    >
                      {icon}
                    </span>
                  </div>
                  <div className="min-w-0">
                    <div className="text-pro-text text-xs font-semibold mb-0.5">
                      <span className="text-pro-text-dim mr-1">{idx + 1}.</span>
                      {title}
                    </div>
                    <div className="text-pro-text-muted text-[11px] leading-snug">
                      {desc}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </section>

          {/* Section 3: Restart tour */}
          <section>
            <h3 className="text-pro-text-muted text-[10px] font-semibold uppercase tracking-widest mb-2">
              온보딩
            </h3>
            <button
              onClick={() => {
                localStorage.removeItem("ep_tour_done");
                window.location.reload();
              }}
              className="w-full flex items-center gap-2 px-3 py-2 rounded bg-pro-elevated hover:bg-pro-hover text-pro-text-muted hover:text-pro-text text-xs transition-all border border-pro-border"
            >
              <span
                className="material-symbols-outlined text-pro-accent"
                style={{ fontSize: "15px" }}
              >
                replay
              </span>
              온보딩 투어 다시 시작
            </button>
          </section>
        </div>

        {/* Footer */}
        <div className="px-4 py-3 border-t border-pro-border flex-shrink-0">
          <p className="text-pro-text-dim text-[10px] text-center">
            easyPreparation Pro Console
          </p>
        </div>
      </div>
    </>
  );
}
