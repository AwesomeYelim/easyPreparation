"use client";

import { useEffect, useRef, useState, type ReactNode } from "react";
import { prefersReducedMotion } from "./live/hooks";

/** Calls `fn` once when `el` first scrolls into view. */
function onceVisible(el: Element, fn: () => void) {
  const io = new IntersectionObserver(
    ([e]) => {
      if (e.isIntersecting) {
        fn();
        io.disconnect();
      }
    },
    { threshold: 0.15 },
  );
  io.observe(el);
  return () => io.disconnect();
}

/**
 * Fades children up as they scroll in. Server HTML stays fully visible; the
 * hidden state is applied only after mount, and only for content below the fold.
 */
export function Reveal({
  children,
  delay = 0,
  className = "",
}: {
  children: ReactNode;
  delay?: number;
  className?: string;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const [state, setState] = useState<"idle" | "hidden" | "shown">("idle");
  useEffect(() => {
    const el = ref.current;
    if (!el || prefersReducedMotion()) return;
    if (el.getBoundingClientRect().top < window.innerHeight) return;
    setState("hidden");
    return onceVisible(el, () => setState("shown"));
  }, []);
  return (
    <div
      ref={ref}
      className={`${state === "idle" ? "" : "reveal"} ${state === "shown" ? "reveal-in" : ""} ${className}`}
      style={state === "idle" ? undefined : { transitionDelay: `${delay}ms` }}
    >
      {children}
    </div>
  );
}

/** Counts from 0 to `to` when scrolled into view. Renders the final value on the server. */
export function CountUp({ to, suffix = "" }: { to: number; suffix?: string }) {
  const ref = useRef<HTMLSpanElement>(null);
  const [n, setN] = useState(to);
  useEffect(() => {
    const el = ref.current;
    if (!el || prefersReducedMotion()) return;
    setN(0);
    let raf = 0;
    const stop = onceVisible(el, () => {
      const start = performance.now();
      const step = (now: number) => {
        const p = Math.min(1, (now - start) / 1200);
        setN(Math.round(to * (1 - Math.pow(1 - p, 3))));
        if (p < 1) raf = requestAnimationFrame(step);
      };
      raf = requestAnimationFrame(step);
    });
    return () => {
      stop();
      cancelAnimationFrame(raf);
    };
  }, [to]);
  return (
    <span ref={ref}>
      {n.toLocaleString("ko-KR")}
      {suffix}
    </span>
  );
}
