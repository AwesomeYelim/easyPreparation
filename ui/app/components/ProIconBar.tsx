"use client";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState, useEffect, useRef } from "react";
import Sidebar from "./SideBar";

const NAV_ITEMS = [
  { href: "/bulletin", icon: "newspaper", label: "주보" },
  { href: "/lyrics", icon: "format_quote", label: "찬양" },
  { href: "/bible", icon: "auto_stories", label: "성경" },
];

function QRPopover({ onClose }: { onClose: () => void }) {
  const ref = useRef<HTMLDivElement>(null);
  const BASE = process.env.NEXT_PUBLIC_API_BASE_URL ?? "";
  const mobileUrl = BASE.replace(/:\d+$/, ":8080") + "/mobile";

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onClose();
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [onClose]);

  return (
    <div
      ref={ref}
      className="absolute left-12 bottom-12 z-50 bg-pro-elevated border border-pro-border rounded-xl shadow-2xl p-3 flex flex-col gap-2"
      style={{ width: 180 }}
    >
      <div className="text-[11px] font-semibold text-pro-text text-center">모바일 리모컨</div>
      <img src="/mobile/qr.png" alt="QR" className="w-full rounded-lg border border-pro-border" />
      <div className="text-[9px] text-pro-text-dim text-center break-all">{mobileUrl}</div>
      <button
        onClick={async () => {
          try {
            await fetch("/api/open-mobile");
          } catch {
            window.open(mobileUrl, "_blank", "noopener");
          }
        }}
        className="text-[10px] text-center text-pro-accent hover:underline cursor-pointer bg-transparent border-none"
      >
        바로 열기 →
      </button>
    </div>
  );
}

export default function ProIconBar() {
  const pathname = usePathname();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [qrOpen, setQrOpen] = useState(false);

  return (
    <>
      <aside
        className="flex flex-col items-center py-3 gap-1 bg-pro-surface border-r border-pro-border relative"
        style={{ gridColumn: "1", gridRow: "2" }}
      >
        {/* 네비 아이콘 */}
        <nav className="flex flex-col gap-1 flex-1 w-full items-center pt-1">
          {NAV_ITEMS.map(({ href, icon, label }) => {
            const isActive = pathname?.startsWith(href);
            return (
              <Link
                key={href}
                href={href}
                title={label}
                className={`w-9 h-9 flex items-center justify-center rounded-lg transition-all ${
                  isActive
                    ? "bg-pro-accent/20 text-pro-accent"
                    : "text-pro-text-muted hover:text-pro-text hover:bg-pro-hover"
                }`}
              >
                <span
                  className="material-symbols-outlined"
                  style={{
                    fontSize: "18px",
                    fontVariationSettings: isActive
                      ? "'FILL' 1, 'wght' 400, 'GRAD' 0, 'opsz' 24"
                      : "'FILL' 0, 'wght' 400, 'GRAD' 0, 'opsz' 24",
                  }}
                >
                  {icon}
                </span>
              </Link>
            );
          })}
        </nav>

        {/* QR 버튼 (설정 위) */}
        <div className="relative">
          <button
            className={`w-9 h-9 flex items-center justify-center rounded-lg transition-all ${
              qrOpen ? "bg-pro-accent/20 text-pro-accent" : "text-pro-text-muted hover:text-pro-text hover:bg-pro-hover"
            }`}
            onClick={() => setQrOpen((v) => !v)}
            title="모바일 리모컨 QR"
          >
            <span
              className="material-symbols-outlined"
              style={{ fontSize: "18px", fontVariationSettings: "'FILL' 0, 'wght' 400, 'GRAD' 0, 'opsz' 24" }}
            >
              qr_code
            </span>
          </button>
          {qrOpen && <QRPopover onClose={() => setQrOpen(false)} />}
        </div>

        {/* 설정 버튼 */}
        <button
          className="w-9 h-9 flex items-center justify-center rounded-lg text-pro-text-muted hover:text-pro-text hover:bg-pro-hover transition-all"
          onClick={() => setSidebarOpen(true)}
          title="설정"
        >
          <span
            className="material-symbols-outlined"
            style={{ fontSize: "18px", fontVariationSettings: "'FILL' 0, 'wght' 400, 'GRAD' 0, 'opsz' 24" }}
          >
            settings
          </span>
        </button>
      </aside>

      <Sidebar open={sidebarOpen} onClose={() => setSidebarOpen(false)} />
    </>
  );
}
