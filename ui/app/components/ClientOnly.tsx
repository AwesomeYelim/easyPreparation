"use client";
import { useState, useEffect, ReactNode } from "react";

/**
 * 서버에서는 null을 렌더링하고 클라이언트에서만 children을 렌더링합니다.
 * React 18 하이드레이션 불일치(hydration mismatch) 방지용.
 */
export default function ClientOnly({ children }: { children: ReactNode }) {
  const [mounted, setMounted] = useState(false);
  useEffect(() => setMounted(true), []);
  if (!mounted) return null;
  return <>{children}</>;
}
