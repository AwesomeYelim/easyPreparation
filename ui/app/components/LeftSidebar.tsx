"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState, useEffect } from "react";
import { useRecoilState } from "recoil";
import { displayPanelOpenState, sidebarCollapsedState } from "@/recoilState";
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
  const [collapsed, setCollapsed] = useRecoilState(sidebarCollapsedState);
  const { church } = useAuth();
  const [mounted, setMounted] = useState(false);
  useEffect(() => { setMounted(true); }, []);

  return (
    <>
      <aside
        className={`fixed left-0 top-0 h-full flex flex-col py-8 bg-white border-r border-slate-200 z-50 ${mounted ? "transition-[width,padding] duration-200 ease-in-out" : ""} ${
          collapsed ? "w-16 px-2" : "w-64 px-5"
        }`}
      >
        {/* 토글 버튼 */}
        <button
          className="absolute -right-4 top-10 w-8 h-8 flex items-center justify-center bg-white border-2 border-slate-300 rounded-full shadow-md text-on-surface-variant hover:text-navy-dark transition-colors z-10"
          onClick={() => setCollapsed(!collapsed)}
          title={collapsed ? "사이드바 열기" : "사이드바 접기"}
        >
          <span className="material-symbols-outlined" style={{ fontSize: "16px" }}>
            {collapsed ? "chevron_right" : "chevron_left"}
          </span>
        </button>

        {/* Logo */}
        <div
          className={`mb-10 flex flex-col gap-2 ${
            collapsed ? "items-center justify-center px-0" : "px-2"
          }`}
        >
          <img
            src="/images/ep-logo.svg"
            alt="EP"
            width={40}
            height={40}
          />
          {!collapsed && (
            <p className="text-[10px] font-black text-on-surface opacity-40 uppercase tracking-[0.25em]">
              easyPreparation
            </p>
          )}
        </div>

        {/* Navigation */}
        <nav className="flex-1 space-y-1.5">
          {NAV_ITEMS.map(({ href, icon, label }) => {
            const isActive = pathname?.startsWith(href);
            return (
              <Link
                key={href}
                href={href}
                title={collapsed ? label : undefined}
                className={`flex items-center gap-4 py-3 rounded-2xl transition-all ${
                  collapsed ? "justify-center px-2" : "px-4"
                } ${
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
                {!collapsed && (
                  <span className="text-sm tracking-tight font-semibold">{label}</span>
                )}
              </Link>
            );
          })}
        </nav>

        {/* Bottom Actions */}
        <div className="mt-auto pt-6 space-y-2 border-t border-slate-100">
          {/* Display Control */}
          <div className={`flex mb-6 ${collapsed ? "flex-col gap-2" : "gap-2"}`}>
            <button
              className={`flex-1 py-3 rounded-2xl font-bold flex items-center justify-center gap-2 transition-all text-sm ${
                collapsed ? "px-2" : "px-4"
              } ${
                panelOpen
                  ? "bg-electric-blue text-white shadow-lg shadow-electric-blue/30"
                  : "bg-slate-100 text-on-surface-variant hover:bg-slate-200"
              }`}
              onClick={() => setPanelOpen(!panelOpen)}
              title={collapsed ? "제어판" : undefined}
            >
              <span className="material-symbols-outlined text-lg">slideshow</span>
              {!collapsed && <span className="tracking-wide">제어판</span>}
            </button>
            <button
              className="p-3 rounded-2xl bg-slate-100 text-on-surface-variant hover:bg-slate-200 transition-all"
              onClick={() => openDisplayWindow(true)}
              title="새 창으로 열기"
            >
              <span className="material-symbols-outlined text-lg">open_in_new</span>
            </button>
          </div>

          <button
            className={`flex items-center gap-4 py-2.5 rounded-xl text-on-surface-variant hover:text-navy-dark transition-colors w-full ${
              collapsed ? "justify-center px-2" : "px-4"
            }`}
            onClick={() => setSidebarOpen(true)}
            title={collapsed ? (church?.name || "설정") : undefined}
          >
            <span className="material-symbols-outlined">settings</span>
            {!collapsed && (
              <span className="text-sm font-semibold">
                {church?.name || "설정"}
              </span>
            )}
          </button>
        </div>
      </aside>

      <Sidebar open={sidebarOpen} onClose={() => setSidebarOpen(false)} />
    </>
  );
}
