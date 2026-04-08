"use client";

import { usePathname } from "next/navigation";

const PAGE_TITLES: Record<string, string> = {
  "/bulletin": "주보",
  "/lyrics": "찬양",
  "/bible": "성경",
};

export default function TopHeader() {
  const pathname = usePathname();
  const title = Object.entries(PAGE_TITLES).find(([path]) =>
    pathname?.startsWith(path)
  )?.[1] || "easyPreparation";

  return (
    <header className="flex justify-between items-center w-full px-8 h-16 bg-white/70 backdrop-blur-xl border-b border-slate-100 z-40 sticky top-0">
      <div className="flex items-center gap-3">
        <h1 className="text-lg font-black text-navy-dark tracking-tight">{title}</h1>
      </div>
    </header>
  );
}
