"use client";

import { useEffect, useRef, useState, type RefObject } from "react";

export function prefersReducedMotion() {
  return (
    typeof window !== "undefined" &&
    window.matchMedia("(prefers-reduced-motion: reduce)").matches
  );
}

/**
 * One-second clock that only runs while `ref` is on screen and the tab is
 * visible. Starts at 0 on both server and client so hydration matches.
 */
export function useTicker(ref: RefObject<Element>, enabled = true) {
  const [t, setT] = useState(0);
  useEffect(() => {
    const el = ref.current;
    if (!el || !enabled || prefersReducedMotion()) return;
    let onScreen = false;
    let id: ReturnType<typeof setInterval> | undefined;
    const sync = () => {
      const run = onScreen && !document.hidden;
      if (run && !id) id = setInterval(() => setT((v) => v + 1), 1000);
      if (!run && id) {
        clearInterval(id);
        id = undefined;
      }
    };
    const io = new IntersectionObserver(([e]) => {
      onScreen = e.isIntersecting;
      sync();
    });
    io.observe(el);
    document.addEventListener("visibilitychange", sync);
    return () => {
      io.disconnect();
      document.removeEventListener("visibilitychange", sync);
      if (id) clearInterval(id);
    };
  }, [ref, enabled]);
  return t;
}

/** Scale factor that fits a fixed `width` design into its container. */
export function useFitScale(width: number) {
  const ref = useRef<HTMLDivElement>(null);
  const [scale, setScale] = useState<number | null>(null);
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const ro = new ResizeObserver(() => setScale(el.clientWidth / width));
    ro.observe(el);
    return () => ro.disconnect();
  }, [width]);
  return { ref, scale };
}
