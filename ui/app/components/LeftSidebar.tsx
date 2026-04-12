"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";
import { useRecoilState } from "recoil";
import { displayPanelOpenState } from "@/recoilState";
import { useAuth } from "@/lib/LocalAuthContext";
import { openDisplayWindow } from "@/lib/apiClient";
import Sidebar from "./SideBar";

const NAV_ITEMS = [
  { href: "/bulletin", icon: "newspaper", label: "주보" },
  { href: "/lyrics", icon: "format_quote", label: "찬양" },
  { href: "/bible", icon: "auto_stories", label: "성경" },
];

export default function LeftSidebar() {
  const pathname = usePathname();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [panelOpen, setPanelOpen] = useRecoilState(displayPanelOpenState);
  const { church } = useAuth();

  return (
    <>
      <aside className="fixed left-0 top-0 h-full flex flex-col py-8 px-5 w-64 bg-white border-r border-slate-200 z-50">
        {/* Logo */}
        <div className="mb-10 px-2 flex flex-col gap-2">
          <img
            src="/images/ep-logo.svg"
            alt="EP"
            width={40}
            height={40}
          />
          <p className="text-[10px] font-black text-on-surface opacity-40 uppercase tracking-[0.25em]">
            easyPreparation
          </p>
        </div>

        {/* Navigation */}
        <nav className="flex-1 space-y-1.5">
          {NAV_ITEMS.map(({ href, icon, label }) => {
            const isActive = pathname?.startsWith(href);
            return (
              <Link
                key={href}
                href={href}
                className={`flex items-center gap-4 py-3 px-4 rounded-2xl transition-all ${
                  isActive
                    ? "bg-electric-blue/10 text-electric-blue font-bold shadow-sm shadow-electric-blue/5"
                    : "text-on-surface-variant hover:text-navy-dark hover:bg-slate-100"
                }`}
              >
                <span
                  className="material-symbols-outlined"
                  style={isActive ? { fontVariationSettings: "'FILL' 1" } : undefined}
                >
                  {icon}
                </span>
                <span className="text-sm tracking-tight font-semibold">{label}</span>
              </Link>
            );
          })}
        </nav>

        {/* Bottom Actions */}
        <div className="mt-auto pt-6 space-y-2 border-t border-slate-100">
          {/* Display Control */}
          <div className="flex gap-2 mb-6">
            <button
              className={`flex-1 py-3 px-4 rounded-2xl font-bold flex items-center justify-center gap-2 transition-all text-sm ${
                panelOpen
                  ? "bg-electric-blue text-white shadow-lg shadow-electric-blue/30"
                  : "bg-slate-100 text-on-surface-variant hover:bg-slate-200"
              }`}
              onClick={() => setPanelOpen(!panelOpen)}
            >
              <span className="material-symbols-outlined text-lg">slideshow</span>
              <span className="tracking-wide">예배 화면</span>
            </button>
            <button
              className="p-3 rounded-2xl bg-slate-100 text-on-surface-variant hover:bg-slate-200 transition-all"
              onClick={() => openDisplayWindow()}
              title="새 창으로 열기"
            >
              <span className="material-symbols-outlined text-lg">open_in_new</span>
            </button>
          </div>

          <button
            className="flex items-center gap-4 py-2.5 px-4 rounded-xl text-on-surface-variant hover:text-navy-dark transition-colors w-full"
            onClick={() => setSidebarOpen(true)}
          >
            <span className="material-symbols-outlined">settings</span>
            <span className="text-sm font-semibold">
              {church?.name || "설정"}
            </span>
          </button>
        </div>
      </aside>

      <Sidebar open={sidebarOpen} onClose={() => setSidebarOpen(false)} />
    </>
  );
}
