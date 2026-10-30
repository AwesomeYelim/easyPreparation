"use client";

import { useRef } from "react";
import Stage from "./Stage";
import { useTicker } from "./hooks";
import { ConsoleMock, PhoneMock } from "./mocks";

const PERIOD = 3; // seconds per worship item

/** Hero mock: desktop console and phone remote advancing in lockstep. */
export default function HeroLive() {
  const ref = useRef<HTMLDivElement>(null);
  const t = useTicker(ref);
  const index = 1 + Math.floor(t / PERIOD);

  return (
    <div ref={ref} className="relative pb-14 pr-[14%] sm:pb-16">
      <div className="overflow-hidden rounded-card border border-white/[0.12] bg-surface shadow-[0_30px_60px_rgba(0,0,0,0.5),0_0_0_1px_rgba(0,0,0,0.4)]">
        <div className="flex h-[34px] items-center gap-1.5 border-b border-line bg-raised px-3">
          <span className="h-2.5 w-2.5 rounded-full bg-[#2a3a55]" />
          <span className="h-2.5 w-2.5 rounded-full bg-[#2a3a55]" />
          <span className="h-2.5 w-2.5 rounded-full bg-[#2a3a55]" />
          <span className="ml-2.5 h-[18px] flex-1 rounded border border-line bg-bg" />
        </div>
        <Stage w={1200} h={750}>
          <ConsoleMock index={index} t={t} period={PERIOD} />
        </Stage>
      </div>
      <div className="absolute bottom-0 right-0 w-[28%] max-w-[150px] overflow-hidden rounded-[26px] border-[5px] border-[#1c2740] bg-surface shadow-[0_24px_48px_rgba(0,0,0,0.55)]">
        <Stage w={390} h={844}>
          <PhoneMock index={index} />
        </Stage>
      </div>
    </div>
  );
}
