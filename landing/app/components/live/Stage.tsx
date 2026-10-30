"use client";

import type { ReactNode } from "react";
import { useFitScale } from "./hooks";

/**
 * Renders `children` at a fixed design size and scales the whole thing to the
 * container width, so the mock UI never reflows on narrow viewports.
 */
export default function Stage({
  w,
  h,
  className = "",
  children,
}: {
  w: number;
  h: number;
  className?: string;
  children: ReactNode;
}) {
  const { ref, scale } = useFitScale(w);
  return (
    <div
      ref={ref}
      className={`relative w-full overflow-hidden ${className}`}
      style={{ aspectRatio: `${w} / ${h}` }}
    >
      <div
        className="absolute left-0 top-0 origin-top-left transition-opacity duration-300"
        style={{
          width: w,
          height: h,
          transform: `scale(${scale ?? 0})`,
          opacity: scale ? 1 : 0,
        }}
      >
        {children}
      </div>
    </div>
  );
}
