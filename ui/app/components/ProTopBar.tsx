"use client";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useState, useEffect } from "react";
import { useRecoilState } from "recoil";
import { inspectorOpenState } from "@/recoilState";
import { useAuth } from "@/lib/LocalAuthContext";

const TABS = [
  { href: "/bulletin", label: "Bulletin", short: "주보", shortcut: "F1" },
  { href: "/lyrics", label: "Hymns", short: "찬양", shortcut: "F2" },
  { href: "/bible", label: "Scripture", short: "성경", shortcut: "F3" },
];

export default function ProTopBar() {
  const pathname = usePathname();
  const router = useRouter();
  const [time, setTime] = useState("--:--:--");
  const [inspOpen, setInspOpen] = useRecoilState(inspectorOpenState);
  const { church } = useAuth();

  // F1/F2/F3 단축키
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      // 입력 필드 포커스 중엔 무시
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
      if (e.key === "F1") { e.preventDefault(); router.push("/bulletin"); }
      if (e.key === "F2") { e.preventDefault(); router.push("/lyrics"); }
      if (e.key === "F3") { e.preventDefault(); router.push("/bible"); }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [router]);

  useEffect(() => {
    const update = () => {
      const now = new Date();
      setTime(
        now.toLocaleTimeString("ko-KR", {
          hour: "2-digit",
          minute: "2-digit",
          second: "2-digit",
          hour12: false,
        })
      );
    };
    update();
    const id = setInterval(update, 1000);
    return () => clearInterval(id);
  }, []);

  return (
    <div
      className="flex items-center bg-pro-surface border-b border-pro-border px-3 gap-3 select-none"
      data-testid="topbar"
      style={{ gridColumn: "1 / -1", gridRow: "1" }}
    >
      {/* 로고 + 브랜드 */}
      <div className="flex items-center gap-2 flex-shrink-0">
        <img src="/images/ep-logo.svg" alt="EP" width={20} height={20} className="opacity-80" />
        <span className="text-pro-text text-xs font-bold tracking-wide hidden md:block">
          easyPreparation
        </span>
        <span className="text-pro-text-dim text-[10px] opacity-60 hidden lg:block">
          · Pro Console
        </span>
      </div>

      {/* 구분선 */}
      <div className="w-px h-5 bg-pro-border flex-shrink-0 hidden sm:block" />

      {/* 상태 표시 */}
      <div className="hidden sm:flex items-center gap-1.5 flex-shrink-0">
        <div className="w-1.5 h-1.5 rounded-full bg-pro-draft" />
        <span className="text-pro-text-dim text-[10px] hidden sm:block">OFF AIR</span>
      </div>

      {/* 시계 */}
      <div className="flex-shrink-0 font-mono text-pro-text text-xs bg-pro-elevated px-2 py-0.5 rounded border border-pro-border tabular-nums">
        {time}
      </div>

      {/* 탭 네비게이션 */}
      <nav className="flex items-center flex-1 overflow-x-auto self-stretch [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
        {TABS.map(({ href, label, short, shortcut }) => {
          const isActive = pathname?.startsWith(href);
          return (
            <Link
              key={href}
              href={href}
              className={`flex items-center gap-1 px-2 sm:px-3 text-xs font-semibold whitespace-nowrap transition-colors border-b-2 h-full ${
                isActive
                  ? "bg-pro-tab-active text-pro-accent border-pro-tab-border"
                  : "text-pro-text-muted hover:text-pro-text hover:bg-pro-hover border-transparent"
              }`}
            >
              <span className="sm:hidden">{short}</span>
              <span className="hidden sm:inline">{label}</span>
              <span className="text-[9px] opacity-30 hidden sm:inline">{shortcut}</span>
            </Link>
          );
        })}
      </nav>

      {/* 우측: 교회명 + Output 버튼 */}
      <div className="flex items-center gap-2 ml-auto flex-shrink-0">
        {church?.name && (
          <span className="text-pro-text-dim text-[10px] opacity-50 hidden xl:block truncate max-w-[6rem]">
            {church.name}
          </span>
        )}
        <button
          className={`flex items-center gap-1.5 px-3 py-2 min-h-[36px] rounded text-xs font-bold transition-all ${
            inspOpen
              ? "bg-pro-accent text-white shadow-lg"
              : "bg-pro-elevated text-pro-text-muted border border-pro-border hover:bg-pro-hover hover:text-pro-text"
          }`}
          onClick={() => setInspOpen((v) => !v)}
          aria-label={inspOpen ? "인스펙터 패널 닫기" : "인스펙터 패널 열기"}
          aria-expanded={inspOpen}
        >
          Broadcast
        </button>
      </div>
    </div>
  );
}
