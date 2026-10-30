"use client";

import { useEffect, useRef } from "react";
import { prefersReducedMotion } from "./hooks";

// Diagonal light beams behind the hero (Raycast-style), drawn with blurred
// gradient strips. They drift on their own and lean toward the pointer.
const beams = [
  { x: 8, w: 70, c: "79,140,201", d: 11, o: 0.55 },
  { x: 20, w: 130, c: "127,179,230", d: 14, o: 0.5 },
  { x: 34, w: 60, c: "232,170,90", d: 9, o: 0.45 },
  { x: 46, w: 170, c: "79,140,201", d: 17, o: 0.6 },
  { x: 60, w: 80, c: "169,205,240", d: 12, o: 0.4 },
  { x: 72, w: 120, c: "232,170,90", d: 15, o: 0.35 },
  { x: 86, w: 70, c: "79,140,201", d: 10, o: 0.5 },
];

export default function Beams() {
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const el = ref.current;
    if (!el || prefersReducedMotion()) return;
    let raf = 0;
    const onMove = (e: PointerEvent) => {
      cancelAnimationFrame(raf);
      raf = requestAnimationFrame(() => {
        const nx = e.clientX / window.innerWidth - 0.5;
        const ny = e.clientY / window.innerHeight - 0.5;
        el.style.setProperty("--px", `${nx * 40}px`);
        el.style.setProperty("--py", `${ny * 30}px`);
      });
    };
    window.addEventListener("pointermove", onMove);
    return () => {
      window.removeEventListener("pointermove", onMove);
      cancelAnimationFrame(raf);
    };
  }, []);

  return (
    <div
      ref={ref}
      aria-hidden
      className="pointer-events-none absolute inset-0 overflow-hidden [mask-image:radial-gradient(ellipse_70%_75%_at_50%_40%,#000_35%,transparent_85%)]"
    >
      <div
        className="absolute -inset-[30%] transition-transform duration-700 ease-out"
        style={{ transform: "translate(var(--px,0px), var(--py,0px)) rotate(-38deg)" }}
      >
        {beams.map((b, i) => (
          <span
            key={i}
            className="beam absolute top-0 h-full"
            style={{
              left: `${b.x}%`,
              width: b.w,
              opacity: b.o,
              animationDuration: `${b.d}s`,
              animationDelay: `${-i * 1.7}s`,
              background: `linear-gradient(180deg, transparent 10%, rgba(${b.c},0.9) 45%, rgba(${b.c},0.25) 60%, transparent 85%)`,
            }}
          />
        ))}
      </div>
      <div className="grain absolute inset-0" />
    </div>
  );
}
