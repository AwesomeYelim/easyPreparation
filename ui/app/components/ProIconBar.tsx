"use client";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";
import Sidebar from "./SideBar";

const NAV_ITEMS = [
  { href: "/bulletin", icon: "newspaper", label: "주보" },
  { href: "/lyrics", icon: "format_quote", label: "찬양" },
  { href: "/bible", icon: "auto_stories", label: "성경" },
];

export default function ProIconBar() {
  const pathname = usePathname();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <>
      <aside
        className="flex flex-col items-center py-3 gap-1 bg-pro-surface border-r border-pro-border"
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
